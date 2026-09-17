package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aurauii/aura-core/internal/models"
)

type Client struct {
	projectURL string
	anonKey    string
	httpClient *http.Client
}

func NewClient(projectURL, anonKey string) *Client {
	return &Client{
		projectURL: strings.TrimRight(projectURL, "/"),
		anonKey:    anonKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetPendingDocuments fetches all documents with status="pending" from Supabase DB
func (c *Client) GetPendingDocuments(ctx context.Context) ([]models.Document, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/documents?status=eq.pending&select=*", c.projectURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get documents request: %w", err)
	}

	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase returned %d: %s", resp.StatusCode, string(b))
	}

	type rawDoc struct {
		ID        string                 `json:"id"`
		Title     string                 `json:"title"`
		FileName  string                 `json:"filename"`
		URL       string                 `json:"url"`
		Status    string                 `json:"status"`
		CreatedAt time.Time              `json:"created_at"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	var rawList []rawDoc
	if err := json.NewDecoder(resp.Body).Decode(&rawList); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	docs := make([]models.Document, len(rawList))
	for i, r := range rawList {
		docs[i] = models.Document{
			ID:        r.ID,
			Title:     r.Title,
			FileName:  r.FileName,
			URL:       r.URL,
			Status:    r.Status,
			CreatedAt: r.CreatedAt,
			Metadata:  r.Metadata,
		}
	}

	return docs, nil
}

// UpdateDocumentStatus updates status to 'done', 'processing', or 'failed'
func (c *Client) UpdateDocumentStatus(ctx context.Context, docID, status string) error {
	reqURL := fmt.Sprintf("%s/rest/v1/documents?id=eq.%s", c.projectURL, docID)
	payload := map[string]string{"status": status}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase status update returned %d: %s", resp.StatusCode, string(b))
	}

	return nil
}

// UploadFile uploads binary data to Supabase Storage bucket and returns its public URL
func (c *Client) UploadFile(ctx context.Context, bucket, storagePath, contentType string, data []byte) (string, error) {
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.projectURL, bucket, storagePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("supabase upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase storage upload error (%d): %s", resp.StatusCode, string(b))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.projectURL, bucket, storagePath)
	return publicURL, nil
}

