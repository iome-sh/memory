package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeCompactionNote(t *testing.T, store *PalaceStore, id string) {
	t.Helper()
	err := store.Write(MemoryEntry{
		ID:        id,
		Type:      "note",
		Tier:      TierContextual,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Content:   MemoryContent{Summary: "sprint note " + id, Full: "sprint planning notes " + id},
		Metrics:   MemoryMetrics{ScoreImpact: 0.5},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPerformCompaction_StampsLastCompaction(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	if !store.GetStats().LastCompaction.IsZero() {
		t.Fatal("expected zero LastCompaction before run")
	}

	writeCompactionNote(t, store, "keep-1")
	before := time.Now().Add(-time.Second)
	store.PerformCompaction(TierContextual, DefaultCompactionConfig, func(string) string {
		return `[{"action":"ARCHIVE","target":["keep-1"],"reason":"test"}]`
	}, nil)

	got := store.GetStats().LastCompaction
	if got.IsZero() {
		t.Fatal("LastCompaction still zero after PerformCompaction")
	}
	if got.Before(before) || got.After(time.Now().Add(time.Second)) {
		t.Fatalf("LastCompaction %v out of range", got)
	}
}

func TestPerformCompaction_EmptyTierDoesNotStamp(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	store.PerformCompaction(TierContextual, DefaultCompactionConfig, func(string) string {
		t.Fatal("generateFn should not run on empty tier")
		return "[]"
	}, nil)
	if !store.GetStats().LastCompaction.IsZero() {
		t.Fatal("empty tier should not stamp LastCompaction")
	}
}

func TestVerifyAction_RejectUnknownAndMissingIDs(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	writeCompactionNote(t, store, "e1")
	writeCompactionNote(t, store, "e2")

	cases := []struct {
		name string
		act  CompactionAction
		ok   bool
	}{
		{name: "unknown action", act: CompactionAction{Action: "DELETE", TargetIDs: []string{"e1"}}},
		{name: "empty action", act: CompactionAction{Action: "", TargetIDs: []string{"e1"}}},
		{name: "missing target ids", act: CompactionAction{Action: "ARCHIVE"}},
		{name: "blank id", act: CompactionAction{Action: "ARCHIVE", TargetIDs: []string{"  "}}},
		{name: "unknown id", act: CompactionAction{Action: "ARCHIVE", TargetIDs: []string{"missing"}}},
		{name: "merge one id", act: CompactionAction{Action: "MERGE", TargetIDs: []string{"e1"}}},
		{name: "archive existing", act: CompactionAction{Action: "ARCHIVE", TargetIDs: []string{"e1"}}, ok: true},
		{name: "lowercase archive", act: CompactionAction{Action: "archive", TargetIDs: []string{"e1"}}, ok: true},
		{name: "merge two existing", act: CompactionAction{Action: "MERGE", TargetIDs: []string{"e1", "e2"}}, ok: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := store.verifyAction(tc.act, TierContextual)
			if got != tc.ok {
				t.Fatalf("verifyAction(%+v) = %v, want %v", tc.act, got, tc.ok)
			}
		})
	}
}

func TestHandleSummarize_StampsValidFrom(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	ts := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	if err := store.Write(MemoryEntry{
		ID:        "p1",
		Type:      "note",
		Tier:      TierContextual,
		Version:   1,
		CreatedAt: ts,
		UpdatedAt: ts,
		SessionID: "sess-compact",
		Timestamp: ts,
		Content:   MemoryContent{Summary: "sprint note p1", Full: "sprint planning notes p1"},
		Metrics:   MemoryMetrics{ScoreImpact: 0.5},
	}); err != nil {
		t.Fatal(err)
	}
	writeCompactionNote(t, store, "p2")

	if err := store.handleSummarize([]string{"p1", "p2"}, TierContextual, DefaultCompactionConfig, nil); err != nil {
		t.Fatal(err)
	}

	var product MemoryEntry
	found := false
	for _, e := range store.ListEntriesInTier(TierContextual) {
		if e.Type == "summary" {
			product = e
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected summary product in contextual tier")
	}
	if product.SessionID != "sess-compact" {
		t.Fatalf("SessionID = %q, want sess-compact", product.SessionID)
	}
	if !product.Timestamp.Equal(ts) {
		t.Fatalf("Timestamp = %v, want %v", product.Timestamp, ts)
	}
	if product.Provenance.SourceStep != "compaction" {
		t.Fatalf("SourceStep = %q", product.Provenance.SourceStep)
	}
	if !hasValidFromTag(product) {
		t.Fatalf("missing valid_from tags=%v", product.TemporalTags)
	}
	from, until := ParseValidityWindow(product)
	if from == nil {
		t.Fatalf("ParseValidityWindow from=nil tags=%v", product.TemporalTags)
	}
	if until != nil {
		t.Fatalf("did not expect valid_until tags=%v", product.TemporalTags)
	}
	now := time.Now().UTC()
	if !EntryValidAt(product, now) {
		t.Fatalf("EntryValidAt(now) false from=%v", from)
	}
	if !EntryValidAt(product, time.Time{}) {
		t.Fatal("EntryValidAt(zero→now) false")
	}
	if _, ok := store.Load("p1", TierContextual); ok {
		t.Fatal("summarized source p1 still in contextual")
	}
	if _, ok := store.Load("p2", TierContextual); ok {
		t.Fatal("summarized source p2 still in contextual")
	}
	if _, ok := store.Load("p1", TierArchival); !ok {
		t.Fatal("expected p1 in archival after summarize")
	}
}

func TestHandleArchive_UnlinksSourceTier(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	writeCompactionNote(t, store, "keep-1")

	if err := store.handleArchive([]string{"keep-1"}, TierContextual, DefaultCompactionConfig); err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Load("keep-1", TierContextual); ok {
		t.Fatal("source JSON still in contextual after ARCHIVE")
	}
	got, ok := store.Load("keep-1", TierArchival)
	if !ok {
		t.Fatal("expected archival copy after ARCHIVE")
	}
	if got.Tier != TierArchival {
		t.Fatalf("Tier = %d, want archival", got.Tier)
	}

	listed := store.ListMemoryWithOptions(ListMemoryOptions{Query: "sprint note keep-1"})
	for _, e := range listed {
		if e.ID == "keep-1" {
			t.Fatal("archived entry still in default ListMemory (archival excluded)")
		}
	}
	withArchival := store.ListMemoryWithOptions(ListMemoryOptions{Query: "sprint note keep-1", IncludeArchival: true})
	found := false
	for _, e := range withArchival {
		if e.ID == "keep-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected archived entry when IncludeArchival is set")
	}
}

func TestHandleArchive_WriteError(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStore(base)
	writeCompactionNote(t, store, "keep-1")
	archDir := filepath.Join(base, "tier-3-archival")
	if err := os.Chmod(archDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(archDir, 0755) })

	err := store.handleArchive([]string{"keep-1"}, TierContextual, DefaultCompactionConfig)
	if err == nil {
		t.Fatal("expected archive Write error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to archive entry") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := store.Load("keep-1", TierContextual); !ok {
		t.Fatal("source should remain in contextual when archive Write fails")
	}
	if _, ok := store.Load("keep-1", TierArchival); ok {
		t.Fatal("did not expect archival file after failed Write")
	}
}

func TestHandleSummarize_WriteError(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStore(base)
	writeCompactionNote(t, store, "p1")
	ctxDir := filepath.Join(base, "tier-2-contextual")
	if err := os.Chmod(ctxDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ctxDir, 0755) })

	err := store.handleSummarize([]string{"p1"}, TierContextual, DefaultCompactionConfig, nil)
	if err == nil {
		t.Fatal("expected summary Write error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to write summary") {
		t.Fatalf("unexpected error: %v", err)
	}
}
