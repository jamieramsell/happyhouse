// Package api holds the HTTP layer for the auth service. For the Phase 0
// skeleton it exposes only health endpoints; domain, application and
// infrastructure packages arrive in Phase 1 (auth proper).
package api

import (
	"encoding/json"
	"net/http"
)

// ServiceName is reported by the health endpoints so callers can confirm
// which service answered through the gateway.
const ServiceName = "auth"

// NewRouter builds the service's HTTP handler. Routes are registered under
// the /api/v1/auth prefix so Traefik can path-route without rewriting.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/auth/healthz", healthz)
	mux.HandleFunc("GET /api/v1/auth/readyz", healthz)
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": ServiceName,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
