# UII-Bench-50 Automated Evaluation Runner — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an automated evaluation module that runs 50 institutional test scenarios through the AURA UII RAG pipeline, scores each with LLM-as-Judge (Context Relevance, Groundedness, Answer Relevance), computes a Harmonic Triad Score (RTS), and exports results to CSV + Markdown for thesis Chapters 4 & 5.

**Architecture:** A standalone CLI (`cmd/bench/main.go`) loads a JSON dataset of 50 scenarios clustered into 5 groups, orchestrates the existing Two-Stage RAG pipeline (Retriever + Generator) for each scenario, sends the results to a DeepSeek-based LLM Judge for 0.0–1.0 scoring on CR/G/AR, computes the harmonic mean RTS, and writes final output to CSV and Markdown files. All new code lives in `internal/benchmark/` and `cmd/bench/`. Every component is unit-tested with `httptest` mocks (no live API calls in tests).

**Tech Stack:** Go 1.27, `net/http/httptest` for mocking, `encoding/json`, `encoding/csv`, `text/template` for Markdown generation. Reuses existing `pkg/cohere`, `pkg/pinecone`, `pkg/deepseek` clients and `internal/rag` pipeline.

**Spec:** `docs/superpowers/specs/2026-09-17-uii-bench-50-design.md` (approved in brainstorming session — LLM-as-Judge approach)

## Global Constraints

- Go 1.22+ (project uses go1.27.1)
- Zero external dependencies — stdlib only (no testify, no RAGAS library)
- `pkg/` packages MUST NOT import `internal/` (existing rule)
- Files ≤ 400 lines (ECC rule, 800 max)
- Test coverage ≥ 80% per new package
- All tests use `httptest.NewServer` mocks — zero live API calls in test suite
- Pinecone index: `aurarags` (1024-dim cosine) — never `imuiirags2`

## File Structure

```
aura-core/
├── internal/benchmark/
│   ├── dataset.go          # JSON dataset loader & validator (TestScenario struct)
│   ├── dataset_test.go     # TDD: parse, validate, cluster filtering
│   ├── judge.go            # LLM-as-Judge: DeepSeek evaluates CR, G, AR
│   ├── judge_test.go       # TDD: mock DeepSeek judge responses
│   ├── evaluator.go        # Orchestrator: loop scenarios → RAG → Judge → collect
│   ├── evaluator_test.go   # TDD: mock pipeline, verify orchestration logic
│   ├── reporter.go         # CSV + Markdown export with aggregate stats
│   └── reporter_test.go    # TDD: verify CSV/MD format & content
├── cmd/bench/main.go       # CLI entrypoint with flags
└── testdata/
    └── uii_bench_50.json   # 50 scenario dataset (5 clusters × 10)
```

---

### Task 1: Dataset Loader (`internal/benchmark/dataset.go`)

**Files:**
- Create: `internal/benchmark/dataset.go`
- Create: `internal/benchmark/dataset_test.go`
- Create: `testdata/uii_bench_50_sample.json` (3 sample scenarios for tests)

**Interfaces:**
- Consumes: Nothing (first task)
- Produces:
  - `type TestScenario struct` — fields: `ID string`, `Cluster int`, `ClusterName string`, `Query string`, `ExpectedKeywords []string`, `ExpectedSourceTitles []string`
  - `type Dataset struct` — fields: `Scenarios []TestScenario`, `Version string`, `Description string`
  - `func LoadDataset(path string) (*Dataset, error)` — reads JSON file, validates all fields
  - `func (d *Dataset) FilterByCluster(cluster int) []TestScenario` — returns subset

- [ ] **Step 1: Write the failing test**

Create `internal/benchmark/dataset_test.go`:

```go
package benchmark

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDataset_ValidJSON(t *testing.T) {
	// Create temp test fixture
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_scenarios.json")
	content := `{
		"version": "1.0",
		"description": "UII-Bench-50 Test Dataset",
		"scenarios": [
			{
				"id": "S01",
				"cluster": 1,
				"cluster_name": "Prasyarat Seminar Proposal",
				"query": "Berapa minimal SKS untuk mendaftar seminar proposal skripsi?",
				"expected_keywords": ["110 SKS", "nilai E", "IPK 2.00"],
				"expected_source_titles": ["Pedoman FTI"]
			},
			{
				"id": "S02",
				"cluster": 1,
				"cluster_name": "Prasyarat Seminar Proposal",
				"query": "Apakah mahasiswa dengan nilai E bisa mendaftar sempro?",
				"expected_keywords": ["tidak boleh", "nilai E", "nol"],
				"expected_source_titles": ["Pedoman FTI"]
			},
			{
				"id": "S03",
				"cluster": 2,
				"cluster_name": "Beban SKS Semester",
				"query": "Berapa SKS maksimal yang bisa diambil jika IPS saya 3.20?",
				"expected_keywords": ["24 SKS", "IPS >= 3.00"],
				"expected_source_titles": ["Pedoman Akademik UII"]
			}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	ds, err := LoadDataset(path)
	if err != nil {
		t.Fatalf("LoadDataset failed: %v", err)
	}

	if ds.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", ds.Version)
	}
	if len(ds.Scenarios) != 3 {
		t.Fatalf("expected 3 scenarios, got %d", len(ds.Scenarios))
	}
	if ds.Scenarios[0].ID != "S01" {
		t.Errorf("expected S01, got %s", ds.Scenarios[0].ID)
	}
	if ds.Scenarios[0].Cluster != 1 {
		t.Errorf("expected cluster 1, got %d", ds.Scenarios[0].Cluster)
	}
	if ds.Scenarios[0].Query == "" {
		t.Error("query should not be empty")
	}
	if len(ds.Scenarios[0].ExpectedKeywords) != 3 {
		t.Errorf("expected 3 keywords, got %d", len(ds.Scenarios[0].ExpectedKeywords))
	}
}

