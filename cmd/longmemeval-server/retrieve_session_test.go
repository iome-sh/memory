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
	baseDir := filepath.Join(t.TempDir(), "lme_hash")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
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
