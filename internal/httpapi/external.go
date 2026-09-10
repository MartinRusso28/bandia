package httpapi

import (
	"github.com/MartinRusso28/bandia/internal/domain"
	"net/http"
	"strings"
	"unicode/utf8"
)

func validText(s string, max int) bool {
	return strings.TrimSpace(s) != "" && !strings.ContainsRune(s, 0) && utf8.RuneCountInString(s) <= max
}

func (a *api) turn(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, "POST") || !validID(w, r) {
		return
	}
	var in domain.TurnInput
	if !decode(w, r, &in) {
		return
	}
	if in.Sequence < 1 || in.Sequence > 11 || in.HistoryThrough != in.Sequence-1 || !validText(in.CharacterID, 100) || !validText(in.AgentRunID, 200) || len(in.ContextHash) != 64 || !validText(in.Prompt, 32000) || !validText(in.Content, 8000) {
		fail(w, r, 422, "invalid_input", "Required: planned sequence, preceding history count, character, context hash, agent run ID, original prompt (up to 32000 chars) and content (up to 8000 chars)")
		return
	}
	// Do not trim or rewrite the original prompt or response.
	out, replayed, err := a.repo.AppendTurn(r.Context(), r.PathValue("id"), r.Header.Get("Idempotency-Key"), in)
	if err != nil {
		result(w, r, nil, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	respond(w, r, 201, out)
}

func (a *api) closeExternal(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, "POST") || !validID(w, r) {
		return
	}
	var in domain.Closure
	if !decode(w, r, &in) {
		return
	}
	if in.LastSequence != 11 || !validText(in.Summary, 8000) || len(in.Decisions) < 1 || len(in.Decisions) > 20 {
		fail(w, r, 422, "invalid_input", "Required: last_sequence 11, summary (up to 8000 chars), and 1–20 decisions")
		return
	}
	for _, decision := range in.Decisions {
		if !validText(decision, 2000) {
			fail(w, r, 422, "invalid_input", "Each decision must contain 1–2000 characters")
			return
		}
	}
	out, replayed, err := a.repo.CloseExternal(r.Context(), r.PathValue("id"), r.Header.Get("Idempotency-Key"), in)
	if err != nil {
		result(w, r, nil, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	respond(w, r, 200, out)
}
