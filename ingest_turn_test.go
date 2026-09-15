package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestIngestTurn_FactChildrenStampValidFrom(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-1",
		SessionID: "sess-1",
		Content:   MemoryContent{Full: "hello world"},
		ExtractedFacts: []string{
			"I live in Seattle",
			"  ",
			"My name is Alice",
		},
	}); err != nil {
		t.Fatal(err)
	}

	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) != 2 {
		t.Fatalf("got %d semantic facts, want 2", len(facts))
	}
	now := time.Now().UTC()
	for _, f := range facts {
		if f.Type != "turn_fact" {
			t.Fatalf("type = %q, want turn_fact", f.Type)
		}
		if f.SessionID != "sess-1" || f.Provenance.SourceStep != "ingest_turn_fact" {
			t.Fatalf("unexpected fact metadata: %+v", f)
		}
		if !hasValidFromTag(f) {
			t.Fatalf("missing valid_from on %s tags=%v", f.ID, f.TemporalTags)
		}
		from, until := ParseValidityWindow(f)
		if from == nil {
			t.Fatalf("ParseValidityWindow from=nil on %s tags=%v", f.ID, f.TemporalTags)
		}
		if until != nil {
			t.Fatalf("did not expect valid_until on new fact %s tags=%v", f.ID, f.TemporalTags)
		}
		if !EntryValidAt(f, now) {
			t.Fatalf("EntryValidAt(now) false for %s from=%v", f.ID, from)
		}
		if !EntryValidAt(f, time.Time{}) {
			t.Fatalf("EntryValidAt(zero→now) false for %s", f.ID)
		}
	}
}

