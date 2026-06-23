// Download the default KnightsAnalytics all-MiniLM-L6-v2 ONNX model for local Palace recall.
//
// Usage:
//
//	go run ./scripts/download_onnx_model.go
//	export MEMORY_ONNX_MODEL_PATH="$(go run ./scripts/download_onnx_model.go)"
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knights-analytics/hugot"
)

const defaultModel = "KnightsAnalytics/all-MiniLM-L6-v2"

func main() {
	dest := filepath.Join("testdata", "models")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", dest, err)
		os.Exit(1)
	}
	dir, err := hugot.DownloadModel(context.Background(), defaultModel, dest, hugot.NewDownloadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "download model: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(dir)
}