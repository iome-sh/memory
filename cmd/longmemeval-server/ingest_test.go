package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iome-sh/memory"
)

func setupIngestHarness(t *testing.T, baseDir string) {
	t.Helper()
	t.Setenv(memory.EnvONNXModelPath, "")
	embeddingDim = memory.DefaultHashEmbeddingDim
	globalStore = memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
		BaseDir:       baseDir,
		EmbeddingFunc: memory.GenerateSimpleEmbedding,
	})
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")
	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = false
	*flagFactAugLevel = 0
	*flagEnableChainOfNote = false
}

func postIngestRaw(t *testing.T, turns int) *http.Response {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	payloadTurns := make([]struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Timestamp string `json:"timestamp"`
		Cycle     int    `json:"cycle"`
	}, turns)
	for i := 0; i < turns; i++ {
		payloadTurns[i] = struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			Timestamp string `json:"timestamp"`
			Cycle     int    `json:"cycle"`
		}{
			Role:      "user",
			Content:   "I adopted a golden retriever named Max in March 2024.",
			Timestamp: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Cycle:     1,
		}
	}
	body, err := json.Marshal(IngestRequest{ConvID: "ingest-honesty", Turns: payloadTurns})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(srv.URL+"/ingest", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func decodeIngestResponse(t *testing.T, resp *http.Response) IngestResponse {
	t.Helper()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var out IngestResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode ingest body %q: %v", raw, err)
	}
	return out
}

func TestHandleIngest_PersistOKIncrements(t *testing.T) {
	setupIngestHarness(t, t.TempDir())

	resp := postIngestRaw(t, 2)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	got := decodeIngestResponse(t, resp)
	if got.Status != "ok" {
		t.Fatalf("status = %q, want ok (error=%q)", got.Status, got.Error)
	}
	if got.Ingested != 2 {
		t.Fatalf("ingested = %d, want 2", got.Ingested)
	}
}

func TestHandleIngest_IngestTurnErrorNotOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	setupIngestHarness(t, path)
	*flagEnableTurnGranularity = true

	resp := postIngestRaw(t, 2)
	if resp.StatusCode == http.StatusOK {
		t.Fatal("failed IngestTurn must not return HTTP 200")
	}
	got := decodeIngestResponse(t, resp)
	if got.Status == "ok" {
		t.Fatal("failed persist must not return blanket status ok")
	}
	if got.Ingested != 0 {
		t.Fatalf("ingested = %d, want 0 on persist failure", got.Ingested)
	}
	if strings.TrimSpace(got.Error) == "" {
		t.Fatal("expected persist error in body")
	}
}

func TestHandleIngest_WriteErrorNotOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	setupIngestHarness(t, path)
	*flagEnableTurnGranularity = false

	resp := postIngestRaw(t, 1)
	if resp.StatusCode == http.StatusOK {
		t.Fatal("failed Write must not return HTTP 200")
	}
	got := decodeIngestResponse(t, resp)
	if got.Status == "ok" {
		t.Fatal("failed persist must not return blanket status ok")
	}
	if got.Ingested != 0 {
		t.Fatalf("ingested = %d, want 0 on persist failure", got.Ingested)
	}
}

func TestHandleIngest_FailedTurnNotCounted(t *testing.T) {
	setupIngestHarness(t, t.TempDir())
	orig := persistIngestTurn
	t.Cleanup(func() { persistIngestTurn = orig })
	n := 0
	persistIngestTurn = func(entry memory.MemoryEntry) error {
		n++
		if n == 2 {
			return errors.New("disk full")
		}
		return orig(entry)
	}

	resp := postIngestRaw(t, 3)
	if resp.StatusCode == http.StatusOK {
		t.Fatal("partial persist must not return HTTP 200")
	}
	got := decodeIngestResponse(t, resp)
	if got.Status == "ok" {
		t.Fatal("partial persist must not return blanket status ok")
	}
	if got.Ingested != 2 {
		t.Fatalf("ingested = %d, want 2 (failed turn not counted)", got.Ingested)
	}
	if !strings.Contains(got.Error, "disk full") {
		t.Fatalf("error = %q, want disk full", got.Error)
	}
}
