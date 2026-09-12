// Download the default KnightsAnalytics bge-small-en-v1.5 ONNX model for local Palace recall.
//
// Hugging Face BGE download may 401/404. That is expected (optional ONNX) and
// is not a broken install. Fail-soft:
//
//  1. print the hugot error on stderr
//  2. if testdata/models/KnightsAnalytics_all-MiniLM-L6-v2 exists, print that
//     path on stdout (so MEMORY_ONNX_MODEL_PATH="$(go run …)" still works)
//  3. else print that TTFH cost-max is the hash embedder and exit 0 with no
//     stdout path
//
// MiniLM is a local 384-d fallback, not the official LongMemEval V1
// BGE-small-en-v1.5 pin. Hash-overlap unpublished. Not Memory GA.
// Mkdir failures still exit 1. This repo does not vendor BGE.
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
	"os"

	"github.com/iome-sh/memory/internal/onnxdownload"
	"github.com/knights-analytics/hugot"
)

func main() {
	os.Exit(onnxdownload.Run(context.Background(), os.Stdout, os.Stderr, onnxdownload.Config{
		Download: func(ctx context.Context, model, dest string) (string, error) {
			return hugot.DownloadModel(ctx, model, dest, hugot.NewDownloadOptions())
		},
	}))
}
