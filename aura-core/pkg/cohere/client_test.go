package cohere

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbed_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or wrong auth header")
		}

		var req EmbedRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Texts) != 2 {
			t.Errorf("expected 2 texts, got %d", len(req.Texts))
		}

		resp := EmbedResponse{
			ID:         "test-id",
			Embeddings: [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient("test-key", mockServer.URL, "embed-multilingual-v3.0", "rerank-v3.5")
	embeddings, err := client.Embed(context.Background(), []string{"hello", "world"}, "search_document")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(embeddings) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(embeddings))
	}
}

func TestEmbed_EmptyTexts(t *testing.T) {
	client := NewClient("key", "http://unused", "model", "rerank")
	result, err := client.Embed(context.Background(), []string{}, "search_query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty texts")
	}
}

func TestEmbed_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer mockServer.Close()

	client := NewClient("key", mockServer.URL, "model", "rerank")
	_, err := client.Embed(context.Background(), []string{"hello"}, "search_query")
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestRerank_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/rerank" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := RerankResponse{
			ID: "test-rerank",
			Results: []RerankResult{
				{Index: 0, RelevanceScore: 0.95},
				{Index: 2, RelevanceScore: 0.80},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient("key", mockServer.URL, "embed-model", "rerank-v3.5")
	results, err := client.Rerank(context.Background(), "query", []string{"doc1", "doc2", "doc3"}, 2)
	if err != nil {
		t.Fatalf("Rerank failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0].RelevanceScore != 0.95 {
		t.Errorf("expected score 0.95, got %v", results[0].RelevanceScore)
	}
}

func TestRerank_EmptyDocuments(t *testing.T) {
	client := NewClient("key", "http://unused", "model", "rerank")
	result, err := client.Rerank(context.Background(), "query", []string{}, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty documents")
	}
}
