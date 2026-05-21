package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

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
		// 1. Store raw conversation turn
		rawEntry := memory.MemoryEntry{
			ID:   memory.GenerateMemoryID(),
			Type: "conversation_turn",
			Tier: memory.TierWorking,
			Content: memory.MemoryContent{
				Full:    t.Content,
				Summary: truncate(t.Content, 280),
			},
			Cycle: t.Cycle,
		}
		if err := globalStore.Write(rawEntry); err != nil {
			log.Printf("write error for conv %s: %v", req.ConvID, err)
		}

		// 2. Extract atomic facts and store them in TierSemantic (high-signal index)
		factEntry := memory.MemoryEntry{
			ID:   memory.GenerateMemoryID(),
			Type: "atomic_fact",
			Tier: memory.TierSemantic,
			Content: memory.MemoryContent{
				Full:    t.Content,
				Summary: truncate(t.Content, 200),
			},
			Cycle: t.Cycle,
			Metrics: memory.MemoryMetrics{
				ScoreImpact: 0.92, // High importance so facts rank well
				UsageCount:  1,
			},
		}

		// Only write if we actually extracted something useful
		facts := memory.ExtractAtomicFacts(factEntry) // Note: using exported version if available, or internal
		if len(facts) > 0 {
			// Create one semantic entry per extracted fact for better granularity
			for _, factText := range facts {
				factID := memory.GenerateMemoryID()
				atomicFact := memory.MemoryEntry{
					ID:        factID,
					Type:      "atomic_fact",
					Tier:      memory.TierSemantic,
					Version:   1,
					CreatedAt: rawEntry.CreatedAt,
					UpdatedAt: rawEntry.CreatedAt,
					Content: memory.MemoryContent{
						Summary: truncate(factText, 180),
						Full:    factText,
						Tags:    []string{"extracted", "personal_fact"},
					},
					Provenance: memory.MemoryProvenance{
						SourceStep: "longmemeval_ingest",
						ParentIDs:  []string{rawEntry.ID},
					},
					Metrics: memory.MemoryMetrics{
						ScoreImpact: 0.95,
						UsageCount:  1,
					},
				}
				_ = globalStore.Write(atomicFact)
			}
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
		req.Limit = 20
	}

	// Generate query embedding for re-ranking
	queryVec := memory.GenerateSimpleEmbedding(req.Query, 768)

	// Keyword + recent hybrid retrieval (high recall for benchmark)
	keywordResults := globalStore.SearchMemory(req.Query, nil, req.Limit*4, queryVec)

	recent := globalStore.ListEntriesInTier(memory.TierWorking)
	if len(recent) > 50 {
		recent = recent[:50]
	}

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

	if len(queryVec) > 0 && len(combined) > 1 {
		sort.SliceStable(combined, func(i, j int) bool {
			iText := combined[i].Content.Summary + " " + combined[i].Content.Full
			jText := combined[j].Content.Summary + " " + combined[j].Content.Full
			iVec := memory.GenerateSimpleEmbedding(iText, len(queryVec))
			jVec := memory.GenerateSimpleEmbedding(jText, len(queryVec))
			return memory.CosineSimilarity(iVec, queryVec) > memory.CosineSimilarity(jVec, queryVec)
		})
	}

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
