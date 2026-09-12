package onnxdownload

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iome-sh/memory"
)

func TestDefaultBAAIDir(t *testing.T) {
	want := filepath.Join("testdata", "models", "BAAI_bge-small-en-v1.5")
	if got := DefaultBAAIDir(); got != want {
		t.Fatalf("DefaultBAAIDir = %q, want %q", got, want)
	}
	if DefaultBAAIDir() == DefaultMiniLMDir() {
		t.Fatal("BAAI dir must not be MiniLM")
	}
	if memory.HugotCacheDirName(memory.BAAIONNXModelHF) != "BAAI_bge-small-en-v1.5" {
		t.Fatalf("cache name = %q", memory.HugotCacheDirName(memory.BAAIONNXModelHF))
	}
}

func TestBAAILayoutComplete(t *testing.T) {
	dir := t.TempDir()
	if BAAILayoutComplete(dir) {
		t.Fatal("empty dir must not be complete")
	}
	write := func(name string, n int) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, n), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("model.onnx", 500)
	write("config.json", 8)
	write("tokenizer.json", 8)
	write("vocab.txt", 8)
	if BAAILayoutComplete(dir) {
		t.Fatal("tiny model.onnx must not count as complete")
	}
	write("model.onnx", 1_000_001)
	if !BAAILayoutComplete(dir) {
		t.Fatal("want complete")
	}
}

func TestFetchBAAI_SkipWhenComplete(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "BAAI_bge-small-en-v1.5")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"config.json", "tokenizer.json", "vocab.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "model.onnx"), make([]byte, 1_000_001), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := fetchBAAI(context.Background(), parent, func(context.Context, string, string) error {
		t.Fatal("network must not run when layout is complete")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("got %q want %q", got, dir)
	}
}

func TestFetchBAAI_ReshapeOnnxSubdir(t *testing.T) {
	parent := t.TempDir()
	var seen []string
	got, err := fetchBAAI(context.Background(), parent, func(_ context.Context, url, dest string) error {
		seen = append(seen, url+" -> "+filepath.Base(dest))
		n := 64
		if strings.HasSuffix(dest, "model.onnx") {
			n = 1_000_001
			if !strings.Contains(url, "/onnx/model.onnx") {
				t.Fatalf("model url = %q, want .../onnx/model.onnx", url)
			}
			if filepath.Base(dest) != "model.onnx" {
				t.Fatalf("dest base = %q", dest)
			}
			if strings.Contains(dest, string(filepath.Separator)+"onnx"+string(filepath.Separator)) {
				t.Fatalf("hugot dest must be root model.onnx, got %q", dest)
			}
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, make([]byte, n), 0o600)
	})
	if err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Join(parent, "BAAI_bge-small-en-v1.5")
	if got != wantDir {
		t.Fatalf("dir = %q want %q", got, wantDir)
	}
	if !BAAILayoutComplete(got) {
		t.Fatal("layout incomplete")
	}
	if len(seen) != len(baaIHugotFiles) {
		t.Fatalf("fetched %d files, want %d: %v", len(seen), len(baaIHugotFiles), seen)
	}
}

func TestHTTPGetToFile_RetryUnauthOn401(t *testing.T) {
	var sawAuth, sawUnauth int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			sawAuth++
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "Invalid username or password.")
			return
		}
		sawUnauth++
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "public-ok")
	}))
	t.Cleanup(srv.Close)

	prev := httpDo
	httpDo = srv.Client().Do
	t.Cleanup(func() { httpDo = prev })

	t.Setenv("HF_TOKEN", "hf_bad_token")
	t.Setenv("HUGGING_FACE_HUB_TOKEN", "")
	dest := filepath.Join(t.TempDir(), "config.json")
	if err := HTTPGetToFile(context.Background(), srv.URL+"/config.json", dest); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if sawAuth != 1 || sawUnauth != 1 {
		t.Fatalf("sawAuth=%d sawUnauth=%d want 1,1", sawAuth, sawUnauth)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "public-ok" {
		t.Fatalf("body = %q", b)
	}
}

func TestHTTPGetToFile_404NotRetried(t *testing.T) {
	var n int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "Repository not found")
	}))
	t.Cleanup(srv.Close)
	prev := httpDo
	httpDo = srv.Client().Do
	t.Cleanup(func() { httpDo = prev })
	t.Setenv("HF_TOKEN", "hf_any")
	err := HTTPGetToFile(context.Background(), srv.URL+"/missing", filepath.Join(t.TempDir(), "x"))
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("err = %v", err)
	}
	if n != 1 {
		t.Fatalf("requests = %d want 1 (404 is missing-repo, not auth miss)", n)
	}
}

func TestHTTPGetToFile_401WithoutTokenNoRetry(t *testing.T) {
	var n int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	prev := httpDo
	httpDo = srv.Client().Do
	t.Cleanup(func() { httpDo = prev })
	t.Setenv("HF_TOKEN", "")
	t.Setenv("HUGGING_FACE_HUB_TOKEN", "")
	err := HTTPGetToFile(context.Background(), srv.URL+"/x", filepath.Join(t.TempDir(), "x"))
	if err == nil || !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("err = %v", err)
	}
	if n != 1 {
		t.Fatalf("requests = %d want 1", n)
	}
}
