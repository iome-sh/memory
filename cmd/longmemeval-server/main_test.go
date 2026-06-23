package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/knights-analytics/hugot"
	"github.com/sudo-jin/memory"
)

func TestLongMemEval_ONNXBeatsHashSemanticSimilarity(t *testing.T) {
	modelDir := onnxModelDirForLongMemEvalTest(t)

	emb, err := memory.NewGONNXEmbedder(memory.GONNXOptions{ModelPath: modelDir, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emb.Close() })

	relatedA := "I adopted a golden retriever named Max in March 2024."
	relatedB := "My dog Max is a golden retriever I got last spring."
	unrelated := "The quarterly revenue report exceeded analyst expectations."

	vecA, err := emb.Embed(relatedA)
	if err != nil {
		t.Fatal(err)
	}
	vecB, err := emb.Embed(relatedB)
	if err != nil {
		t.Fatal(err)
	}
	vecU, err := emb.Embed(unrelated)
	if err != nil {
		t.Fatal(err)
	}

	onnxRelated := memory.CosineSimilarity(vecA, vecB)
	onnxUnrelated := memory.CosineSimilarity(vecA, vecU)
	if onnxRelated <= onnxUnrelated {
		t.Fatalf("onnx related %.4f should exceed unrelated %.4f", onnxRelated, onnxUnrelated)
	}

	dim := memory.MiniLMEmbeddingDim
	hashA := memory.GenerateSimpleEmbedding(relatedA, dim)
	hashB := memory.GenerateSimpleEmbedding(relatedB, dim)
	hashU := memory.GenerateSimpleEmbedding(unrelated, dim)
	hashRelated := memory.CosineSimilarity(hashA, hashB)
	hashUnrelated := memory.CosineSimilarity(hashA, hashU)

	if onnxRelated <= hashRelated {
		t.Fatalf("onnx related %.4f should beat hash related %.4f", onnxRelated, hashRelated)
	}
	if hashRelated > hashUnrelated+0.05 {
		t.Fatalf("hash should not separate related/unrelated strongly: related %.4f unrelated %.4f", hashRelated, hashUnrelated)
	}
}

func TestLongMemEval_ResolveEmbeddingDimMatchesMode(t *testing.T) {
	if got := memory.ResolveEmbeddingDim(""); got != memory.DefaultHashEmbeddingDim {
		t.Fatalf("hash dim = %d, want %d", got, memory.DefaultHashEmbeddingDim)
	}
	if got := memory.ResolveEmbeddingDim(t.TempDir()); got != memory.MiniLMEmbeddingDim {
		t.Fatalf("onnx dim = %d, want %d", got, memory.MiniLMEmbeddingDim)
	}
}

func onnxModelDirForLongMemEvalTest(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv(memory.EnvONNXModelPath); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	cached := filepath.Join("..", "..", "testdata", "models", "KnightsAnalytics_all-MiniLM-L6-v2")
	if info, err := os.Stat(cached); err == nil && info.IsDir() {
		return cached
	}
	ctx := context.Background()
	dir, err := hugot.DownloadModel(ctx, "KnightsAnalytics/all-MiniLM-L6-v2", filepath.Join("..", "..", "testdata", "models"), hugot.NewDownloadOptions())
	if err != nil {
		t.Skipf("onnx model unavailable: %v", err)
	}
	return dir
}