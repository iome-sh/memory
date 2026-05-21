package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
		now := time.Now()

		// 1. Store raw conversation turn (Working tier for recency)
		rawEntry := memory.MemoryEntry{
			ID:   memory.GenerateMemoryID(),
			Type: "conversation_turn",
			Tier: memory.TierWorking,
			Content: memory.MemoryContent{
				Full:    t.Content,
				Summary: truncate(t.Content, 280),
			},
			Cycle:     t.Cycle,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := globalStore.Write(rawEntry); err != nil {
			log.Printf("write error for conv %s: %v", req.ConvID, err)
		}

		// 2. Extract atomic facts -> TierSemantic (high-signal)
		facts := memory.ExtractAtomicFacts(rawEntry)
		for _, factText := range facts {
			factID := memory.GenerateMemoryID()
			atomicFact := memory.MemoryEntry{
				ID:        factID,
				Type:      "atomic_fact",
				Tier:      memory.TierSemantic,
				Version:   1,
				CreatedAt: now,
				UpdatedAt: now,
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

		// 3. ALWAYS add at least one high-ScoreImpact semantic entry per turn
		// This guarantees every conversation contributes to TierSemantic for recall
		turnSemantic := memory.MemoryEntry{
			ID:        memory.GenerateMemoryID(),
			Type:      "turn_semantic",
			Tier:      memory.TierSemantic,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
			Content: memory.MemoryContent{
				Summary: truncate(t.Content, 220),
				Full:    t.Content,
				Tags:    []string{"raw_turn", "guaranteed_semantic"},
			},
			Provenance: memory.MemoryProvenance{
				SourceStep: "longmemeval_ingest_always",
				ParentIDs:  []string{rawEntry.ID},
			},
			Metrics: memory.MemoryMetrics{
				ScoreImpact: 0.88, // High but slightly below pure atomic facts
				UsageCount:  1,
			},
		}
		_ = globalStore.Write(turnSemantic)
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
		req.Limit = 40 // Raised for better recall on LongMemEval
	}

	queryVec := memory.GenerateSimpleEmbedding(req.Query, 768)

	// Deep refactor: aggressive high-recall for LongMemEval
	// 1. All TierSemantic facts (highest priority for atomic + guaranteed entries)
	semanticFacts := globalStore.ListEntriesInTier(memory.TierSemantic)

	// 2. Broad keyword search (larger pool)
	keywordResults := globalStore.SearchMemory(req.Query, nil, req.Limit*8, queryVec)

	// 3. Recent Working tier (recency bias)
	recent := globalStore.ListEntriesInTier(memory.TierWorking)
	if len(recent) > 100 {
		recent = recent[:100]
	}

	// Merge with semantic first, then others (dedup)
	seen := make(map[string]bool)
	combined := make([]memory.MemoryEntry, 0, len(semanticFacts)+len(keywordResults)+len(recent))

	// Prepend all semantic (core of high-recall strategy)
	for _, e := range semanticFacts {
		if !seen[e.ID] {
			seen[e.ID] = true
			combined = append(combined, e)
		}
	}

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

	// Re-rank with composite score: heavy boost for TierSemantic + relevance + overlap
	if len(queryVec) > 0 && len(combined) > 1 {
		sort.SliceStable(combined, func(i, j int) bool {
			iEntry := combined[i]
			jEntry := combined[j]

			// TierSemantic gets strong priority
			iBoost := 0.0
			if iEntry.Tier == memory.TierSemantic {
				iBoost = 0.45
			}
			jBoost := 0.0
			if jEntry.Tier == memory.TierSemantic {
				jBoost = 0.45
			}

			iText := iEntry.Content.Summary + " " + iEntry.Content.Full
			jText := jEntry.Content.Summary + " " + jEntry.Content.Full

			iVec := memory.GenerateSimpleEmbedding(iText, len(queryVec))
			jVec := memory.GenerateSimpleEmbedding(jText, len(queryVec))

			iSim := memory.CosineSimilarity(iVec, queryVec)
			jSim := memory.CosineSimilarity(jVec, queryVec)

			iRel := memory.CalculateRelevanceScore(iEntry)
			jRel := memory.CalculateRelevanceScore(jEntry)

			// Simple token overlap bonus
			iOverlap := tokenOverlapScore(req.Query, iText)
			jOverlap := tokenOverlapScore(req.Query, jText)

			iScore := iBoost + iSim*0.25 + iRel*0.2 + iOverlap*0.1
			jScore := jBoost + jSim*0.25 + jRel*0.2 + jOverlap*0.1

			return iScore > jScore
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

// tokenOverlapScore gives bonus for query words appearing in content (loose match)
func tokenOverlapScore(query, content string) float64 {
	qLower := strings.ToLower(query)
	cLower := strings.ToLower(content)
	qWords := strings.FieldsFunc(qLower, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}
	if len(qWords) == 0 {
		return 0
	}
	matches := 0
	for _, w := range qWords {
		if len(w) >= 2 && strings.Contains(cLower, w) {
			matches++
		}
	}
	return float64(matches) / float64(len(qWords))
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
