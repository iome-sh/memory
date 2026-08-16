package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iome-sh/memory"
)

func TestHandleIngest_OfficialHaystackDate(t *testing.T) {
	globalStore = memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
		BaseDir:       t.TempDir(),
		EmbeddingFunc: memory.GenerateSimpleEmbedding,
	})
	*flagEnableTurnGranularity = true

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	body, err := json.Marshal(IngestRequest{
		ConvID: "official-date",
		Turns: []struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			Timestamp string `json:"timestamp"`
			Cycle     int    `json:"cycle"`
		}{
			{
				Role:      "user",
				Content:   "I adopted a golden retriever named Max.",
				Timestamp: "2023/04/10 (Mon) 17:50",
				Cycle:     1,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(srv.URL+"/ingest", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	hits := globalStore.SearchMemory("Max golden retriever", nil, 5, nil)
	if len(hits) == 0 {
		t.Fatal("expected ingested turn")
	}
	want := time.Date(2023, 4, 10, 17, 50, 0, 0, time.UTC)
	if !hits[0].Timestamp.Equal(want) {
		t.Fatalf("timestamp = %s, want official %s (RFC3339-only parse would have used now)", hits[0].Timestamp, want)
	}
	if !strings.Contains(strings.ToLower(hits[0].Content.Full), "max") {
		t.Fatalf("content = %q", hits[0].Content.Full)
	}
}
