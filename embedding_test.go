package memory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knights-analytics/hugot"
)

func TestResolveEmbeddingDim(t *testing.T) {
	if got := ResolveEmbeddingDim(""); got != DefaultHashEmbeddingDim {
		t.Fatalf("empty path dim = %d, want %d", got, DefaultHashEmbeddingDim)
	}
	if got := ResolveEmbeddingDim("   "); got != DefaultHashEmbeddingDim {
		t.Fatalf("whitespace path dim = %d, want %d", got, DefaultHashEmbeddingDim)
	}
	if got := ResolveEmbeddingDim("/models/minilm"); got != MiniLMEmbeddingDim {
		t.Fatalf("onnx path dim = %d, want %d", got, MiniLMEmbeddingDim)
	}
}

func TestNewGONNXEmbeddingFunc_EmptyPathUsesHash(t *testing.T) {
	fn, err := NewGONNXEmbeddingFunc("")
	if err != nil {
		t.Fatal(err)
	}
	a := fn("hello", 8)
	b := fn("hello", 8)
	c := fn("world", 8)
	if !vectorsEqual(a, b) {
		t.Fatal("hash embedding should be deterministic for identical input")
	}
	if vectorsEqual(a, c) {
		t.Fatal("hash embedding should differ for different input")
	}
}

func TestResolveONNXModelDir(t *testing.T) {
	dir := t.TempDir()
	onnx := filepath.Join(dir, "model.onnx")
	if err := os.WriteFile(onnx, []byte("fake"), 0o600); err != nil {
		t.Fatal(err)
	}

	gotDir, gotFile, err := resolveONNXModelDir(dir)
	if err != nil || gotDir != dir || gotFile != "" {
		t.Fatalf("dir resolve = (%q, %q, %v)", gotDir, gotFile, err)
	}

	gotDir, gotFile, err = resolveONNXModelDir(onnx)
	if err != nil || gotDir != dir || gotFile != "model.onnx" {
		t.Fatalf("file resolve = (%q, %q, %v)", gotDir, gotFile, err)
	}
}

func TestFitEmbeddingDim(t *testing.T) {
	vec := []float32{1, 2, 3}
	if got := fitEmbeddingDim(vec, 0); len(got) != 3 {
		t.Fatalf("dim 0: got len %d", len(got))
	}
	if got := fitEmbeddingDim(vec, 2); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("truncate: %+v", got)
	}
	if got := fitEmbeddingDim(vec, 5); len(got) != 5 || got[3] != 0 {
		t.Fatalf("pad: %+v", got)
	}
}

func TestGONNXEmbedding_SemanticSimilarity(t *testing.T) {
	modelDir := onnxModelDirForTest(t)

	emb, err := NewGONNXEmbedder(GONNXOptions{ModelPath: modelDir, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emb.Close() })

	if emb.Dimension() != MiniLMEmbeddingDim {
		t.Fatalf("dimension = %d, want %d", emb.Dimension(), MiniLMEmbeddingDim)
	}

	projectA := "The Phoenix release ships memory sidecar ingest and vector retrieval."
	projectB := "Phoenix milestone covers Palace memory ingest with hybrid retrieval."
	weather := "Rain forecast for Vancouver with scattered afternoon showers."

	vecA, err := emb.Embed(projectA)
	if err != nil {
		t.Fatal(err)
	}
	vecB, err := emb.Embed(projectB)
	if err != nil {
		t.Fatal(err)
	}
	vecW, err := emb.Embed(weather)
	if err != nil {
		t.Fatal(err)
	}

	simRelated := CosineSimilarity(vecA, vecB)
	simUnrelated := CosineSimilarity(vecA, vecW)
	if simRelated <= simUnrelated {
		t.Fatalf("related similarity %.4f should exceed unrelated %.4f", simRelated, simUnrelated)
	}
	if simRelated < 0.45 {
		t.Fatalf("related similarity too low: %.4f", simRelated)
	}

	// Palace injectable path must not equal hash fallback for semantic text.
	hash := GenerateSimpleEmbedding(projectA, MiniLMEmbeddingDim)
	if vectorsEqual(vecA, hash) {
		t.Fatal("onnx embedding must differ from hash fallback")
	}
}

