package benchmark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"aurauii/aura-core/pkg/deepseek"
)

func TestHarmonicMean(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  float64
		expected float64
	}{
		{"all ones", 1.0, 1.0, 1.0, 1.0},
		{"all same", 0.8, 0.8, 0.8, 0.8},
		{"mixed", 0.9, 0.8, 0.7, 0.789},
		{"has zero", 0.9, 0.0, 0.7, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HarmonicMean(tt.a, tt.b, tt.c)
			diff := result - tt.expected
			if diff < -0.01 || diff > 0.01 {
				t.Errorf("HarmonicMean(%v,%v,%v) = %v, want ~%v", tt.a, tt.b, tt.c, result, tt.expected)
			}
		})
	}
}

func TestJudge_Evaluate(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return different scores for each of the 3 evaluation calls
		var score float64
		switch callCount {
		case 1: // Context Relevance
			score = 0.85
		case 2: // Groundedness
			score = 0.90
		case 3: // Answer Relevance
			score = 0.80
		}

		resp := map[string]interface{}{
			"id": "mock",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": mustMarshal(map[string]float64{"score": score}),
					},
					"finish_reason": "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := deepseek.NewClient("test-key", mockServer.URL, "test-model")
	judge := NewJudge(client)

	input := JudgeInput{
		Query:             "Berapa minimal SKS untuk sempro?",
		RetrievedPassages: []string{"Minimal 110 SKS lulus tanpa nilai E"},
		GeneratedAnswer:   "Berdasarkan Pedoman FTI, minimal 110 SKS.",
	}

	scores, err := judge.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if scores.ContextRelevance != 0.85 {
		t.Errorf("CR = %v, want 0.85", scores.ContextRelevance)
	}
	if scores.Groundedness != 0.90 {
		t.Errorf("G = %v, want 0.90", scores.Groundedness)
	}
	if scores.AnswerRelevance != 0.80 {
		t.Errorf("AR = %v, want 0.80", scores.AnswerRelevance)
	}
	if scores.TriadScore == 0 {
		t.Error("TriadScore should not be 0")
	}
	if callCount != 3 {
		t.Errorf("expected 3 DeepSeek calls, got %d", callCount)
	}
}

func TestJudge_Evaluate_MalformedResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"id": "mock",
			"choices": []map[string]interface{}{
				{
					"index":         0,
					"message":       map[string]string{"role": "assistant", "content": "not json"},
					"finish_reason": "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := deepseek.NewClient("test-key", mockServer.URL, "test-model")
	judge := NewJudge(client)

	_, err := judge.Evaluate(context.Background(), JudgeInput{
		Query:             "test",
		RetrievedPassages: []string{"passage"},
		GeneratedAnswer:   "answer",
	})
	if err == nil {
		t.Error("expected error for malformed judge response")
	}
}

func mustMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
