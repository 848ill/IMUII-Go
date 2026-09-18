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

	// Web UI Static Files & Pages
	if staticDir == "" {
		staticDir = "web"
	}

	// Page routes
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Exact page routes
		switch path {
		case "/", "/index.html":
			serveHTML(w, r, staticDir, "index.html")
			return
		case "/chat":
			serveHTML(w, r, staticDir, "chat.html")
			return
		case "/login":
			serveAuthPage(w, r, staticDir, "login.html")
			return
		case "/auth/callback":
			serveAuthPage(w, r, staticDir, "callback.html")
			return
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

// serveHTML serves a plain HTML file from the templates directory.
func serveHTML(w http.ResponseWriter, r *http.Request, staticDir, filename string) {
	pagePath := filepath.Join(staticDir, "templates", filename)
	if _, err := os.Stat(pagePath); err == nil {
		http.ServeFile(w, r, pagePath)
		return
	}
	http.NotFound(w, r)
}

// serveAuthPage serves an HTML file and injects Supabase config via string replacement.
func serveAuthPage(w http.ResponseWriter, r *http.Request, staticDir, filename string) {
	pagePath := filepath.Join(staticDir, "templates", filename)
	data, err := os.ReadFile(pagePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	html := string(data)
	html = strings.Replace(html, `data-url=""`, `data-url="`+supabaseURL+`"`, 1)
	html = strings.Replace(html, `data-key=""`, `data-key="`+supabaseKey+`"`, 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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
