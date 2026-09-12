package onnxdownload

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iome-sh/memory"
)

func TestDefaultMiniLMDir(t *testing.T) {
	want := filepath.Join("testdata", "models", "KnightsAnalytics_all-MiniLM-L6-v2")
	if got := DefaultMiniLMDir(); got != want {
		t.Fatalf("DefaultMiniLMDir = %q, want %q", got, want)
	}
	if DefaultMiniLMDir() == filepath.Join("testdata", "models", memory.HugotCacheDirName(memory.DefaultONNXModelHF)) {
		t.Fatal("MiniLM path must not be the official BGE-small-en-v1.5 cache dir")
	}
}

func TestDefaultModel(t *testing.T) {
	t.Setenv(memory.EnvEmbeddingModelHF, "")
	if got := DefaultModel(); got != memory.DefaultONNXModelHF {
		t.Fatalf("default model = %q, want %q", got, memory.DefaultONNXModelHF)
	}
	t.Setenv(memory.EnvEmbeddingModelHF, "  org/custom  ")
	if got := DefaultModel(); got != "org/custom" {
		t.Fatalf("env override = %q, want org/custom", got)
	}
}

func TestRun_FetchBAAIAfter404(t *testing.T) {
	var stdout, stderr bytes.Buffer
	baai := filepath.Join(t.TempDir(), "BAAI_bge-small-en-v1.5")
	got := Run(context.Background(), &stdout, &stderr, Config{
		DestDir: t.TempDir(),
		Download: func(context.Context, string, string) (string, error) {
			return "", fmt.Errorf("hugot: 404 Not Found")
		},
		FetchBAAI: func(context.Context, string) (string, error) {
			return baai, nil
		},
	})
	if got != 0 {
		t.Fatalf("exit = %d stderr=%q", got, stderr.String())
	}
	if s := strings.TrimSpace(stdout.String()); s != baai {
		t.Fatalf("stdout = %q want %q", stdout.String(), baai)
	}
	errOut := stderr.String()
	for _, needle := range []string{"404", "BAAI/bge-small-en-v1.5", "not a KnightsAnalytics export", "Not a leaderboard number"} {
		if !strings.Contains(errOut, needle) {
			t.Fatalf("stderr missing %q: %q", needle, errOut)
		}
	}
}

func TestRun_DownloadSuccess(t *testing.T) {
	var stdout, stderr bytes.Buffer
	got := Run(context.Background(), &stdout, &stderr, Config{
		DestDir: t.TempDir(),
		Download: func(context.Context, string, string) (string, error) {
			return "/models/KnightsAnalytics_bge-small-en-v1.5", nil
		},
	})
	if got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}
	if s := strings.TrimSpace(stdout.String()); s != "/models/KnightsAnalytics_bge-small-en-v1.5" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty on success", stderr.String())
	}
}

