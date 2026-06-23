package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/knights-analytics/hugot"
	"github.com/sudo-jin/memory"
)

func TestBench_SubsetPasses(t *testing.T) {
	modelDir := onnxModelDirForBenchTest(t)
	t.Setenv(memory.EnvONNXModelPath, modelDir)

	dataset := filepath.Join(findRepoRoot(), "testdata", "longmemeval_oracle_subset.json")

	report, err := RunBench(BenchOptions{
		DatasetPath: dataset,
		TopK:        5,
		MinRecall:   0.6,
		ModelPath:   modelDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	if report.Total != 3 {
		t.Fatalf("total = %d, want 3", report.Total)
	}
	if report.Hits < 2 {
		t.Fatalf("recall gate: %d/%d hits, need at least 2/3 (%.2f)", report.Hits, report.Total, report.Recall)
	}
	if report.Recall < 0.6 {
		t.Fatalf("recall %.2f below min-recall 0.60", report.Recall)
	}
}

func onnxModelDirForBenchTest(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv(memory.EnvONNXModelPath); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	cached := filepath.Join(findRepoRoot(), "testdata", "models", "KnightsAnalytics_all-MiniLM-L6-v2")
	if info, err := os.Stat(cached); err == nil && info.IsDir() {
		return cached
	}
	ctx := context.Background()
	dir, err := hugot.DownloadModel(ctx, "KnightsAnalytics/all-MiniLM-L6-v2", filepath.Join(findRepoRoot(), "testdata", "models"), hugot.NewDownloadOptions())
	if err != nil {
		t.Skipf("onnx model unavailable: %v", err)
	}
	return dir
}