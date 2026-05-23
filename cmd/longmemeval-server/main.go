package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sudo-jin/memory"
)

// LongMemEvalServer - Production harness for LongMemEval-S / LongMemEval-M
// All four LongMemEval optimizations are toggleable via command-line flags.
//
// Flags:
//   -enable-turn-granularity     Use IngestTurn + turn-level metadata (default true)
//   -enable-time-aware-expansion Time-aware query expansion + timestamp filtering
//   -fact-augmentation-level     0=off, 1=facts, 2=facts+keyphrases (default 2)
//   -enable-chain-of-note        Use ReadWithChainOfNote for answer synthesis
//
// Endpoints:
//   POST /ingest     - Ingest conversation turns (uses IngestTurn when enabled)
//   POST /retrieve   - Hybrid retrieval with all active optimizations
//   POST /synthesize - Chain-of-Note reading stage (when enabled)
//   GET  /health
//
// Output format is compatible with LongMemEval official JSONL submission.

// MemoryHit is the benchmark DTO.
type MemoryHit struct {
	ID        string  `json:"id"`
	Summary   string  `json:"summary"`
	Full      string  `json:"full"`
	Score     float64 `json:"score,omitempty"`
	Timestamp string  `json:"timestamp,omitempty"`
	TurnID    string  `json:"turn_id,omitempty"`
}

// LongMemEval official answer format (JSONL line)
type LongMemEvalAnswer struct {
	QuestionID string  `json:"question_id"`
	Answer     string  `json:"answer"`
	Confidence float64 `json:"confidence"`
}

type IngestRequest struct {
	ConvID string `json:"conv_id"`
	Turns  []struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Timestamp string `json:"timestamp"`
		Cycle     int    `json:"cycle"`
	} `json:"turns"`
	History []struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Timestamp string `json:"timestamp"`
		Cycle     int    `json:"cycle"`
	} `json:"history"`
}

type RetrieveRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type RetrieveResponse struct {
	Memories []MemoryHit `json:"memories"`
}

type SynthesizeRequest struct {
	Query     string      `json:"query"`
	Retrieved []MemoryHit `json:"retrieved"`
}

type SynthesizeResponse struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer,omitempty"`
	JSON   string `json:"json,omitempty"`
}

var (
	globalStore       *memory.PalaceStore
	globalVectorStore *memory.VectorStore

	// Toggleable LongMemEval features
	flagEnableTurnGranularity = flag.Bool("enable-turn-granularity", true, "Use IngestTurn with turn-level metadata")
	flagEnableTimeAware       = flag.Bool("enable-time-aware-expansion", true, "Enable time-aware query expansion")
	flagFactAugLevel          = flag.Int("fact-augmentation-level", 2, "0=off, 1=facts only, 2=facts+keyphrases")
	flagEnableChainOfNote     = flag.Bool("enable-chain-of-note", true, "Use ReadWithChainOfNote for synthesis")
)

func main() {
	flag.Parse()

	baseDir := filepath.Join(os.TempDir(), "longmemeval_palace_v2")
	_ = os.MkdirAll(baseDir, 0755)

	cfg := memory.PalaceConfig{
		BaseDir: baseDir,
	}
	globalStore = memory.NewPalaceStoreWithConfig(cfg)

	globalVectorStore = memory.NewVectorStore("localhost:6334", "longmemeval_memory")
	if globalVectorStore.Enabled {
		_ = globalVectorStore.CreateCollection(768)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"features": map[string]any{
				"turn_granularity":  *flagEnableTurnGranularity,
				"time_aware":        *flagEnableTimeAware,
				"fact_augmentation": *flagFactAugLevel,
				"chain_of_note":     *flagEnableChainOfNote,
			},
		})
	})

	http.HandleFunc("/ingest", handleIngest)
	http.HandleFunc("/retrieve", handleRetrieve)
	http.HandleFunc("/synthesize", handleSynthesize)

	log.Printf("LongMemEval server :8765 | turn=%v time=%v fact=%d con=%v",
		*flagEnableTurnGranularity, *flagEnableTimeAware, *flagFactAugLevel, *flagEnableChainOfNote)
	log.Fatal(http.ListenAndServe(":8765", nil))
}

func handleIngest(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req IngestRequest
	json.Unmarshal(body, &req)

	turns := req.Turns
	if len(turns) == 0 {
		turns = req.History
	}

	for _, t := range turns {
		now := time.Now()
		ts := now
		if t.Timestamp != "" {
			if parsed, err := time.Parse(time.RFC3339, t.Timestamp); err == nil {
				ts = parsed
			}
		}

		entry := memory.MemoryEntry{
			ID:           memory.GenerateMemoryID(),
			Type:         "conversation_turn",
			Tier:         memory.TierWorking,
			Content:      memory.MemoryContent{Full: t.Content, Summary: truncate(t.Content, 280)},
			Cycle:        t.Cycle,
			CreatedAt:    now,
			UpdatedAt:    now,
			Timestamp:    ts,
			TurnID:       memory.GenerateMemoryID(),
			SessionID:    req.ConvID,
			OriginalText: t.Content,
		}

		if *flagEnableTurnGranularity {
			_ = globalStore.IngestTurn(entry)
		} else {
			_ = globalStore.Write(entry)
		}

		// Also store in Qdrant for hybrid retrieval
		if globalVectorStore != nil && globalVectorStore.Enabled {
			vec := globalStore.Config.EmbeddingFunc(t.Content, 768)
			payload := map[string]interface{}{
				"conv_id":   req.ConvID,
				"turn_id":   entry.TurnID,
				"timestamp": ts.Unix(),
			}
			_ = globalVectorStore.StoreVector(uuid.NewString(), vec, payload)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "ingested": fmt.Sprintf("%d turns", len(turns))})
}

