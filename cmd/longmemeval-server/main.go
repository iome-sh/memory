package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sudo-jin/memory"
)

// LongMemEvalServer wraps PalaceStore for the LongMemEval benchmark harness.
// Run with: go run cmd/longmemeval-server/main.go

// MemoryHit is a lightweight DTO for the benchmark.
type MemoryHit struct {
	ID      string  `json:"id"`
	Summary string  `json:"summary"`
	Full    string  `json:"full"`
	Score   float64 `json:"score,omitempty"`
}

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
	Memories []MemoryHit `json:"memories"`
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
			ID:   memory.GenerateMemoryID(),
			Type: "conversation_turn",
			Tier: memory.TierWorking,
			Content: memory.MemoryContent{
				Full:    t.Content,
				Summary: truncate(t.Content, 280),
			},
			Cycle: t.Cycle,
		}
		if err := globalStore.Write(entry); err != nil {
			log.Printf("write error for conv %s: %v", req.ConvID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "ok",
		"ingested": fmt.Sprintf("%d turns", len(req.Turns)),
	})
}

func handleRetrieve(w http.ResponseWriter, r *http.Request) {
	var req RetrieveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 16
	}

	// Generate query embedding (used for re-ranking)
	queryVec := memory.GenerateSimpleEmbedding(req.Query, 768)

	// 1. Keyword + vector search (higher internal limit for recall)
	keywordResults := globalStore.SearchMemory(req.Query, nil, req.Limit*3, queryVec)

	// 2. Also include recent Working tier entries (recency bias helps many personal questions)
	recent := globalStore.listEntriesInTier(memory.TierWorking)
	if len(recent) > 40 {
		recent = recent[:40] // take the top 40 by current relevance score (includes recency)
	}

	// Merge + dedup by ID
	seen := make(map[string]bool)
	combined := make([]memory.MemoryEntry, 0, len(keywordResults)+len(recent))
	for _, e := range keywordResults {
		if !seen[e.ID] {
			seen[e.ID] = true
			combined = append(combined, e)
		}
	}
	for _, e := range recent {
		if !seen[e.ID] {
			seen[e.ID] = true
			combined = append(combined, e)
		}
	}

	// Re-rank the combined set with vector similarity if we have a query vector
	if len(queryVec) > 0 && len(combined) > 0 {
		sort.Slice(combined, func(i, j int) bool {
			iVec := memory.GenerateSimpleEmbedding(combined[i].Content.Summary+" "+combined[i].Content.Full, len(queryVec))
			jVec := memory.GenerateSimpleEmbedding(combined[j].Content.Summary+" "+combined[j].Content.Full, len(queryVec))
			return memory.CosineSimilarity(iVec, queryVec) > memory.CosineSimilarity(jVec, queryVec)
		})
	}

	// Take top Limit
	if len(combined) > req.Limit {
		combined = combined[:req.Limit]
	}

	hits := make([]MemoryHit, 0, len(combined))
	for _, e := range combined {
		hits = append(hits, MemoryHit{
			ID:      e.ID,
			Summary: e.Content.Summary,
			Full:    e.Content.Full,
		})
	}

	resp := RetrieveResponse{Memories: hits}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCompact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "compaction triggered (stub)"})
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
