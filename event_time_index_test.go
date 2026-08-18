package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDurableEventTimeIndex_PersistsAndReloads(t *testing.T) {
	baseDir := t.TempDir()
	seed := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	base := writeListFixture(t, seed)
	_ = seed.ListMemoryWithOptions(ListMemoryOptions{Limit: 100})

	idxPath := filepath.Join(baseDir, "indexes", "event-time.json")
	if _, err := os.Stat(idxPath); err != nil {
		t.Fatalf("expected durable index at %s: %v", idxPath, err)
	}

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	from := base
	to := base.Add(3 * time.Hour)
	got := reopened.ListMemoryWithOptions(ListMemoryOptions{
		SessionID: "sess-A",
		TimeFrom:  &from,
		TimeTo:    &to,
		Limit:     100,
	})
	scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir, DisableMetaIndex: true})
	want := scan.ListMemoryWithOptions(ListMemoryOptions{
		SessionID: "sess-A",
		TimeFrom:  &from,
		TimeTo:    &to,
		Limit:     100,
	})
	if !sameIDsInOrder(got, want) {
		t.Fatalf("reload vs scan: got %v want %v", idsOf(got), idsOf(want))
	}
	if reopened.MetaIndexLen() == 0 {
		t.Fatal("expected meta index to load from durable snapshot")
	}
}

func TestDurableEventTimeIndex_StaleStampRebuilds(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	if err := store.Write(MemoryEntry{
		ID: "first", Tier: TierContextual,
		Timestamp: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Content:   MemoryContent{Summary: "first"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})

	// Second process writes a new entry (stamp must change).
	writer := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	if err := writer.Write(MemoryEntry{
		ID: "second", Tier: TierContextual,
		Timestamp: time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		Content:   MemoryContent{Summary: "second"},
	}); err != nil {
		t.Fatal(err)
	}

	reader := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	got := reader.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if !listHasID(got, "first") || !listHasID(got, "second") {
		t.Fatalf("stale snapshot must rebuild; got %v", idsOf(got))
	}
}

func TestDurableEventTimeIndex_DisabledSkipsFile(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir, DisableDurableIndex: true})
	if err := store.Write(MemoryEntry{
		ID: "x", Tier: TierContextual, Content: MemoryContent{Summary: "only"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})
	idxPath := filepath.Join(baseDir, "indexes", "event-time.json")
	if _, err := os.Stat(idxPath); !os.IsNotExist(err) {
		t.Fatalf("DisableDurableIndex should not write %s (err=%v)", idxPath, err)
	}
}

func TestDurableEventTimeIndex_CorruptFileFallsBack(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	if err := store.Write(MemoryEntry{
		ID: "ok", Tier: TierContextual, Content: MemoryContent{Summary: "ok"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})

	idxPath := filepath.Join(baseDir, "indexes", "event-time.json")
	if err := os.WriteFile(idxPath, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	got := reopened.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})
	if !listHasID(got, "ok") {
		t.Fatalf("corrupt snapshot must fall back to FS rebuild; got %v", idsOf(got))
	}
	// Rebuild should have rewritten a valid snapshot.
	raw, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("rewritten snapshot invalid: %v", err)
	}
	if snap.Version != durableEventTimeIndexVersion || snap.JSONCount < 1 {
		t.Fatalf("rewritten snapshot = %+v", snap)
	}
}

func TestDurableEventTimeIndex_IncrementalWritePersists(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	if err := store.Write(MemoryEntry{
		ID: "first", Tier: TierContextual,
		Timestamp: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Content:   MemoryContent{Summary: "first"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if err := store.Write(MemoryEntry{
		ID: "second", Tier: TierContextual,
		Timestamp: time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		Content:   MemoryContent{Summary: "second"},
	}); err != nil {
		t.Fatal(err)
	}
	if store.MetaIndexLen() != 2 {
		t.Fatalf("incremental stamp dirty; len=%d", store.MetaIndexLen())
	}

	raw, err := os.ReadFile(filepath.Join(baseDir, "indexes", "event-time.json"))
	if err != nil {
		t.Fatal(err)
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("durable after incremental write: %v", err)
	}
	if snap.JSONCount != 2 {
		t.Fatalf("durable JSONCount=%d want 2", snap.JSONCount)
	}

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	got := reopened.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if !listHasID(got, "first") || !listHasID(got, "second") {
		t.Fatalf("fresh process should load patched snapshot; got %v", idsOf(got))
	}
	if reopened.MetaIndexRebuilds() != 0 {
		t.Fatalf("fresh process rebuilt (%d); stamp should match incremental persist", reopened.MetaIndexRebuilds())
	}
}

func listHasID(entries []MemoryEntry, id string) bool {
	for _, e := range entries {
		if e.ID == id {
			return true
		}
	}
	return false
}