func TestIngestTurn_FactWriteError(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStore(base)
	semDir := filepath.Join(base, "tier-4-semantic")
	if err := os.Chmod(semDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(semDir, 0755) })

	err := store.IngestTurn(MemoryEntry{
		Content:        MemoryContent{Full: "turn body"},
		ExtractedFacts: []string{"I graduated from MIT"},
	})
	if err == nil {
		t.Fatal("expected fact Write error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to write turn fact") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Contract (IngestTurn godoc): partial persist. Parent is written first (default
	// tier 0 → contextual) and remains after a child Write error; the failed fact is absent.
	turns := store.ListEntriesInTier(TierContextual)
	if len(turns) != 1 {
		t.Fatalf("contextual turns = %d, want 1 (parent persisted before fact failure)", len(turns))
	}
	if got := store.ListEntriesInTier(TierSemantic); len(got) != 0 {
		t.Fatalf("semantic facts after failed write = %d, want 0", len(got))
	}
}

func TestInheritTurnFactTags(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "empty parent",
			in:   nil,
			want: []string{"fact_augmented", "from_turn"},
		},
		{
			name: "mcp source and role",
			in:   []string{"source:iomesh-memory-mcp", "role:user"},
			want: []string{"source:iomesh-memory-mcp", "role:user", "fact_augmented", "from_turn"},
		},
		{
			name: "caller longmemeval inherited",
			in:   []string{"longmemeval"},
			want: []string{"longmemeval", "fact_augmented", "from_turn"},
		},
		{
			name: "blanks and duplicates dropped",
			in:   []string{"", "  ", "role:user", "role:user", "fact_augmented"},
			want: []string{"role:user", "fact_augmented", "from_turn"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := inheritTurnFactTags(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("inheritTurnFactTags(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestIngestTurn_FactChildrenInheritParentTagsNoDefaultLongmemeval(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       base,
		EmbeddingFunc: GenerateSimpleEmbedding,
	})
	parentTags := []string{"source:iomesh-memory-mcp", "role:user"}
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-mcp-1",
		SessionID: "sess-mcp",
		Content: MemoryContent{
			Full: "I live in Seattle. My name is Alice.",
			Tags: parentTags,
		},
	}); err != nil {
		t.Fatal(err)
	}

	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) == 0 {
		t.Fatal("expected auto-extracted turn_fact children")
	}

	wantChild := []string{"source:iomesh-memory-mcp", "role:user", "source_hint:private", "fact_augmented", "from_turn"}
	for _, f := range facts {
		if f.Type != "turn_fact" {
			t.Fatalf("type = %q, want turn_fact", f.Type)
		}
		if !reflect.DeepEqual(f.Content.Tags, wantChild) {
			t.Fatalf("child tags = %v, want %v", f.Content.Tags, wantChild)
		}
		if EntryHasTag(f, "longmemeval") {
			t.Fatalf("library ingest must not stamp longmemeval; tags=%v", f.Content.Tags)
		}
		if !EntryHasTag(f, "source:iomesh-memory-mcp") || !EntryHasTag(f, "role:user") {
			t.Fatalf("child missing inherited source/role tags: %v", f.Content.Tags)
		}

		// Disk contract: semantic JSON carries inherited tags, not the benchmark label.
		raw, err := os.ReadFile(filepath.Join(base, "tier-4-semantic", f.ID+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var disk MemoryEntry
		if err := json.Unmarshal(raw, &disk); err != nil {
			t.Fatal(err)
		}
		if disk.Type != "turn_fact" {
			t.Fatalf("disk type = %q, want turn_fact", disk.Type)
		}
		if !reflect.DeepEqual(disk.Content.Tags, wantChild) {
			t.Fatalf("disk tags = %v, want %v", disk.Content.Tags, wantChild)
		}
		if disk.Provenance.SourceStep != "ingest_turn_fact" {
			t.Fatalf("disk source_step = %q", disk.Provenance.SourceStep)
		}
	}

	parent, ok := store.Load("turn-mcp-1", TierContextual)
	if !ok {
		t.Fatal("parent turn missing")
	}
	if parent.Provenance.SourceHint != SourceHintPrivate || !EntryHasTag(parent, FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("parent missing private source class; hint=%q tags=%v", parent.Provenance.SourceHint, parent.Content.Tags)
	}

	bySource := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "source:iomesh-memory-mcp", Limit: 50})
	if len(bySource) != 1+len(facts) {
		t.Fatalf("tag=source:iomesh-memory-mcp got %d, want parent+children %d", len(bySource), 1+len(facts))
	}
	byPrefix := store.ListMemoryWithOptions(ListMemoryOptions{TagPrefix: "source:", Limit: 50})
	if len(byPrefix) != 1+len(facts) {
		t.Fatalf("tag_prefix=source: got %d, want parent+children %d", len(byPrefix), 1+len(facts))
	}
	byHint := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "source_hint:private", Limit: 50})
	if len(byHint) != 1+len(facts) {
		t.Fatalf("tag=source_hint:private got %d, want parent+children %d", len(byHint), 1+len(facts))
	}
	byBench := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "longmemeval", Limit: 50})
	if len(byBench) != 0 {
		t.Fatalf("tag=longmemeval got %d, want 0 on library ingest", len(byBench))
	}
}

func TestIngestTurn_FactChildrenInheritCallerLongmemeval(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       t.TempDir(),
		EmbeddingFunc: GenerateSimpleEmbedding,
	})
	if err := store.IngestTurn(MemoryEntry{
		Content: MemoryContent{
			Full: "I live in Seattle.",
			Tags: []string{"longmemeval"},
		},
		ExtractedFacts: []string{"I live in Seattle"},
	}); err != nil {
		t.Fatal(err)
	}
	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) != 1 {
		t.Fatalf("got %d semantic facts, want 1", len(facts))
	}
	if !EntryHasTag(facts[0], "longmemeval") {
		t.Fatalf("caller-supplied longmemeval must be inherited; tags=%v", facts[0].Content.Tags)
	}
	if !EntryHasTag(facts[0], "fact_augmented") || !EntryHasTag(facts[0], "from_turn") {
		t.Fatalf("missing structural markers: %v", facts[0].Content.Tags)
	}
	if facts[0].Provenance.SourceHint != SourceHintPrivate || !EntryHasTag(facts[0], FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("local-palace ingest must stamp private source class; hint=%q tags=%v", facts[0].Provenance.SourceHint, facts[0].Content.Tags)
	}
}

