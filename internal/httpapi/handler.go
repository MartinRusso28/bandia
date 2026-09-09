// Package httpapi exposes the bootstrap service's operational endpoints.
package httpapi

import (
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// NewHandler reports process health separately from band workflow readiness.
// Readiness remains unavailable until the durable runtime is implemented.
func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path != "/healthz" && r.URL.Path != "/readyz" {
			writeStatus(w, http.StatusNotFound, statusResponse{Status: "not_found"})
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeStatus(w, http.StatusMethodNotAllowed, statusResponse{Status: "method_not_allowed"})
			return
		}
		code := http.StatusOK
		body := statusResponse{Status: "ok"}
		if r.URL.Path == "/readyz" {
			code = http.StatusServiceUnavailable
			body = statusResponse{Status: "not_ready", Reason: "band_runtime_not_implemented"}
		}
		if r.Method == http.MethodHead {
			w.WriteHeader(code)
			return
		}
		writeStatus(w, code, body)
	})
}

func writeStatus(w http.ResponseWriter, code int, body statusResponse) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
