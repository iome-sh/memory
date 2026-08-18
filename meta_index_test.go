package memory

import (
	"fmt"
	"testing"
	"time"
)

// writeListFixture writes a multi-filter fixture shared by parity + invalidation tests.
func writeListFixture(t *testing.T, store *PalaceStore) time.Time {
	t.Helper()
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "a1", Tier: TierContextual, SessionID: "sess-A",
			Timestamp: base, TemporalTags: []string{"subject:auth", "cycle:1"},
			Content: MemoryContent{Summary: "Alpha auth notes"},
		},
		{
			ID: "a2", Tier: TierContextual, SessionID: "sess-A",
			Timestamp: base.Add(time.Hour),
			Content:   MemoryContent{Summary: "other A", Tags: []string{"subject:billing"}},
		},
		{
			ID: "b1", Tier: TierWorking, SessionID: "sess-B",
			Timestamp: base.Add(2 * time.Hour),
			Content:   MemoryContent{Summary: "Beta work item", Full: "details BETA release"},
		},
		{
			ID: "s1", Tier: TierSemantic, SessionID: "sess-A",
			Timestamp: base.Add(3 * time.Hour), OriginalText: "raw GAMMA fact",
			Content: MemoryContent{Summary: "semantic fact"},
		},
		{
			ID: "arch", Tier: TierArchival, SessionID: "sess-A",
			Timestamp: base.Add(4 * time.Hour),
			Content:   MemoryContent{Summary: "archived only"},
		},
		{
			ID: "old", Tier: TierContextual, SessionID: "sess-A",
			Timestamp: base.Add(-24 * time.Hour),
			Content:   MemoryContent{Summary: "too old for window"},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	return base
}

func TestListMemoryMetaIndex_ParityWithScan(t *testing.T) {
	baseDir := t.TempDir()
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	// Populate once via index-enabled store.
	seed := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	_ = writeListFixture(t, seed)

	from := base
	to := base.Add(3 * time.Hour)

	cases := []struct {
		name string
		opts ListMemoryOptions
	}{
		{"default", ListMemoryOptions{Limit: 100}},
		{"session", ListMemoryOptions{SessionID: "sess-A", Limit: 100}},
		{"time_window", ListMemoryOptions{TimeFrom: &from, TimeTo: &to, Limit: 100}},
		{"tag", ListMemoryOptions{Tag: "subject:auth", Limit: 100}},
		{"tag_prefix", ListMemoryOptions{TagPrefix: "subject:", Limit: 100}},
		{"query", ListMemoryOptions{Query: "beta", Limit: 100}},
		{"query_orig", ListMemoryOptions{Query: "gamma", Limit: 100}},
		{"ascending", ListMemoryOptions{Limit: 100, Ascending: true}},
		{"include_archival", ListMemoryOptions{IncludeArchival: true, Limit: 100}},
		{"tier_archival", ListMemoryOptions{Tier: tierPtr(TierArchival), Limit: 100}},
		{"limit_2_newest", ListMemoryOptions{SessionID: "sess-A", Limit: 2}},
		{
			"combo",
			ListMemoryOptions{
				SessionID: "sess-A",
				TimeFrom:  &from,
				TimeTo:    &to,
				TagPrefix: "subject:",
				Limit:     10,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withIdx := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
			scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir, DisableMetaIndex: true})

			gotIdx := withIdx.ListMemoryWithOptions(tc.opts)
			gotScan := scan.ListMemoryWithOptions(tc.opts)

			if !sameIDsInOrder(gotIdx, gotScan) {
				t.Fatalf("index vs scan parity: index=%v scan=%v", idsOf(gotIdx), idsOf(gotScan))
			}
			if withIdx.MetaIndexLen() == 0 {
				t.Fatal("expected meta index to be built after ListMemoryWithOptions")
			}
		})
	}
}

