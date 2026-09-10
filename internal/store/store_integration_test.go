package store

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MartinRusso28/bandia/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Integration tests require an explicit disposable database. Each test owns a
// fresh schema; cleanup never truncates or drops user tables in public.
func database(t *testing.T) (*Store, *pgxpool.Config) {
	t.Helper()
	dsn := os.Getenv("BANDIA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set BANDIA_TEST_DATABASE_URL to run real PostgreSQL tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid test database configuration")
	}
	admin, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal("test pool unavailable")
	}
	if err = admin.Ping(ctx); err != nil {
		admin.Close()
		t.Fatal("test database unreachable")
	}
	schema := "bandia_test_" + id()
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	config = config.Copy()
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	s := &Store{pool: pool}
	t.Cleanup(func() {
		s.Close()
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(cleanup, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Error("test schema cleanup failed", err)
		}
		admin.Close()
	})
	return s, config
}

func TestPostgresLifecycleAndRestart(t *testing.T) {
	s, config := database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Ready(ctx); err == nil {
		t.Fatal("unmigrated database ready")
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatal("migration not idempotent", err)
	}
	if err := s.Ready(ctx); err != nil {
		t.Fatal(err)
	}
	cast, err := s.Characters(ctx)
	if err != nil || len(cast) != 5 {
		t.Fatal("seed cast", cast, err)
	}
	instruction, replay, err := s.CreateInstruction(ctx, "brief-1", "Elegir el rumbo de la banda")
	if err != nil || replay {
		t.Fatal(instruction, replay, err)
	}
	again, replay, err := s.CreateInstruction(ctx, "brief-1", instruction.Text)
	if err != nil || !replay || again.ID != instruction.ID {
		t.Fatal("instruction replay", err)
	}
	if _, _, err = s.CreateInstruction(ctx, "brief-1", "Different"); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("instruction conflict", err)
	}
	in := domain.MeetingInput{Topic: "Primer encuentro", InstructionIDs: []string{instruction.ID}}
	accepted, replay, err := s.CreateMeeting(ctx, "meeting-1", in)
	if err != nil || replay {
		t.Fatal(accepted, replay, err)
	}
	meeting, err := s.Meeting(ctx, accepted.MeetingID)
	if err != nil || meeting.Status != "queued" || len(meeting.Participants) != 5 || len(meeting.Instructions) != 1 {
		t.Fatal("snapshot", meeting, err)
	}
	if _, err = s.pool.Exec(ctx, `UPDATE band SET name='Nueva identidad'`); err != nil {
		t.Fatal(err)
	}
	if err = s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	band, err := s.Band(ctx)
	if err != nil || band.Name != "Nueva identidad" {
		t.Fatal("migration reset band", err)
	}
	meeting, err = s.Meeting(ctx, accepted.MeetingID)
	if err != nil || meeting.Band.Name != "Fuera de Hora" {
		t.Fatal("snapshot changed", err)
	}
	// A completely new connection pool must find and process the queued job.
	s.Close()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	s.pool = pool
	worked, err := s.ProcessNext(ctx)
	if err != nil || !worked {
		t.Fatal("queued work lost on restart", err)
	}
	job, err := s.Job(ctx, accepted.JobID)
	if err != nil || job.Status != "blocked" || job.Attempts != 1 || job.BlockedReason != "agent_provider_not_implemented" {
		t.Fatal("job", job, err)
	}
	meeting, err = s.Meeting(ctx, accepted.MeetingID)
	if err != nil || meeting.Status != "blocked" || meeting.BlockedReason != job.BlockedReason {
		t.Fatal("meeting not updated atomically", err)
	}
	messages, err := s.Messages(ctx, meeting.ID)
	if err != nil || messages == nil || len(messages) != 0 {
		t.Fatal("blocked meeting must have no invented messages", err)
	}
	if worked, err = s.ProcessNext(ctx); err != nil || worked {
		t.Fatal("blocked job executed twice", err)
	}
	result, replay, err := s.CreateMeeting(ctx, "meeting-1", in)
	if err != nil || !replay || result != accepted {
		t.Fatal("replay must retain original acceptance", err)
	}
	if _, err = s.Meeting(ctx, id()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("missing meeting", err)
	}
	if _, err = s.Messages(ctx, id()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("missing messages", err)
	}
	if _, err = s.Job(ctx, id()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("missing job", err)
	}
}

func TestPostgresConcurrentIdempotency(t *testing.T) {
	s, _ := database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	const n = 16
	results := make(chan domain.MeetingAccepted, n)
	errorsCh := make(chan error, n)
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, _, err := s.CreateMeeting(ctx, "shared-key", domain.MeetingInput{Topic: "Rumbo", InstructionIDs: []string{}})
			results <- v
			errorsCh <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	var first domain.MeetingAccepted
	for v := range results {
		if first.MeetingID == "" {
			first = v
		}
		if v != first {
			t.Fatal("duplicate meeting", v, first)
		}
	}
	for _, table := range []string{"meetings", "jobs", "idempotency_keys"} {
		var count int
		if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 1 {
			t.Fatal(table, count, err)
		}
	}
	if _, _, err := s.CreateMeeting(ctx, "shared-key", domain.MeetingInput{Topic: "Changed"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("conflict missing", err)
	}
	// Concurrent workers must transition one job exactly once.
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.ProcessNext(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	job, err := s.Job(ctx, first.JobID)
	if err != nil || job.Attempts != 1 {
		t.Fatal("worker duplication", job, err)
	}
}

func TestPostgresRollbackAndChecksum(t *testing.T) {
	s, _ := database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	_, _, err := s.CreateMeeting(ctx, "rollback", domain.MeetingInput{Topic: "Invalid reference", InstructionIDs: []string{id()}})
	if !errors.Is(err, domain.ErrInstruction) {
		t.Fatal(err)
	}
	for _, table := range []string{"meetings", "jobs", "idempotency_keys"} {
		var n int
		if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil || n != 0 {
			t.Fatal("partial write", table, n, err)
		}
	}
	if _, _, err = s.CreateMeeting(ctx, "rollback", domain.MeetingInput{Topic: "Valid after rollback"}); err != nil {
		t.Fatal("rolled back key was reserved", err)
	}
	if _, err = s.pool.Exec(ctx, `UPDATE schema_migrations SET checksum='tampered'`); err != nil {
		t.Fatal(err)
	}
	if err = s.Migrate(ctx); err == nil {
		t.Fatal("changed migration accepted")
	}
	if err = s.Ready(ctx); err == nil {
		t.Fatal("incompatible schema ready")
	}
}
