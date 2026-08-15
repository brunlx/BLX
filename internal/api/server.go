// Package api exposes the HTTP interface for the pentest-commander backend.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Brunlx/BLX/internal/tools"
)

// Server holds the dependencies for the HTTP handlers.
type Server struct {
	catalog    *tools.Catalog
	corsOrigin string
}

// New creates a Server backed by the given tool catalog.
// CORS is only enabled when the CORS_ORIGIN environment variable is set
// (useful for separated frontend development); same-origin requests always work.
func New(catalog *tools.Catalog) *Server {
	return &Server{
		catalog:    catalog,
		corsOrigin: os.Getenv("CORS_ORIGIN"),
	}
}

// Routes assembles the full HTTP handler with middleware and static serving.
func (s *Server) Routes(static http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/tools", s.handleListTools)
	mux.HandleFunc("GET /api/tools/{id}", s.handleGetTool)
	mux.HandleFunc("POST /api/generate", s.handleGenerate)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "endpoint não encontrado")
	})
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "endpoint não encontrado")
	})

	if static != nil {
		mux.Handle("/", static)
	}

	return s.withMiddleware(mux)
}

// writeJSON encodes a value as JSON with the given status code.
// Responses are never cached (they may contain credentials and hashes).
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api: falha ao codificar resposta: %v", err)
	}
}

// writeError encodes an error payload.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// loggingResponseWriter captures the status code for request logs.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")

		// A panic in a handler must not crash the whole process: recover,
		// log the stack and answer 500.
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("api: panic em %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				writeError(w, http.StatusInternalServerError, "erro interno")
			}
		}()

		// CORS is opt-in (CORS_ORIGIN env); disabled by default for same-origin use.
		if s.corsOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", s.corsOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			// Only answer preflights that actually ask for CORS and target the API.
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" && strings.HasPrefix(r.URL.Path, "/api/") {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		lw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)

		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s -> %d (%s)", r.Method, r.URL.EscapedPath(), lw.status, time.Since(start).Round(time.Microsecond))
		}
	})
}