func handleRetrieve(w http.ResponseWriter, r *http.Request) {
	var req RetrieveRequest
	json.NewDecoder(r.Body).Decode(&req)
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// Time-aware expansion (when enabled)
	filter := map[string]interface{}{}
	if *flagEnableTimeAware {
		if hasTemporal, start, end := classifyTemporalIntent(req.Query); hasTemporal {
			filter["timestamp"] = map[string]interface{}{
				"gte": float64(start.Unix()),
				"lte": float64(end.Unix()),
			}
		}
	}

	var combined []memory.MemoryEntry
	seen := make(map[string]bool)

	// Vector path (with filter when time-aware is active)
	if globalVectorStore != nil && globalVectorStore.Enabled {
		queryVec := globalStore.Config.EmbeddingFunc(req.Query, 768)
		vecResults, _ := globalVectorStore.SearchSimilar(queryVec, req.Limit*2, filter, true)
		for _, vr := range vecResults {
			if entry, ok := globalStore.Load(vr.ID, memory.TierSemantic); ok {
				if !seen[entry.ID] {
					seen[entry.ID] = true
					combined = append(combined, entry)
				}
			}
		}
	}

	// File-based hybrid (keyword + semantic)
	keywordResults := globalStore.SearchMemory(req.Query, nil, req.Limit*3, nil)
	for _, e := range keywordResults {
		if !seen[e.ID] {
			seen[e.ID] = true
			combined = append(combined, e)
		}
	}

	// Fact augmentation level filtering
	if *flagFactAugLevel >= 1 {
		semantic := globalStore.ListEntriesInTier(memory.TierSemantic)
		for _, e := range semantic {
			if !seen[e.ID] {
				seen[e.ID] = true
				combined = append(combined, e)
			}
		}
	}

	if len(combined) > req.Limit {
		combined = combined[:req.Limit]
	}

	hits := make([]MemoryHit, 0, len(combined))
	for _, e := range combined {
		hits = append(hits, MemoryHit{
			ID:        e.ID,
			Summary:   e.Content.Summary,
			Full:      e.Content.Full,
			Timestamp: e.Timestamp.Format(time.RFC3339),
			TurnID:    e.TurnID,
		})
	}

	json.NewEncoder(w).Encode(RetrieveResponse{Memories: hits})
}

func handleSynthesize(w http.ResponseWriter, r *http.Request) {
	var req SynthesizeRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Convert MemoryHit back to MemoryEntry for the CoN method
	entries := make([]memory.MemoryEntry, 0, len(req.Retrieved))
	for _, h := range req.Retrieved {
		entries = append(entries, memory.MemoryEntry{
			ID:        h.ID,
			TurnID:    h.TurnID,
			Timestamp: parseTime(h.Timestamp),
			Content:   memory.MemoryContent{Summary: h.Summary, Full: h.Full},
		})
	}

	prompt := ""
	if *flagEnableChainOfNote {
		p, _ := globalStore.ReadWithChainOfNote(entries, req.Query)
		prompt = p
	} else {
		// Fallback simple prompt
		prompt = "Query: " + req.Query + "\nContext: " + fmt.Sprintf("%d items", len(entries))
	}

	json.NewEncoder(w).Encode(SynthesizeResponse{
		Prompt: prompt,
		// In real usage the ego LLM would be called here with the prompt
	})
}

// classifyTemporalIntent - lightweight version for server (reuses logic from memory package in real impl)
func classifyTemporalIntent(query string) (bool, time.Time, time.Time) {
	q := strings.ToLower(query)
	now := time.Now()

	if strings.Contains(q, "last week") || strings.Contains(q, "past week") {
		return true, now.AddDate(0, 0, -7), now
	}
	if strings.Contains(q, "last month") {
		return true, now.AddDate(0, -1, 0), now
	}
	if strings.Contains(q, "this year") || strings.Contains(q, "in 2025") || strings.Contains(q, "in 2026") {
		return true, time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC), now
	}
	return false, time.Time{}, time.Time{}
}

func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Example helper to output LongMemEval JSONL line
func outputLongMemEvalJSONL(questionID, answer string, confidence float64) string {
	line, _ := json.Marshal(LongMemEvalAnswer{
		QuestionID: questionID,
		Answer:     answer,
		Confidence: confidence,
	})
	return string(line)
}
