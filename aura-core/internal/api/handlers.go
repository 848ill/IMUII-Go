package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/internal/rag"
	"aurauii/aura-core/pkg/jina"
	"aurauii/aura-core/pkg/supabase"
)

type Handler struct {
	generator      *rag.Generator
	ingestor       *rag.Ingestor
	supabaseClient *supabase.Client
	jinaClient     *jina.Client
	adminKey       string
}

func NewHandler(
	generator *rag.Generator,
	ingestor *rag.Ingestor,
	supabaseClient *supabase.Client,
	jinaClient *jina.Client,
	adminKey string,
) *Handler {
	return &Handler{
		generator:      generator,
		ingestor:       ingestor,
		supabaseClient: supabaseClient,
		jinaClient:     jinaClient,
		adminKey:       adminKey,
	}
}

func (h *Handler) verifyAdmin(r *http.Request) bool {
	if h.adminKey == "" {
		return true
	}
	if r.Header.Get("X-Admin-Key") == h.adminKey {
		return true
	}
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") && strings.TrimPrefix(authHeader, "Bearer ") == h.adminKey {
		return true
	}
	if r.FormValue("adminKey") == h.adminKey {
		return true
	}
	return false
}

// Health checks system status and operational telemetry
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"service":   "AURA Core Native Orchestrator",
		"version":   "2.0.0 (Pure Go Native RAG)",
		"timestamp": time.Now().Format(time.RFC3339),
		"pipeline":  "AURA Sovereign Two-Stage RAG Engine",
	})
}

// Chat handles synchronous question-answering
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	resp, err := h.generator.Generate(r.Context(), req)
	if err != nil {
		http.Error(w, "Generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ChatStream handles Server-Sent Events (SSE) real-time token streaming
func (h *Handler) ChatStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	var req models.ChatRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	} else {
		req.Query = r.URL.Query().Get("query")
		req.SessionID = r.URL.Query().Get("sessionId")
	}

	if req.Query == "" {
		http.Error(w, "Query parameter required", http.StatusBadRequest)
		return
	}

	// Set SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 1. Callback for retrieved sources
	onSources := func(sources []models.SourceCitation) error {
		data, _ := json.Marshal(sources)
		fmt.Fprintf(w, "event: sources\ndata: %s\n\n", string(data))
		flusher.Flush()
		return nil
	}

	// 2. Callback for real-time tokens
	onToken := func(token string) error {
		data, _ := json.Marshal(map[string]string{"token": token})
		fmt.Fprintf(w, "event: token\ndata: %s\n\n", string(data))
		flusher.Flush()
		return nil
	}

	// 3. Callback for final timing metrics
	onMetrics := func(metrics models.TimingMetrics) error {
		data, _ := json.Marshal(metrics)
		fmt.Fprintf(w, "event: metrics\ndata: %s\n\n", string(data))
		flusher.Flush()
		return nil
	}

	// Execute streaming RAG pipeline
	err := h.generator.StreamGenerate(r.Context(), req, onSources, onToken, onMetrics)
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(errData))
		flusher.Flush()
		return
	}

	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

// Ingest triggers batch document ingestion from Supabase or manual input
func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.verifyAdmin(r) {
		http.Error(w, "Unauthorized: Kunci admin diperlukan untuk memicu ingesti regulasi", http.StatusUnauthorized)
		return
	}

	var req struct {
		Document *models.Document `json:"document,omitempty"`
		FromDB   bool             `json:"fromDB"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.FromDB || req.Document == nil {
		count, err := h.ingestor.IngestPendingFromSupabase(r.Context())
		if err != nil {
			http.Error(w, "Ingestion from DB failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Successfully processed %d pending documents from Supabase", count),
			"count":   count,
		})
		return
	}

	if err := h.ingestor.IngestDocument(r.Context(), *req.Document); err != nil {
		http.Error(w, "Ingestion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Document ingested successfully into knowledge base",
	})
}
