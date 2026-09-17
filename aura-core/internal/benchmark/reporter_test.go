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