func TestRun_FailSoft(t *testing.T) {
	mini := t.TempDir()
	missing := filepath.Join(t.TempDir(), "absent-minilm")

	tests := []struct {
		name       string
		cfg        Config
		wantExit   int
		wantStdout string
		wantErr    []string
		notErr     []string
	}{
		{
			name: "401 MiniLM present",
			cfg: Config{
				MiniLMDir: mini,
				Download: func(context.Context, string, string) (string, error) {
					return "", fmt.Errorf("hugot: 401 Unauthorized")
				},
			},
			wantExit:   0,
			wantStdout: mini,
			wantErr: []string{
				"401",
				"Hugging Face 401/404 is expected",
				"KnightsAnalytics/bge-small-en-v1.5",
				mini,
				"not the official LongMemEval V1",
				"Hash-overlap unpublished",
				"Not a leaderboard number",
				"using MiniLM fallback",
			},
			notErr: []string{"TTFH cost-max is the hash embedder"},
		},
		{
			name: "404 MiniLM missing",
			cfg: Config{
				MiniLMDir: missing,
				Download: func(context.Context, string, string) (string, error) {
					return "", fmt.Errorf("hugot: 404 Not Found")
				},
			},
			wantExit:   0,
			wantStdout: "",
			wantErr: []string{
				"404",
				"Hugging Face 401/404 is expected",
				"not the official LongMemEval V1",
				"Hash-overlap unpublished",
				"Not a leaderboard number",
				"TTFH cost-max is the hash embedder (no ONNX required)",
			},
			notErr: []string{"using MiniLM fallback"},
		},
		{
			name: "mkdir failure stays exit 1",
			cfg: Config{
				DestDir: filepath.Join(t.TempDir(), "dest"),
				MkdirAll: func(string, os.FileMode) error {
					return fmt.Errorf("permission denied")
				},
				Download: func(context.Context, string, string) (string, error) {
					t.Fatal("download must not run after mkdir failure")
					return "", nil
				},
			},
			wantExit:   1,
			wantStdout: "",
			wantErr:    []string{"mkdir", "permission denied"},
			notErr:     []string{"Hugging Face 401/404", "MiniLM fallback"},
		},
		{
			name: "nil download is wiring, not fail-soft",
			cfg: Config{
				DestDir: t.TempDir(),
			},
			wantExit:   1,
			wantStdout: "",
			wantErr:    []string{"download func required"},
			notErr:     []string{"Hugging Face 401/404"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cfg.DestDir == "" {
				tc.cfg.DestDir = t.TempDir()
			}
			var stdout, stderr bytes.Buffer
			got := Run(context.Background(), &stdout, &stderr, tc.cfg)
			if got != tc.wantExit {
				t.Fatalf("exit = %d, want %d; stderr=%q", got, tc.wantExit, stderr.String())
			}
			if tc.wantStdout == "" {
				if stdout.Len() != 0 {
					t.Fatalf("stdout = %q, want empty", stdout.String())
				}
			} else if s := strings.TrimSpace(stdout.String()); s != tc.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tc.wantStdout)
			}
			errOut := stderr.String()
			for _, needle := range tc.wantErr {
				if !strings.Contains(errOut, needle) {
					t.Fatalf("stderr missing %q: %q", needle, errOut)
				}
			}
			for _, needle := range tc.notErr {
				if strings.Contains(errOut, needle) {
					t.Fatalf("stderr has unexpected %q: %q", needle, errOut)
				}
			}
		})
	}
}

func TestMiniLMExists_RequiresDirectory(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if miniLMExists(os.Stat, file) {
		t.Fatalf("file must not count as MiniLM dir")
	}
	if miniLMExists(os.Stat, filepath.Join(t.TempDir(), "missing")) {
		t.Fatal("missing path must not count as MiniLM dir")
	}
}

// statNotExist is a Stat seam that always misses (no network, no disk MiniLM).
func TestRun_StatSeamMissingMiniLM(t *testing.T) {
	var stdout, stderr bytes.Buffer
	got := Run(context.Background(), &stdout, &stderr, Config{
		DestDir:   t.TempDir(),
		MiniLMDir: "testdata/models/KnightsAnalytics_all-MiniLM-L6-v2",
		Stat: func(string) (os.FileInfo, error) {
			return nil, fs.ErrNotExist
		},
		Download: func(context.Context, string, string) (string, error) {
			return "", fmt.Errorf("401 Unauthorized")
		},
	})
	if got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty (hash default)", stdout.String())
	}
	if !strings.Contains(stderr.String(), "hash embedder") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

type fakeInfo struct {
	dir bool
}

func (f fakeInfo) Name() string       { return "KnightsAnalytics_all-MiniLM-L6-v2" }
func (f fakeInfo) Size() int64        { return 0 }
func (f fakeInfo) Mode() fs.FileMode  { return fs.ModeDir }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return f.dir }
func (f fakeInfo) Sys() any           { return nil }

func TestRun_StatSeamMiniLMPresent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	mini := DefaultMiniLMDir()
	got := Run(context.Background(), &stdout, &stderr, Config{
		DestDir:   t.TempDir(),
		MiniLMDir: mini,
		Stat: func(string) (os.FileInfo, error) {
			return fakeInfo{dir: true}, nil
		},
		Download: func(context.Context, string, string) (string, error) {
			return "", fmt.Errorf("401 Unauthorized")
		},
	})
	if got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}
	if s := strings.TrimSpace(stdout.String()); s != mini {
		t.Fatalf("stdout = %q, want %q", stdout.String(), mini)
	}
}
