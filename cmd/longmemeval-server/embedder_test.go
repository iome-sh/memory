package main

import (
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

func fakeBatchEmbed(texts []string, dim int) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, s := range texts {
		out[i] = memory.GenerateSimpleEmbedding(s, dim)
	}
	return out, nil
}

func TestPalaceConfigFromEmbed_AttachesBatchFunc(t *testing.T) {
	h := harnessEmbed{
		Func:    memory.GenerateSimpleEmbedding,
		Batch:   fakeBatchEmbed,
		Dim:     memory.MiniLMEmbeddingDim,
		ModelID: "test-onnx",
		ONNX:    true,
	}
	cfg := palaceConfigFromEmbed(t.TempDir(), h)
	if cfg.EmbeddingFunc == nil {
		t.Fatal("EmbeddingFunc must be set")
	}
	if cfg.BatchEmbeddingFunc == nil {
		t.Fatal("ONNX harness must set BatchEmbeddingFunc")
	}
	if cfg.PersistEmbeddings {
		t.Fatal("PersistEmbeddings default must be off")
	}
	if cfg.EmbeddingModel != "test-onnx" {
		t.Fatalf("EmbeddingModel = %q", cfg.EmbeddingModel)
	}
}

func TestPalaceConfigFromEmbed_PersistEmbeddingsOnlyWhenONNX(t *testing.T) {
	t.Setenv(envPersistEmbeddings, "1")

	hashCfg := palaceConfigFromEmbed(t.TempDir(), hashHarnessEmbed())
	if hashCfg.PersistEmbeddings {
		t.Fatal("hash embedder must not persist embeddings")
	}
	if hashCfg.BatchEmbeddingFunc != nil {
		t.Fatal("hash embedder must not set BatchEmbeddingFunc")
	}
	if hashCfg.EmbeddingModel != "" {
		t.Fatalf("hash EmbeddingModel = %q, want empty", hashCfg.EmbeddingModel)
	}

	onnx := harnessEmbed{
		Func:    memory.GenerateSimpleEmbedding,
		Batch:   fakeBatchEmbed,
		Dim:     memory.MiniLMEmbeddingDim,
		ModelID: "bge-small-en-v1.5",
		ONNX:    true,
	}
	onnxCfg := palaceConfigFromEmbed(t.TempDir(), onnx)
	if !onnxCfg.PersistEmbeddings {
		t.Fatal("ONNX + LONGMEMEVAL_PERSIST_EMBEDDINGS=1 should persist")
	}
	if onnxCfg.EmbeddingModel != "bge-small-en-v1.5" {
		t.Fatalf("EmbeddingModel = %q", onnxCfg.EmbeddingModel)
	}

	t.Setenv(envPersistEmbeddings, "")
	off := palaceConfigFromEmbed(t.TempDir(), onnx)
	if off.PersistEmbeddings {
		t.Fatal("unset LONGMEMEVAL_PERSIST_EMBEDDINGS must leave persist off")
	}
}

func TestEmbeddingModelID(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/models/BAAI_bge-small-en-v1.5", "bge-small-en-v1.5"},
		{"/models/KnightsAnalytics_all-MiniLM-L6-v2", "all-MiniLM-L6-v2"},
		{"/models/custom", "onnx"},
	}
	for _, tc := range cases {
		if got := embeddingModelID(tc.path); got != tc.want {
			t.Fatalf("embeddingModelID(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestLoadHarnessEmbedder_HashWhenModelMissing(t *testing.T) {
	t.Setenv(memory.EnvONNXModelPath, "")
	h, err := loadHarnessEmbedder()
	if err != nil {
		t.Fatal(err)
	}
	if h.ONNX {
		t.Fatal("empty model path must be hash")
	}
	if h.Batch != nil {
		t.Fatal("hash fallback must not set BatchEmbeddingFunc")
	}
	if h.Dim != memory.DefaultHashEmbeddingDim {
		t.Fatalf("dim = %d, want %d", h.Dim, memory.DefaultHashEmbeddingDim)
	}
}

func TestHandleRetrieve_UsesBatchEmbeddingFunc(t *testing.T) {
	var batchCalls int
	var batchTexts int
	batchFn := func(texts []string, dim int) ([][]float32, error) {
		batchCalls++
		batchTexts += len(texts)
		return fakeBatchEmbed(texts, dim)
	}
	embeddingDim = memory.DefaultHashEmbeddingDim
	globalStore = memory.NewPalaceStoreWithConfig(palaceConfigFromEmbed(t.TempDir(), harnessEmbed{
		Func:    memory.GenerateSimpleEmbedding,
		Batch:   batchFn,
		Dim:     embeddingDim,
		ModelID: "test-onnx",
		ONNX:    true,
	}))
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")
	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = false
	*flagFactAugLevel = 0
	*flagEnableChainOfNote = false

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	turns := []ingestTurn{
		{Role: "user", Content: "I adopted a golden retriever named Max in March 2024.", Timestamp: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC), Cycle: 1},
		{Role: "user", Content: "Max loves fetching tennis balls at the park every weekend.", Timestamp: time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC), Cycle: 2},
		{Role: "user", Content: "Quarterly revenue exceeded analyst expectations this spring.", Timestamp: time.Date(2024, 3, 17, 9, 0, 0, 0, time.UTC), Cycle: 3},
	}
	if err := postIngest(srv.URL, "batch-retrieve", turns); err != nil {
		t.Fatal(err)
	}

	out, err := postRetrieve(srv.URL, "What is the name of my golden retriever dog?", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Memories) == 0 {
		t.Fatal("expected retrieved memories")
	}
	if batchCalls < 1 {
		t.Fatal("retrieve should batch-score candidates via BatchEmbeddingFunc")
	}
	if batchTexts < 2 {
		t.Fatalf("batch scored %d texts, want at least 2 candidates", batchTexts)
	}
	if !retrieveHasNeedle(out, "max") && !retrieveHasNeedle(out, "retriever") {
		t.Fatalf("retrieve missed gold; got %#v", out.Memories)
	}
}

