package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/iome-sh/memory"
	"github.com/iome-sh/memory/internal/longmemeval"
)

func main() {
	dataRoot := flag.String("data-root", "", "official V2 data root (questions.jsonl + trajectories.jsonl + haystacks/)")
	tier := flag.String("tier", "small", "haystack tier: small or medium")
	limit := flag.Int("limit", 0, "max questions (0 = all loaded)")
	topK := flag.Int("topk", 5, "top-k text context items")
	flag.Parse()

	if strings.TrimSpace(*dataRoot) == "" {
		fmt.Fprintln(os.Stderr, "error: -data-root is required (do not vendor the 7GB snapshot)")
		os.Exit(2)
	}

	fmt.Fprintln(os.Stderr, "longmemeval-v2-bench: Insert/Query adapter over Palace (text steps only).")
	fmt.Fprintln(os.Stderr, "longmemeval-v2-bench: printed overlap is judge-free string overlap. Not official V2 LAFS. Not V1 gpt-4o QA. Hash default. dual_write OFF. Not Memory GA.")

	ds, err := longmemeval.LoadV2(*dataRoot, *tier)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(2)
	}
	questions := ds.Questions
	if *limit > 0 && *limit < len(questions) {
		questions = questions[:*limit]
	}

	hits := 0
	for _, q := range questions {
		store := memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
			BaseDir:       mustTemp(),
			EmbeddingFunc: memory.GenerateSimpleEmbedding,
		})
		adapter := &longmemeval.PalaceMemory{Store: store, TopK: *topK}
		for _, trajID := range ds.Haystack[q.ID] {
			tr, ok := ds.Trajectories[trajID]
			if !ok {
				fmt.Fprintf(os.Stderr, "MISS %s: unknown trajectory %s\n", q.ID, trajID)
				continue
			}
			if err := adapter.Insert(tr); err != nil {
				fmt.Fprintf(os.Stderr, "MISS %s: insert %s: %v\n", q.ID, trajID, err)
				continue
			}
		}
		items := adapter.Query(q.Question, q.Image)
		hit := overlapHit(q.Answer, items)
		if hit {
			hits++
			fmt.Printf("[HIT] %s overlap answer=%q\n", q.ID, q.Answer)
		} else {
			fmt.Printf("[MISS] %s overlap answer=%q\n", q.ID, q.Answer)
		}
	}
	total := len(questions)
	frac := 0.0
	if total > 0 {
		frac = float64(hits) / float64(total)
	}
	fmt.Printf("aggregate overlap (not V2 LAFS, not V1 gpt-4o): %d/%d = %.2f\n", hits, total, frac)
}

func mustTemp() string {
	dir, err := os.MkdirTemp("", "lme_v2_bench_*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tempdir: %v\n", err)
		os.Exit(2)
	}
	return dir
}

func overlapHit(answer string, items []longmemeval.ContextItem) bool {
	gold := strings.ToLower(strings.TrimSpace(answer))
	if gold == "" {
		return false
	}
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Value), gold) {
			return true
		}
	}
	return false
}