// TTFH-shaped walking skeleton (README / examples/ttfh_rca): three RCA turns,
// retrieve in the same process, facts-as-of, observable source_hint=private.
// A green unit test is not a live laptop PULSE run.
func TestIngestTurn_TTFHShapedWalkingSkeleton(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	session := "inc-webhook-5xx"
	turns := []MemoryEntry{
		{
			SessionID: session,
			Content: MemoryContent{
				Summary: "PagerDuty page: webhook ingress 5xx",
				Full:    "On-call: webhook ingress returned 5xx.",
				Tags:    []string{"pagerduty"},
			},
			ExtractedFacts: []string{"PagerDuty page fired for webhook ingress 5xx"},
		},
		{
			SessionID: session,
			Content: MemoryContent{
				Summary: "Signed delivery HTTP 200; dashboard still waiting",
				Full:    "HMAC-verified delivery returned 200. Consume receipt is a different clock.",
				Tags:    []string{"hmac"},
			},
			ExtractedFacts: []string{"HMAC 200 is not a consume receipt"},
		},
		{
			SessionID: session,
			Content: MemoryContent{
				Summary: "CreateConsumer failed: mode column NULL",
				Full:    "Durable consumer insert wrote SQL NULL into a NOT NULL mode column.",
				Tags:    []string{"storage"},
			},
			ExtractedFacts: []string{"CreateConsumer 500 when consumers.mode is NULL"},
		},
	}
	for i, turn := range turns {
		if err := store.IngestTurn(turn); err != nil {
			t.Fatalf("ingest %d: %v", i+1, err)
		}
	}

	hits := store.SearchMemoryWithOptions("hmac consume receipt", SearchMemoryOptions{
		SessionID: session,
		Limit:     10,
	})
	if len(hits) == 0 {
		t.Fatal("retrieve-after-ingest empty (same process)")
	}
	sawPrivate := false
	sawHMAC := false
	for _, h := range hits {
		if h.Provenance.SourceHint == SourceHintPrivate {
			sawPrivate = true
		}
		if strings.Contains(strings.ToLower(h.Content.Summary+h.Content.Full), "hmac") {
			sawHMAC = true
		}
	}
	if !sawPrivate {
		t.Fatalf("expected source_hint=private on retrieve hits: %+v", hits[0].Provenance)
	}
	if !sawHMAC {
		t.Fatalf("expected HMAC RCA turn in same-session retrieve, got %d hits", len(hits))
	}

	facts := store.ListFactsAsOf(FactsAsOfOptions{SessionID: session, Limit: 10})
	if len(facts) < 3 {
		t.Fatalf("facts-as-of got %d, want >= 3 extracted facts", len(facts))
	}
	for _, f := range facts {
		if f.Provenance.SourceHint != SourceHintPrivate {
			t.Fatalf("fact %s source_hint=%q want private", f.ID, f.Provenance.SourceHint)
		}
	}
}

