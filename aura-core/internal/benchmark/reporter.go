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
