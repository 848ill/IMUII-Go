package benchmark

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
)

// AggregateStats holds mean scores for a group of scenarios.
type AggregateStats struct {
	MeanCR  float64 `json:"mean_cr"`
	MeanG   float64 `json:"mean_g"`
	MeanAR  float64 `json:"mean_ar"`
	MeanRTS float64 `json:"mean_rts"`
	Count   int     `json:"count"`
}

// WriteCSV writes evaluation results as a CSV file.
func WriteCSV(results []ScenarioResult, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"ID", "Cluster", "ClusterName", "Query", "CR", "G", "AR", "RTS",
		"EmbedMs", "DenseMs", "RerankMs", "GenMs", "TotalMs", "Error"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, r := range results {
		row := []string{
			r.Scenario.ID,
			fmt.Sprintf("%d", r.Scenario.Cluster),
			r.Scenario.ClusterName,
			r.Scenario.Query,
			fmt.Sprintf("%.4f", r.Scores.ContextRelevance),
			fmt.Sprintf("%.4f", r.Scores.Groundedness),
			fmt.Sprintf("%.4f", r.Scores.AnswerRelevance),
			fmt.Sprintf("%.4f", r.Scores.TriadScore),
			fmt.Sprintf("%d", r.Metrics.EmbeddingMs),
			fmt.Sprintf("%d", r.Metrics.DenseANNMs),
			fmt.Sprintf("%d", r.Metrics.RerankMs),
			fmt.Sprintf("%d", r.Metrics.GenerationMs),
			fmt.Sprintf("%d", r.Metrics.TotalMs),
			r.Error,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row %s: %w", r.Scenario.ID, err)
		}
	}

	// Legend / Keterangan Metrik RAG Triad & Latensi (14 columns matching CSV header)
	emptyRow := make([]string, 14)
	_ = writer.Write(emptyRow)

	legendHeader := []string{
		"# METRIK", "SINGKATAN", "NAMA LENGKAP", "DEFINISI & INDIKATOR", "RENTANG", "TARGET SKRIPSI",
		"FORMULA / ALAT UKUR", "", "", "", "", "", "", "",
	}
	_ = writer.Write(legendHeader)

	legends := [][]string{
		{"# CR", "CR", "Context Relevance", "Mengukur apakah pasal/dokumen yang diretrieve oleh Pinecone & Reranker benar-benar relevan dengan pertanyaan mahasiswa.", "0.00 - 1.00", ">= 0.70", "LLM-as-Judge (DeepSeek)", "", "", "", "", "", "", ""},
		{"# G", "G", "Groundedness (Anti-Halusinasi)", "Mengukur kesetiaan klaim jawaban terhadap dokumen rujukan resmi. 1.00 = 100% fakta bersumber dari regulasi tanpa halusinasi.", "0.00 - 1.00", ">= 0.85", "LLM-as-Judge (DeepSeek)", "", "", "", "", "", "", ""},
		{"# AR", "AR", "Answer Relevance", "Mengukur seberapa tepat, tuntas, dan langsung jawaban sistem dalam menyelesaikan masalah pertanyaan mahasiswa.", "0.00 - 1.00", ">= 0.90", "LLM-as-Judge (DeepSeek)", "", "", "", "", "", "", ""},
		{"# RTS", "RTS", "RAG Triad Score", "Skor komposit holistik RAG: rata-rata harmonik dari CR, G, dan AR.", "0.00 - 1.00", ">= 0.80", "3 / (1/CR + 1/G + 1/AR)", "", "", "", "", "", "", ""},
		{"# EmbedMs", "EmbedMs", "Embedding Latency", "Waktu pembuatan vektor representasi teks query menggunakan Cohere Multilingual v3.", "Milidetik (ms)", "< 500 ms", "Cohere Embed API", "", "", "", "", "", "", ""},
		{"# DenseMs", "DenseMs", "Dense ANN Retrieval", "Waktu pencarian kemiripan vektor Top-20 kandidat pada Pinecone Vector Space.", "Milidetik (ms)", "< 500 ms", "Pinecone Vector Index", "", "", "", "", "", "", ""},
		{"# RerankMs", "RerankMs", "Neural Reranking", "Waktu penyortiran ulang akurasi tinggi Top-8 kandidat oleh Cohere Cross-Encoder.", "Milidetik (ms)", "< 500 ms", "Cohere Rerank v3.5", "", "", "", "", "", "", ""},
		{"# GenMs", "GenMs", "Generation Latency", "Waktu inferensi dan penalaran sistemik anti-halusinasi oleh LLM DeepSeek.", "Milidetik (ms)", "< 5000 ms", "DeepSeek Chat API", "", "", "", "", "", "", ""},
		{"# TotalMs", "TotalMs", "End-to-End Latency", "Total waktu eksekusi pipeline Pure Go dari query diterima hingga respons selesai.", "Milidetik (ms)", "< 6000 ms", "Wall-clock Time", "", "", "", "", "", "", ""},
	}
	for _, l := range legends {
		_ = writer.Write(l)
	}

	return nil
}

