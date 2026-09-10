package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/MartinRusso28/bandia/internal/domain"
	"github.com/jackc/pgx/v5"
)

func readExternal(ctx context.Context, q queryer, meetingID string) (domain.ExternalContext, error) {
	var out domain.ExternalContext
	m, err := readMeeting(ctx, q, meetingID)
	if err != nil {
		return out, err
	}
	if m.Mode != "external" {
		return out, domain.ErrTurn
	}
	msgs, err := readMessages(ctx, q, meetingID)
	if err != nil {
		return out, err
	}
	out = domain.ExternalContext{Meeting: m, Messages: msgs, NextSequence: len(msgs) + 1}
	if len(msgs) < len(m.TurnPlan) && m.Status == "waiting_external" {
		out.NextCharacterID = m.TurnPlan[len(msgs)]
	}
	// Hash the complete persisted context, without its own hash field.
	b, err := json.Marshal(out)
	if err != nil {
		return out, err
	}
	sum := sha256.Sum256(b)
	out.ContextHash = hex.EncodeToString(sum[:])
	return out, nil
}

func (s *Store) ExternalContext(ctx context.Context, meetingID string) (domain.ExternalContext, error) {
	var out domain.ExternalContext
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return out, err
	}
	defer tx.Rollback(context.Background())
	out, err = readExternal(ctx, tx, meetingID)
	if err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func lockMeeting(ctx context.Context, tx pgx.Tx, meetingID string) error {
	var id string
	err := tx.QueryRow(ctx, `SELECT id FROM meetings WHERE id=$1 FOR UPDATE`, meetingID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (s *Store) AppendTurn(ctx context.Context, meetingID, key string, in domain.TurnInput) (domain.Message, bool, error) {
	var out domain.Message
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, false, err
	}
	defer tx.Rollback(context.Background())
	operation := "external.turn:" + meetingID
	previous, err := reserve(ctx, tx, operation, key, in)
	if err != nil {
		return out, false, err
	}
	if previous != nil {
		err = json.Unmarshal(previous, &out)
		return out, true, err
	}
	if err = lockMeeting(ctx, tx, meetingID); err != nil {
		return out, false, err
	}
	state, err := readExternal(ctx, tx, meetingID)
	if err != nil {
		return out, false, err
	}
	if state.Meeting.Status != "waiting_external" || state.NextCharacterID == "" || in.Sequence != state.NextSequence || in.HistoryThrough != state.NextSequence-1 || in.CharacterID != state.NextCharacterID || in.ContextHash != state.ContextHash {
		return out, false, domain.ErrTurn
	}
	var used bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM meeting_messages WHERE meeting_id=$1 AND agent_run_id=$2)`, meetingID, in.AgentRunID).Scan(&used)
	if err != nil {
		return out, false, err
	}
	if used {
		return out, false, domain.ErrTurn
	}
	out = domain.Message{ID: id(), CharacterID: in.CharacterID, Sequence: in.Sequence, Content: in.Content, Prompt: in.Prompt, AgentRunID: in.AgentRunID, ContextHash: in.ContextHash, HistoryThrough: in.HistoryThrough, Origin: "external_agent_reported"}
	err = tx.QueryRow(ctx, `INSERT INTO meeting_messages(id,meeting_id,character_id,sequence,content,prompt,agent_run_id,context_hash,history_through,origin) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING created_at`, out.ID, meetingID, out.CharacterID, out.Sequence, out.Content, out.Prompt, out.AgentRunID, out.ContextHash, out.HistoryThrough, out.Origin).Scan(&out.CreatedAt)
	if err != nil {
		return out, false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE meetings SET updated_at=now() WHERE id=$1`, meetingID); err != nil {
		return out, false, err
	}
	err = finish(ctx, tx, operation, key, out)
	return out, false, err
}

func (s *Store) CloseExternal(ctx context.Context, meetingID, key string, in domain.Closure) (domain.Meeting, bool, error) {
	var out domain.Meeting
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, false, err
	}
	defer tx.Rollback(context.Background())
	operation := "external.close:" + meetingID
	previous, err := reserve(ctx, tx, operation, key, in)
	if err != nil {
		return out, false, err
	}
	if previous != nil {
		err = json.Unmarshal(previous, &out)
		return out, true, err
	}
	if err = lockMeeting(ctx, tx, meetingID); err != nil {
		return out, false, err
	}
	state, err := readExternal(ctx, tx, meetingID)
	if err != nil {
		return out, false, err
	}
	if state.Meeting.Status != "waiting_external" || in.LastSequence != len(state.Meeting.TurnPlan) || len(state.Messages) != in.LastSequence {
		return out, false, domain.ErrTurn
	}
	body, err := json.Marshal(in)
	if err != nil {
		return out, false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE meetings SET status='completed',closure=$2,updated_at=now() WHERE id=$1`, meetingID, body); err != nil {
		return out, false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE jobs SET status='completed',updated_at=now() WHERE meeting_id=$1`, meetingID); err != nil {
		return out, false, err
	}
	out, err = readMeeting(ctx, tx, meetingID)
	if err != nil {
		return out, false, err
	}
	err = finish(ctx, tx, operation, key, out)
	return out, false, err
}
