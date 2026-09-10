package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/MartinRusso28/bandia/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
	"testing"
	"time"
)

// Test fixtures are intentionally not real agent output.
func fixture(c domain.ExternalContext) domain.TurnInput {
	return domain.TurnInput{Sequence: c.NextSequence, HistoryThrough: c.NextSequence - 1, CharacterID: c.NextCharacterID, ContextHash: c.ContextHash, AgentRunID: fmt.Sprintf("fixture-run-%d", c.NextSequence), Prompt: "  TEST FIXTURE prompt\n", Content: "TEST FIXTURE response\n"}
}

func TestExternalMeetingLifecycle(t *testing.T) {
	s, cfg := database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	accepted, _, err := s.CreateMeeting(ctx, "external", domain.MeetingInput{Mode: "external", Topic: "Test fixtures only"})
	if err != nil {
		t.Fatal(err)
	}
	if worked, err := s.ProcessNext(ctx); worked || err != nil {
		t.Fatal("internal worker touched external job", err)
	}
	c, err := s.ExternalContext(ctx, accepted.MeetingID)
	if err != nil {
		t.Fatal(err)
	}
	if c.NextSequence != 1 || c.NextCharacterID != "luna" || len(c.Meeting.TurnPlan) != 11 || len(c.ContextHash) != 64 {
		t.Fatal(c)
	}
	closure := domain.Closure{LastSequence: 11, Summary: "Fixture summary", Decisions: []string{"Fixture decision"}}
	if _, _, err = s.CloseExternal(ctx, accepted.MeetingID, "early", closure); !errors.Is(err, domain.ErrTurn) {
		t.Fatal("early close", err)
	}
	var first domain.Message
	var original domain.TurnInput
	for n := 1; n <= 11; n++ {
		c, err = s.ExternalContext(ctx, accepted.MeetingID)
		if err != nil {
			t.Fatal(err)
		}
		in := fixture(c)
		if n == 1 {
			original = in
			bad := in
			bad.CharacterID = "productor"
			if _, _, err = s.AppendTurn(ctx, accepted.MeetingID, "wrong", bad); !errors.Is(err, domain.ErrTurn) {
				t.Fatal("wrong speaker", err)
			}
		}
		msg, replay, err := s.AppendTurn(ctx, accepted.MeetingID, fmt.Sprint(n), in)
		if err != nil || replay || msg.Prompt != in.Prompt || msg.Content != in.Content {
			t.Fatal("turn not preserved", err)
		}
		if n == 1 {
			first = msg
		}
		if n == 3 {
			s.Close()
			pool, err := pgxpool.NewWithConfig(ctx, cfg)
			if err != nil {
				t.Fatal(err)
			}
			s.pool = pool
		}
	}
	if _, _, err = s.AppendTurn(ctx, accepted.MeetingID, "stale", original); !errors.Is(err, domain.ErrTurn) {
		t.Fatal("stale context accepted", err)
	}
	m, replay, err := s.CloseExternal(ctx, accepted.MeetingID, "finish", closure)
	if err != nil || replay || m.Status != "completed" || m.Closure.Summary != closure.Summary {
		t.Fatal("close", err)
	}
	job, err := s.Job(ctx, accepted.JobID)
	if err != nil || job.Status != "completed" || job.Attempts != 0 {
		t.Fatal("job", job, err)
	}
	again, replay, err := s.AppendTurn(ctx, accepted.MeetingID, "1", original)
	if err != nil || !replay || again.ID != first.ID {
		t.Fatal("replay after close", err)
	}
	if _, replay, err = s.CloseExternal(ctx, accepted.MeetingID, "finish", closure); err != nil || !replay {
		t.Fatal("close replay", err)
	}
	if _, _, err = s.CloseExternal(ctx, accepted.MeetingID, "other-close", closure); !errors.Is(err, domain.ErrTurn) {
		t.Fatal("second closure accepted", err)
	}
	c, err = s.ExternalContext(ctx, accepted.MeetingID)
	if err != nil || len(c.Messages) != 11 || c.NextCharacterID != "" {
		t.Fatal("final context", err)
	}
}

func TestExternalTurnConcurrency(t *testing.T) {
	s, _ := database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	a, _, err := s.CreateMeeting(ctx, "external", domain.MeetingInput{Mode: "external", Topic: "Concurrent fixture"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.ExternalContext(ctx, a.MeetingID)
	if err != nil {
		t.Fatal(err)
	}
	in := fixture(c)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, _, e := s.AppendTurn(ctx, a.MeetingID, "same", in); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	c, err = s.ExternalContext(ctx, a.MeetingID)
	if err != nil || len(c.Messages) != 1 {
		t.Fatal("duplicate turns", err)
	}
	in = fixture(c)
	in.AgentRunID = "fixture-run-1"
	if _, _, err = s.AppendTurn(ctx, a.MeetingID, "reused-execution", in); !errors.Is(err, domain.ErrTurn) {
		t.Fatal("reused execution ID", err)
	}
	in = fixture(c)
	errs = make(chan error, 8)
	for n := range 8 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _, e := s.AppendTurn(ctx, a.MeetingID, fmt.Sprintf("race-%d", n), in)
			errs <- e
		}(n)
	}
	wg.Wait()
	close(errs)
	success := 0
	for e := range errs {
		if e == nil {
			success++
		} else if !errors.Is(e, domain.ErrTurn) {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatal("concurrent winners", success)
	}
}
