package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sudo-jin/memory"
)

// FlexAnswer unmarshals LongMemEval answers that may be JSON strings or numbers.
type FlexAnswer string

func (a *FlexAnswer) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*a = ""
		return nil
	}
	switch data[0] {
	case '"':
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*a = FlexAnswer(s)
		return nil
	case 'n':
		*a = ""
		return nil
	default:
		var n json.Number
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		*a = FlexAnswer(n.String())
		return nil
	}
}

func (a FlexAnswer) String() string {
	return string(a)
}

// LongMemEvalInstance mirrors the official LongMemEval JSON schema (oracle subset).
type LongMemEvalInstance struct {
	QuestionID         string           `json:"question_id"`
	QuestionType       string           `json:"question_type,omitempty"`
	Question           string           `json:"question"`
	Answer             FlexAnswer       `json:"answer"`
	QuestionDate       string           `json:"question_date,omitempty"`
	HaystackSessionIDs []string         `json:"haystack_session_ids,omitempty"`
	HaystackDates      []string         `json:"haystack_dates,omitempty"`
	HaystackSessions   [][]HaystackTurn `json:"haystack_sessions"`
	AnswerSessionIDs   []string         `json:"answer_session_ids,omitempty"`
}

type HaystackTurn struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	HasAnswer bool   `json:"has_answer,omitempty"`
}

// BenchOptions configures an offline recall run (no OpenAI).
type BenchOptions struct {
	DatasetPath string
	Limit       int
	TopK        int
	MinRecall   float64
	ModelPath   string
}

// QuestionResult is the per-question recall outcome.
type QuestionResult struct {
	QuestionID string
	Question   string
	Answer     string
	Hit        bool
	TopSummary string
	TopFull    string
}

// BenchReport aggregates recall metrics.
type BenchReport struct {
	Total          int
	Hits           int
	Recall         float64
	TotalDuration  time.Duration
	AvgQuestionMs  float64
	EmbedCallsNote string
	Results        []QuestionResult
}

func main() {
	dataset := flag.String("dataset", "", "path to LongMemEval JSON array (required)")
	limit := flag.Int("limit", 0, "max examples to run (0 = all)")
	topK := flag.Int("topk", 5, "top-k memories to check for answer overlap")
	minRecall := flag.Float64("min-recall", 0.6, "minimum aggregate recall; exit 1 if below")
	jsonReport := flag.String("json-report", "", "write aggregate JSON report to this path")
	quiet := flag.Bool("quiet", false, "only print aggregate recall line")
	flag.Parse()

	if strings.TrimSpace(*dataset) == "" {
		fmt.Fprintln(os.Stderr, "error: -dataset is required")
		os.Exit(2)
	}

	start := time.Now()
	report, err := RunBench(BenchOptions{
		DatasetPath: *dataset,
		Limit:       *limit,
		TopK:        *topK,
		MinRecall:   *minRecall,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "bench error: %v\n", err)
		os.Exit(2)
	}
	report.TotalDuration = time.Since(start)
	if report.Total > 0 {
		report.AvgQuestionMs = float64(report.TotalDuration.Milliseconds()) / float64(report.Total)
	}
	report.EmbedCallsNote = "SearchMemory precomputes one embed per entry (batch ONNX when configured); avoids O(n log n) re-embeds in sort comparator"

	if strings.TrimSpace(*jsonReport) != "" {
		payload := struct {
			Total          int     `json:"total"`
			Hits           int     `json:"hits"`
			Recall         float64 `json:"recall"`
			TotalDuration  string  `json:"total_duration"`
			AvgQuestionMs  float64 `json:"avg_question_ms"`
			EmbedCallsNote string  `json:"embed_calls_note"`
			MinRecall      float64 `json:"min_recall"`
		}{
			Total:          report.Total,
			Hits:           report.Hits,
			Recall:         report.Recall,
			TotalDuration:  report.TotalDuration.String(),
			AvgQuestionMs:  report.AvgQuestionMs,
			EmbedCallsNote: report.EmbedCallsNote,
			MinRecall:      *minRecall,
		}
		data, marshalErr := json.MarshalIndent(payload, "", "  ")
		if marshalErr != nil {
			fmt.Fprintf(os.Stderr, "json report error: %v\n", marshalErr)
			os.Exit(2)
		}
		if writeErr := os.WriteFile(*jsonReport, append(data, '\n'), 0o644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "json report write error: %v\n", writeErr)
			os.Exit(2)
		}
	}

	if !*quiet {
		for _, r := range report.Results {
			status := "MISS"
			if r.Hit {
				status = "HIT"
			}
			fmt.Printf("[%s] %s recall=%s answer=%q\n", status, r.QuestionID, status, r.Answer)
			if !r.Hit {
				fmt.Printf("  question: %s\n", r.Question)
				fmt.Printf("  top summary: %s\n", r.TopSummary)
				fmt.Printf("  top full: %s\n", r.TopFull)
			}
		}
		fmt.Println()
	}

	fmt.Printf("aggregate recall: %d/%d = %.2f (min-recall=%.2f, total=%s, avg=%.1fms/q)\n",
		report.Hits, report.Total, report.Recall, *minRecall, report.TotalDuration.Round(time.Millisecond), report.AvgQuestionMs)

	if report.Recall < *minRecall {
		os.Exit(1)
	}
}

