package rag

import (
	"context"
	"fmt"
	"sort"
	"time"

	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/pinecone"
)

type Retriever struct {
	cohereClient   *cohere.Client
	pineconeClient *pinecone.Client
	denseTopK      int
	rerankTopN     int
}

func NewRetriever(cohereClient *cohere.Client, pineconeClient *pinecone.Client) *Retriever {
	return &Retriever{
		cohereClient:   cohereClient,
		pineconeClient: pineconeClient,
		denseTopK:      20, // Stage-1 ANN
		rerankTopN:     8,  // Stage-2 Cross-Encoder
	}
}

type RetrievalResult struct {
	Passages []models.RetrievalCandidate
	Sources  []models.SourceCitation
	Metrics  models.TimingMetrics
}

// Retrieve executes the Two-Stage Neural Retrieval pipeline:
// Query -> Cohere 1024-dim -> Pinecone Top-20 -> Cohere Rerank Top-8
func (r *Retriever) Retrieve(ctx context.Context, query string) (*RetrievalResult, error) {
	var metrics models.TimingMetrics

	// Stage 1a: Query Dense Vector Embedding
	t0 := time.Now()
	qVectors, err := r.cohereClient.Embed(ctx, []string{query}, "search_query")
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	if len(qVectors) == 0 {
		return nil, fmt.Errorf("no query embedding generated")
	}
	metrics.EmbeddingMs = time.Since(t0).Milliseconds()

	// Stage 1b: Dense Approximate Nearest Neighbor (ANN) Retrieval
	t1 := time.Now()
	denseCandidates, err := r.pineconeClient.Query(ctx, qVectors[0], r.denseTopK)
	if err != nil {
		return nil, fmt.Errorf("pinecone dense retrieval failed: %w", err)
	}
	metrics.DenseANNMs = time.Since(t1).Milliseconds()

	if len(denseCandidates) == 0 {
		return &RetrievalResult{Metrics: metrics}, nil
	}

	// Stage 2: Cross-Encoder Neural Reranking
	t2 := time.Now()
	candidateTexts := make([]string, len(denseCandidates))
	for i, c := range denseCandidates {
		candidateTexts[i] = c.Text
	}

	rerankResults, err := r.cohereClient.Rerank(ctx, query, candidateTexts, r.rerankTopN)
	if err != nil {
		// Fallback: return dense candidates sorted by ANN score if rerank fails
		sort.Slice(denseCandidates, func(i, j int) bool {
			return denseCandidates[i].Score > denseCandidates[j].Score
		})
		passages := denseCandidates
		if len(passages) > r.rerankTopN {
			passages = passages[:r.rerankTopN]
		}
		metrics.RerankMs = time.Since(t2).Milliseconds()
		return &RetrievalResult{
			Passages: passages,
			Sources:  extractSources(passages),
			Metrics:  metrics,
		}, nil
	}
	metrics.RerankMs = time.Since(t2).Milliseconds()

	// Assemble final reranked passages
	finalPassages := make([]models.RetrievalCandidate, len(rerankResults))
	for i, rr := range rerankResults {
		orig := denseCandidates[rr.Index]
		finalPassages[i] = models.RetrievalCandidate{
			ID:       orig.ID,
			Text:     orig.Text,
			Score:    rr.RelevanceScore,
			Metadata: orig.Metadata,
		}
	}

	return &RetrievalResult{
		Passages: finalPassages,
		Sources:  extractSources(finalPassages),
		Metrics:  metrics,
	}, nil
}

func extractSources(candidates []models.RetrievalCandidate) []models.SourceCitation {
	seen := make(map[string]bool)
	var sources []models.SourceCitation

	for _, c := range candidates {
		title := "Pedoman Akademik UII"
		fileName := "pedoman.pdf"
		url := ""

		if t, ok := c.Metadata["title"].(string); ok && t != "" {
			title = t
		}
		if fn, ok := c.Metadata["file_name"].(string); ok && fn != "" {
			fileName = fn
		}
		if u, ok := c.Metadata["url"].(string); ok && u != "" {
			url = u
		}

		key := fmt.Sprintf("%s:%s", title, fileName)
		if !seen[key] {
			seen[key] = true
			sources = append(sources, models.SourceCitation{
				Title:    title,
				FileName: fileName,
				URL:      url,
				Score:    c.Score,
				Excerpt:  truncate(c.Text, 180),
			})
		}
	}

	return sources
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
