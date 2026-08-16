package memory

import (
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
