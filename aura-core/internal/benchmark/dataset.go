package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
)

// TestScenario represents a single evaluation case from the UII-Bench-50 dataset.
type TestScenario struct {
	ID                   string   `json:"id"`
	Cluster              int      `json:"cluster"`
	ClusterName          string   `json:"cluster_name"`
	Query                string   `json:"query"`
	ExpectedKeywords     []string `json:"expected_keywords"`
	ExpectedSourceTitles []string `json:"expected_source_titles"`
}

// Dataset holds the complete benchmark scenario collection.
type Dataset struct {
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Scenarios   []TestScenario `json:"scenarios"`
}

// LoadDataset reads and validates a JSON file containing test scenarios.
func LoadDataset(path string) (*Dataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset file %s: %w", path, err)
	}

	var ds Dataset
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, fmt.Errorf("failed to parse dataset JSON: %w", err)
	}

	if len(ds.Scenarios) == 0 {
		return nil, fmt.Errorf("dataset contains no scenarios")
	}

	// Validate each scenario has required fields
	for i, s := range ds.Scenarios {
		if s.ID == "" {
			return nil, fmt.Errorf("scenario at index %d has empty ID", i)
		}
		if s.Query == "" {
			return nil, fmt.Errorf("scenario %s has empty query", s.ID)
		}
		if s.Cluster < 1 || s.Cluster > 10 {
			return nil, fmt.Errorf("scenario %s has invalid cluster %d (must be 1-10)", s.ID, s.Cluster)
		}
	}

	return &ds, nil
}

// FilterByCluster returns scenarios belonging to the specified cluster.
func (d *Dataset) FilterByCluster(cluster int) []TestScenario {
	var filtered []TestScenario
	for _, s := range d.Scenarios {
		if s.Cluster == cluster {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
