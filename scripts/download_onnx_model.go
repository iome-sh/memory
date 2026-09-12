// Download the default KnightsAnalytics bge-small-en-v1.5 ONNX model for local Palace recall.
//
// Hugging Face KnightsAnalytics/bge-small-en-v1.5 401/404 is expected (repo
// not published). Fail-soft:
//
//  1. print the hugot error on stderr
//  2. download public BAAI/bge-small-en-v1.5 ONNX and reshape to hugot layout
//     (testdata/models/BAAI_bge-small-en-v1.5/model.onnx) — not vendored
//  3. else MiniLM testdata/models/KnightsAnalytics_all-MiniLM-L6-v2 if present
//  4. else empty stdout and TTFH cost-max hash (no ONNX required)
//
// MiniLM is not the official LongMemEval V1 BGE pin. BAAI reshape is the
// comparable BGE-small-en-v1.5 ONNX. Hash-overlap unpublished. Not a leaderboard number.
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
		FetchBAAI: onnxdownload.FetchBAAIToHugotLayout,
	}))
}
