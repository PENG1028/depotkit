// Package api provides the HTTP API for DBManager's Runner injection
// and management interfaces.
//
// Routes:
//   GET /api/v1/health          — Health check
//   GET /api/v1/bindings        — Query service env bindings
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/depotly/depotly/pkg/binding"
	"github.com/depotly/depotly/pkg/store"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	bindings *binding.Service
	store    *store.DB
	addr     string
}

// NewServer creates an API server with the given store.
func NewServer(db *store.DB, addr string) *Server {
	return &Server{
		bindings: binding.NewService(db),
		store:    db,
		addr:     addr,
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	// API v1 routes
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/bindings", s.handleBindings)

	// Wrap with CORS and logging
	handler := withLogging(withCORS(mux))

	log.Printf("DBManager API server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, handler)
}

// withLogging wraps a handler with request logging.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("→ %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("← %s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// withCORS adds permissive CORS headers for local development.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeJSON sends a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// handleHealth responds to health checks.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "depotly-dbmanager",
		"version":   "0.1.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleBindings responds to Runner env injection queries.
// Query params:
//
//	service (required) — service name (e.g. "poofnote")
//	env     (optional) — environment (default: "default")
//
// Response:
//
//	200: { "service": "...", "environment": "...", "status": "ready", "env": {...} }
//	400: { "error": "..." }
//	404: { "error": "...", "status": "missing_resource", "missing": [...] }
func (s *Server) handleBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service := r.URL.Query().Get("service")
	env := r.URL.Query().Get("env")
	if env == "" {
		env = "default"
	}

	if service == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'service' is required")
		return
	}

	entries, err := s.bindings.QueryServiceEnv(service, env)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	envMap := make(map[string]string)
	var missing []string

	if len(entries) == 0 {
		// Check if the service has any resources bound at all
		allEntries, _ := s.bindings.QueryServiceEnv(service, "")
		if len(allEntries) == 0 {
			// No bindings at all for this service
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"service":     service,
				"environment": env,
				"status":      "no_bindings",
				"env":         envMap,
			})
			return
		}
	}

	for _, e := range entries {
		if e.Required {
			envMap[e.EnvKey] = e.SecretRef
		} else {
			envMap[e.EnvKey] = e.SecretRef
		}
	}

	if len(missing) > 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service":     service,
			"environment": env,
			"status":      "missing_resource",
			"env":         envMap,
			"missing":     missing,
		})
		return
	}

	status := "ready"
	if len(entries) == 0 {
		status = "no_bindings"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"service":     service,
		"environment": env,
		"status":      status,
		"env":         envMap,
	})
}