func TestLoadDataset_InvalidPath(t *testing.T) {
	_, err := LoadDataset("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadDataset_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.json")
	_ = os.WriteFile(path, []byte("{invalid json"), 0644)

	_, err := LoadDataset(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadDataset_EmptyScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.json")
	_ = os.WriteFile(path, []byte(`{"version":"1.0","scenarios":[]}`), 0644)

	_, err := LoadDataset(path)
	if err == nil {
		t.Error("expected error for empty scenarios")
	}
}

func TestDataset_FilterByCluster(t *testing.T) {
	ds := &Dataset{
		Scenarios: []TestScenario{
			{ID: "S01", Cluster: 1, Query: "q1"},
			{ID: "S02", Cluster: 1, Query: "q2"},
			{ID: "S03", Cluster: 2, Query: "q3"},
		},
	}

	cluster1 := ds.FilterByCluster(1)
	if len(cluster1) != 2 {
		t.Errorf("expected 2 scenarios in cluster 1, got %d", len(cluster1))
	}

	cluster2 := ds.FilterByCluster(2)
	if len(cluster2) != 1 {
		t.Errorf("expected 1 scenario in cluster 2, got %d", len(cluster2))
	}

	cluster99 := ds.FilterByCluster(99)
	if len(cluster99) != 0 {
		t.Errorf("expected 0 scenarios in cluster 99, got %d", len(cluster99))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v`
Expected: FAIL — `package benchmark` does not exist yet

- [ ] **Step 3: Write minimal implementation**

Create `internal/benchmark/dataset.go`:

```go
package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
)

// TestScenario represents a single evaluation case from the UII-Bench-50 dataset.
type TestScenario struct {
	ID                   string   `json:"id"`
	Cluster              int      `json:"cluster"`
	ClusterName          string   `json:"cluster_name"`
	Query                string   `json:"query"`
	ExpectedKeywords     []string `json:"expected_keywords"`
	ExpectedSourceTitles []string `json:"expected_source_titles"`
}

// Dataset holds the complete benchmark scenario collection.
type Dataset struct {
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Scenarios   []TestScenario `json:"scenarios"`
}

// LoadDataset reads and validates a JSON file containing test scenarios.
func LoadDataset(path string) (*Dataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset file %s: %w", path, err)
	}

	var ds Dataset
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, fmt.Errorf("failed to parse dataset JSON: %w", err)
	}

	if len(ds.Scenarios) == 0 {
		return nil, fmt.Errorf("dataset contains no scenarios")
	}

	// Validate each scenario has required fields
	for i, s := range ds.Scenarios {
		if s.ID == "" {
			return nil, fmt.Errorf("scenario at index %d has empty ID", i)
		}
		if s.Query == "" {
			return nil, fmt.Errorf("scenario %s has empty query", s.ID)
		}
		if s.Cluster < 1 || s.Cluster > 5 {
			return nil, fmt.Errorf("scenario %s has invalid cluster %d (must be 1-5)", s.ID, s.Cluster)
		}
	}

	return &ds, nil
}

