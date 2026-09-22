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

func TestDurableEventTimeIndex_V2StoresValidityAndEntityKeys(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if err := store.Write(MemoryEntry{
		ID: "f", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{
			"entity:person:alice",
			"subject:auth",
			"valid_from:" + from.Format(time.RFC3339),
			"valid_until:" + until.Format(time.RFC3339),
		},
		Content: MemoryContent{
			Summary: "hello",
			Tags:    []string{"entity:org:acme", "valid_until:2099-01-01T00:00:00Z"},
		},
		Relations: MemoryRelations{RelatedConcepts: []string{"Project:Widget"}},
	}); err != nil {
		t.Fatal(err)
	}
	// First list rebuilds from tier JSON (writes landed while the index was dirty).
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})

	raw, err := os.ReadFile(filepath.Join(baseDir, "indexes", "event-time.json"))
	if err != nil {
		t.Fatal(err)
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Version != 2 {
		t.Fatalf("version = %d, want 2", snap.Version)
	}
	var row durableEntryMeta
	found := false
	for _, e := range snap.Entries {
		if e.ID == "f" {
			row = e
			found = true
		}
	}
	if !found {
		t.Fatalf("missing f in %+v", snap.Entries)
	}
	if !row.HasValidity || row.ValidFrom == nil || !row.ValidFrom.Equal(from) {
		t.Fatalf("valid_from = %v has=%v", row.ValidFrom, row.HasValidity)
	}
	if row.ValidUntil == nil || !row.ValidUntil.Equal(until) {
		t.Fatalf("valid_until = %v, want %v (content tag must not win)", row.ValidUntil, until)
	}
	if len(row.EntityTags) != 1 || row.EntityTags[0] != "entity:person:alice" {
		t.Fatalf("entity_tags = %v", row.EntityTags)
	}
	if !row.EntityKeysKnown {
		t.Fatal("entity_keys_known false after rebuild")
	}
	for _, k := range []string{"person:alice", "entity:person:alice", "subject:auth", "org:acme", "entity:org:acme", "project:widget"} {
		if !idsInclude(row.EntityKeys, k) {
			t.Fatalf("entity_keys %v missing %s", row.EntityKeys, k)
		}
	}
}

func TestDurableEventTimeIndex_V1RebuildsNotMisread(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	closedUntil := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	vf := "valid_from:" + from.Format(time.RFC3339)
	if err := store.Write(MemoryEntry{
		ID: "closed", Tier: TierContextual, Timestamp: asOf.Add(-2 * time.Hour),
		TemporalTags: []string{vf, "valid_until:" + closedUntil.Format(time.RFC3339), "entity:person:alice"},
		Content:      MemoryContent{Summary: "old"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(MemoryEntry{
		ID: "open", Tier: TierContextual, Timestamp: asOf.Add(-time.Hour),
		TemporalTags: []string{vf, "entity:person:alice"},
		Content:      MemoryContent{Summary: "current"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 10})

	idxPath := filepath.Join(baseDir, "indexes", "event-time.json")
	raw, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Version != durableEventTimeIndexVersion {
		t.Fatalf("seed version %d", snap.Version)
	}
	// Keep the tier stamp. Drop v2 fields and claim v1 so a naive decoder
	// would treat the closed fact as known-by event time.
	snap.Version = 1
	for i := range snap.Entries {
		snap.Entries[i].HasValidity = false
		snap.Entries[i].ValidFrom = nil
		snap.Entries[i].ValidUntil = nil
		snap.Entries[i].EntityTags = nil
		snap.Entries[i].EntityKeys = nil
		snap.Entries[i].EntityKeysKnown = false
	}
	downgraded, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(idxPath, downgraded, 0o600); err != nil {
		t.Fatal(err)
	}

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	got := reopened.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 10})
	if ids := idsOf(got); !stringSliceEq(ids, []string{"open"}) {
		t.Fatalf("v1 snapshot misread: %v", ids)
	}
	if reopened.MetaIndexRebuilds() != 1 {
		t.Fatalf("v1 file must rebuild once, rebuilds=%d", reopened.MetaIndexRebuilds())
	}
	raw, err = os.ReadFile(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	var rewritten durableEventTimeIndex
	if err := json.Unmarshal(raw, &rewritten); err != nil {
		t.Fatal(err)
	}
	if rewritten.Version != durableEventTimeIndexVersion {
		t.Fatalf("rewritten version %d", rewritten.Version)
	}
	var closed durableEntryMeta
	for _, e := range rewritten.Entries {
		if e.ID == "closed" {
			closed = e
		}
	}
	if !closed.HasValidity || closed.ValidUntil == nil || !closed.ValidUntil.Equal(closedUntil) {
		t.Fatalf("rebuilt closed meta = %+v", closed)
	}
	if len(closed.EntityTags) != 1 || closed.EntityTags[0] != "entity:person:alice" {
		t.Fatalf("rebuilt entity_tags = %v", closed.EntityTags)
	}
}

func TestDurableEventTimeIndex_WritePatchesValidity(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if err := store.Write(MemoryEntry{
		ID: "alice", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{"entity:person:alice", "valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "open"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(MemoryEntry{
		ID: "noise", Tier: TierContextual, Timestamp: from,
		Content: MemoryContent{Summary: "noise"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	rebuilds := store.MetaIndexRebuilds()
	if rebuilds == 0 {
		t.Fatal("expected initial rebuild")
	}
	n, err := store.SupersedeEntityFacts("person:alice", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("supersede count = %d, want 1", n)
	}
	if got := store.MetaIndexRebuilds(); got != rebuilds {
		t.Fatalf("validity patch rebuilt index: %d → %d", rebuilds, got)
	}

	raw, err := os.ReadFile(filepath.Join(baseDir, "indexes", "event-time.json"))
	if err != nil {
		t.Fatal(err)
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	var row durableEntryMeta
	for _, e := range snap.Entries {
		if e.ID == "alice" {
			row = e
		}
	}
	if !row.HasValidity || row.ValidUntil == nil || !row.ValidUntil.Equal(asOf) {
		t.Fatalf("patched durable validity = %+v", row)
	}
	if !row.EntityKeysKnown || !idsInclude(row.EntityKeys, "person:alice") {
		t.Fatalf("patched entity keys = %v known=%v", row.EntityKeys, row.EntityKeysKnown)
	}

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	reads := trackEntryJSONReads(reopened)
	got := reopened.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 10})
	if len(got) != 0 {
		t.Fatalf("closed fact still listed: %v", idsOf(got))
	}
	if reopened.MetaIndexRebuilds() != 0 {
		t.Fatalf("reopen rebuilt %d; patched snapshot stamp should match", reopened.MetaIndexRebuilds())
	}
	if idsInclude(*reads, "alice") || idsInclude(*reads, "noise") {
		t.Fatalf("closed/noise bodies read from patched snapshot: %v", *reads)
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