// RunBench executes ingest+retrieve recall for each LongMemEval instance.
func RunBench(opts BenchOptions) (BenchReport, error) {
	instances, err := loadDataset(opts.DatasetPath)
	if err != nil {
		return BenchReport{}, err
	}
	if opts.Limit > 0 && opts.Limit < len(instances) {
		instances = instances[:opts.Limit]
	}
	if opts.TopK <= 0 {
		opts.TopK = 5
	}

	modelPath, err := resolveONNXModelPath(opts.ModelPath)
	if err != nil {
		return BenchReport{}, err
	}

	prev := os.Getenv(memory.EnvONNXModelPath)
	if err := os.Setenv(memory.EnvONNXModelPath, modelPath); err != nil {
		return BenchReport{}, err
	}
	defer func() {
		if prev == "" {
			_ = os.Unsetenv(memory.EnvONNXModelPath)
		} else {
			_ = os.Setenv(memory.EnvONNXModelPath, prev)
		}
	}()

	var embedFn memory.EmbeddingFunc
	var batchFn memory.BatchEmbeddingFunc
	if strings.TrimSpace(modelPath) != "" {
		emb, embErr := memory.NewGONNXEmbedder(memory.GONNXOptions{
			ModelPath: modelPath,
			Strict:    os.Getenv(memory.EnvEmbeddingStrict) == "1" || strings.EqualFold(os.Getenv(memory.EnvEmbeddingStrict), "true"),
		})
		if embErr != nil {
			return BenchReport{}, fmt.Errorf("onnx embedding init: %w", embErr)
		}
		defer func() { _ = emb.Close() }()
		embedFn = emb.Func()
		batchFn = emb.BatchFunc()
	} else {
		var fnErr error
		embedFn, fnErr = memory.NewGONNXEmbeddingFuncFromEnv()
		if fnErr != nil {
			return BenchReport{}, fmt.Errorf("onnx embedding init: %w", fnErr)
		}
	}
	embeddingDim := memory.ResolveEmbeddingDim(modelPath)

	report := BenchReport{Total: len(instances)}
	for _, inst := range instances {
		qr, err := evalInstance(inst, embedFn, batchFn, embeddingDim, opts.TopK)
		if err != nil {
			return BenchReport{}, fmt.Errorf("%s: %w", inst.QuestionID, err)
		}
		report.Results = append(report.Results, qr)
		if qr.Hit {
			report.Hits++
		}
	}
	if report.Total > 0 {
		report.Recall = float64(report.Hits) / float64(report.Total)
	}
	return report, nil
}

func loadDataset(path string) ([]LongMemEvalInstance, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read dataset: %w", err)
	}
	var instances []LongMemEvalInstance
	if err := json.Unmarshal(data, &instances); err != nil {
		return nil, fmt.Errorf("parse dataset: %w", err)
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("dataset is empty")
	}
	return instances, nil
}

