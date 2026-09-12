// Package onnxdownload is the fail-soft helper for scripts/download_onnx_model.go.
//
// Hugging Face 401/404 for KnightsAnalytics/bge-small-en-v1.5 is expected
// (that repo does not exist). Next: reshape public BAAI/bge-small-en-v1.5
// ONNX into hugot layout. MiniLM is a local unpublished fallback, not the
// official LongMemEval V1 BGE pin. TTFH cost-max is the hash embedder.
// Hash-overlap unpublished. Not Memory GA. dual_write OFF.
package onnxdownload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iome-sh/memory"
)

const (
	// DefaultDestRel is the hugot cache parent under the repo root.
	DefaultDestRel = "testdata/models"
)

// DefaultModel is MEMORY_EMBEDDING_MODEL or KnightsAnalytics/bge-small-en-v1.5.
func DefaultModel() string {
	return memory.DefaultEmbeddingModelFromEnv()
}

// DefaultDestDir is testdata/models (repo-root relative).
func DefaultDestDir() string {
	return filepath.FromSlash(DefaultDestRel)
}

// DefaultMiniLMDir is the in-tree MiniLM ONNX path. Not the official V1 BGE pin.
func DefaultMiniLMDir() string {
	return filepath.Join("testdata", "models", memory.HugotCacheDirName(memory.LegacyONNXModelHF))
}

// Config is one download attempt. Download is required; MkdirAll/Stat default
// to os. Mkdir failures are exit 1. Hugot download errors fail soft.
type Config struct {
	Model     string
	DestDir   string
	MiniLMDir string
	Download  func(ctx context.Context, model, dest string) (string, error)
	// FetchBAAI downloads BAAI/bge-small-en-v1.5 into hugot layout under dest.
	// Nil skips this step (unit tests). KnightsAnalytics/bge-small-en-v1.5 is 404.
	FetchBAAI func(ctx context.Context, destParent string) (string, error)
	MkdirAll  func(path string, perm os.FileMode) error
	Stat      func(name string) (os.FileInfo, error)
}

// Run writes a model directory to stdout on BGE success or MiniLM fallback so
// MEMORY_ONNX_MODEL_PATH="$(go run ./scripts/download_onnx_model.go)" still
// works. Hugging Face 401/404 (and other hugot download errors) fail soft:
// MiniLM if present, else empty stdout and hash. Mkdir failures return 1.
func Run(ctx context.Context, stdout, stderr io.Writer, cfg Config) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = DefaultModel()
	}
	dest := strings.TrimSpace(cfg.DestDir)
	if dest == "" {
		dest = DefaultDestDir()
	}
	miniLM := strings.TrimSpace(cfg.MiniLMDir)
	if miniLM == "" {
		miniLM = DefaultMiniLMDir()
	}
	mkdir := cfg.MkdirAll
	if mkdir == nil {
		mkdir = os.MkdirAll
	}
	stat := cfg.Stat
	if stat == nil {
		stat = os.Stat
	}

	if err := mkdir(dest, 0o755); err != nil {
		fmt.Fprintf(stderr, "mkdir %s: %v\n", dest, err)
		return 1
	}

	baaiDir := filepath.Join(dest, memory.HugotCacheDirName(memory.BAAIONNXModelHF))
	if BAAILayoutComplete(baaiDir) {
		fmt.Fprintf(stderr, "using existing BAAI BGE-small-en-v1.5 hugot layout: %s\n", baaiDir)
		fmt.Fprintln(stderr, "source is BAAI/bge-small-en-v1.5 (onnx/model.onnx reshaped). KnightsAnalytics/bge-small-en-v1.5 does not exist.")
		fmt.Fprintln(stdout, baaiDir)
		return 0
	}

	if cfg.Download == nil {
		fmt.Fprintln(stderr, "download model: download func required")
		return 1
	}

	dir, err := cfg.Download(ctx, model, dest)
	if err == nil {
		path := strings.TrimSpace(dir)
		if path != "" {
			fmt.Fprintln(stdout, path)
		}
		return 0
	}

	fmt.Fprintf(stderr, "download model: %v\n", err)
	fmt.Fprintf(stderr, "Hugging Face 401/404 is expected for %s (repo not published).\n", memory.DefaultONNXModelHF)

	if cfg.FetchBAAI != nil {
		baai, berr := cfg.FetchBAAI(ctx, dest)
		if berr == nil && strings.TrimSpace(baai) != "" {
			fmt.Fprintf(stderr, "using BAAI/bge-small-en-v1.5 hugot layout: %s\n", baai)
			fmt.Fprintln(stderr, "this is BAAI ONNX (onnx/model.onnx reshaped to root model.onnx), not a KnightsAnalytics export.")
			fmt.Fprintln(stderr, "Hash-overlap unpublished. Not Memory GA. dual_write OFF.")
			fmt.Fprintln(stdout, strings.TrimSpace(baai))
			return 0
		}
		if berr != nil {
			fmt.Fprintf(stderr, "BAAI BGE-small-en-v1.5 fetch: %v\n", berr)
		}
	}

	fmt.Fprintf(stderr, "MiniLM fallback %s is not the official LongMemEval V1 BGE-small-en-v1.5 pin.\n", miniLM)
	fmt.Fprintln(stderr, "Hash-overlap unpublished. Not Memory GA. dual_write OFF.")

	if miniLMExists(stat, miniLM) {
		fmt.Fprintf(stderr, "using MiniLM fallback: %s\n", miniLM)
		fmt.Fprintln(stdout, miniLM)
		return 0
	}

	fmt.Fprintln(stderr, "MiniLM fallback missing; TTFH cost-max is the hash embedder (no ONNX required).")
	return 0
}

func miniLMExists(stat func(string) (os.FileInfo, error), dir string) bool {
	if stat == nil || strings.TrimSpace(dir) == "" {
		return false
	}
	info, err := stat(dir)
	return err == nil && info != nil && info.IsDir()
}
