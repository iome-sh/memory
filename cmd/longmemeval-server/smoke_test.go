package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iome-sh/memory"
)

// setupTestHarness mirrors main() initialization: ONNX from env or hugot download,
// PalaceStore on a temp base dir, and Qdrant disabled so retrieval uses file-based SearchMemory.
func setupTestHarness(t *testing.T) {
	t.Helper()

	modelDir := onnxModelDirForLongMemEvalTest(t)
	t.Setenv(memory.EnvONNXModelPath, modelDir)

	baseDir := filepath.Join(t.TempDir(), "longmemeval_palace_v2")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}

	modelPath := strings.TrimSpace(os.Getenv(memory.EnvONNXModelPath))
	embeddingDim = memory.ResolveEmbeddingDim(modelPath)
	embedFn, err := memory.NewGONNXEmbeddingFuncFromEnv()
	if err != nil {
		t.Fatalf("onnx embedding init failed: %v", err)
	}
	if embeddingDim != memory.MiniLMEmbeddingDim {
		t.Fatalf("embedding dim = %d, want %d", embeddingDim, memory.MiniLMEmbeddingDim)
	}

	cfg := memory.PalaceConfig{
		BaseDir:       baseDir,
		EmbeddingFunc: embedFn,
	}
	globalStore = memory.NewPalaceStoreWithConfig(cfg)

	// Smoke gate must not depend on Qdrant; file-based hybrid SearchMemory is enough.
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")

	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = true
	*flagFactAugLevel = 2
	*flagEnableChainOfNote = true
}

func TestLongMemEval_IngestRetrieveRecall(t *testing.T) {
	setupTestHarness(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ingestBody, err := json.Marshal(IngestRequest{
		ConvID: "smoke-golden-retriever",
		Turns: []struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			Timestamp string `json:"timestamp"`
			Cycle     int    `json:"cycle"`
		}{
			{
				Role:      "user",
				Content:   "I adopted a golden retriever named Max in March 2024.",
				Timestamp: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
				Cycle:     1,
			},
			{
				Role:      "assistant",
				Content:   "Max sounds like a wonderful companion!",
				Timestamp: time.Date(2024, 3, 15, 10, 0, 30, 0, time.UTC).Format(time.RFC3339),
				Cycle:     1,
			},
			{
				Role:      "user",
				Content:   "Max loves fetching tennis balls at the park every weekend.",
				Timestamp: time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
				Cycle:     2,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ingestResp, err := http.Post(srv.URL+"/ingest", "application/json", bytes.NewReader(ingestBody))
	if err != nil {
		t.Fatal(err)
	}
	defer ingestResp.Body.Close()
	if ingestResp.StatusCode != http.StatusOK {
		t.Fatalf("ingest status = %d, want %d", ingestResp.StatusCode, http.StatusOK)
	}

	retrieveBody, err := json.Marshal(RetrieveRequest{
		Query: "What is the name of my golden retriever dog?",
		Limit: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	retrieveResp, err := http.Post(srv.URL+"/retrieve", "application/json", bytes.NewReader(retrieveBody))
	if err != nil {
		t.Fatal(err)
	}
	defer retrieveResp.Body.Close()
	if retrieveResp.StatusCode != http.StatusOK {
		t.Fatalf("retrieve status = %d, want %d", retrieveResp.StatusCode, http.StatusOK)
	}

	var out RetrieveResponse
	if err := json.NewDecoder(retrieveResp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Memories) == 0 {
		t.Fatal("expected at least one retrieved memory")
	}

	top := strings.ToLower(out.Memories[0].Full + " " + out.Memories[0].Summary)
	if !strings.Contains(top, "max") && !strings.Contains(top, "retriever") {
		t.Fatalf("top memory should mention Max or retriever; got summary=%q full=%q",
			out.Memories[0].Summary, out.Memories[0].Full)
	}
}
