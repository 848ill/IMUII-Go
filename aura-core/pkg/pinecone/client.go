package pinecone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"aurauii/aura-core/internal/models"
)

type Client struct {
	apiKey     string
	indexName  string
	host       string
	hostMu     sync.RWMutex
	httpClient *http.Client
}

type DescribeIndexResponse struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Dimension int    `json:"dimension"`
	Status    struct {
		Ready bool   `json:"ready"`
		State string `json:"state"`
	} `json:"status"`
}

type VectorItem struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata"`
}

type UpsertRequest struct {
	Vectors   []VectorItem `json:"vectors"`
	Namespace string       `json:"namespace,omitempty"`
}

type QueryRequest struct {
	Vector          []float32 `json:"vector"`
	TopK            int       `json:"topK"`
	IncludeMetadata bool      `json:"includeMetadata"`
	IncludeValues   bool      `json:"includeValues"`
	Namespace       string    `json:"namespace,omitempty"`
}

type QueryMatch struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type QueryResponse struct {
	Matches []QueryMatch `json:"matches"`
}

func NewClient(apiKey, indexName, host string) *Client {
	if host != "" && !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	return &Client{
		apiKey:     apiKey,
		indexName:  indexName,
		host:       strings.TrimRight(host, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ensureHost resolves the Pinecone data plane host URL dynamically if not configured
func (c *Client) ensureHost(ctx context.Context) (string, error) {
	c.hostMu.RLock()
	if c.host != "" {
		h := c.host
		c.hostMu.RUnlock()
		return h, nil
	}
	c.hostMu.RUnlock()

	c.hostMu.Lock()
	defer c.hostMu.Unlock()

	if c.host != "" {
		return c.host, nil
	}

	if c.apiKey == "" {
		return "", fmt.Errorf("pinecone API key is required")
	}

	// Call Pinecone Control Plane API to describe index
	descURL := fmt.Sprintf("https://api.pinecone.io/indexes/%s", c.indexName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, descURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create describe index request: %w", err)
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query pinecone control plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("describe index returned %d: %s", resp.StatusCode, string(b))
	}

	var desc DescribeIndexResponse
	if err := json.NewDecoder(resp.Body).Decode(&desc); err != nil {
		return "", fmt.Errorf("failed to decode index describe: %w", err)
	}

	host := desc.Host
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	c.host = strings.TrimRight(host, "/")
	return c.host, nil
}

// Upsert inserts or updates vectors with metadata in Pinecone
func (c *Client) Upsert(ctx context.Context, chunks []models.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	host, err := c.ensureHost(ctx)
	if err != nil {
		return err
	}

	vectors := make([]VectorItem, len(chunks))
	for i, chunk := range chunks {
		meta := chunk.Metadata
		if meta == nil {
			meta = make(map[string]interface{})
		}
		meta["text"] = chunk.Text
		meta["doc_id"] = chunk.DocID
		meta["chunk_index"] = chunk.Index

		vectors[i] = VectorItem{
			ID:       chunk.ID,
			Values:   chunk.Vector,
			Metadata: meta,
		}
	}

	payload := UpsertRequest{Vectors: vectors}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal upsert payload: %w", err)
	}

	upsertURL := fmt.Sprintf("%s/vectors/upsert", host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upsertURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create upsert request: %w", err)
	}

	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pinecone upsert request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pinecone upsert returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Query retrieves the top-K nearest neighbors for a dense vector
func (c *Client) Query(ctx context.Context, vector []float32, topK int) ([]models.RetrievalCandidate, error) {
	host, err := c.ensureHost(ctx)
	if err != nil {
		return nil, err
	}

	payload := QueryRequest{
		Vector:          vector,
		TopK:            topK,
		IncludeMetadata: true,
		IncludeValues:   false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query payload: %w", err)
	}

	queryURL := fmt.Sprintf("%s/query", host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pinecone query request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pinecone query returned %d: %s", resp.StatusCode, string(respBody))
	}

	var qResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qResp); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	candidates := make([]models.RetrievalCandidate, len(qResp.Matches))
	for i, match := range qResp.Matches {
		text := ""
		if t, ok := match.Metadata["text"].(string); ok {
			text = t
		}
		candidates[i] = models.RetrievalCandidate{
			ID:       match.ID,
			Text:     text,
			Score:    match.Score,
			Metadata: match.Metadata,
		}
	}

	return candidates, nil
}
