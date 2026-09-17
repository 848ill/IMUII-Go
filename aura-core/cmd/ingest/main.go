package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"aurauii/aura-core/internal/config"
	"aurauii/aura-core/internal/models"
	"aurauii/aura-core/internal/rag"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/jina"
	"aurauii/aura-core/pkg/pinecone"
	"aurauii/aura-core/pkg/supabase"
)

func main() {
	fromSupabase := flag.Bool("from-supabase", false, "Ingest all pending documents from Supabase")
	docURL := flag.String("url", "", "Direct document URL (PDF/Web) to ingest")
	fileName := flag.String("filename", "dokumen.pdf", "Document filename")
	docTitle := flag.String("title", "Pedoman Akademik UII", "Document Title")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()

	jinaClient := jina.NewClient(cfg.JinaAPIKey, cfg.JinaReaderBaseURL)
	cohereClient := cohere.NewClient(cfg.CohereAPIKey, "", cfg.CohereEmbedModel, cfg.CohereRerankModel)
	pineconeClient := pinecone.NewClient(cfg.PineconeAPIKey, cfg.PineconeIndex, cfg.PineconeHost)

	var supabaseClient *supabase.Client
	if cfg.SupabaseURL != "" && cfg.SupabaseAnonKey != "" {
		supabaseClient = supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	}

	ingestor := rag.NewIngestor(jinaClient, cohereClient, pineconeClient, supabaseClient)

	if *fromSupabase {
		log.Println("[CLI Ingest] Fetching pending documents from Supabase...")
		count, err := ingestor.IngestPendingFromSupabase(ctx)
		if err != nil {
			log.Fatalf("[CLI Ingest] Error ingesting from Supabase: %v", err)
		}
		log.Printf("[CLI Ingest] Ingestion complete. Successfully ingested %d documents.", count)
		return
	}

	if *docURL != "" {
		log.Printf("[CLI Ingest] Ingesting single document: %s (%s)...", *fileName, *docURL)
		doc := models.Document{
			ID:       fmt.Sprintf("doc_%d", os.Getpid()),
			Title:    *docTitle,
			FileName: *fileName,
			URL:      *docURL,
		}
		if err := ingestor.IngestDocument(ctx, doc); err != nil {
			log.Fatalf("[CLI Ingest] Ingestion failed: %v", err)
		}
		log.Println("[CLI Ingest] Ingestion completed successfully!")
		return
	}

	fmt.Println("AURA Core Ingestion CLI Tool")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/ingest/main.go --from-supabase")
	fmt.Println("  go run cmd/ingest/main.go --url <URL> --filename <name> --title <title>")
}
