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
	"github.com/iome-sh/memory"
	"github.com/iome-sh/memory/internal/longmemeval"
)

// LongMemEval local bench harness for LongMemEval-S / LongMemEval-M.
// Residual-honest: not a production memory service, not Memory GA, not live ingest.
// dual_write OFF. Judge-free overlap / fixtures ≠ official APPLY.
//
// Flags:
//   -enable-turn-granularity     Use IngestTurn + turn-level metadata (default true)
//   -enable-time-aware-expansion Time-aware query expansion + timestamp filtering
//   -fact-augmentation-level     0=off, 1=facts, 2=facts+keyphrases (default 2)
//   -enable-chain-of-note        Use ReadWithChainOfNote for answer synthesis
//
// Endpoints:
//   POST /ingest     - Persist conversation turns (IngestTurn when enabled). Failed palace persist is not status ok.
//   POST /retrieve   - File-based hybrid retrieval; Qdrant only if LONGMEMEVAL_QDRANT_URL is set
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
	SessionID string  `json:"session_id,omitempty"`
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

// IngestResponse is the /ingest JSON body. Failed palace persist is never status ok.
type IngestResponse struct {
	Status   string `json:"status"`
	Ingested int    `json:"ingested"`
	Error    string `json:"error,omitempty"`
}

type RetrieveRequest struct {
	Query     string `json:"query"`
	Limit     int    `json:"limit"`
	SessionID string `json:"session_id,omitempty"`
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
	embeddingDim      int

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

	modelPath := strings.TrimSpace(os.Getenv(memory.EnvONNXModelPath))
	embeddingDim = memory.ResolveEmbeddingDim(modelPath)
	embedFn, err := memory.NewGONNXEmbeddingFuncFromEnv()
	if err != nil {
		log.Printf("onnx embedding init failed (%v); falling back to hash dim=%d", err, memory.DefaultHashEmbeddingDim)
		embedFn = memory.GenerateSimpleEmbedding
		embeddingDim = memory.DefaultHashEmbeddingDim
	} else if modelPath == "" {
		log.Printf("embedding mode=hash dim=%d (set %s for ONNX MiniLM)", embeddingDim, memory.EnvONNXModelPath)
	} else {
		log.Printf("embedding mode=onnx dim=%d model=%s", embeddingDim, modelPath)
	}

	cfg := memory.PalaceConfig{
		BaseDir:       baseDir,
		EmbeddingFunc: embedFn,
	}
	globalStore = memory.NewPalaceStoreWithConfig(cfg)

	// Qdrant is opt-in. Empty URL keeps ingest file-based (not live hybrid ingest).
	// dual_write OFF · not Memory GA · palace persist is the ingest truth.
	qdrantURL := strings.TrimSpace(os.Getenv("LONGMEMEVAL_QDRANT_URL"))
	globalVectorStore = memory.NewVectorStore(qdrantURL, "longmemeval_memory")
	if globalVectorStore.Enabled {
		_ = globalVectorStore.CreateCollection(embeddingDim)
		log.Printf("qdrant opt-in url=%s (sidecar only; not live ingest)", qdrantURL)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status":     "ok",
			"embed_mode": embedMode(),
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

	ingested := 0
	var persistErrs []string
	for _, t := range turns {
		now := time.Now()
		ts := now
		if t.Timestamp != "" {
			if parsed, ok := longmemeval.ParseTime(t.Timestamp); ok {
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

		if err := persistIngestTurn(entry); err != nil {
			persistErrs = append(persistErrs, err.Error())
			continue
		}
		ingested++

		// Optional Qdrant sidecar only when LONGMEMEVAL_QDRANT_URL is set.
		// dual_write OFF: palace persist is ingest truth, not live hybrid ingest.
		if globalVectorStore != nil && globalVectorStore.Enabled {
			vec := globalStore.Config.EmbeddingFunc(t.Content, embeddingDim)
			payload := map[string]interface{}{
				"conv_id":   req.ConvID,
				"turn_id":   entry.TurnID,
				"timestamp": ts.Unix(),
			}
			if err := globalVectorStore.StoreVector(uuid.NewString(), vec, payload); err != nil {
				log.Printf("qdrant store skipped (opt-in sidecar, not live ingest): %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := IngestResponse{Status: "ok", Ingested: ingested}
	if len(persistErrs) > 0 {
		resp.Status = "error"
		resp.Error = persistErrs[0]
		if len(persistErrs) > 1 {
			resp.Error = fmt.Sprintf("%s (%d persist errors)", persistErrs[0], len(persistErrs))
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(resp)
}

var persistIngestTurn = persistIngestTurnDefault

func persistIngestTurnDefault(entry memory.MemoryEntry) error {
	if *flagEnableTurnGranularity {
		return globalStore.IngestTurn(entry)
	}
	return globalStore.Write(entry)
}

func handleRetrieve(w http.ResponseWriter, r *http.Request) {
	var req RetrieveRequest
	json.NewDecoder(r.Body).Decode(&req)
	if req.Limit <= 0 {
		req.Limit = 50
	}
	sessionID := strings.TrimSpace(req.SessionID)

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
		queryVec := globalStore.Config.EmbeddingFunc(req.Query, embeddingDim)
		vecResults, _ := globalVectorStore.SearchSimilar(queryVec, req.Limit*2, filter, true)
		for _, vr := range vecResults {
			if entry, ok := globalStore.Load(vr.ID, memory.TierSemantic); ok {
				if sessionID != "" && entry.SessionID != sessionID {
					continue
				}
				if !seen[entry.ID] {
					seen[entry.ID] = true
					combined = append(combined, entry)
				}
			}
		}
	}

	// File-based hybrid. Official QA must pass session_id so a shared palace
	// is not other-session dominated (#55). Overlap-ranked keyword hits keep
	// gold phrases ahead of OR-any-token flood (#56).
	queryVec := globalStore.Config.EmbeddingFunc(req.Query, embeddingDim)
	keywordResults := globalStore.SearchMemoryWithOptions(req.Query, memory.SearchMemoryOptions{
		SessionID: sessionID,
		Limit:     req.Limit,
		QueryVec:  queryVec,
	})
	for _, e := range keywordResults {
		if !seen[e.ID] {
			seen[e.ID] = true
			combined = append(combined, e)
		}
	}

	// Fact augmentation: session-scoped only. Unscoped dump of TierSemantic
	// re-introduces other-session domination on official shared-palace runs.
	if *flagFactAugLevel >= 1 && sessionID == "" {
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
		ts := ""
		if !e.Timestamp.IsZero() {
			ts = e.Timestamp.Format(time.RFC3339)
		}
		hits = append(hits, MemoryHit{
			ID:        e.ID,
			Summary:   e.Content.Summary,
			Full:      e.Content.Full,
			Timestamp: ts,
			TurnID:    e.TurnID,
			SessionID: e.SessionID,
		})
	}

	json.NewEncoder(w).Encode(RetrieveResponse{Memories: hits})
}

func embedMode() string {
	if strings.TrimSpace(os.Getenv(memory.EnvONNXModelPath)) != "" && embeddingDim == memory.MiniLMEmbeddingDim {
		return "onnx"
	}
	return "hash"
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
	if t, ok := longmemeval.ParseTime(s); ok {
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
