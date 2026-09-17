package benchmark

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDataset_ValidJSON(t *testing.T) {
	// Create temp test fixture
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_scenarios.json")
	content := `{
		"version": "1.0",
		"description": "UII-Bench-50 Test Dataset",
		"scenarios": [
			{
				"id": "S01",
				"cluster": 1,
				"cluster_name": "Prasyarat Seminar Proposal",
				"query": "Berapa minimal SKS untuk mendaftar seminar proposal skripsi?",
				"expected_keywords": ["110 SKS", "nilai E", "IPK 2.00"],
				"expected_source_titles": ["Pedoman FTI"]
			},
			{
				"id": "S02",
				"cluster": 1,
				"cluster_name": "Prasyarat Seminar Proposal",
				"query": "Apakah mahasiswa dengan nilai E bisa mendaftar sempro?",
				"expected_keywords": ["tidak boleh", "nilai E", "nol"],
				"expected_source_titles": ["Pedoman FTI"]
			},
			{
				"id": "S03",
				"cluster": 2,
				"cluster_name": "Beban SKS Semester",
				"query": "Berapa SKS maksimal yang bisa diambil jika IPS saya 3.20?",
				"expected_keywords": ["24 SKS", "IPS >= 3.00"],
				"expected_source_titles": ["Pedoman Akademik UII"]
			}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	ds, err := LoadDataset(path)
	if err != nil {
		t.Fatalf("LoadDataset failed: %v", err)
	}

	if ds.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", ds.Version)
	}
	if len(ds.Scenarios) != 3 {
		t.Fatalf("expected 3 scenarios, got %d", len(ds.Scenarios))
	}
	if ds.Scenarios[0].ID != "S01" {
		t.Errorf("expected S01, got %s", ds.Scenarios[0].ID)
	}
	if ds.Scenarios[0].Cluster != 1 {
		t.Errorf("expected cluster 1, got %d", ds.Scenarios[0].Cluster)
	}
	if ds.Scenarios[0].Query == "" {
		t.Error("query should not be empty")
	}
	if len(ds.Scenarios[0].ExpectedKeywords) != 3 {
		t.Errorf("expected 3 keywords, got %d", len(ds.Scenarios[0].ExpectedKeywords))
	}
}

func TestLoadDataset_InvalidPath(t *testing.T) {
	_, err := LoadDataset("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadDataset_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.json")
	_ = os.WriteFile(path, []byte("{invalid json"), 0644)

	_, err := LoadDataset(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadDataset_EmptyScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.json")
	_ = os.WriteFile(path, []byte(`{"version":"1.0","scenarios":[]}`), 0644)

	_, err := LoadDataset(path)
	if err == nil {
		t.Error("expected error for empty scenarios")
	}
}

func TestDataset_FilterByCluster(t *testing.T) {
	ds := &Dataset{
		Scenarios: []TestScenario{
			{ID: "S01", Cluster: 1, Query: "q1"},
			{ID: "S02", Cluster: 1, Query: "q2"},
			{ID: "S03", Cluster: 2, Query: "q3"},
		},
	}

	cluster1 := ds.FilterByCluster(1)
	if len(cluster1) != 2 {
		t.Errorf("expected 2 scenarios in cluster 1, got %d", len(cluster1))
	}

	cluster2 := ds.FilterByCluster(2)
	if len(cluster2) != 1 {
		t.Errorf("expected 1 scenario in cluster 2, got %d", len(cluster2))
	}

	cluster99 := ds.FilterByCluster(99)
	if len(cluster99) != 0 {
		t.Errorf("expected 0 scenarios in cluster 99, got %d", len(cluster99))
	}
}
