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
