package memory

import (
	"os"
	"path/filepath"
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
