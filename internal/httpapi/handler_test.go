package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOperationalContract(t *testing.T) {
	cases := []struct {
		name string
		method string
		path string
		code int
		status string
	}{
		{"process alive", "GET", "/healthz", 200, "ok"},
		{"runtime not ready", "GET", "/readyz", 503, "not_ready"},
		{"unknown route", "GET", "/meetings", 404, "not_found"},
		{"no catch-all health", "GET", "/healthz/extra", 404, "not_found"},
		{"no write operations", "POST", "/healthz", 405, "method_not_allowed"},
		{"head health", "HEAD", "/healthz", 200, ""},
		{"head readiness", "HEAD", "/readyz", 503, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			NewHandler().ServeHTTP(rr, httptest.NewRequest(tc.method, tc.path, nil))
			if rr.Code != tc.code {
				t.Fatalf("code = %d, want %d", rr.Code, tc.code)
			}
			if rr.Header().Get("Content-Type") != "application/json" {
				t.Fatal("expected JSON content type")
			}
			if tc.method == http.MethodHead {
				if rr.Body.Len() != 0 { t.Fatal("HEAD must have no body") }
				return
			}
			var body statusResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatal(err) }
			if body.Status != tc.status { t.Fatalf("status = %q, want %q", body.Status, tc.status) }
			if tc.path == "/readyz" && body.Reason != "band_runtime_not_implemented" {
				t.Fatal("readiness must explain the missing runtime")
			}
			if tc.code == 405 && rr.Header().Get("Allow") != "GET, HEAD" {
				t.Fatal("method rejection must advertise supported methods")
			}
		})
	}
}
