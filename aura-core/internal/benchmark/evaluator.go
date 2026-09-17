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
	Scenario TestScenario           `json:"scenario"`
	Answer   string                 `json:"answer"`
	Sources  []models.SourceCitation `json:"sources"`
	Scores   JudgeScores            `json:"scores"`
	Metrics  models.TimingMetrics   `json:"metrics"`
	Error    string                 `json:"error,omitempty"`
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