func TestListMemoryMetaIndex_IncrementalWriteNoRebuild(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	e1 := MemoryEntry{
		ID: "first", Tier: TierContextual, SessionID: "s",
		Timestamp: base, Content: MemoryContent{Summary: "first entry"},
	}
	if err := store.Write(e1); err != nil {
		t.Fatal(err)
	}

	r1 := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if len(r1) != 1 || r1[0].ID != "first" {
		t.Fatalf("before second write: got %v, want [first]", idsOf(r1))
	}
	if n := store.MetaIndexLen(); n != 1 {
		t.Fatalf("MetaIndexLen after list = %d, want 1", n)
	}
	rebuilds := store.MetaIndexRebuilds()
	if rebuilds == 0 {
		t.Fatal("first ListMemoryWithOptions should full-rebuild an unbuilt index")
	}

	// Write patches in place; List must see the new entry without a rebuild walk.
	e2 := MemoryEntry{
		ID: "second", Tier: TierContextual, SessionID: "s",
		Timestamp: base.Add(time.Hour), Content: MemoryContent{Summary: "second entry"},
	}
	if err := store.Write(e2); err != nil {
		t.Fatal(err)
	}
	if n := store.MetaIndexLen(); n != 2 {
		t.Fatalf("MetaIndexLen after Write should stay clean (len 2), got %d", n)
	}
	if got := store.MetaIndexRebuilds(); got != rebuilds {
		t.Fatalf("Write must not full-rebuild; rebuilds %d → %d", rebuilds, got)
	}

	r2 := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if len(r2) != 2 {
		t.Fatalf("after second write len = %d, want 2; got %v", len(r2), idsOf(r2))
	}
	if r2[0].ID != "second" || r2[1].ID != "first" {
		t.Fatalf("newest-first after incremental write: got %v, want [second first]", idsOf(r2))
	}
	if got := store.MetaIndexRebuilds(); got != rebuilds {
		t.Fatalf("List after incremental Write must not rebuild; rebuilds %d → %d", rebuilds, got)
	}
	if n := store.MetaIndexLen(); n != 2 {
		t.Fatalf("MetaIndexLen after list = %d, want 2", n)
	}
}

func TestListMemoryMetaIndex_InvalidateHook(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	if err := store.Write(MemoryEntry{
		ID: "x", Tier: TierContextual, Timestamp: base,
		Content: MemoryContent{Summary: "x"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})
	if store.MetaIndexLen() != 1 {
		t.Fatal("expected built index")
	}
	store.InvalidateMetaIndex()
	if store.MetaIndexLen() != 0 {
		t.Fatal("expected dirty after InvalidateMetaIndex")
	}
	// Rebuild still correct
	got := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})
	if len(got) != 1 || got[0].ID != "x" {
		t.Fatalf("after force invalidate: got %v", idsOf(got))
	}
}

func TestListMemoryMetaIndex_WriteLatentDoesNotListOrDirty(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	if err := store.Write(MemoryEntry{
		ID: "visible", Tier: TierContextual, Timestamp: base,
		Content: MemoryContent{Summary: "visible"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})
	if store.MetaIndexLen() != 1 {
		t.Fatal("expected index len 1")
	}
	rebuilds := store.MetaIndexRebuilds()
	if err := store.WriteLatent(MemoryEntry{
		ID: "latent", Tier: TierWorking, Timestamp: base.Add(time.Hour),
		Content: MemoryContent{Summary: "latent only"},
	}); err != nil {
		t.Fatal(err)
	}
	if store.MetaIndexLen() != 1 {
		t.Fatal("WriteLatent must not dirty the listed-tier meta index")
	}
	if got := store.MetaIndexRebuilds(); got != rebuilds {
		t.Fatalf("WriteLatent must not rebuild; rebuilds %d → %d", rebuilds, got)
	}
	got := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	// Latent is not in tier dirs used by ListMemory.
	if len(got) != 1 || got[0].ID != "visible" {
		t.Fatalf("latent must not appear in ListMemory: got %v", idsOf(got))
	}
	if store.MetaIndexRebuilds() != rebuilds {
		t.Fatal("List after WriteLatent must not rebuild")
	}
}

func TestListMemoryMetaIndex_IncrementalOverwrite(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := store.Write(MemoryEntry{
		ID: "same", Tier: TierContextual, Timestamp: base,
		Content: MemoryContent{Summary: "alpha notes"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 5})
	rebuilds := store.MetaIndexRebuilds()

	if err := store.Write(MemoryEntry{
		ID: "same", Tier: TierContextual, Timestamp: base.Add(time.Hour),
		Content: MemoryContent{Summary: "beta notes", Tags: []string{"subject:billing"}},
	}); err != nil {
		t.Fatal(err)
	}
	if store.MetaIndexLen() != 1 {
		t.Fatalf("overwrite should replace in place, len=%d", store.MetaIndexLen())
	}
	if store.MetaIndexRebuilds() != rebuilds {
		t.Fatal("overwrite must not full-rebuild")
	}

	got := store.ListMemoryWithOptions(ListMemoryOptions{Query: "beta", Limit: 5})
	if len(got) != 1 || got[0].ID != "same" {
		t.Fatalf("query after overwrite: got %v", idsOf(got))
	}
	tagged := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "subject:billing", Limit: 5})
	if len(tagged) != 1 || tagged[0].ID != "same" {
		t.Fatalf("tag after overwrite: got %v", idsOf(tagged))
	}
	if store.MetaIndexRebuilds() != rebuilds {
		t.Fatal("List after overwrite must not rebuild")
	}
}

