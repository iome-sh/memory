package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sudo-jin/memory"
)

// LongMemEvalServer wraps PalaceStore for the LongMemEval benchmark harness.
// Run with: go run cmd/longmemeval-server/main.go
// Then use the Python orchestrator to drive ingestion + retrieval.

type IngestRequest struct {
	ConvID string `json:"conv_id"`
	Turns  []struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Timestamp string `json:"timestamp"`
		Cycle     int    `json:"cycle"`
	} `json:"turns"`
}

type RetrieveRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type RetrieveResponse struct {
	Memories []memory.MemoryEntry `json:"memories"`
	Scores   []float64            `json:"scores,omitempty"`
}


var globalStore *memory.PalaceStore

func main() {
	baseDir := filepath.Join(os.TempDir(), "longmemeval_palace")
	_ = os.MkdirAll(baseDir, 0755)
	globalStore = memory.NewPalaceStore(baseDir)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/ingest", handleIngest)
	http.HandleFunc("/retrieve", handleRetrieve)
	http.HandleFunc("/compact", handleCompact)

	log.Println("LongMemEval benchmark server listening on :8765")
	log.Fatal(http.ListenAndServe(":8765", nil))
}

func handleIngest(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, t := range req.Turns {
		entry := memory.MemoryEntry{
			ID:      memory.GenerateMemoryID(),
			Type:    "conversation_turn",
			Tier:    memory.TierWorking,
			Content: memory.MemoryContent{
				Full:    t.Content,
				Summary: truncate(t.Content, 280),
			},
			Cycle: t.Cycle,
			// TemporalTags can be populated here if timestamps are parsed
		}
		if err := globalStore.Write(entry); err != nil {
			log.Printf("write error: %v", err)
		}
	}

	// Optional: trigger compaction every N turns for RecMem testing
	// if len(req.Turns) > 0 { _ = globalStore.AutoRecMemCompaction(...) }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "ingested": fmt.Sprintf("%d turns", len(req.Turns))})
}

func handleRetrieve(w http.ResponseWriter, r *http.Request) {
	var req RetrieveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 8
	}

	// Use the package's hybrid SearchMemory (keyword + vector re-rank)
	results := globalStore.SearchMemory(req.Query, nil, req.Limit, nil)

	resp := RetrieveResponse{Memories: results}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCompact(w http.ResponseWriter, r *http.Request) {
	// In a full run you would pass a real generateFn (LLM call) here.
	// For benchmark we expose the hook so the orchestrator can decide when to compact.
	// Example: globalStore.AutoRecMemCompaction(yourGenerateFn, nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "compaction triggered (stub)"})
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