// V1.6 support-department overlay kit (examples/dept-rca/support): ticket
// export + policy + macro as private overlay. Facts-as-of the ticket finds
// the 14-day unused-seat rule. source_hint stays private (not mesh). A green
// unit test is not E-G1.
func TestDeptRCASupportKit_WalkingSkeleton(t *testing.T) {
	kit := filepath.Join("examples", "dept-rca", "support")
	for _, name := range []string{"README.md", "ticket-export.md", "policy.md", "macro.md", "main.go"} {
		path := filepath.Join(kit, name)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("kit file %s: %v", path, err)
		}
		if len(strings.TrimSpace(string(b))) == 0 {
			t.Fatalf("kit file %s empty", path)
		}
	}

	policy, err := os.ReadFile(filepath.Join(kit, "policy.md"))
	if err != nil {
		t.Fatal(err)
	}
	policyBody := string(policy)
	if !strings.Contains(policyBody, "14 days") {
		t.Fatal("policy.md must state the 14-day unused-seat rule")
	}
	if !strings.Contains(policyBody, "valid_from 2026-01-01") {
		t.Fatal("policy.md must carry valid_from 2026-01-01")
	}

	ticket, err := os.ReadFile(filepath.Join(kit, "ticket-export.md"))
	if err != nil {
		t.Fatal(err)
	}
	ticketBody := string(ticket)
	if !strings.Contains(ticketBody, "ZD-1001") {
		t.Fatal("ticket-export.md must be ticket ZD-1001")
	}
	if !strings.Contains(ticketBody, "2026-06-15T14:22:00Z") {
		t.Fatal("ticket-export.md must record created 2026-06-15T14:22:00Z")
	}
	if !strings.Contains(ticketBody, "Example workspace") {
		t.Fatal("ticket-export.md must use Example workspace (no live customer)")
	}

	macro, err := os.ReadFile(filepath.Join(kit, "macro.md"))
	if err != nil {
		t.Fatal(err)
	}
	macroBody := string(macro)

	ticketAt, err := time.Parse(time.RFC3339, "2026-06-15T14:22:00Z")
	if err != nil {
		t.Fatal(err)
	}
	policyFrom, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	macroAt, err := time.Parse(time.RFC3339, "2026-06-15T16:05:00Z")
	if err != nil {
		t.Fatal(err)
	}

	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       t.TempDir(),
		EmbeddingFunc: GenerateSimpleEmbedding,
	})
	session := "dept-support-zd-1001"
	tags := []string{"dept:support", "scenario:support"}
	turns := []MemoryEntry{
		{
			SessionID: session,
			Timestamp: policyFrom,
			TemporalTags: []string{
				"valid_from:" + policyFrom.UTC().Format(time.RFC3339),
			},
			Content: MemoryContent{
				Summary: "Unused-seat refund policy",
				Full:    policyBody,
				Tags:    tags,
			},
			ExtractedFacts: []string{"Unused seats may be refunded within 14 days of invoice"},
		},
		{
			SessionID: session,
			Timestamp: ticketAt,
			Content: MemoryContent{
				Summary: "ZD-1001 unused-seat refund",
				Full:    ticketBody,
				Tags:    tags,
			},
			ExtractedFacts: []string{"Example workspace requested a refund for unused seats on ticket ZD-1001"},
		},
		{
			SessionID: session,
			Timestamp: macroAt,
			Content: MemoryContent{
				Summary: "Macro: unused-seat refund points at 14-day policy",
				Full:    macroBody,
				Tags:    tags,
			},
			ExtractedFacts: []string{"Agent reply points at the 14-day unused-seat refund policy"},
		},
	}
	for i, turn := range turns {
		if err := store.IngestTurn(turn); err != nil {
			t.Fatalf("ingest %d: %v", i+1, err)
		}
	}

	hits := store.SearchMemoryWithOptions("refund unused seats policy", SearchMemoryOptions{
		SessionID: session,
		Limit:     10,
	})
	if len(hits) == 0 {
		t.Fatal("retrieve-after-ingest empty (same process)")
	}
	sawPrivate := false
	sawPolicy := false
	for _, h := range hits {
		if h.Provenance.SourceHint == "mesh" {
			t.Fatalf("overlay must not stamp mesh; hit %s source_hint=%q", h.ID, h.Provenance.SourceHint)
		}
		if h.Provenance.SourceHint != SourceHintPrivate {
			t.Fatalf("hit %s source_hint=%q want private", h.ID, h.Provenance.SourceHint)
		}
		sawPrivate = true
		hay := strings.ToLower(h.Content.Summary + " " + h.Content.Full)
		if strings.Contains(hay, "14 day") {
			sawPolicy = true
		}
	}
	if !sawPrivate {
		t.Fatal("expected source_hint=private on retrieve hits")
	}
	if !sawPolicy {
		t.Fatalf("expected 14-day refund policy in same-session retrieve, got %d hits", len(hits))
	}

	facts := store.ListFactsAsOf(FactsAsOfOptions{
		AsOf:      ticketAt,
		SessionID: session,
		Query:     "refund",
		Limit:     10,
	})
	if len(facts) == 0 {
		t.Fatal("facts-as-of ticket created time empty")
	}
	found14 := false
	for _, f := range facts {
		if f.Provenance.SourceHint == "mesh" {
			t.Fatalf("fact %s stamped mesh", f.ID)
		}
		if f.Provenance.SourceHint != SourceHintPrivate {
			t.Fatalf("fact %s source_hint=%q want private", f.ID, f.Provenance.SourceHint)
		}
		hay := strings.ToLower(f.Content.Summary + " " + f.Content.Full + " " + f.OriginalText)
		if strings.Contains(hay, "14 day") {
			found14 = true
		}
	}
	if !found14 {
		t.Fatal("facts-as-of ticket created time should find the 14-day unused-seat rule")
	}
}