// WriteMarkdown writes a thesis-ready Markdown report with per-scenario and aggregate tables.
func WriteMarkdown(results []ScenarioResult, w io.Writer) error {
	aggs := ComputeAggregates(results)

	// Title
	fmt.Fprintf(w, "# Hasil Evaluasi UII-Bench-50 (Automated RAG Triad Scoring)\n\n")

	// Per-scenario table
	fmt.Fprintf(w, "## Tabel Skor Per-Skenario\n\n")
	fmt.Fprintf(w, "| ID | Cluster | Nama Kluster | Query | CR | G | AR | RTS | Total (ms) |\n")
	fmt.Fprintf(w, "|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range results {
		errNote := ""
		if r.Error != "" {
			errNote = " ⚠️"
		}
		fmt.Fprintf(w, "| %s | %d | %s | %s | %.4f | %.4f | %.4f | %.4f | %d%s |\n",
			r.Scenario.ID, r.Scenario.Cluster, r.Scenario.ClusterName,
			truncateQuery(r.Scenario.Query, 50),
			r.Scores.ContextRelevance, r.Scores.Groundedness,
			r.Scores.AnswerRelevance, r.Scores.TriadScore,
			r.Metrics.TotalMs, errNote)
	}

	// Aggregate table
	fmt.Fprintf(w, "\n## Agregat Per-Kluster\n\n")
	fmt.Fprintf(w, "| Kluster | N | Mean CR | Mean G | Mean AR | Mean RTS |\n")
	fmt.Fprintf(w, "|---|---|---|---|---|---|\n")

	// Sort cluster keys
	var keys []int
	for k := range aggs {
		if k > 0 {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)

	for _, k := range keys {
		a := aggs[k]
		fmt.Fprintf(w, "| Kluster %d | %d | %.4f | %.4f | %.4f | %.4f |\n",
			k, a.Count, a.MeanCR, a.MeanG, a.MeanAR, a.MeanRTS)
	}

	// Overall
	if overall, ok := aggs[0]; ok {
		fmt.Fprintf(w, "| **Rata-rata Keseluruhan** | **%d** | **%.4f** | **%.4f** | **%.4f** | **%.4f** |\n",
			overall.Count, overall.MeanCR, overall.MeanG, overall.MeanAR, overall.MeanRTS)
	}

	// Metrics Explanation / Legend
	fmt.Fprintf(w, "\n## Penjelasan & Indikator Metrik RAG Triad\n\n")
	fmt.Fprintf(w, "| Metrik | Singkatan | Nama Lengkap | Definisi & Indikator | Rentang Nilai | Target Ideal Skripsi |\n")
	fmt.Fprintf(w, "|---|---|---|---|---|---|\n")
	fmt.Fprintf(w, "| **CR** | CR | *Context Relevance* | Mengukur apakah pasal/dokumen yang diretrieve oleh Pinecone & Reranker benar-benar relevan dengan pertanyaan mahasiswa. | 0.00 – 1.00 (0–100%%) | $\\ge 0.70$ (70%%) |\n")
	fmt.Fprintf(w, "| **G** | G | *Groundedness* (Anti-Halusinasi) | Mengukur kesetiaan klaim jawaban terhadap dokumen rujukan resmi. Nilai 1.00 berarti 100%% fakta bersumber dari regulasi tanpa halusinasi. | 0.00 – 1.00 (0–100%%) | $\\ge 0.85$ (85%%) |\n")
	fmt.Fprintf(w, "| **AR** | AR | *Answer Relevance* | Mengukur seberapa tepat, tuntas, dan langsung jawaban sistem dalam menyelesaikan masalah pertanyaan mahasiswa tanpa bertele-tele. | 0.00 – 1.00 (0–100%%) | $\\ge 0.90$ (90%%) |\n")
	fmt.Fprintf(w, "| **RTS** | RTS | *RAG Triad Score* | Skor komposit holistik RAG yang dihitung dari rata-rata harmonik: $\\frac{3}{\\frac{1}{\\text{CR}} + \\frac{1}{\\text{G}} + \\frac{1}{\\text{AR}}}$. | 0.00 – 1.00 (0–100%%) | $\\ge 0.80$ (80%%) |\n")

	fmt.Fprintf(w, "\n### Indikator Latensi Pipeline (Pure Go Native Orchestrator)\n")
	fmt.Fprintf(w, "- **EmbedMs**: Waktu embedding teks query menjadi vektor 1024-dimensi oleh Cohere Multilingual v3.\n")
	fmt.Fprintf(w, "- **DenseMs**: Waktu pencarian kemiripan vektor ANN (*Approximate Nearest Neighbor*) Top-20 di Pinecone Vector Database.\n")
	fmt.Fprintf(w, "- **RerankMs**: Waktu penyortiran ulang akurasi tinggi Top-8 kandidat oleh Cohere Cross-Encoder Neural Reranker v3.5.\n")
	fmt.Fprintf(w, "- **GenMs**: Waktu inferensi teks dan penalaran anti-halusinasi oleh LLM DeepSeek.\n")
	fmt.Fprintf(w, "- **TotalMs**: Total waktu eksekusi pipeline end-to-end dari query diterima hingga respons selesai (*wall-clock time*).\n")

	return nil
}

// ComputeAggregates computes mean scores per cluster and overall (key 0).
func ComputeAggregates(results []ScenarioResult) map[int]AggregateStats {
	type accumulator struct {
		sumCR, sumG, sumAR, sumRTS float64
		count                      int
	}

	clusters := make(map[int]*accumulator)
	overall := &accumulator{}

	for _, r := range results {
		if r.Error != "" {
			continue
		}

		if _, ok := clusters[r.Scenario.Cluster]; !ok {
			clusters[r.Scenario.Cluster] = &accumulator{}
		}
		c := clusters[r.Scenario.Cluster]
		c.sumCR += r.Scores.ContextRelevance
		c.sumG += r.Scores.Groundedness
		c.sumAR += r.Scores.AnswerRelevance
		c.sumRTS += r.Scores.TriadScore
		c.count++

		overall.sumCR += r.Scores.ContextRelevance
		overall.sumG += r.Scores.Groundedness
		overall.sumAR += r.Scores.AnswerRelevance
		overall.sumRTS += r.Scores.TriadScore
		overall.count++
	}

	result := make(map[int]AggregateStats)
	for k, acc := range clusters {
		if acc.count > 0 {
			result[k] = AggregateStats{
				MeanCR:  acc.sumCR / float64(acc.count),
				MeanG:   acc.sumG / float64(acc.count),
				MeanAR:  acc.sumAR / float64(acc.count),
				MeanRTS: acc.sumRTS / float64(acc.count),
				Count:   acc.count,
			}
		}
	}
	if overall.count > 0 {
		result[0] = AggregateStats{
			MeanCR:  overall.sumCR / float64(overall.count),
			MeanG:   overall.sumG / float64(overall.count),
			MeanAR:  overall.sumAR / float64(overall.count),
			MeanRTS: overall.sumRTS / float64(overall.count),
			Count:   overall.count,
		}
	}

	return result
}

func truncateQuery(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
