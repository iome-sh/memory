// TTFH-shaped walking skeleton for the palace kernel (not Memory GA).
//
// Ingest three RCA-shaped turns, retrieve in the same process, list facts-as-of,
// and print provenance.source_hint. Hash embedder · no Qdrant · no cloud palace.
// dual_write OFF.
//
//	go run ./examples/ttfh_rca
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iome-sh/memory"
)

func main() {
	dir := strings.TrimSpace(os.Getenv("PALACE_ROOT"))
	if dir == "" {
		var err error
		dir, err = os.MkdirTemp("", "ttfh-palace-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "palace dir: %v\n", err)
			os.Exit(1)
		}
	} else if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "palace dir: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("palace:", dir)
	store := memory.NewPalaceStore(dir)
	session := "inc-webhook-5xx"

	turns := []memory.MemoryEntry{
		{
			SessionID: session,
			Content: memory.MemoryContent{
				Summary: "PagerDuty page: webhook ingress 5xx",
				Full:    "On-call: webhook ingress returned 5xx. Start RCA from the signed delivery, not the dashboard chrome.",
				Tags:    []string{"pagerduty"},
			},
			ExtractedFacts: []string{"PagerDuty page fired for webhook ingress 5xx"},
		},
		{
			SessionID: session,
			Content: memory.MemoryContent{
				Summary: "Signed delivery HTTP 200; dashboard still waiting",
				Full:    "HMAC-verified delivery returned 200. Consume receipt is a different clock. Dashboard waiting is not a 200 miss.",
				Tags:    []string{"hmac"},
			},
			ExtractedFacts: []string{"HMAC 200 is not a consume receipt"},
		},
		{
			SessionID: session,
			Content: memory.MemoryContent{
				Summary: "CreateConsumer failed: mode column NULL",
				Full:    "Durable consumer insert wrote SQL NULL into a NOT NULL mode column. Persist trimmed mode (empty → durable) on insert.",
				Tags:    []string{"storage"},
			},
			ExtractedFacts: []string{"CreateConsumer 500 when consumers.mode is NULL"},
		},
	}

	for i, turn := range turns {
		if err := store.IngestTurn(turn); err != nil {
			fmt.Fprintf(os.Stderr, "ingest turn %d: %v\n", i+1, err)
			os.Exit(1)
		}
	}

	hits := store.SearchMemoryWithOptions("hmac consume receipt", memory.SearchMemoryOptions{
		SessionID: session,
		Limit:     10,
	})
	fmt.Println("retrieve (same process):")
	for _, h := range hits {
		fmt.Printf("  %s  source_hint=%s\n", h.Content.Summary, h.Provenance.SourceHint)
	}

	facts := store.ListFactsAsOf(memory.FactsAsOfOptions{
		AsOf:      time.Now().UTC(),
		SessionID: session,
		Limit:     10,
	})
	fmt.Println("facts-as-of:")
	for _, f := range facts {
		fmt.Printf("  %s  source_hint=%s\n", f.Content.Summary, f.Provenance.SourceHint)
	}
}
