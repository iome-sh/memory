package memory

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultONNXModelHF is the recommended hugot ONNX export for agent memory (384-d, CPU ORT).
	DefaultONNXModelHF = "KnightsAnalytics/bge-small-en-v1.5"
	// LegacyONNXModelHF is the previous default (384-d MiniLM).
	LegacyONNXModelHF = "KnightsAnalytics/all-MiniLM-L6-v2"
	// BGESmallEmbeddingDim is the output width for BAAI/bge-small-en-v1.5 ONNX exports.
	BGESmallEmbeddingDim = 384
	// EnvEmbeddingModelHF overrides the Hugging Face model id used for documentation and download scripts.
	EnvEmbeddingModelHF = "MEMORY_EMBEDDING_MODEL"
)

// DefaultEmbeddingModelFromEnv returns the configured HF model id or DefaultONNXModelHF.
func DefaultEmbeddingModelFromEnv() string {
	if raw := strings.TrimSpace(getenv(EnvEmbeddingModelHF)); raw != "" {
		return raw
	}
	return DefaultONNXModelHF
}

// HugotCacheDirName returns the local directory name hugot uses for a Hugging Face model id.
func HugotCacheDirName(hfModel string) string {
	return strings.ReplaceAll(strings.TrimSpace(hfModel), "/", "_")
}

// DefaultONNXModelCacheDirName is the testdata/models subdirectory for the default ONNX export.
func DefaultONNXModelCacheDirName() string {
	return HugotCacheDirName(DefaultEmbeddingModelFromEnv())
}

func getenv(key string) string {
	return strings.TrimSpace(osGetenv(key))
}

// osGetenv is a test seam over os.Getenv.
var osGetenv = func(key string) string {
	return ""
}

func init() {
	osGetenv = func(key string) string {
		return os.Getenv(key)
	}
}

// inferEmbeddingDimFromModelPath returns a best-effort dimension for a hugot model directory name.
func inferEmbeddingDimFromModelPath(modelPath string) int {
	base := strings.ToLower(filepath.Base(strings.TrimSpace(modelPath)))
	switch {
	case strings.Contains(base, "minilm"), strings.Contains(base, "bge-small"):
		return BGESmallEmbeddingDim
	default:
		return 0
	}
}
