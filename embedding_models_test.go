package memory

import (
	"testing"
)

func TestDefaultEmbeddingModelFromEnv(t *testing.T) {
	t.Setenv(EnvEmbeddingModelHF, "")
	if got := DefaultEmbeddingModelFromEnv(); got != DefaultONNXModelHF {
		t.Fatalf("default model = %q, want %q", got, DefaultONNXModelHF)
	}
	t.Setenv(EnvEmbeddingModelHF, "  custom/model  ")
	if got := DefaultEmbeddingModelFromEnv(); got != "custom/model" {
		t.Fatalf("env override = %q, want custom/model", got)
	}
}

func TestInferEmbeddingDimFromModelPath(t *testing.T) {
	cases := []struct {
		path string
		want int
	}{
		{"/models/KnightsAnalytics_all-MiniLM-L6-v2", BGESmallEmbeddingDim},
		{"/models/KnightsAnalytics_bge-small-en-v1.5", BGESmallEmbeddingDim},
		{"/models/unknown-model", 0},
	}
	for _, tc := range cases {
		if got := inferEmbeddingDimFromModelPath(tc.path); got != tc.want {
			t.Fatalf("infer(%q) = %d, want %d", tc.path, got, tc.want)
		}
	}
}

func TestDefaultONNXModelCacheDirName(t *testing.T) {
	t.Setenv(EnvEmbeddingModelHF, "")
	if got := DefaultONNXModelCacheDirName(); got != "KnightsAnalytics_bge-small-en-v1.5" {
		t.Fatalf("cache dir = %q, want KnightsAnalytics_bge-small-en-v1.5", got)
	}
	if got := HugotCacheDirName("org/model-id"); got != "org_model-id" {
		t.Fatalf("HugotCacheDirName = %q, want org_model-id", got)
	}
}

func TestResolveEmbeddingDimUsesModelPath(t *testing.T) {
	if got := ResolveEmbeddingDim("/models/KnightsAnalytics_bge-small-en-v1.5"); got != BGESmallEmbeddingDim {
		t.Fatalf("bge dim = %d, want %d", got, BGESmallEmbeddingDim)
	}
	if got := ResolveEmbeddingDim("/models/custom-embedder"); got != BGESmallEmbeddingDim {
		t.Fatalf("unknown onnx dim = %d, want default %d", got, BGESmallEmbeddingDim)
	}
}
