package httpapi

import (
	"context"
	"encoding/json"
	"github.com/MartinRusso28/bandia/internal/domain"
	"strings"
	"testing"
)

type externalFake struct {
	fakeRepo
	turn domain.TurnInput
}

func (f *externalFake) AppendTurn(_ context.Context, _, _ string, in domain.TurnInput) (domain.Message, bool, error) {
	f.calls++
	f.turn = in
	return domain.Message{Prompt: in.Prompt, Content: in.Content}, f.replayed, f.createErr
}
func (f *externalFake) CloseExternal(context.Context, string, string, domain.Closure) (domain.Meeting, bool, error) {
	f.calls++
	return domain.Meeting{Status: "completed"}, false, f.createErr
}

func TestExternalHTTPValidation(t *testing.T) {
	path := "/v1/meetings/" + strings.Repeat("a", 32) + "/turns"
	f := &externalFake{}
	h := NewHandler(f, testToken)
	in := domain.TurnInput{Sequence: 1, CharacterID: "luna", ContextHash: strings.Repeat("a", 64), AgentRunID: "run-1", Prompt: "  original prompt\n", Content: " original response\n"}
	b, _ := json.Marshal(in)
	w := request(h, "POST", path, string(b), "turn-1", testToken)
	if w.Code != 201 || f.turn.Prompt != in.Prompt || f.turn.Content != in.Content {
		t.Fatal(w.Code, f.turn)
	}
	in.HistoryThrough = 2
	b, _ = json.Marshal(in)
	if w = request(h, "POST", path, string(b), "bad", testToken); w.Code != 422 || f.calls != 1 {
		t.Fatal(w.Code, "bad input reached store")
	}
	in.HistoryThrough = 0
	in.Prompt = "\x00"
	b, _ = json.Marshal(in)
	if w = request(h, "POST", path, string(b), "nul", testToken); w.Code != 422 {
		t.Fatal(w.Code)
	}
	in.Prompt = "valid"
	b, _ = json.Marshal(in)
	f.createErr = domain.ErrTurn
	if w = request(h, "POST", path, string(b), "conflict", testToken); w.Code != 409 || !strings.Contains(w.Body.String(), "turn_conflict") {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = request(h, "POST", path, string(b), "unauth", ""); w.Code != 401 {
		t.Fatal(w.Code)
	}
	closePath := strings.TrimSuffix(path, "turns") + "close"
	if w = request(h, "POST", closePath, `{"last_sequence":0,"summary":"x","decisions":["x"]}`, "early", testToken); w.Code != 422 {
		t.Fatal(w.Code)
	}
	f.createErr = nil
	if w = request(h, "POST", closePath, `{"last_sequence":11,"summary":"x","decisions":["x"]}`, "close", testToken); w.Code != 200 {
		t.Fatal(w.Code)
	}
}