// FilterByCluster returns scenarios belonging to the specified cluster.
func (d *Dataset) FilterByCluster(cluster int) []TestScenario {
	var filtered []TestScenario
	for _, s := range d.Scenarios {
		if s.Cluster == cluster {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v`
Expected: All 5 tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/dataset.go internal/benchmark/dataset_test.go
git commit -m "feat(benchmark): add dataset loader with JSON parsing and cluster filtering"
```

---

### Task 2: LLM-as-Judge (`internal/benchmark/judge.go`)

**Files:**
- Create: `internal/benchmark/judge.go`
- Create: `internal/benchmark/judge_test.go`

**Interfaces:**
- Consumes: `pkg/deepseek.Client` (existing — `Complete(ctx, messages, temperature) (string, error)`)
- Produces:
  - `type JudgeScores struct` — fields: `ContextRelevance float64`, `Groundedness float64`, `AnswerRelevance float64`, `TriadScore float64`
  - `type JudgeInput struct` — fields: `Query string`, `RetrievedPassages []string`, `GeneratedAnswer string`
  - `type Judge struct` — constructor: `NewJudge(deepseekClient *deepseek.Client) *Judge`
  - `func (j *Judge) Evaluate(ctx context.Context, input JudgeInput) (*JudgeScores, error)` — calls DeepSeek 3× (one per metric), parses JSON scores, computes RTS
  - `func HarmonicMean(a, b, c float64) float64` — 3/(1/a+1/b+1/c), returns 0 if any input is 0

- [ ] **Step 1: Write the failing test**

Create `internal/benchmark/judge_test.go`:

```go
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
		Query:              "Berapa minimal SKS untuk sempro?",
		RetrievedPassages:  []string{"Minimal 110 SKS lulus tanpa nilai E"},
		GeneratedAnswer:    "Berdasarkan Pedoman FTI, minimal 110 SKS.",
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v -run "TestHarmonicMean|TestJudge"`
Expected: FAIL — `HarmonicMean`, `NewJudge`, `JudgeInput`, `JudgeScores` undefined

- [ ] **Step 3: Write minimal implementation**

Create `internal/benchmark/judge.go`:

```go
package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"aurauii/aura-core/pkg/deepseek"
)

// JudgeScores holds the three RAG Triad metrics plus harmonic mean.
type JudgeScores struct {
	ContextRelevance float64 `json:"context_relevance"`
	Groundedness     float64 `json:"groundedness"`
	AnswerRelevance  float64 `json:"answer_relevance"`
	TriadScore       float64 `json:"triad_score"`
}

// JudgeInput contains everything the LLM judge needs to evaluate.
type JudgeInput struct {
	Query             string
	RetrievedPassages []string
	GeneratedAnswer   string
}

// Judge uses DeepSeek as an LLM-as-Judge to score RAG output quality.
type Judge struct {
	client *deepseek.Client
}

// NewJudge creates a Judge backed by a DeepSeek client.
func NewJudge(client *deepseek.Client) *Judge {
	return &Judge{client: client}
}

// Evaluate scores the RAG output on three metrics (CR, G, AR) and computes RTS.
func (j *Judge) Evaluate(ctx context.Context, input JudgeInput) (*JudgeScores, error) {
	passages := strings.Join(input.RetrievedPassages, "\n---\n")

	cr, err := j.scoreMetric(ctx, contextRelevancePrompt(input.Query, passages))
	if err != nil {
		return nil, fmt.Errorf("context relevance scoring failed: %w", err)
	}

	g, err := j.scoreMetric(ctx, groundednessPrompt(passages, input.GeneratedAnswer))
	if err != nil {
		return nil, fmt.Errorf("groundedness scoring failed: %w", err)
	}

	ar, err := j.scoreMetric(ctx, answerRelevancePrompt(input.Query, input.GeneratedAnswer))
	if err != nil {
		return nil, fmt.Errorf("answer relevance scoring failed: %w", err)
	}

	return &JudgeScores{
		ContextRelevance: cr,
		Groundedness:     g,
		AnswerRelevance:  ar,
		TriadScore:       HarmonicMean(cr, g, ar),
	}, nil
}

// scoreMetric sends a judge prompt to DeepSeek and parses the numeric score.
func (j *Judge) scoreMetric(ctx context.Context, messages []deepseek.Message) (float64, error) {
	raw, err := j.client.Complete(ctx, messages, 0.0)
	if err != nil {
		return 0, err
	}

	// Clean markdown code fences if present
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var result struct {
		Score float64 `json:"score"`
	}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return 0, fmt.Errorf("failed to parse judge score from response %q: %w", raw, err)
	}

	return result.Score, nil
}

// HarmonicMean computes 3 / (1/a + 1/b + 1/c). Returns 0 if any input is 0.
func HarmonicMean(a, b, c float64) float64 {
	if a <= 0 || b <= 0 || c <= 0 {
		return 0
	}
	return 3.0 / (1.0/a + 1.0/b + 1.0/c)
}

func contextRelevancePrompt(query, passages string) []deepseek.Message {
	return []deepseek.Message{
		{Role: "system", Content: `You are an evaluation judge for a Retrieval-Augmented Generation (RAG) system.
Evaluate CONTEXT RELEVANCE: how relevant are the retrieved passages to the user's query?
Score from 0.0 (completely irrelevant) to 1.0 (perfectly relevant).
Respond ONLY with a JSON object: {"score": <float>}`},
		{Role: "user", Content: fmt.Sprintf("QUERY:\n%s\n\nRETRIEVED PASSAGES:\n%s", query, passages)},
	}
}

func groundednessPrompt(passages, answer string) []deepseek.Message {
	return []deepseek.Message{
		{Role: "system", Content: `You are an evaluation judge for a Retrieval-Augmented Generation (RAG) system.
Evaluate GROUNDEDNESS: is every claim in the generated answer supported by the retrieved passages?
A hallucinated or unsupported claim scores low. Score from 0.0 (completely hallucinated) to 1.0 (fully grounded).
Respond ONLY with a JSON object: {"score": <float>}`},
		{Role: "user", Content: fmt.Sprintf("RETRIEVED PASSAGES:\n%s\n\nGENERATED ANSWER:\n%s", passages, answer)},
	}
}

func answerRelevancePrompt(query, answer string) []deepseek.Message {
	return []deepseek.Message{
		{Role: "system", Content: `You are an evaluation judge for a Retrieval-Augmented Generation (RAG) system.
Evaluate ANSWER RELEVANCE: does the generated answer actually address the user's query?
Score from 0.0 (completely off-topic) to 1.0 (directly and fully answers the question).
Respond ONLY with a JSON object: {"score": <float>}`},
		{Role: "user", Content: fmt.Sprintf("QUERY:\n%s\n\nGENERATED ANSWER:\n%s", query, answer)},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v -run "TestHarmonicMean|TestJudge"`
Expected: All 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/judge.go internal/benchmark/judge_test.go
git commit -m "feat(benchmark): add LLM-as-Judge scoring for CR, G, AR, RTS"
```

---

### Task 3: Evaluation Orchestrator (`internal/benchmark/evaluator.go`)

**Files:**
- Create: `internal/benchmark/evaluator.go`
- Create: `internal/benchmark/evaluator_test.go`

**Interfaces:**
- Consumes:
  - `benchmark.Dataset` — from Task 1: `LoadDataset(path) (*Dataset, error)`
  - `benchmark.Judge` — from Task 2: `NewJudge(client) *Judge`, `Evaluate(ctx, input) (*JudgeScores, error)`
  - `rag.Retriever` — existing: `Retrieve(ctx, query) (*RetrievalResult, error)`
  - `rag.Generator` — existing: `Generate(ctx, req) (*ChatResponse, error)`
- Produces:
  - `type ScenarioResult struct` — fields: `Scenario TestScenario`, `Answer string`, `Sources []models.SourceCitation`, `Scores JudgeScores`, `Metrics models.TimingMetrics`, `Error string`
  - `type EvalConfig struct` — fields: `DatasetPath string`, `OutputDir string`, `ClusterFilter int` (0 = all)
  - `type Evaluator struct` — constructor: `NewEvaluator(generator *rag.Generator, judge *Judge) *Evaluator`
  - `func (e *Evaluator) Run(ctx context.Context, ds *Dataset, clusterFilter int) ([]ScenarioResult, error)` — iterates scenarios, calls Generate + Judge, collects results
  - `func (e *Evaluator) RunSingle(ctx context.Context, scenario TestScenario) (*ScenarioResult, error)` — single scenario evaluation

- [ ] **Step 1: Write the failing test**

Create `internal/benchmark/evaluator_test.go`:

```go
package benchmark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"aurauii/aura-core/internal/models"
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

	cohereClient := cohere.NewClient("test", "embed-multilingual-v3.0", "rerank-v3.5")
	pineconeClient := pinecone.NewClient("test", "aurarags", pineconeServer.URL)
	deepseekClient := deepseek.NewClient("test", deepseekServer.URL, "test-model")

	// Override Cohere base URL — we need to patch the client to use mock server
	// Since Cohere client has hardcoded URL, we'll test at the evaluator level
	// using a pre-built retriever with mock infrastructure
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

	cohereClient := cohere.NewClient("test", "embed-multilingual-v3.0", "rerank-v3.5")
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v -run "TestEvaluator"`
Expected: FAIL — `NewEvaluator`, `ScenarioResult` undefined

- [ ] **Step 3: Write minimal implementation**

Create `internal/benchmark/evaluator.go`:

```go
package benchmark

import (
	"context"
	"fmt"
	"log"

	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/internal/rag"
)

// ScenarioResult holds the complete output from evaluating one test scenario.
type ScenarioResult struct {
	Scenario TestScenario          `json:"scenario"`
	Answer   string                `json:"answer"`
	Sources  []models.SourceCitation `json:"sources"`
	Scores   JudgeScores           `json:"scores"`
	Metrics  models.TimingMetrics  `json:"metrics"`
	Error    string                `json:"error,omitempty"`
}

// EvalConfig holds runtime configuration for the evaluation.
type EvalConfig struct {
	DatasetPath   string
	OutputDir     string
	ClusterFilter int // 0 = run all clusters
}

// Evaluator orchestrates running scenarios through the RAG pipeline and scoring them.
type Evaluator struct {
	generator *rag.Generator
	judge     *Judge
}

// NewEvaluator creates an evaluator with the given RAG generator and judge.
func NewEvaluator(generator *rag.Generator, judge *Judge) *Evaluator {
	return &Evaluator{
		generator: generator,
		judge:     judge,
	}
}

// Run executes all scenarios (or a filtered cluster) and returns results.
func (e *Evaluator) Run(ctx context.Context, ds *Dataset, clusterFilter int) ([]ScenarioResult, error) {
	scenarios := ds.Scenarios
	if clusterFilter > 0 {
		scenarios = ds.FilterByCluster(clusterFilter)
	}

	if len(scenarios) == 0 {
		return nil, fmt.Errorf("no scenarios to evaluate (cluster filter: %d)", clusterFilter)
	}

	results := make([]ScenarioResult, 0, len(scenarios))
	for i, scenario := range scenarios {
		log.Printf("[Evaluator] Running scenario %d/%d: %s (%s)",
			i+1, len(scenarios), scenario.ID, scenario.ClusterName)

		result, err := e.RunSingle(ctx, scenario)
		if err != nil {
			log.Printf("[Evaluator] Scenario %s failed: %v", scenario.ID, err)
			results = append(results, ScenarioResult{
				Scenario: scenario,
				Error:    err.Error(),
			})
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// RunSingle evaluates one scenario end-to-end: RAG generation → LLM Judge scoring.
func (e *Evaluator) RunSingle(ctx context.Context, scenario TestScenario) (*ScenarioResult, error) {
	// 1. Run the full RAG pipeline
	req := models.ChatRequest{
		Query: scenario.Query,
	}

	resp, err := e.generator.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("RAG generation failed for %s: %w", scenario.ID, err)
	}

	// 2. Extract passage texts for the judge
	passages := make([]string, len(resp.Sources))
	for i, src := range resp.Sources {
		passages[i] = src.Excerpt
	}
	// If sources have no excerpts, use a placeholder
	if len(passages) == 0 {
		passages = []string{"(no passages retrieved)"}
	}

	// 3. Score with LLM Judge
	judgeInput := JudgeInput{
		Query:             scenario.Query,
		RetrievedPassages: passages,
		GeneratedAnswer:   resp.Answer,
	}

	scores, err := e.judge.Evaluate(ctx, judgeInput)
	if err != nil {
		return nil, fmt.Errorf("judge evaluation failed for %s: %w", scenario.ID, err)
	}

	return &ScenarioResult{
		Scenario: scenario,
		Answer:   resp.Answer,
		Sources:  resp.Sources,
		Scores:   *scores,
		Metrics:  resp.Metrics,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v -run "TestEvaluator"`
Expected: All 2 Evaluator tests PASS

**Note:** These tests will partially fail because the Cohere client has hardcoded base URLs (`https://api.cohere.com`). The mock server URLs won't be used by the Cohere client. This is a known constraint — the evaluator tests function as integration-level tests that verify orchestration logic. The Cohere calls will fail during test (no API key), and the evaluator should gracefully handle the error path. If needed, refactor will be done in Step 4a to make the test green by either: (a) adding a `BaseURL` field to `cohere.Client`, or (b) restructuring the test to use interfaces.

**Step 4a (if needed): Make Cohere client testable**

If tests fail due to hardcoded Cohere URLs, add a `BaseURL` field to `pkg/cohere/client.go`:

Patch `pkg/cohere/client.go` — add `baseURL string` field to `Client`, default to `https://api.cohere.com`, and use `c.baseURL` in `Embed()` and `Rerank()` instead of hardcoded URLs.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/evaluator.go internal/benchmark/evaluator_test.go
git commit -m "feat(benchmark): add evaluation orchestrator with RAG pipeline integration"
```

---

### Task 4: Make Cohere Client Testable (`pkg/cohere/client.go`)

**Files:**
- Modify: `pkg/cohere/client.go` — add `baseURL` field, use it in Embed/Rerank
- Create: `pkg/cohere/client_test.go` — TDD tests with httptest mocks

**Interfaces:**
- Consumes: Nothing new
- Produces:
  - Modified `cohere.NewClient(apiKey, embedModel, rerankModel string) *Client` → `cohere.NewClient(apiKey, baseURL, embedModel, rerankModel string) *Client`
  - **Breaking change**: callers of `cohere.NewClient` must add `baseURL` parameter. Update call sites: `internal/rag/ingestor.go`, `cmd/server/main.go`, `cmd/ingest/main.go`, `cmd/seed/main.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/cohere/client_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./pkg/cohere/ -v`
Expected: FAIL — `NewClient` signature mismatch (3 params vs expected 4)

- [ ] **Step 3: Write minimal implementation (modify existing)**

Patch `pkg/cohere/client.go`:

1. Add `baseURL string` field to `Client` struct
2. Update `NewClient` to accept `baseURL` parameter:
```go
func NewClient(apiKey, baseURL, embedModel, rerankModel string) *Client {
	if baseURL == "" {
		baseURL = "https://api.cohere.com"
	}
	// ... rest of constructor
}
```
3. Replace hardcoded `"https://api.cohere.com/v1/embed"` with `fmt.Sprintf("%s/v1/embed", c.baseURL)` in `Embed()`
4. Replace hardcoded `"https://api.cohere.com/v1/rerank"` with `fmt.Sprintf("%s/v1/rerank", c.baseURL)` in `Rerank()`
5. Update all call sites to pass `""` as baseURL (uses default):
   - `cmd/server/main.go`
   - `cmd/ingest/main.go`
   - `cmd/seed/main.go`

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/hanifadam/AURAUII/aura-core
go test ./pkg/cohere/ -v
go build ./...   # verify no call site breaks
```
Expected: All cohere tests PASS, build succeeds

- [ ] **Step 5: Commit**

```bash
git add pkg/cohere/client.go pkg/cohere/client_test.go cmd/
git commit -m "refactor(cohere): add configurable baseURL for testability"
```

---

### Task 5: CSV + Markdown Reporter (`internal/benchmark/reporter.go`)

**Files:**
- Create: `internal/benchmark/reporter.go`
- Create: `internal/benchmark/reporter_test.go`

**Interfaces:**
- Consumes:
  - `benchmark.ScenarioResult` — from Task 3
  - `benchmark.JudgeScores` — from Task 2
- Produces:
  - `func WriteCSV(results []ScenarioResult, w io.Writer) error` — writes CSV with columns: ID, Cluster, ClusterName, Query, CR, G, AR, RTS, EmbedMs, DenseMs, RerankMs, GenMs, TotalMs, Error
  - `func WriteMarkdown(results []ScenarioResult, w io.Writer) error` — writes thesis-ready Markdown with: per-scenario table, per-cluster aggregate table (mean CR/G/AR/RTS), overall aggregate row
  - `type AggregateStats struct` — fields: `MeanCR, MeanG, MeanAR, MeanRTS float64`, `Count int`
  - `func ComputeAggregates(results []ScenarioResult) map[int]AggregateStats` — keyed by cluster (0 = overall)

- [ ] **Step 1: Write the failing test**

Create `internal/benchmark/reporter_test.go`:

```go
package benchmark

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"aurauii/aura-core/internal/models"
)

func sampleResults() []ScenarioResult {
	return []ScenarioResult{
		{
			Scenario: TestScenario{ID: "S01", Cluster: 1, ClusterName: "Sempro", Query: "q1"},
			Answer:   "answer1",
			Scores:   JudgeScores{ContextRelevance: 0.90, Groundedness: 0.85, AnswerRelevance: 0.80, TriadScore: HarmonicMean(0.90, 0.85, 0.80)},
			Metrics:  models.TimingMetrics{EmbeddingMs: 100, DenseANNMs: 50, RerankMs: 80, GenerationMs: 500, TotalMs: 730},
		},
		{
			Scenario: TestScenario{ID: "S02", Cluster: 1, ClusterName: "Sempro", Query: "q2"},
			Answer:   "answer2",
			Scores:   JudgeScores{ContextRelevance: 0.80, Groundedness: 0.90, AnswerRelevance: 0.85, TriadScore: HarmonicMean(0.80, 0.90, 0.85)},
			Metrics:  models.TimingMetrics{EmbeddingMs: 110, DenseANNMs: 55, RerankMs: 75, GenerationMs: 480, TotalMs: 720},
		},
		{
			Scenario: TestScenario{ID: "S03", Cluster: 2, ClusterName: "SKS", Query: "q3"},
			Answer:   "answer3",
			Scores:   JudgeScores{ContextRelevance: 0.70, Groundedness: 0.95, AnswerRelevance: 0.75, TriadScore: HarmonicMean(0.70, 0.95, 0.75)},
			Metrics:  models.TimingMetrics{EmbeddingMs: 95, DenseANNMs: 45, RerankMs: 70, GenerationMs: 520, TotalMs: 730},
		},
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	results := sampleResults()

	err := WriteCSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// Header + 3 data rows
	if len(records) != 4 {
		t.Errorf("expected 4 rows (header + 3 data), got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"ID", "Cluster", "ClusterName", "Query", "CR", "G", "AR", "RTS", "EmbedMs", "DenseMs", "RerankMs", "GenMs", "TotalMs", "Error"}
	if len(records[0]) != len(expectedHeaders) {
		t.Errorf("expected %d columns, got %d", len(expectedHeaders), len(records[0]))
	}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// First data row ID
	if records[1][0] != "S01" {
		t.Errorf("first row ID = %q, want S01", records[1][0])
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	results := sampleResults()

	err := WriteMarkdown(results, &buf)
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}

	md := buf.String()

	// Should contain table headers
	if !strings.Contains(md, "| ID |") {
		t.Error("markdown should contain table header with ID column")
	}
	if !strings.Contains(md, "S01") {
		t.Error("markdown should contain scenario S01")
	}
	if !strings.Contains(md, "Rata-rata Keseluruhan") || !strings.Contains(md, "Overall") {
		// Check for aggregate section
		if !strings.Contains(md, "Agregat") && !strings.Contains(md, "Aggregate") {
			t.Error("markdown should contain aggregate statistics section")
		}
	}
}

func TestComputeAggregates(t *testing.T) {
	results := sampleResults()
	aggs := ComputeAggregates(results)

	// Should have cluster 1, cluster 2, and overall (key 0)
	if len(aggs) != 3 {
		t.Errorf("expected 3 aggregate groups, got %d", len(aggs))
	}

	overall, ok := aggs[0]
	if !ok {
		t.Fatal("missing overall aggregate (key 0)")
	}
	if overall.Count != 3 {
		t.Errorf("overall count = %d, want 3", overall.Count)
	}
	if overall.MeanCR < 0.5 || overall.MeanCR > 1.0 {
		t.Errorf("overall MeanCR out of range: %v", overall.MeanCR)
	}

	c1, ok := aggs[1]
	if !ok {
		t.Fatal("missing cluster 1 aggregate")
	}
	if c1.Count != 2 {
		t.Errorf("cluster 1 count = %d, want 2", c1.Count)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v -run "TestWrite|TestCompute"`
Expected: FAIL — `WriteCSV`, `WriteMarkdown`, `ComputeAggregates`, `AggregateStats` undefined

- [ ] **Step 3: Write minimal implementation**

Create `internal/benchmark/reporter.go`:

```go
package benchmark

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
)

// AggregateStats holds mean scores for a group of scenarios.
type AggregateStats struct {
	MeanCR  float64 `json:"mean_cr"`
	MeanG   float64 `json:"mean_g"`
	MeanAR  float64 `json:"mean_ar"`
	MeanRTS float64 `json:"mean_rts"`
	Count   int     `json:"count"`
}

// WriteCSV writes evaluation results as a CSV file.
func WriteCSV(results []ScenarioResult, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"ID", "Cluster", "ClusterName", "Query", "CR", "G", "AR", "RTS",
		"EmbedMs", "DenseMs", "RerankMs", "GenMs", "TotalMs", "Error"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, r := range results {
		row := []string{
			r.Scenario.ID,
			fmt.Sprintf("%d", r.Scenario.Cluster),
			r.Scenario.ClusterName,
			r.Scenario.Query,
			fmt.Sprintf("%.4f", r.Scores.ContextRelevance),
			fmt.Sprintf("%.4f", r.Scores.Groundedness),
			fmt.Sprintf("%.4f", r.Scores.AnswerRelevance),
			fmt.Sprintf("%.4f", r.Scores.TriadScore),
			fmt.Sprintf("%d", r.Metrics.EmbeddingMs),
			fmt.Sprintf("%d", r.Metrics.DenseANNMs),
			fmt.Sprintf("%d", r.Metrics.RerankMs),
			fmt.Sprintf("%d", r.Metrics.GenerationMs),
			fmt.Sprintf("%d", r.Metrics.TotalMs),
			r.Error,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row %s: %w", r.Scenario.ID, err)
		}
	}

	return nil
}

// WriteMarkdown writes a thesis-ready Markdown report with per-scenario and aggregate tables.
func WriteMarkdown(results []ScenarioResult, w io.Writer) error {
	aggs := ComputeAggregates(results)

	// Title
	fmt.Fprintf(w, "# Hasil Evaluasi UII-Bench-50 (Automated RAG Triad Scoring)\n\n")

	// Per-scenario table
	fmt.Fprintf(w, "## Tabel Skor Per-Skenario\n\n")
	fmt.Fprintf(w, "| ID | Cluster | Nama Kluster | Query | CR | G | AR | RTS | Total (ms) |\n")
	fmt.Fprintf(w, "|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range results {
		errNote := ""
		if r.Error != "" {
			errNote = " ⚠️"
		}
		fmt.Fprintf(w, "| %s | %d | %s | %s | %.4f | %.4f | %.4f | %.4f | %d%s |\n",
			r.Scenario.ID, r.Scenario.Cluster, r.Scenario.ClusterName,
			truncateQuery(r.Scenario.Query, 50),
			r.Scores.ContextRelevance, r.Scores.Groundedness,
			r.Scores.AnswerRelevance, r.Scores.TriadScore,
			r.Metrics.TotalMs, errNote)
	}

	// Aggregate table
	fmt.Fprintf(w, "\n## Agregat Per-Kluster\n\n")
	fmt.Fprintf(w, "| Kluster | N | Mean CR | Mean G | Mean AR | Mean RTS |\n")
	fmt.Fprintf(w, "|---|---|---|---|---|---|\n")

	// Sort cluster keys
	var keys []int
	for k := range aggs {
		if k > 0 {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)

	for _, k := range keys {
		a := aggs[k]
		fmt.Fprintf(w, "| Kluster %d | %d | %.4f | %.4f | %.4f | %.4f |\n",
			k, a.Count, a.MeanCR, a.MeanG, a.MeanAR, a.MeanRTS)
	}

	// Overall
	if overall, ok := aggs[0]; ok {
		fmt.Fprintf(w, "| **Rata-rata Keseluruhan** | **%d** | **%.4f** | **%.4f** | **%.4f** | **%.4f** |\n",
			overall.Count, overall.MeanCR, overall.MeanG, overall.MeanAR, overall.MeanRTS)
	}

	return nil
}

// ComputeAggregates computes mean scores per cluster and overall (key 0).
func ComputeAggregates(results []ScenarioResult) map[int]AggregateStats {
	type accumulator struct {
		sumCR, sumG, sumAR, sumRTS float64
		count                      int
	}

	clusters := make(map[int]*accumulator)
	overall := &accumulator{}

	for _, r := range results {
		if r.Error != "" {
			continue
		}

		if _, ok := clusters[r.Scenario.Cluster]; !ok {
			clusters[r.Scenario.Cluster] = &accumulator{}
		}
		c := clusters[r.Scenario.Cluster]
		c.sumCR += r.Scores.ContextRelevance
		c.sumG += r.Scores.Groundedness
		c.sumAR += r.Scores.AnswerRelevance
		c.sumRTS += r.Scores.TriadScore
		c.count++

		overall.sumCR += r.Scores.ContextRelevance
		overall.sumG += r.Scores.Groundedness
		overall.sumAR += r.Scores.AnswerRelevance
		overall.sumRTS += r.Scores.TriadScore
		overall.count++
	}

	result := make(map[int]AggregateStats)
	for k, acc := range clusters {
		if acc.count > 0 {
			result[k] = AggregateStats{
				MeanCR:  acc.sumCR / float64(acc.count),
				MeanG:   acc.sumG / float64(acc.count),
				MeanAR:  acc.sumAR / float64(acc.count),
				MeanRTS: acc.sumRTS / float64(acc.count),
				Count:   acc.count,
			}
		}
	}
	if overall.count > 0 {
		result[0] = AggregateStats{
			MeanCR:  overall.sumCR / float64(overall.count),
			MeanG:   overall.sumG / float64(overall.count),
			MeanAR:  overall.sumAR / float64(overall.count),
			MeanRTS: overall.sumRTS / float64(overall.count),
			Count:   overall.count,
		}
	}

	return result
}

func truncateQuery(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go test ./internal/benchmark/ -v -run "TestWrite|TestCompute"`
Expected: All 3 reporter tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/reporter.go internal/benchmark/reporter_test.go
git commit -m "feat(benchmark): add CSV and Markdown reporter with aggregate statistics"
```

---

### Task 6: CLI Entrypoint (`cmd/bench/main.go`)

**Files:**
- Create: `cmd/bench/main.go`

**Interfaces:**
- Consumes:
  - `benchmark.LoadDataset(path)` — Task 1
  - `benchmark.NewEvaluator(generator, judge)` — Task 3
  - `benchmark.NewJudge(deepseekClient)` — Task 2
  - `benchmark.WriteCSV(results, w)` — Task 5
  - `benchmark.WriteMarkdown(results, w)` — Task 5
  - `config.Load()` — existing
  - `cohere.NewClient(key, baseURL, embedModel, rerankModel)` — Task 4
  - `pinecone.NewClient(key, index, host)` — existing
  - `deepseek.NewClient(key, baseURL, model)` — existing
  - `rag.NewRetriever(cohereClient, pineconeClient)` — existing
  - `rag.NewGenerator(deepseekClient, retriever)` — existing
- Produces: Executable `cmd/bench/main.go` with CLI flags: `-dataset`, `-output`, `-cluster`, `-format`

- [ ] **Step 1: Write the implementation**

Create `cmd/bench/main.go`:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"aurauii/aura-core/internal/benchmark"
	"aurauii/aura-core/internal/config"
	"aurauii/aura-core/internal/rag"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/deepseek"
	"aurauii/aura-core/pkg/pinecone"
)

func main() {
	datasetPath := flag.String("dataset", "testdata/uii_bench_50.json", "Path to the JSON dataset file")
	outputDir := flag.String("output", "benchmark_results", "Output directory for results")
	clusterFilter := flag.Int("cluster", 0, "Filter by cluster number (0 = all)")
	format := flag.String("format", "both", "Output format: csv, markdown, or both")
	flag.Parse()

	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║       AURA UII — UII-Bench-50 Evaluation Runner        ║")
	log.Println("║     LLM-as-Judge RAG Triad Scoring (CR, G, AR, RTS)    ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")

	// 1. Load config & dataset
	cfg := config.Load()

	ds, err := benchmark.LoadDataset(*datasetPath)
	if err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}
	log.Printf("Loaded %d scenarios from %s", len(ds.Scenarios), *datasetPath)

	// 2. Initialize clients
	cohereClient := cohere.NewClient(cfg.CohereAPIKey, "", cfg.CohereEmbedModel, cfg.CohereRerankModel)
	pineconeClient := pinecone.NewClient(cfg.PineconeAPIKey, cfg.PineconeIndex, cfg.PineconeHost)
	deepseekClient := deepseek.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL, cfg.DeepSeekModel)

	// 3. Build pipeline
	retriever := rag.NewRetriever(cohereClient, pineconeClient)
	generator := rag.NewGenerator(deepseekClient, retriever)
	judge := benchmark.NewJudge(deepseekClient)
	evaluator := benchmark.NewEvaluator(generator, judge)

	// 4. Run evaluation
	ctx := context.Background()
	startTime := time.Now()
	log.Printf("Starting evaluation (cluster filter: %d)...", *clusterFilter)

	results, err := evaluator.Run(ctx, ds, *clusterFilter)
	if err != nil {
		log.Fatalf("Evaluation failed: %v", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Evaluation complete: %d scenarios in %v", len(results), elapsed)

	// 5. Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02_150405")

	// 6. Write CSV
	if *format == "csv" || *format == "both" {
		csvPath := filepath.Join(*outputDir, fmt.Sprintf("uii_bench_results_%s.csv", timestamp))
		csvFile, err := os.Create(csvPath)
		if err != nil {
			log.Fatalf("Failed to create CSV file: %v", err)
		}
		if err := benchmark.WriteCSV(results, csvFile); err != nil {
			csvFile.Close()
			log.Fatalf("Failed to write CSV: %v", err)
		}
		csvFile.Close()
		log.Printf("CSV results written to: %s", csvPath)
	}

	// 7. Write Markdown
	if *format == "markdown" || *format == "both" {
		mdPath := filepath.Join(*outputDir, fmt.Sprintf("uii_bench_results_%s.md", timestamp))
		mdFile, err := os.Create(mdPath)
		if err != nil {
			log.Fatalf("Failed to create Markdown file: %v", err)
		}
		if err := benchmark.WriteMarkdown(results, mdFile); err != nil {
			mdFile.Close()
			log.Fatalf("Failed to write Markdown: %v", err)
		}
		mdFile.Close()
		log.Printf("Markdown results written to: %s", mdPath)
	}

	// 8. Print summary
	aggs := benchmark.ComputeAggregates(results)
	if overall, ok := aggs[0]; ok {
		log.Println("═══════════════════════════════════════")
		log.Printf("  Overall Mean CR:  %.4f", overall.MeanCR)
		log.Printf("  Overall Mean G:   %.4f", overall.MeanG)
		log.Printf("  Overall Mean AR:  %.4f", overall.MeanAR)
		log.Printf("  Overall Mean RTS: %.4f", overall.MeanRTS)
		log.Printf("  Scenarios: %d | Duration: %v", overall.Count, elapsed)
		log.Println("═══════════════════════════════════════")
	}
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/hanifadam/AURAUII/aura-core && go build ./cmd/bench/`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add cmd/bench/main.go
git commit -m "feat(benchmark): add CLI entrypoint for UII-Bench-50 evaluation runner"
```

---

### Task 7: UII-Bench-50 Dataset (`testdata/uii_bench_50.json`)

**Files:**
- Create: `testdata/uii_bench_50.json`

**Interfaces:**
- Consumes: Domain knowledge from HANDOVER.md Sections 1.3 (institutional rules)
- Produces: JSON file with 50 scenarios across 5 clusters (10 per cluster)

**Cluster definitions:**
1. **Prasyarat Seminar Proposal (Sempro) Skripsi** — 10 scenarios about SKS, nilai E, IPK, Metodologi Penelitian
2. **Beban SKS Semester & IPS** — 10 scenarios about credit limits per IPS bracket
3. **Masa Berlaku SK & Dosen Pembimbing** — 10 scenarios about SK duration, perpanjangan, logbook
4. **Batas Turnitin & Plagiarisme** — 10 scenarios about similarity index, bibliography exclusion
5. **Yudisium & Kelulusan Sarjana** — 10 scenarios about 144 SKS, nilai D limit, TOEFL, hafalan

- [ ] **Step 1: Create the full 50-scenario dataset**

Create `testdata/uii_bench_50.json` with all 50 scenarios. Each scenario must have a realistic bahasa Indonesia academic query that a UII student would actually ask, grounded in the rules documented in HANDOVER.md Section 1.3.

- [ ] **Step 2: Validate dataset loads**

Run:
```bash
cd /Users/hanifadam/AURAUII/aura-core
go test ./internal/benchmark/ -v -run "TestLoadDataset"
```
Also write a quick inline validation:
```bash
go run -exec "echo" cmd/bench/main.go -dataset testdata/uii_bench_50.json 2>&1 | head -5
```

- [ ] **Step 3: Commit**

```bash
git add testdata/uii_bench_50.json
git commit -m "feat(benchmark): add UII-Bench-50 dataset with 50 institutional scenarios"
```

---

### Task 8: Coverage Verification & Final Integration Test

**Files:**
- Modify: All `*_test.go` files if needed for coverage gaps
- No new files

**Interfaces:**
- Consumes: All tasks above
- Produces: Coverage report ≥ 80% for `internal/benchmark/` package

- [ ] **Step 1: Run full test suite**

```bash
cd /Users/hanifadam/AURAUII/aura-core
go test ./internal/benchmark/ -v -count=1
```
Expected: All tests PASS

- [ ] **Step 2: Check coverage**

```bash
cd /Users/hanifadam/AURAUII/aura-core
go test -coverprofile=coverage_benchmark.out ./internal/benchmark/
go tool cover -func=coverage_benchmark.out
```
Expected: ≥ 80% coverage for `internal/benchmark/` package

- [ ] **Step 3: Run full project build and tests**

```bash
cd /Users/hanifadam/AURAUII/aura-core
go build ./...
go test ./... -v
```
Expected: Build succeeds, all tests pass across all packages

- [ ] **Step 4: Commit final state**

```bash
git add -A
git commit -m "test(benchmark): verify ≥80% coverage for UII-Bench-50 evaluation module"
```