func evalInstance(inst LongMemEvalInstance, embedFn memory.EmbeddingFunc, batchFn memory.BatchEmbeddingFunc, embeddingDim, topK int) (QuestionResult, error) {
	baseDir, err := os.MkdirTemp("", "longmemeval_bench_*")
	if err != nil {
		return QuestionResult{}, err
	}
	defer os.RemoveAll(baseDir)

	cfg := memory.PalaceConfig{
		BaseDir:            baseDir,
		EmbeddingFunc:      embedFn,
		BatchEmbeddingFunc: batchFn,
	}
	store := memory.NewPalaceStoreWithConfig(cfg)

	convID := inst.QuestionID
	for sessIdx, session := range inst.HaystackSessions {
		sessionID := fmt.Sprintf("%s-sess-%d", convID, sessIdx)
		if sessIdx < len(inst.HaystackSessionIDs) && inst.HaystackSessionIDs[sessIdx] != "" {
			sessionID = inst.HaystackSessionIDs[sessIdx]
		}

		var sessionTS time.Time
		if sessIdx < len(inst.HaystackDates) && inst.HaystackDates[sessIdx] != "" {
			if parsed, parseErr := time.Parse(time.RFC3339, inst.HaystackDates[sessIdx]); parseErr == nil {
				sessionTS = parsed
			}
		}
		if sessionTS.IsZero() {
			sessionTS = time.Now().Add(-time.Duration(sessIdx) * time.Hour)
		}

		for turnIdx, turn := range session {
			ts := sessionTS.Add(time.Duration(turnIdx) * time.Minute)
			entry := memory.MemoryEntry{
				ID:           memory.GenerateMemoryID(),
				Type:         "conversation_turn",
				Tier:         memory.TierWorking,
				Content:      memory.MemoryContent{Full: turn.Content, Summary: truncate(turn.Content, 280)},
				Cycle:        turnIdx + 1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
				Timestamp:    ts,
				TurnID:       memory.GenerateMemoryID(),
				SessionID:    sessionID,
				OriginalText: turn.Content,
			}
			if err := store.IngestTurn(entry); err != nil {
				return QuestionResult{}, fmt.Errorf("ingest turn %d/%d: %w", sessIdx, turnIdx, err)
			}
		}
	}

	queryVec := embedFn(inst.Question, embeddingDim)
	memories := store.SearchMemory(inst.Question, nil, topK, queryVec)

	qr := QuestionResult{
		QuestionID: inst.QuestionID,
		Question:   inst.Question,
		Answer:     inst.Answer.String(),
	}
	if len(memories) > 0 {
		qr.TopSummary = memories[0].Content.Summary
		qr.TopFull = memories[0].Content.Full
	}
	qr.Hit = answerInMemories(inst.Answer.String(), memories, topK)
	return qr, nil
}

func answerInMemories(answer string, memories []memory.MemoryEntry, topK int) bool {
	if strings.TrimSpace(answer) == "" || len(memories) == 0 {
		return false
	}
	answerLower := strings.ToLower(strings.TrimSpace(answer))
	if len(memories) > topK {
		memories = memories[:topK]
	}

	var corpus strings.Builder
	for _, m := range memories {
		corpus.WriteString(strings.ToLower(m.Content.Full))
		corpus.WriteByte(' ')
		corpus.WriteString(strings.ToLower(m.Content.Summary))
		corpus.WriteByte(' ')
	}
	text := corpus.String()

	if strings.Contains(text, answerLower) {
		return true
	}

	tokens := significantAnswerTokens(answer)
	if len(tokens) == 0 {
		return false
	}
	matched := 0
	for _, tok := range tokens {
		if strings.Contains(text, tok) {
			matched++
		}
	}
	threshold := len(tokens)
	if threshold > 1 {
		threshold = (len(tokens) + 1) / 2
	}
	return matched >= threshold
}

func significantAnswerTokens(answer string) []string {
	stop := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "to": true,
		"in": true, "on": true, "at": true, "of": true, "for": true, "is": true,
		"are": true, "was": true, "were": true, "my": true, "i": true,
	}
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(answer), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}) {
		if len(w) < 3 || stop[w] {
			continue
		}
		out = append(out, w)
	}
	return out
}

func resolveONNXModelPath(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("model path %s: %w", explicit, err)
		}
		return explicit, nil
	}
	if dir := strings.TrimSpace(os.Getenv(memory.EnvONNXModelPath)); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		}
	}

	repoRoot := findRepoRoot()
	cached := filepath.Join(repoRoot, "testdata", "models", "KnightsAnalytics_all-MiniLM-L6-v2")
	if info, err := os.Stat(cached); err == nil && info.IsDir() {
		return cached, nil
	}

	switch strings.ToLower(strings.TrimSpace(os.Getenv(memory.EnvEmbeddingStrict))) {
	case "1", "true", "yes", "on":
		return "", fmt.Errorf("%s is set but no ONNX model found (set %s or cache under testdata/models/)",
			memory.EnvEmbeddingStrict, memory.EnvONNXModelPath)
	}

	dir, err := downloadONNXModel(repoRoot)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}
		dir = parent
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}