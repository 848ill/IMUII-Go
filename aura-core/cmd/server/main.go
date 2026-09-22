package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aurauii/aura-core/internal/api"
	"aurauii/aura-core/internal/config"
	"aurauii/aura-core/internal/rag"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/deepseek"
	"aurauii/aura-core/pkg/jina"
	"aurauii/aura-core/pkg/pinecone"
	"aurauii/aura-core/pkg/supabase"
)

func main() {
	log.Println("==================================================")
	log.Println("  AURA Core: Academic Universal Regulatory Assistant")
	log.Println("  Architecture: Pure Go Native Two-Stage RAG Engine")
	log.Println("==================================================")

	cfg := config.Load()

	// Initialize API Clients
	jinaClient := jina.NewClient(cfg.JinaAPIKey, cfg.JinaReaderBaseURL)
	cohereClient := cohere.NewClient(cfg.CohereAPIKey, "", cfg.CohereEmbedModel, cfg.CohereRerankModel)
	pineconeClient := pinecone.NewClient(cfg.PineconeAPIKey, cfg.PineconeIndex, cfg.PineconeHost)
	deepseekClient := deepseek.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL, cfg.DeepSeekModel)

	var supabaseClient *supabase.Client
	if cfg.SupabaseURL != "" && cfg.SupabaseAnonKey != "" {
		supabaseClient = supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	}

	// Initialize RAG Orchestrator
	ingestor := rag.NewIngestor(jinaClient, cohereClient, pineconeClient, supabaseClient)
	retriever := rag.NewRetriever(cohereClient, pineconeClient)
	generator := rag.NewGenerator(deepseekClient, retriever)

	// Initialize API Router
	handler := api.NewHandler(generator, ingestor, supabaseClient, jinaClient, cfg.AdminAPIKey)
	router := api.NewRouter(handler, "web")

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[AURA Core] Server listening on http://localhost:%d\n", cfg.Port)
		log.Printf("[AURA Core] Health endpoint: http://localhost:%d/api/health\n", cfg.Port)
		log.Printf("[AURA Core] Chat SSE endpoint: http://localhost:%d/api/chat/stream\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[AURA Core] Server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("[AURA Core] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[AURA Core] Server shutdown error: %v", err)
	}
	log.Println("[AURA Core] Server stopped cleanly.")
}
