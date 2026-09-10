package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/MartinRusso28/bandia/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func Open(ctx context.Context, dsn string) (*Store, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	config.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("cannot initialize database pool")
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("database connection failed")
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func id() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func (s *Store) Band(ctx context.Context) (domain.Band, error) {
	var b domain.Band
	err := s.pool.QueryRow(ctx, `SELECT id,name,identity FROM band WHERE id='bandia'`).Scan(&b.ID, &b.Name, &b.Identity)
	return b, err
}

func (s *Store) Characters(ctx context.Context) ([]domain.Character, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,name,role,profile FROM characters ORDER BY position`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Character])
}

// The key reservation and result commit together. Concurrent duplicates wait on
// the unique index, then replay the original response without creating work.
func reserve(ctx context.Context, tx pgx.Tx, operation, key string, input any) (json.RawMessage, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(encoded)
	hash := hex.EncodeToString(sum[:])
	tag, err := tx.Exec(ctx, `INSERT INTO idempotency_keys(operation,key,request_hash,response) VALUES($1,$2,$3,'{}') ON CONFLICT DO NOTHING`, operation, key, hash)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 1 {
		return nil, nil
	}
	var existing string
	var response json.RawMessage
	err = tx.QueryRow(ctx, `SELECT request_hash,response FROM idempotency_keys WHERE operation=$1 AND key=$2`, operation, key).Scan(&existing, &response)
	if err != nil {
		return nil, err
	}
	if existing != hash {
		return nil, domain.ErrConflict
	}
	return response, nil
}

func finish(ctx context.Context, tx pgx.Tx, operation, key string, response any) error {
	body, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE idempotency_keys SET response=$3 WHERE operation=$1 AND key=$2`, operation, key, body); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CreateInstruction(ctx context.Context, key, text string) (domain.Instruction, bool, error) {
	var out domain.Instruction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, false, err
	}
	defer tx.Rollback(context.Background())
	previous, err := reserve(ctx, tx, "instruction.create", key, text)
	if err != nil {
		return out, false, err
	}
	if previous != nil {
		err = json.Unmarshal(previous, &out)
		return out, true, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO manager_instructions(id,text) VALUES($1,$2) RETURNING id,text,created_at`, id(), text).Scan(&out.ID, &out.Text, &out.CreatedAt)
	if err != nil {
		return out, false, err
	}
	err = finish(ctx, tx, "instruction.create", key, out)
	return out, false, err
}

func (s *Store) CreateMeeting(ctx context.Context, key string, input domain.MeetingInput) (domain.MeetingAccepted, bool, error) {
	var out domain.MeetingAccepted
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, false, err
	}
	defer tx.Rollback(context.Background())
	previous, err := reserve(ctx, tx, "meeting.create", key, input)
	if err != nil {
		return out, false, err
	}
	if previous != nil {
		err = json.Unmarshal(previous, &out)
		return out, true, err
	}
	instructions := make([]domain.Instruction, 0, len(input.InstructionIDs))
	for _, instructionID := range input.InstructionIDs {
		var v domain.Instruction
		err = tx.QueryRow(ctx, `SELECT id,text,created_at FROM manager_instructions WHERE id=$1`, instructionID).Scan(&v.ID, &v.Text, &v.CreatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return out, false, domain.ErrInstruction
		}
		if err != nil {
			return out, false, err
		}
		instructions = append(instructions, v)
	}
	snapshot, err := json.Marshal(instructions)
	if err != nil {
		return out, false, err
	}
	out = domain.MeetingAccepted{MeetingID: id(), JobID: id(), Status: "queued"}
	// Snapshot band and cast in the same statement as the meeting insert.
	tag, err := tx.Exec(ctx, `INSERT INTO meetings(id,topic,status,band_snapshot,participants,instructions)
	SELECT $1,$2,'queued',jsonb_build_object('id',b.id,'name',b.name,'identity',b.identity),
	(SELECT jsonb_agg(jsonb_build_object('id',id,'name',name,'role',role,'profile',profile) ORDER BY position) FROM characters),$3
	FROM band b WHERE b.id='bandia'`, out.MeetingID, input.Topic, snapshot)
	if err != nil {
		return out, false, err
	}
	if tag.RowsAffected() != 1 {
		return out, false, errors.New("band seed missing")
	}
	_, err = tx.Exec(ctx, `INSERT INTO jobs(id,meeting_id,type,status) VALUES($1,$2,'meeting.execute','queued')`, out.JobID, out.MeetingID)
	if err != nil {
		return out, false, err
	}
	err = finish(ctx, tx, "meeting.create", key, out)
	return out, false, err
}

func (s *Store) Meeting(ctx context.Context, meetingID string) (domain.Meeting, error) {
	var m domain.Meeting
	err := s.pool.QueryRow(ctx, `SELECT m.id,j.id,m.topic,m.status,coalesce(m.blocked_reason,''),m.band_snapshot,m.participants,m.instructions,m.created_at,m.updated_at FROM meetings m JOIN jobs j ON j.meeting_id=m.id WHERE m.id=$1`, meetingID).Scan(&m.ID, &m.JobID, &m.Topic, &m.Status, &m.BlockedReason, &m.Band, &m.Participants, &m.Instructions, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = domain.ErrNotFound
	}
	return m, err
}

func (s *Store) Messages(ctx context.Context, meetingID string) ([]domain.Message, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM meetings WHERE id=$1)`, meetingID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrNotFound
	}
	rows, err := s.pool.Query(ctx, `SELECT id,character_id,sequence,content,created_at FROM meeting_messages WHERE meeting_id=$1 ORDER BY sequence`, meetingID)
	if err != nil {
		return nil, err
	}
	messages, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Message])
	if messages == nil {
		messages = []domain.Message{}
	}
	return messages, err
}

func (s *Store) Job(ctx context.Context, jobID string) (domain.Job, error) {
	var j domain.Job
	err := s.pool.QueryRow(ctx, `SELECT id,meeting_id,type,status,coalesce(blocked_reason,''),attempts,created_at,updated_at FROM jobs WHERE id=$1`, jobID).Scan(&j.ID, &j.MeetingID, &j.Type, &j.Status, &j.BlockedReason, &j.Attempts, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = domain.ErrNotFound
	}
	return j, err
}

// ProcessNext records an honest capability failure. This transaction performs no
// external I/O, so row locks suffice. Real provider work will require leases and
// a running state in a subsequent migration; never hold a transaction over it.
func (s *Store) ProcessNext(ctx context.Context) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(context.Background())
	var jobID, meetingID string
	err = tx.QueryRow(ctx, `SELECT id,meeting_id FROM jobs WHERE status='queued' ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&jobID, &meetingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, err = tx.Exec(ctx, `UPDATE jobs SET status='blocked',blocked_reason='agent_provider_not_implemented',attempts=attempts+1,updated_at=now() WHERE id=$1`, jobID)
	if err != nil {
		return false, err
	}
	_, err = tx.Exec(ctx, `UPDATE meetings SET status='blocked',blocked_reason='agent_provider_not_implemented',updated_at=now() WHERE id=$1`, meetingID)
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
