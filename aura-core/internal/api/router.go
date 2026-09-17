package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func NewRouter(h *Handler, staticDir string) http.Handler {
	mux := http.NewServeMux()

	// API Endpoints
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("POST /api/chat", h.Chat)
	mux.HandleFunc("POST /api/chat/stream", h.ChatStream)
	mux.HandleFunc("GET /api/chat/stream", h.ChatStream)
	mux.HandleFunc("POST /api/ingest", h.Ingest)
	mux.HandleFunc("POST /api/upload", h.Upload)

	// Web UI Static Files & Index
	if staticDir == "" {
		staticDir = "web"
	}

	// Serve UI
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "/index.html" || path == "/chat" {
			indexPath := filepath.Join(staticDir, "templates", "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// Check static files in web/static/
		staticFile := filepath.Join(staticDir, "static", strings.TrimPrefix(path, "/static/"))
		if _, err := os.Stat(staticFile); err == nil {
			http.ServeFile(w, r, staticFile)
			return
		}

		http.NotFound(w, r)
	})

	// Wrap with Global CORS & Logging Middleware
	return withMiddleware(mux)
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
