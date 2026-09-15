// Support-department RCA overlay kit (V1.6 Wave 1).
//
// Ingest ticket export + policy + macro as private overlay, retrieve the
// refund policy, list facts-as-of the ticket created time, print source_hint.
// Hash embedder · no Qdrant · no cloud palace.
//
//	go run ./examples/dept-rca/support
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/iome-sh/memory"
)

const (
	sessionID        = "dept-support-zd-1001"
	ticketCreatedRFC = "2026-06-15T14:22:00Z"
	policyFromRFC    = "2026-01-01T00:00:00Z"
	macroReplyRFC    = "2026-06-15T16:05:00Z"
)

func main() {
	dir := strings.TrimSpace(os.Getenv("PALACE_ROOT"))
	if dir == "" {
		var err error
		dir, err = os.MkdirTemp("", "dept-support-palace-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "palace dir: %v\n", err)
			os.Exit(1)
		}
	} else if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "palace dir: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("palace:", dir)

	kit, err := kitDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "kit dir: %v\n", err)
		os.Exit(1)
	}

	ticketAt, err := time.Parse(time.RFC3339, ticketCreatedRFC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ticket created: %v\n", err)
		os.Exit(1)
	}
	policyFrom, err := time.Parse(time.RFC3339, policyFromRFC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "policy valid_from: %v\n", err)
		os.Exit(1)
	}
	macroAt, err := time.Parse(time.RFC3339, macroReplyRFC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macro time: %v\n", err)
		os.Exit(1)
	}

	bodies := map[string]string{}
	for _, name := range []string{"ticket-export.md", "policy.md", "macro.md"} {
		b, err := os.ReadFile(filepath.Join(kit, name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", name, err)
			os.Exit(1)
		}
		bodies[name] = string(b)
	}

	store := memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
		BaseDir:       dir,
		EmbeddingFunc: memory.GenerateSimpleEmbedding,
	})
	tags := []string{"dept:support", "scenario:support"}

	turns := []memory.MemoryEntry{
		{
			SessionID: sessionID,
			Timestamp: policyFrom,
			TemporalTags: []string{
				"valid_from:" + policyFrom.UTC().Format(time.RFC3339),
			},
			Content: memory.MemoryContent{
				Summary: "Unused-seat refund policy",
				Full:    bodies["policy.md"],
				Tags:    tags,
			},
			ExtractedFacts: []string{"Unused seats may be refunded within 14 days of invoice"},
		},
		{
			SessionID: sessionID,
			Timestamp: ticketAt,
			Content: memory.MemoryContent{
				Summary: "ZD-1001 unused-seat refund",
				Full:    bodies["ticket-export.md"],
				Tags:    tags,
			},
			ExtractedFacts: []string{"Example workspace requested a refund for unused seats on ticket ZD-1001"},
		},
		{
			SessionID: sessionID,
			Timestamp: macroAt,
			Content: memory.MemoryContent{
				Summary: "Macro: unused-seat refund points at 14-day policy",
				Full:    bodies["macro.md"],
				Tags:    tags,
			},
			ExtractedFacts: []string{"Agent reply points at the 14-day unused-seat refund policy"},
		},
	}

	for i, turn := range turns {
		if err := store.IngestTurn(turn); err != nil {
			fmt.Fprintf(os.Stderr, "ingest turn %d: %v\n", i+1, err)
			os.Exit(1)
		}
	}

	hits := store.SearchMemoryWithOptions("refund unused seats policy", memory.SearchMemoryOptions{
		SessionID: sessionID,
		Limit:     10,
	})
	fmt.Println("retrieve (same process):")
	for _, h := range hits {
		fmt.Printf("  %s  source_hint=%s\n", h.Content.Summary, h.Provenance.SourceHint)
	}

	facts := store.ListFactsAsOf(memory.FactsAsOfOptions{
		AsOf:      ticketAt,
		SessionID: sessionID,
		Query:     "refund",
		Limit:     10,
	})
	fmt.Println("facts-as-of ticket", ticketCreatedRFC+":")
	for _, f := range facts {
		fmt.Printf("  %s  source_hint=%s\n", f.Content.Summary, f.Provenance.SourceHint)
	}
}

func kitDir() (string, error) {
	var candidates []string
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Dir(file))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd)
		candidates = append(candidates, filepath.Join(wd, "examples", "dept-rca", "support"))
	}
	for _, dir := range candidates {
		if kitFilesPresent(dir) {
			return dir, nil
		}
	}
	return "", fmt.Errorf("examples/dept-rca/support markdown not found")
}

func kitFilesPresent(dir string) bool {
	for _, name := range []string{"ticket-export.md", "policy.md", "macro.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return false
		}
	}
	return true
}