func TestListMemoryMetaIndex_IncrementalUnlink(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	for _, e := range []MemoryEntry{
		{ID: "keep", Tier: TierContextual, Timestamp: base, Content: MemoryContent{Summary: "keep me"}},
		{ID: "drop", Tier: TierContextual, Timestamp: base.Add(time.Hour), Content: MemoryContent{Summary: "drop me"}},
	} {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if store.MetaIndexLen() != 2 {
		t.Fatalf("warm len=%d want 2", store.MetaIndexLen())
	}
	rebuilds := store.MetaIndexRebuilds()

	if err := store.unlinkEntry("drop", TierContextual); err != nil {
		t.Fatal(err)
	}
	if n := store.MetaIndexLen(); n != 1 {
		t.Fatalf("MetaIndexLen after unlink = %d, want 1", n)
	}
	if store.MetaIndexRebuilds() != rebuilds {
		t.Fatal("unlink must not full-rebuild")
	}

	got := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if len(got) != 1 || got[0].ID != "keep" {
		t.Fatalf("after unlink: got %v, want [keep]", idsOf(got))
	}
	if store.MetaIndexRebuilds() != rebuilds {
		t.Fatal("List after unlink must not rebuild")
	}

	scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: store.BaseDir, DisableMetaIndex: true})
	want := scan.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if !sameIDsInOrder(got, want) {
		t.Fatalf("unlink parity: index=%v scan=%v", idsOf(got), idsOf(want))
	}
}

func TestListMemoryMetaIndex_LargeNUnderfill(t *testing.T) {
	// Micro-style: many out-of-window + few in-window; Limit after filter.
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	from := base
	to := base.Add(time.Hour)

	for i := 0; i < 200; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("out-%03d", i),
			Tier:      TierContextual,
			Timestamp: base.Add(-time.Duration(i+1) * time.Hour),
			Content:   MemoryContent{Summary: fmt.Sprintf("outside %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 5; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("in-%d", i),
			Tier:      TierContextual,
			Timestamp: base.Add(time.Duration(i) * 10 * time.Minute),
			Content:   MemoryContent{Summary: fmt.Sprintf("inside %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Warm index
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 1})
	if n := store.MetaIndexLen(); n != 205 {
		t.Fatalf("MetaIndexLen = %d, want 205", n)
	}

	// Second call uses warm index: underfill still holds.
	results := store.ListMemoryWithOptions(ListMemoryOptions{
		TimeFrom: &from,
		TimeTo:   &to,
		Limit:    10,
	})
	if len(results) != 5 {
		t.Fatalf("underfill with meta index: len=%d want 5; got %v", len(results), idsOf(results))
	}
	// Newest first within window: in-4 .. in-0
	if results[0].ID != "in-4" {
		t.Fatalf("newest in-window first: got %v", idsOf(results))
	}

	// Parity with scan
	scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: store.BaseDir, DisableMetaIndex: true})
	scanRes := scan.ListMemoryWithOptions(ListMemoryOptions{
		TimeFrom: &from,
		TimeTo:   &to,
		Limit:    10,
	})
	if !sameIDsInOrder(results, scanRes) {
		t.Fatalf("large-N parity: index=%v scan=%v", idsOf(results), idsOf(scanRes))
	}
}

func tierPtr(t MemoryTier) *MemoryTier {
	return &t
}
