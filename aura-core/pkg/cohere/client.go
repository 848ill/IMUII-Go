package cohere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiKey      string
	baseURL     string
	embedModel  string
	rerankModel string
	httpClient  *http.Client
}

func NewClient(apiKey, baseURL, embedModel, rerankModel string) *Client {
	if baseURL == "" {
		baseURL = "https://api.cohere.com"
	}
	// Trim trailing slashes for consistent URL construction
	for len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	if embedModel == "" {
		embedModel = "embed-multilingual-v3.0"
	}
	if rerankModel == "" {
		rerankModel = "rerank-v3.5"
	}
	return &Client{
		apiKey:      apiKey,
		baseURL:     baseURL,
		embedModel:  embedModel,
		rerankModel: rerankModel,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// EmbedRequest represents Cohere embed payload
type EmbedRequest struct {
	Model     string   `json:"model"`
	Texts     []string `json:"texts"`
	InputType string   `json:"input_type"` // "search_document" or "search_query"
}

// EmbedResponse represents Cohere embed result
type EmbedResponse struct {
	ID         string      `json:"id"`
	Texts      []string    `json:"texts"`
	Embeddings [][]float32 `json:"embeddings"`
}

// RerankRequest represents Cohere rerank payload
type RerankRequest struct {
	Model           string   `json:"model"`
	Query           string   `json:"query"`
	Documents       []string `json:"documents"`
	TopN            int      `json:"top_n,omitempty"`
	ReturnDocuments bool     `json:"return_documents"`
}

type RerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
	Document       struct {
		Text string `json:"text"`
	} `json:"document,omitempty"`
}

type RerankResponse struct {
	ID      string         `json:"id"`
	Results []RerankResult `json:"results"`
}

// Embed generates high-dimensional (1024-dim) dense vector embeddings
func (c *Client) Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	payload := EmbedRequest{
		Model:     c.embedModel,
		Texts:     texts,
		InputType: inputType,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/embed", c.baseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create embed request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cohere embed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cohere embed returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode cohere embed response: %w", err)
	}

	return embedResp.Embeddings, nil
}

// Rerank performs second-stage neural cross-encoder reranking
func (c *Client) Rerank(ctx context.Context, query string, documents []string, topN int) ([]RerankResult, error) {
	if len(documents) == 0 {
		return nil, nil
	}

	payload := RerankRequest{
		Model:           c.rerankModel,
		Query:           query,
		Documents:       documents,
		TopN:            topN,
		ReturnDocuments: false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rerank request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/rerank", c.baseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create rerank request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cohere rerank request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cohere rerank returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var rerankResp RerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&rerankResp); err != nil {
		return nil, fmt.Errorf("failed to decode cohere rerank response: %w", err)
	}

	return rerankResp.Results, nil
}
