package onnxdownload

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iome-sh/memory"
)

// Hugging Face resolve URLs for BAAI/bge-small-en-v1.5 → hugot layout
// (root model.onnx, matching testdata MiniLM). KnightsAnalytics/bge-small-en-v1.5
// does not exist (auth 404). This is BAAI weights, not a KnightsAnalytics export.
// testdata/models/ is gitignored — do not vendor the ~127 MB ONNX.
var baaIHugotFiles = []struct {
	src string
	dst string
}{
	{src: "onnx/model.onnx", dst: "model.onnx"},
	{src: "config.json", dst: "config.json"},
	{src: "tokenizer.json", dst: "tokenizer.json"},
	{src: "tokenizer_config.json", dst: "tokenizer_config.json"},
	{src: "special_tokens_map.json", dst: "special_tokens_map.json"},
	{src: "vocab.txt", dst: "vocab.txt"},
}

// DefaultBAAIDir is testdata/models/BAAI_bge-small-en-v1.5 (hugot layout).
func DefaultBAAIDir() string {
	return filepath.Join("testdata", "models", memory.HugotCacheDirName(memory.BAAIONNXModelHF))
}

// BAAILayoutComplete reports whether dest looks like a usable hugot BGE dir.
func BAAILayoutComplete(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "model.onnx"))
	if err != nil || info.IsDir() || info.Size() < 1_000_000 {
		return false
	}
	for _, name := range []string{"config.json", "tokenizer.json", "vocab.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return false
		}
	}
	return true
}

// HTTPGetToFile streams url to destPath. Optional HF_TOKEN as Bearer.
func HTTPGetToFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "github.com/iome-sh/memory onnxdownload")
	if tok := strings.TrimSpace(os.Getenv("HF_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	} else if tok := strings.TrimSpace(os.Getenv("HUGGING_FACE_HUB_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	tmp := destPath + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, destPath)
}

func baaIResolveURL(rel string) string {
	return "https://huggingface.co/" + memory.BAAIONNXModelHF + "/resolve/main/" + rel
}

// FetchBAAIToHugotLayout downloads BAAI/bge-small-en-v1.5 into destParent/BAAI_bge-small-en-v1.5
// with root model.onnx. Skips the network when the layout is already complete.
func FetchBAAIToHugotLayout(ctx context.Context, destParent string) (string, error) {
	return fetchBAAI(ctx, destParent, HTTPGetToFile)
}

func fetchBAAI(ctx context.Context, destParent string, get func(context.Context, string, string) error) (string, error) {
	if get == nil {
		return "", fmt.Errorf("BAAI fetch: get func required")
	}
	parent := strings.TrimSpace(destParent)
	if parent == "" {
		parent = DefaultDestDir()
	}
	dir := filepath.Join(parent, memory.HugotCacheDirName(memory.BAAIONNXModelHF))
	if BAAILayoutComplete(dir) {
		return dir, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for _, f := range baaIHugotFiles {
		dst := filepath.Join(dir, f.dst)
		if f.dst == "model.onnx" {
			if info, err := os.Stat(dst); err == nil && !info.IsDir() && info.Size() >= 1_000_000 {
				continue
			}
		} else if _, err := os.Stat(dst); err == nil {
			continue
		}
		url := baaIResolveURL(f.src)
		if err := get(ctx, url, dst); err != nil {
			return "", fmt.Errorf("BAAI %s: %w", f.src, err)
		}
	}
	if !BAAILayoutComplete(dir) {
		return "", fmt.Errorf("BAAI hugot layout incomplete under %s", dir)
	}
	return dir, nil
}
