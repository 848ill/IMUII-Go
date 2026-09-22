package models

import "time"

// Document represents an academic regulatory document
type Document struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	FileName  string                 `json:"fileName"`
	URL       string                 `json:"url"`
	Status    string                 `json:"status"` // pending, processing, done, failed
	CreatedAt time.Time              `json:"createdAt"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Chunk represents a segmented text chunk from a document
type Chunk struct {
	ID       string                 `json:"id"`
	DocID    string                 `json:"docId"`
	Index    int                    `json:"index"`
	Text     string                 `json:"text"`
	Vector   []float32              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SourceCitation represents a cited academic rule/clause
type SourceCitation struct {
	Title    string  `json:"title"`
	FileName string  `json:"fileName"`
	URL      string  `json:"url"`
	Excerpt  string  `json:"excerpt"`
	Score    float64 `json:"score"`
}

// RetrievalCandidate represents a passage returned by Pinecone or Cohere Rerank
type RetrievalCandidate struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"text"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TimingMetrics records detailed execution latency (vital for the thesis/paper)
type TimingMetrics struct {
	EmbeddingMs  int64 `json:"embeddingMs"`
	DenseANNMs   int64 `json:"denseANNMs"`
	RerankMs     int64 `json:"rerankMs"`
	GenerationMs int64 `json:"generationMs"`
	TotalMs      int64 `json:"totalMs"`
}

// ChatRequest represents an incoming query from the web UI or API
type ChatRequest struct {
	Query       string `json:"query"`
	SessionID   string `json:"sessionId,omitempty"`
	UserID      string `json:"userId,omitempty"`
	FileURL     string `json:"fileUrl,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	FileContent string `json:"fileContent,omitempty"`
	FileSummary string `json:"fileSummary,omitempty"`
}

// ChatResponse represents the final response from AURA Core
type ChatResponse struct {
	Answer    string           `json:"answer"`
	Sources   []SourceCitation `json:"sources"`
	Passages  []string         `json:"passages,omitempty"`
	Metrics   TimingMetrics    `json:"metrics"`
	SessionID string           `json:"sessionId"`
	Timestamp time.Time        `json:"timestamp"`
}

// StreamToken represents an SSE token event
type StreamToken struct {
	Type    string           `json:"type"` // "token", "sources", "metrics", "done", "error"
	Content string           `json:"content,omitempty"`
	Sources []SourceCitation `json:"sources,omitempty"`
	Metrics *TimingMetrics   `json:"metrics,omitempty"`
	Error   string           `json:"error,omitempty"`
}