func TestHandleRetrieve_CountQueryKeepsKeywordFirstWithBatch(t *testing.T) {
	var batchCalls int
	batchFn := func(texts []string, dim int) ([][]float32, error) {
		batchCalls++
		return fakeBatchEmbed(texts, dim)
	}
	embeddingDim = memory.DefaultHashEmbeddingDim
	globalStore = memory.NewPalaceStoreWithConfig(palaceConfigFromEmbed(t.TempDir(), harnessEmbed{
		Func:    memory.GenerateSimpleEmbedding,
		Batch:   batchFn,
		Dim:     embeddingDim,
		ModelID: "test-onnx",
		ONNX:    true,
	}))
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")
	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = false
	*flagFactAugLevel = 0
	*flagEnableChainOfNote = false

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	conv := "count-batch-conv"
	if err := postIngest(srv.URL, conv, []ingestTurn{{
		Role:      "user",
		Content:   "I've had some experience from my Marketing Research class project, where I led the data analysis team and we did a comprehensive market analysis for a new product launch.",
		Timestamp: time.Date(2023, 5, 28, 17, 25, 0, 0, time.UTC),
		Cycle:     1,
		SessionID: "hay-led",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := postIngest(srv.URL, conv, []ingestTurn{{
		Role:      "user",
		Content:   "I've been working on a solo project for my Data Mining class, and I'm really interested in applying some of these techniques to my customer purchase data.",
		Timestamp: time.Date(2023, 5, 24, 9, 36, 0, 0, time.UTC),
		Cycle:     1,
		SessionID: "hay-solo",
	}}); err != nil {
		t.Fatal(err)
	}

	out, err := postRetrieveSession(srv.URL, "How many projects have I led", 8, conv)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Memories) == 0 {
		t.Fatal("expected retrieved memories")
	}
	if out.Memories[0].ID != "count-evidence" {
		t.Fatalf("count query should lead with synthetic assembly, got id=%q summary=%q", out.Memories[0].ID, out.Memories[0].Summary)
	}
	if batchCalls < 1 {
		t.Fatal("count retrieve should still pass QueryVec so batch scoring runs")
	}
	blob := strings.ToLower(out.Memories[0].Summary + " " + out.Memories[0].Full)
	if !strings.Contains(blob, "led the data analysis") {
		t.Fatalf("assembly missing led-team gold: %#v", out.Memories[0])
	}
	if !strings.Contains(blob, "solo project") {
		t.Fatalf("assembly missing solo-project gold: %#v", out.Memories[0])
	}
}

func TestHandleIngest_HashNeverPersistsEmbeddings(t *testing.T) {
	t.Setenv(envPersistEmbeddings, "1")
	t.Setenv(memory.EnvONNXModelPath, "")
	base := t.TempDir()
	h := hashHarnessEmbed()
	embeddingDim = h.Dim
	globalStore = memory.NewPalaceStoreWithConfig(palaceConfigFromEmbed(base, h))
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")
	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = false
	*flagFactAugLevel = 0
	*flagEnableChainOfNote = false

	resp := postIngestRaw(t, 1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if palaceJSONHasEmbedding(t, base) {
		t.Fatal("hash embeddings must not be persisted even when LONGMEMEVAL_PERSIST_EMBEDDINGS=1")
	}
}

func TestHandleIngest_ONNXPersistWhenEnvSet(t *testing.T) {
	t.Setenv(envPersistEmbeddings, "1")
	base := t.TempDir()
	dim := 4
	embeddingDim = dim
	globalStore = memory.NewPalaceStoreWithConfig(palaceConfigFromEmbed(base, harnessEmbed{
		Func: func(text string, d int) []float32 {
			v := make([]float32, d)
			if d > 0 {
				v[0] = 1
			}
			return v
		},
		Batch:   fakeBatchEmbed,
		Dim:     dim,
		ModelID: "test-onnx",
		ONNX:    true,
	}))
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")
	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = false
	*flagFactAugLevel = 0
	*flagEnableChainOfNote = false

	resp := postIngestRaw(t, 1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if !palaceJSONHasEmbedding(t, base) {
		t.Fatal("ONNX persist env should store embedding on palace JSON")
	}
}

func palaceJSONHasEmbedding(t *testing.T, base string) bool {
	t.Helper()
	found := false
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		var raw map[string]any
		if json.Unmarshal(data, &raw) != nil {
			return nil
		}
		content, _ := raw["content"].(map[string]any)
		if content == nil {
			return nil
		}
		if _, ok := content["embedding"]; ok {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}
