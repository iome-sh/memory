package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/knights-analytics/hugot"
	"github.com/iome-sh/memory"
)

func TestFlexAnswer_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "string", raw: `"Paris"`, want: "Paris"},
		{name: "integer", raw: `3`, want: "3"},
		{name: "float", raw: `17.5`, want: "17.5"},
		{name: "null", raw: `null`, want: ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var a FlexAnswer
			if err := json.Unmarshal([]byte(tc.raw), &a); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := a.String(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadDataset_NumericAnswer(t *testing.T) {
	t.Parallel()
	raw := `[{"question_id":"x","question":"How many?","answer":3,"haystack_sessions":[[{"role":"user","content":"I have 3 cats"}]]}]`
	var instances []LongMemEvalInstance
	if err := json.Unmarshal([]byte(raw), &instances); err != nil {
		t.Fatalf("parse dataset: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("len = %d, want 1", len(instances))
	}
	if got := instances[0].Answer.String(); got != "3" {
		t.Fatalf("answer = %q, want %q", got, "3")
	}
}

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