func TestTruncateForEmbedding(t *testing.T) {
	short := "hello world"
	if got := truncateForEmbedding(short); got != short {
		t.Fatalf("short text changed: %q", got)
	}
	long := strings.Repeat("word ", miniLMEmbedRuneBudget+50)
	got := truncateForEmbedding(long)
	if len([]rune(got)) != miniLMEmbedRuneBudget {
		t.Fatalf("rune len = %d, want %d", len([]rune(got)), miniLMEmbedRuneBudget)
	}
}

func TestGONNXEmbedding_LongInput(t *testing.T) {
	modelDir := onnxModelDirForTest(t)

	emb, err := NewGONNXEmbedder(GONNXOptions{ModelPath: modelDir, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emb.Close() })

	long := strings.Repeat("The user discussed project milestones and deployment schedules. ", 200)
	if _, err := emb.Embed(long); err != nil {
		t.Fatalf("embed long text: %v", err)
	}
	batch, err := emb.EmbedBatch([]string{long, long})
	if err != nil {
		t.Fatalf("embed batch long text: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("batch len = %d, want 2", len(batch))
	}
}

func TestGONNXEmbedBatch(t *testing.T) {
	modelDir := onnxModelDirForTest(t)

	emb, err := NewGONNXEmbedder(GONNXOptions{ModelPath: modelDir, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emb.Close() })

	texts := []string{
		"The Phoenix release ships memory sidecar ingest and vector retrieval.",
		"Phoenix milestone covers Palace memory ingest with hybrid retrieval.",
		"Rain forecast for Vancouver with scattered afternoon showers.",
	}
	batch, err := emb.EmbedBatch(texts)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != len(texts) {
		t.Fatalf("batch len = %d, want %d", len(batch), len(texts))
	}

	for i, text := range texts {
		single, err := emb.Embed(text)
		if err != nil {
			t.Fatal(err)
		}
		if sim := CosineSimilarity(batch[i], single); sim < 0.99 {
			t.Fatalf("batch[%d] vs single cosine %.4f < 0.99", i, sim)
		}
	}

	simRelated := CosineSimilarity(batch[0], batch[1])
	simUnrelated := CosineSimilarity(batch[0], batch[2])
	if simRelated <= simUnrelated {
		t.Fatalf("related similarity %.4f should exceed unrelated %.4f", simRelated, simUnrelated)
	}
}

func TestPalaceStore_SearchMemory_ONNXRerank(t *testing.T) {
	modelDir := onnxModelDirForTest(t)
	emb, err := NewGONNXEmbedder(GONNXOptions{ModelPath: modelDir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emb.Close() })

	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:            t.TempDir(),
		EmbeddingFunc:      emb.Func(),
		BatchEmbeddingFunc: emb.BatchFunc(),
	})

	entries := []MemoryEntry{
		{ID: "a", Tier: TierContextual, Content: MemoryContent{Summary: "Kubernetes pod restart runbook for broker outage"}},
		{ID: "b", Tier: TierContextual, Content: MemoryContent{Summary: "Chocolate cake recipe with espresso buttercream frosting"}},
		{ID: "c", Tier: TierContextual, Content: MemoryContent{Summary: "Broker outage playbook: restart pods and verify NATS connectivity"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	query := "broker pod restart procedure"
	queryVec, err := emb.Embed(query)
	if err != nil {
		t.Fatal(err)
	}

	results := store.SearchMemory(query, nil, 3, queryVec)
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
	if results[0].ID == "b" {
		t.Fatalf("cake recipe ranked first for infra query: %+v", results)
	}
}

func onnxModelDirForTest(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv(EnvONNXModelPath); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	cached := filepath.Join("testdata", "models", "KnightsAnalytics_all-MiniLM-L6-v2")
	if info, err := os.Stat(cached); err == nil && info.IsDir() {
		return cached
	}
	ctx := context.Background()
	dir, err := hugot.DownloadModel(ctx, "KnightsAnalytics/all-MiniLM-L6-v2", filepath.Join("testdata", "models"), hugot.NewDownloadOptions())
	if err != nil {
		t.Skipf("download onnx model: %v", err)
	}
	return dir
}

func vectorsEqual(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
