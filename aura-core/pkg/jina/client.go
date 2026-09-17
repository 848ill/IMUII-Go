package jina

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type ReaderResponse struct {
	Code   int    `json:"code"`
	Status int    `json:"status"`
	Data   struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Content     string `json:"content"`
	} `json:"data"`
}

func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://r.jina.ai"
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ReadURL extracts clean, structured Markdown from any PDF or web document URL
func (c *Client) ReadURL(ctx context.Context, targetURL string) (*ReaderResponse, error) {
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, targetURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create jina request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Return-Format", "markdown")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jina request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jina reader returned status %d: %s", resp.StatusCode, string(body))
	}

	var readerResp ReaderResponse
	if err := json.NewDecoder(resp.Body).Decode(&readerResp); err != nil {
		return nil, fmt.Errorf("failed to decode jina response: %w", err)
	}

	return &readerResp, nil
}
