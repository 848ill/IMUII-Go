package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"aurauii/aura-core/internal/benchmark"
	"aurauii/aura-core/internal/config"
	"aurauii/aura-core/internal/rag"
	"aurauii/aura-core/pkg/cohere"
	"aurauii/aura-core/pkg/deepseek"
	"aurauii/aura-core/pkg/pinecone"
)

func main() {
	datasetPath := flag.String("dataset", "testdata/uii_bench_50.json", "Path to the JSON dataset file")
	outputDir := flag.String("output", "benchmark_results", "Output directory for results")
	clusterFilter := flag.Int("cluster", 0, "Filter by cluster number (0 = all)")
	format := flag.String("format", "both", "Output format: csv, markdown, or both")
	rerank := flag.Bool("rerank", true, "Enable Stage-2 Cross-Encoder reranking (set false for baseline Single-Stage)")
	flag.Parse()

	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║       AURA UII — UII-Bench-50 Evaluation Runner        ║")
	log.Println("║     LLM-as-Judge RAG Triad Scoring (CR, G, AR, RTS)    ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")

	// 1. Load config & dataset
	cfg := config.Load()

	ds, err := benchmark.LoadDataset(*datasetPath)
	if err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}
	log.Printf("Loaded %d scenarios from %s", len(ds.Scenarios), *datasetPath)

	// 2. Initialize clients
	cohereClient := cohere.NewClient(cfg.CohereAPIKey, "", cfg.CohereEmbedModel, cfg.CohereRerankModel)
	pineconeClient := pinecone.NewClient(cfg.PineconeAPIKey, cfg.PineconeIndex, cfg.PineconeHost)
	deepseekClient := deepseek.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL, cfg.DeepSeekModel)

	// 3. Build pipeline
	retriever := rag.NewRetriever(cohereClient, pineconeClient)
	if !*rerank {
		retriever.SetDisableRerank(true)
		log.Println(">>> MODE: Baseline Single-Stage RAG (Tanpa Neural Reranker) <<<")
	} else {
		log.Println(">>> MODE: Proposed Two-Stage RAG (Dengan Neural Reranker) <<<")
	}
	generator := rag.NewGenerator(deepseekClient, retriever)
	judge := benchmark.NewJudge(deepseekClient)
	evaluator := benchmark.NewEvaluator(generator, judge)

	// 4. Run evaluation
	ctx := context.Background()
	startTime := time.Now()
	log.Printf("Starting evaluation (cluster filter: %d)...", *clusterFilter)

	results, err := evaluator.Run(ctx, ds, *clusterFilter)
	if err != nil {
		log.Fatalf("Evaluation failed: %v", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Evaluation complete: %d scenarios in %v", len(results), elapsed)

	// 5. Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}

	modeTag := "two_stage"
	if !*rerank {
		modeTag = "single_stage_baseline"
	}
	timestamp := time.Now().Format("2006-01-02_150405")

	// 6. Write CSV
	if *format == "csv" || *format == "both" {
		csvPath := filepath.Join(*outputDir, fmt.Sprintf("uii_bench_%s_%s.csv", modeTag, timestamp))
		csvFile, err := os.Create(csvPath)
		if err != nil {
			log.Fatalf("Failed to create CSV file: %v", err)
		}
		if err := benchmark.WriteCSV(results, csvFile); err != nil {
			csvFile.Close()
			log.Fatalf("Failed to write CSV: %v", err)
		}
		csvFile.Close()
		log.Printf("CSV results written to: %s", csvPath)
	}

	// 7. Write Markdown
	if *format == "markdown" || *format == "both" {
		mdPath := filepath.Join(*outputDir, fmt.Sprintf("uii_bench_%s_%s.md", modeTag, timestamp))
		mdFile, err := os.Create(mdPath)
		if err != nil {
			log.Fatalf("Failed to create Markdown file: %v", err)
		}
		if err := benchmark.WriteMarkdown(results, mdFile); err != nil {
			mdFile.Close()
			log.Fatalf("Failed to write Markdown: %v", err)
		}
		mdFile.Close()
		log.Printf("Markdown results written to: %s", mdPath)
	}

	// 8. Print summary
	aggs := benchmark.ComputeAggregates(results)
	if overall, ok := aggs[0]; ok {
		log.Println("═══════════════════════════════════════")
		log.Printf("  Overall Mean CR:  %.4f", overall.MeanCR)
		log.Printf("  Overall Mean G:   %.4f", overall.MeanG)
		log.Printf("  Overall Mean AR:  %.4f", overall.MeanAR)
		log.Printf("  Overall Mean RTS: %.4f", overall.MeanRTS)
		log.Printf("  Scenarios: %d | Duration: %v", overall.Count, elapsed)
		log.Println("═══════════════════════════════════════")
	}
}
