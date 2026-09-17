package benchmark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"aurauii/aura-core/internal/rag"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/deepseek"
	"aurauii/aura-core/pkg/pinecone"
)

// mockCohereServer returns a server that handles embed and rerank requests.
func mockCohereServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/embed" {
			resp := cohere.EmbedResponse{
				ID:         "mock-embed",
				Embeddings: [][]float32{make([]float32, 1024)},
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.URL.Path == "/v1/rerank" {
			resp := cohere.RerankResponse{
				ID: "mock-rerank",
				Results: []cohere.RerankResult{
					{Index: 0, RelevanceScore: 0.95},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
}

// mockPineconeServer returns a server that handles query requests.
func mockPineconeServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := pinecone.QueryResponse{
			Matches: []pinecone.QueryMatch{
				{
					ID:    "chunk-1",
					Score: 0.88,
					Metadata: map[string]interface{}{
						"text":      "Minimal 110 SKS lulus tanpa nilai E, IPK >= 2.00 [Pedoman FTI Pasal 2]",
						"title":     "Pedoman FTI",
						"file_name": "pedoman_fti.pdf",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// mockDeepSeekServer returns a server for both generation and judge calls.
func mockDeepSeekServer() (*httptest.Server, *atomic.Int32) {
	callCount := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		var content string
		if n == 1 {
			// First call = generation
			content = "Berdasarkan Pedoman FTI Pasal 2, minimal 110 SKS."
		} else {
			// Subsequent calls = judge scoring
			content = `{"score": 0.85}`
		}

		resp := map[string]interface{}{
			"id": "mock",
			"choices": []map[string]interface{}{
				{
					"index":         0,
					"message":       map[string]string{"role": "assistant", "content": content},
					"finish_reason": "stop",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	return server, callCount
}

func TestEvaluator_RunSingle(t *testing.T) {
	cohereServer := mockCohereServer()
	defer cohereServer.Close()
	pineconeServer := mockPineconeServer()
	defer pineconeServer.Close()
	deepseekServer, callCount := mockDeepSeekServer()
	defer deepseekServer.Close()

	// Pass mock server URLs as baseURL so no real API calls are made
	cohereClient := cohere.NewClient("test", cohereServer.URL, "embed-multilingual-v3.0", "rerank-v3.5")
	pineconeClient := pinecone.NewClient("test", "aurarags", pineconeServer.URL)
	deepseekClient := deepseek.NewClient("test", deepseekServer.URL, "test-model")

	retriever := rag.NewRetriever(cohereClient, pineconeClient)
	generator := rag.NewGenerator(deepseekClient, retriever)
	judge := NewJudge(deepseekClient)
	evaluator := NewEvaluator(generator, judge)

	scenario := TestScenario{
		ID:          "S01",
		Cluster:     1,
		ClusterName: "Prasyarat Seminar Proposal",
		Query:       "Berapa minimal SKS untuk mendaftar sempro?",
	}

	result, err := evaluator.RunSingle(context.Background(), scenario)
	if err != nil {
		t.Fatalf("RunSingle failed: %v", err)
	}

	if result.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if result.Scores.ContextRelevance == 0 {
		t.Error("expected non-zero CR score")
	}
	if result.Scores.TriadScore == 0 {
		t.Error("expected non-zero Triad Score")
	}
	// 1 generation + 3 judge calls = 4 DeepSeek calls
	if callCount.Load() != 4 {
		t.Errorf("expected 4 DeepSeek calls, got %d", callCount.Load())
	}
}

func TestEvaluator_Run(t *testing.T) {
	cohereServer := mockCohereServer()
	defer cohereServer.Close()
	pineconeServer := mockPineconeServer()
	defer pineconeServer.Close()
	deepseekServer, _ := mockDeepSeekServer()
	defer deepseekServer.Close()

	cohereClient := cohere.NewClient("test", cohereServer.URL, "embed-multilingual-v3.0", "rerank-v3.5")
	pineconeClient := pinecone.NewClient("test", "aurarags", pineconeServer.URL)
	deepseekClient := deepseek.NewClient("test", deepseekServer.URL, "test-model")

	retriever := rag.NewRetriever(cohereClient, pineconeClient)
	generator := rag.NewGenerator(deepseekClient, retriever)
	judge := NewJudge(deepseekClient)
	evaluator := NewEvaluator(generator, judge)

	ds := &Dataset{
		Scenarios: []TestScenario{
			{ID: "S01", Cluster: 1, Query: "q1"},
			{ID: "S02", Cluster: 1, Query: "q2"},
			{ID: "S03", Cluster: 2, Query: "q3"},
		},
	}

	// Run all clusters
	results, err := evaluator.Run(context.Background(), ds, 0)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Run cluster 1 only
	results1, err := evaluator.Run(context.Background(), ds, 1)
	if err != nil {
		t.Fatalf("Run cluster 1 failed: %v", err)
	}
	if len(results1) != 2 {
		t.Errorf("expected 2 results for cluster 1, got %d", len(results1))
	}
}
