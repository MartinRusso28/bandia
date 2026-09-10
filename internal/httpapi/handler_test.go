package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MartinRusso28/bandia/internal/domain"
)

// Fakes are confined to HTTP unit tests. Runtime always requires PostgreSQL.
type fakeRepo struct {
	Repository
	readyErr  error
	createErr error
	replayed  bool
	calls     int
	input     domain.MeetingInput
}

func (f *fakeRepo) Ready(context.Context) error { return f.readyErr }
func (f *fakeRepo) CreateMeeting(_ context.Context, _ string, in domain.MeetingInput) (domain.MeetingAccepted, bool, error) {
	f.calls++
	f.input = in
	return domain.MeetingAccepted{MeetingID: strings.Repeat("a", 32), JobID: strings.Repeat("b", 32), Status: "queued"}, f.replayed, f.createErr
}

const testToken = "test-only-manager-token-32-characters"

func request(h http.Handler, method, path, body, key, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestOperationalContract(t *testing.T) {
	f := &fakeRepo{}
	h := NewHandler(f, testToken)
	for _, tc := range []struct {
		method, path string
		code         int
	}{
		{"GET", "/healthz", 200}, {"GET", "/readyz", 200}, {"HEAD", "/healthz", 200},
		{"HEAD", "/readyz", 200}, {"POST", "/healthz", 405}, {"GET", "/healthz/extra", 404},
		{"GET", "/unknown", 404}, {"GET", "/v1/system", 401}, {"GET", "/v1/meetings", 401},
	} {
		w := request(h, tc.method, tc.path, "", "", "")
		if w.Code != tc.code {
			t.Fatalf("%s %s: %d", tc.method, tc.path, w.Code)
		}
		if w.Header().Get("Content-Type") != "application/json" || w.Header().Get("X-Request-ID") == "" {
			t.Fatal("missing response headers")
		}
		if tc.method == "HEAD" && w.Body.Len() != 0 {
			t.Fatal("HEAD has body")
		}
	}
	f.readyErr = errors.New("private database detail")
	w := request(h, "GET", "/readyz", "", "", "")
	if w.Code != 503 || strings.Contains(w.Body.String(), "private") {
		t.Fatal(w.Body.String())
	}
	w = request(h, "HEAD", "/readyz", "", "", "")
	if w.Code != 503 || w.Body.Len() != 0 {
		t.Fatal("HEAD readiness contract")
	}
}

func TestMeetingValidationAndAuthentication(t *testing.T) {
	for _, tc := range []struct {
		name, body, key, token string
		code                   int
	}{
		{"missing auth", `{"topic":"hola"}`, "key", "", 401},
		{"wrong auth", `{"topic":"hola"}`, "key", "wrong", 401},
		{"missing key", `{"topic":"hola"}`, "", testToken, 400},
		{"invalid key", `{"topic":"hola"}`, "bad key", testToken, 400},
		{"unknown field", `{"topic":"hola","run":true}`, "key", testToken, 400},
		{"multiple values", `{"topic":"hola"} {}`, "key", testToken, 400},
		{"blank topic", `{"topic":"  "}`, "key", testToken, 422},
		{"null", `null`, "key", testToken, 422},
		{"bad instruction", `{"topic":"hola","instruction_ids":["bogus"]}`, "key", testToken, 422},
		{"oversize", `{"topic":"` + strings.Repeat("x", 66000) + `"}`, "key", testToken, 413},
		{"valid", `{"topic":"hola"}`, "key", testToken, 202},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeRepo{}
			w := request(NewHandler(f, testToken), "POST", "/v1/meetings", tc.body, tc.key, tc.token)
			if w.Code != tc.code {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			if tc.code != 202 && f.calls != 0 {
				t.Fatal("invalid request reached storage")
			}
		})
	}
}

func TestMeetingReplayAndErrors(t *testing.T) {
	f := &fakeRepo{replayed: true}
	h := NewHandler(f, testToken)
	one, two := strings.Repeat("a", 32), strings.Repeat("b", 32)
	w := request(h, "POST", "/v1/meetings", `{"topic":" hola ","instruction_ids":["`+two+`","`+one+`","`+one+`"]}`, "key", testToken)
	if w.Code != 202 || w.Header().Get("Idempotency-Replayed") != "true" || w.Header().Get("Location") == "" {
		t.Fatal(w.Code, w.Header())
	}
	if f.input.Topic != "hola" || len(f.input.InstructionIDs) != 2 || f.input.InstructionIDs[0] != one {
		t.Fatal("request not normalized")
	}
	for _, tc := range []struct {
		err  error
		code int
	}{{domain.ErrConflict, 409}, {domain.ErrInstruction, 422}, {errors.New("postgres://secret"), 503}} {
		f.createErr = tc.err
		w = request(h, "POST", "/v1/meetings", `{"topic":"hola"}`, "key", testToken)
		if w.Code != tc.code || strings.Contains(w.Body.String(), "secret") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
