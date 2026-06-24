package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iome-sh/memory"
	"github.com/knights-analytics/hugot"
)

func downloadONNXModel(repoRoot string) (string, error) {
	dest := filepath.Join(repoRoot, "testdata", "models")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", fmt.Errorf("mkdir testdata/models: %w", err)
	}
	ctx := context.Background()
	dir, err := hugot.DownloadModel(ctx, memory.DefaultEmbeddingModelFromEnv(), dest, hugot.NewDownloadOptions())
	if err != nil {
		return "", fmt.Errorf("download onnx model: %w", err)
	}
	return dir, nil
}