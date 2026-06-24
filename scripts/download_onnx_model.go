// Download the default KnightsAnalytics bge-small-en-v1.5 ONNX model for local Palace recall.
//
// Usage:
//
//	go run ./scripts/download_onnx_model.go
//	export MEMORY_ONNX_MODEL_PATH="$(go run ./scripts/download_onnx_model.go)"
//
// Optional env:
//
//	MEMORY_EMBEDDING_MODEL — Hugging Face model id (default: KnightsAnalytics/bge-small-en-v1.5)
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knights-analytics/hugot"
)

func defaultModel() string {
	if raw := strings.TrimSpace(os.Getenv("MEMORY_EMBEDDING_MODEL")); raw != "" {
		return raw
	}
	return "KnightsAnalytics/bge-small-en-v1.5"
}

func main() {
	dest := filepath.Join("testdata", "models")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", dest, err)
		os.Exit(1)
	}
	dir, err := hugot.DownloadModel(context.Background(), defaultModel(), dest, hugot.NewDownloadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "download model: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(dir)
}