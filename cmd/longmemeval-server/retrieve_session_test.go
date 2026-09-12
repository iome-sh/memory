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

func setupHashHarness(t *testing.T) {
	t.Helper()
	t.Setenv(memory.EnvONNXModelPath, "")
	t.Setenv(envPersistEmbeddings, "")
	baseDir := filepath.Join(t.TempDir(), "lme_hash")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	h := hashHarnessEmbed()
	embeddingDim = h.Dim
	globalStore = memory.NewPalaceStoreWithConfig(palaceConfigFromEmbed(baseDir, h))
	globalVectorStore = memory.NewVectorStore("", "longmemeval_memory")
	*flagEnableTurnGranularity = true
	*flagEnableTimeAware = false
	*flagFactAugLevel = 0
	*flagEnableChainOfNote = false
}

func TestLongMemEval_RetrieveScopesSessionID(t *testing.T) {
	setupHashHarness(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	goldTurns := []ingestTurn{{
		Role:      "user",
		Content:   "I last cleaned my white Adidas sneakers on Sunday after the park.",
		Timestamp: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		Cycle:     1,
	}}
	if err := postIngest(srv.URL, "08f4fc43-gold", goldTurns); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		noise := []ingestTurn{{
			Role:      "user",
			Content:   "when did I last check email and update weekly notes " + string(rune('a'+i)),
			Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Cycle:     1,
		}}
		if err := postIngest(srv.URL, "noise-"+string(rune('a'+i)), noise); err != nil {
			t.Fatal(err)
		}
	}

	q := "when did I last clean my white Adidas sneakers"
	scoped, err := postRetrieveSession(srv.URL, q, 15, "08f4fc43-gold")
	if err != nil {
		t.Fatal(err)
	}
	if !retrieveHasNeedle(scoped, "adidas") {
		t.Fatalf("session-scoped retrieve missed gold; got %#v", scoped.Memories)
	}
	for _, m := range scoped.Memories {
		if m.SessionID != "" && m.SessionID != "08f4fc43-gold" {
			t.Fatalf("session leak %q in %#v", m.SessionID, m)
		}
	}
}

func TestLongMemEval_HealthReportsEmbedMode(t *testing.T) {
	setupHashHarness(t)
	if got := embedMode(); got != "hash" {
		t.Fatalf("embedMode = %q want hash", got)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if got, _ := body["embed_mode"].(string); got != "hash" {
		t.Fatalf("health embed_mode = %q want hash", got)
	}
}

func TestLongMemEval_HealthEmbedModeDistinguishesONNX(t *testing.T) {
	t.Setenv(memory.EnvONNXModelPath, "")
	embeddingDim = memory.DefaultHashEmbeddingDim
	if got := embedMode(); got != "hash" {
		t.Fatalf("empty path embedMode = %q want hash", got)
	}

	embeddingDim = memory.MiniLMEmbeddingDim
	t.Setenv(memory.EnvONNXModelPath, "/tmp/KnightsAnalytics_all-MiniLM-L6-v2")
	if got := embedMode(); got != "onnx-minilm-l6-v2" {
		t.Fatalf("minilm embedMode = %q", got)
	}

	t.Setenv(memory.EnvONNXModelPath, "/tmp/BAAI_bge-small-en-v1.5")
	if got := embedMode(); got != "onnx-bge-small-en-v1.5" {
		t.Fatalf("bge embedMode = %q", got)
	}

	embeddingDim = memory.DefaultHashEmbeddingDim
	t.Setenv(memory.EnvONNXModelPath, "/tmp/BAAI_bge-small-en-v1.5")
	if got := embedMode(); got != "hash" {
		t.Fatalf("onnx path with hash dim embedMode = %q want hash (fallback)", got)
	}
}

func postRetrieveSession(baseURL, query string, limit int, sessionID string) (RetrieveResponse, error) {
	body, err := json.Marshal(RetrieveRequest{Query: query, Limit: limit, SessionID: sessionID})
	if err != nil {
		return RetrieveResponse{}, err
	}
	resp, err := http.Post(baseURL+"/retrieve", "application/json", bytes.NewReader(body))
	if err != nil {
		return RetrieveResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return RetrieveResponse{}, errRetrieveStatus(resp.StatusCode)
	}
	var out RetrieveResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return RetrieveResponse{}, err
	}
	return out, nil
}

func retrieveHasNeedle(out RetrieveResponse, needle string) bool {
	n := strings.ToLower(needle)
	for _, m := range out.Memories {
		if strings.Contains(strings.ToLower(m.Full+" "+m.Summary), n) {
			return true
		}
	}
	return false
}

func TestLongMemEval_RetrieveCountQueryAssemblesCrossSessionFacts(t *testing.T) {
	setupHashHarness(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	conv := "count-assembly-conv"
	led := []ingestTurn{{
		Role:      "user",
		Content:   "I've had some experience from my Marketing Research class project, where I led the data analysis team and we did a comprehensive market analysis for a new product launch.",
		Timestamp: time.Date(2023, 5, 28, 17, 25, 0, 0, time.UTC),
		Cycle:     1,
		SessionID: "hay-led",
	}}
	solo := []ingestTurn{{
		Role:      "user",
		Content:   "I've been working on a solo project for my Data Mining class, and I'm really interested in applying some of these techniques to my customer purchase data.",
		Timestamp: time.Date(2023, 5, 24, 9, 36, 0, 0, time.UTC),
		Cycle:     1,
		SessionID: "hay-solo",
	}}
	if err := postIngest(srv.URL, conv, led); err != nil {
		t.Fatal(err)
	}
	if err := postIngest(srv.URL, conv, solo); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		noise := []ingestTurn{{
			Role:      "user",
			Content:   "The elbow method is an excellent choice for clustering how many analysis techniques appear in project dashboards and number of projects metrics " + string(rune('a'+i)) + ".",
			Timestamp: time.Date(2023, 5, 20, 6, 16, 0, 0, time.UTC),
			Cycle:     1,
			SessionID: "hay-noise",
		}}
		if err := postIngest(srv.URL, conv, noise); err != nil {
			t.Fatal(err)
		}
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
	blob := strings.ToLower(out.Memories[0].Summary + " " + out.Memories[0].Full)
	if !strings.Contains(blob, "led the data analysis") {
		t.Fatalf("assembly missing led-team gold: %#v", out.Memories[0])
	}
	if !strings.Contains(blob, "solo project") {
		t.Fatalf("assembly missing solo-project gold: %#v", out.Memories[0])
	}
	if !retrieveHasNeedle(out, "led the data analysis") {
		t.Fatalf("retrieve missed led-team gold; got %#v", out.Memories)
	}
	if !retrieveHasNeedle(out, "solo project") {
		t.Fatalf("retrieve missed solo-project gold; got %#v", out.Memories)
	}
}
