package rag

import (
	"context"
	"fmt"
	"log"

	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/pkg/chunker"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/jina"
	"aurauii/aura-core/pkg/pinecone"
	"aurauii/aura-core/pkg/supabase"
)

type Ingestor struct {
	jinaClient     *jina.Client
	cohereClient   *cohere.Client
	pineconeClient *pinecone.Client
	supabaseClient *supabase.Client
	splitter       *chunker.Splitter
}

func NewIngestor(
	jinaClient *jina.Client,
	cohereClient *cohere.Client,
	pineconeClient *pinecone.Client,
	supabaseClient *supabase.Client,
) *Ingestor {
	return &Ingestor{
		jinaClient:     jinaClient,
		cohereClient:   cohereClient,
		pineconeClient: pineconeClient,
		supabaseClient: supabaseClient,
		splitter:       chunker.NewDefaultSplitter(),
	}
}

// IngestDocument processes a single document end-to-end:
// Jina (PDF to Markdown) -> Go Chunker -> Cohere 1024-dim Embeddings -> Pinecone Upsert -> Supabase done
func (ing *Ingestor) IngestDocument(ctx context.Context, doc models.Document) error {
	log.Printf("[Ingestor] Starting ingestion for: %s (ID: %s, URL: %s)", doc.FileName, doc.ID, doc.URL)

	if ing.supabaseClient != nil && doc.ID != "" {
		_ = ing.supabaseClient.UpdateDocumentStatus(ctx, doc.ID, "processing")
	}

	// 1. Jina Reader: Extract clean Markdown from PDF/Web URL
	readerResp, err := ing.jinaClient.ReadURL(ctx, doc.URL)
	if err != nil {
		if ing.supabaseClient != nil && doc.ID != "" {
			_ = ing.supabaseClient.UpdateDocumentStatus(ctx, doc.ID, "failed")
		}
		return fmt.Errorf("jina extraction failed: %w", err)
	}

	markdown := readerResp.Data.Content
	if len(markdown) == 0 {
		return fmt.Errorf("extracted markdown is empty for %s", doc.FileName)
	}
	docTitle := readerResp.Data.Title
	if docTitle == "" {
		docTitle = doc.FileName
	}

	// 2. Recursive Chunker: Split text into 1000-char chunks with 200 overlap
	rawChunks := ing.splitter.SplitText(markdown)
	log.Printf("[Ingestor] Generated %d chunks for %s", len(rawChunks), doc.FileName)
	if len(rawChunks) == 0 {
		return fmt.Errorf("no chunks created from document %s", doc.FileName)
	}

	// 3. Batch Cohere Embeddings (in batches of 48 chunks max)
	batchSize := 48
	chunks := make([]models.Chunk, len(rawChunks))

	for i := 0; i < len(rawChunks); i += batchSize {
		end := i + batchSize
		if end > len(rawChunks) {
			end = len(rawChunks)
		}

		subBatch := rawChunks[i:end]
		embeddings, err := ing.cohereClient.Embed(ctx, subBatch, "search_document")
		if err != nil {
			if ing.supabaseClient != nil && doc.ID != "" {
				_ = ing.supabaseClient.UpdateDocumentStatus(ctx, doc.ID, "failed")
			}
			return fmt.Errorf("cohere embedding failed for batch %d-%d: %w", i, end, err)
		}

		for j, emb := range embeddings {
			idx := i + j
			chunkID := fmt.Sprintf("%s_chunk_%d", doc.ID, idx)
			chunks[idx] = models.Chunk{
				ID:     chunkID,
				DocID:  doc.ID,
				Index:  idx,
				Text:   subBatch[j],
				Vector: emb,
				Metadata: map[string]interface{}{
					"title":       docTitle,
					"file_name":   doc.FileName,
					"url":         doc.URL,
					"chunk_index": idx,
					"total_chunks": len(rawChunks),
				},
			}
		}
	}

	// 4. Pinecone Upsert
	if err := ing.pineconeClient.Upsert(ctx, chunks); err != nil {
		if ing.supabaseClient != nil && doc.ID != "" {
			_ = ing.supabaseClient.UpdateDocumentStatus(ctx, doc.ID, "failed")
		}
		return fmt.Errorf("pinecone upsert failed: %w", err)
	}

	// 5. Update Supabase status to "done"
	if ing.supabaseClient != nil && doc.ID != "" {
		if err := ing.supabaseClient.UpdateDocumentStatus(ctx, doc.ID, "done"); err != nil {
			log.Printf("[Ingestor] Warning: failed to mark status done in Supabase: %v", err)
		}
	}

	log.Printf("[Ingestor] Successfully ingested %s (%d vectors upserted to Pinecone)", doc.FileName, len(chunks))
	return nil
}

// IngestPendingFromSupabase queries Supabase for all pending docs and ingests them sequentially
func (ing *Ingestor) IngestPendingFromSupabase(ctx context.Context) (int, error) {
	if ing.supabaseClient == nil {
		return 0, fmt.Errorf("supabase client is not configured")
	}

	pendingDocs, err := ing.supabaseClient.GetPendingDocuments(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch pending documents: %w", err)
	}

	successCount := 0
	for _, doc := range pendingDocs {
		if err := ing.IngestDocument(ctx, doc); err != nil {
			log.Printf("[Ingestor] Failed to ingest %s: %v", doc.FileName, err)
			continue
		}
		successCount++
	}

	return successCount, nil
}
