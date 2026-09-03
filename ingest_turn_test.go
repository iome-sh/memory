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

	wantChild := []string{"source:iomesh-memory-mcp", "role:user", "fact_augmented", "from_turn"}
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

	bySource := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "source:iomesh-memory-mcp", Limit: 50})
	if len(bySource) != 1+len(facts) {
		t.Fatalf("tag=source:iomesh-memory-mcp got %d, want parent+children %d", len(bySource), 1+len(facts))
	}
	byPrefix := store.ListMemoryWithOptions(ListMemoryOptions{TagPrefix: "source:", Limit: 50})
	if len(byPrefix) != 1+len(facts) {
		t.Fatalf("tag_prefix=source: got %d, want parent+children %d", len(byPrefix), 1+len(facts))
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
}
