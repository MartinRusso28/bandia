package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MartinRusso28/bandia/internal/domain"
)

type Repository interface {
	Ready(context.Context) error
	Band(context.Context) (domain.Band, error)
	Characters(context.Context) ([]domain.Character, error)
	CreateInstruction(context.Context, string, string) (domain.Instruction, bool, error)
	CreateMeeting(context.Context, string, domain.MeetingInput) (domain.MeetingAccepted, bool, error)
	Meeting(context.Context, string) (domain.Meeting, error)
	Messages(context.Context, string) ([]domain.Message, error)
	Job(context.Context, string) (domain.Job, error)
}

type api struct{ repo Repository }

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
var idPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func NewHandler(repo Repository, managerToken string) http.Handler {
	a := &api{repo: repo}
	tokenHash := sha256.Sum256([]byte(managerToken))
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET", "HEAD") {
			return
		}
		respond(w, r, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET", "HEAD") {
			return
		}
		if err := a.repo.Ready(r.Context()); err != nil {
			fail(w, r, 503, "not_ready", "Database or schema unavailable")
			return
		}
		respond(w, r, 200, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/v1/system", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET") {
			return
		}
		respond(w, r, 200, map[string]any{"version": "0.1.0", "capabilities": map[string]bool{"persistent_meetings": true, "agents": false, "scheduler": false, "music": false, "social_publishing": false, "social_feedback": false}})
	})
	mux.HandleFunc("/v1/band", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET") {
			return
		}
		v, err := a.repo.Band(r.Context())
		result(w, r, v, err)
	})
	mux.HandleFunc("/v1/characters", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET") {
			return
		}
		v, err := a.repo.Characters(r.Context())
		result(w, r, map[string]any{"data": v}, err)
	})
	mux.HandleFunc("/v1/manager-instructions", a.instruction)
	mux.HandleFunc("/v1/meetings", a.meeting)
	mux.HandleFunc("/v1/meetings/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET") || !validID(w, r) {
			return
		}
		v, err := a.repo.Meeting(r.Context(), r.PathValue("id"))
		result(w, r, v, err)
	})
	mux.HandleFunc("/v1/meetings/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET") || !validID(w, r) {
			return
		}
		v, err := a.repo.Messages(r.Context(), r.PathValue("id"))
		result(w, r, map[string]any{"data": v}, err)
	})
	mux.HandleFunc("/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !allow(w, r, "GET") || !validID(w, r) {
			return
		}
		v, err := a.repo.Job(r.Context(), r.PathValue("id"))
		result(w, r, v, err)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fail(w, r, 404, "not_found", "Route not found") })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		var requestID [16]byte
		_, _ = rand.Read(requestID[:])
		w.Header().Set("X-Request-ID", hex.EncodeToString(requestID[:]))
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		if strings.HasPrefix(r.URL.Path, "/v1/") || r.URL.Path == "/v1" {
			provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			hash := sha256.Sum256([]byte(provided))
			if len(managerToken) < 32 || !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") || subtle.ConstantTimeCompare(hash[:], tokenHash[:]) != 1 {
				w.Header().Set("WWW-Authenticate", "Bearer")
				fail(w, r, 401, "unauthorized", "Valid manager token required")
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func (a *api) instruction(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, "POST") {
		return
	}
	var input struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &input) {
		return
	}
	input.Text = strings.TrimSpace(input.Text)
	if input.Text == "" || utf8.RuneCountInString(input.Text) > 8000 {
		fail(w, r, 422, "invalid_input", "text must contain 1–8000 characters")
		return
	}
	out, replayed, err := a.repo.CreateInstruction(r.Context(), r.Header.Get("Idempotency-Key"), input.Text)
	if err != nil {
		result(w, r, nil, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	respond(w, r, 201, out)
}

func (a *api) meeting(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, "POST") {
		return
	}
	var input domain.MeetingInput
	if !decode(w, r, &input) {
		return
	}
	input.Topic = strings.TrimSpace(input.Topic)
	if input.Topic == "" || utf8.RuneCountInString(input.Topic) > 2000 || len(input.InstructionIDs) > 50 {
		fail(w, r, 422, "invalid_input", "topic must contain 1–2000 characters; at most 50 instruction_ids")
		return
	}
	for _, id := range input.InstructionIDs {
		if !idPattern.MatchString(id) {
			fail(w, r, 422, "invalid_input", "Invalid instruction ID")
			return
		}
	}
	// Instructions form a set: order and duplicates do not change the request.
	slices.Sort(input.InstructionIDs)
	input.InstructionIDs = slices.Compact(input.InstructionIDs)
	if input.InstructionIDs == nil {
		input.InstructionIDs = []string{}
	}
	out, replayed, err := a.repo.CreateMeeting(r.Context(), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		result(w, r, nil, err)
		return
	}
	w.Header().Set("Location", "/v1/meetings/"+out.MeetingID)
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	respond(w, r, 202, out)
}

func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	if !keyPattern.MatchString(r.Header.Get("Idempotency-Key")) {
		fail(w, r, 400, "invalid_idempotency_key", "Idempotency-Key must contain 1–128 letters, digits, dots, colons, underscores or hyphens")
		return false
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		fail(w, r, 415, "unsupported_media_type", "Use application/json")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(out)
	if err == nil {
		var extra any
		if err = decoder.Decode(&extra); errors.Is(err, io.EOF) {
			return true
		}
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		fail(w, r, 413, "body_too_large", "Maximum body size is 64 KiB")
		return false
	}
	fail(w, r, 400, "invalid_json", "Expected one JSON object with documented fields")
	return false
}

func validID(w http.ResponseWriter, r *http.Request) bool {
	if !idPattern.MatchString(r.PathValue("id")) {
		fail(w, r, 404, "not_found", "Resource not found")
		return false
	}
	return true
}

func allow(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	if slices.Contains(methods, r.Method) {
		return true
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	fail(w, r, 405, "method_not_allowed", "Method not supported")
	return false
}

func result(w http.ResponseWriter, r *http.Request, v any, err error) {
	switch {
	case err == nil:
		respond(w, r, 200, v)
	case errors.Is(err, domain.ErrNotFound):
		fail(w, r, 404, "not_found", "Resource not found")
	case errors.Is(err, domain.ErrConflict):
		fail(w, r, 409, "idempotency_conflict", "This key was already used with different input")
	case errors.Is(err, domain.ErrInstruction):
		fail(w, r, 422, "unknown_instruction", "An instruction_id does not exist")
	default:
		fail(w, r, 503, "storage_unavailable", "Storage operation failed; retry with the same idempotency key")
	}
}

func fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	respond(w, r, status, map[string]any{"error": map[string]string{"code": code, "message": message, "request_id": w.Header().Get("X-Request-ID")}})
}

func respond(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_ = json.NewEncoder(w).Encode(body)
	}
}
