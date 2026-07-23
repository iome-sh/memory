package memory

import (
	"fmt"
	"testing"
	"time"
)

func TestListMemoryWithOptions_SessionFilter(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{ID: "s1a", Tier: TierContextual, SessionID: "sess-A", Timestamp: base, Content: MemoryContent{Summary: "notes A1"}},
		{ID: "s1b", Tier: TierContextual, SessionID: "sess-A", Timestamp: base.Add(time.Hour), Content: MemoryContent{Summary: "notes A2"}},
		{ID: "s2a", Tier: TierContextual, SessionID: "sess-B", Timestamp: base.Add(2 * time.Hour), Content: MemoryContent{Summary: "notes B1"}},
		{ID: "s0", Tier: TierContextual, SessionID: "", Timestamp: base.Add(3 * time.Hour), Content: MemoryContent{Summary: "notes none"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	results := store.ListMemoryWithOptions(ListMemoryOptions{SessionID: "sess-A"})
	if len(results) != 2 {
		t.Fatalf("session filter len = %d, want 2; got %v", len(results), idsOf(results))
	}
	for _, r := range results {
		if r.SessionID != "sess-A" {
			t.Fatalf("got SessionID %q, want sess-A", r.SessionID)
		}
	}
}

func TestListMemoryWithOptions_TimeWindowUnderfill(t *testing.T) {
	// Underfill fix: many entries outside the window + few inside; Limit must still
	// return all in-window (or up to limit of in-window), not first N of unfiltered scan.
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	from := base
	to := base.Add(2 * time.Hour)

	// 30 outside (older), then 3 inside window, then 20 outside (newer).
	// ListEntriesInTier sorts by relevance (default metrics ≈ equal), so FS order can
	// surface outside entries first if Limit were applied before the time filter.
	for i := 0; i < 30; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("old-%02d", i),
			Tier:      TierContextual,
			Timestamp: base.Add(-time.Duration(i+1) * time.Hour),
			Content:   MemoryContent{Summary: fmt.Sprintf("old event %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	inWindowIDs := []string{"win-0", "win-1", "win-2"}
	for i, id := range inWindowIDs {
		e := MemoryEntry{
			ID:        id,
			Tier:      TierContextual,
			Timestamp: base.Add(time.Duration(i) * 30 * time.Minute),
			Content:   MemoryContent{Summary: fmt.Sprintf("window event %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 20; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("new-%02d", i),
			Tier:      TierContextual,
			Timestamp: base.Add(3*time.Hour + time.Duration(i)*time.Hour),
			Content:   MemoryContent{Summary: fmt.Sprintf("new event %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Limit=5 > 3 in-window: must return all 3 in-window, not 5 from unfiltered set.
	results := store.ListMemoryWithOptions(ListMemoryOptions{
		TimeFrom: &from,
		TimeTo:   &to,
		Limit:    5,
	})
	got := idsOf(results)
	if len(results) != 3 {
		t.Fatalf("time window underfill len = %d, want 3; got %v", len(results), got)
	}
	want := map[string]bool{"win-0": true, "win-1": true, "win-2": true}
	for _, id := range got {
		if !want[id] {
			t.Fatalf("unexpected id %q in results %v", id, got)
		}
	}

	// Limit=2 of in-window: return 2 newest-first within window.
	results2 := store.ListMemoryWithOptions(ListMemoryOptions{
		TimeFrom: &from,
		TimeTo:   &to,
		Limit:    2,
	})
	if len(results2) != 2 {
		t.Fatalf("limited window len = %d, want 2; got %v", len(results2), idsOf(results2))
	}
	for _, r := range results2 {
		if !want[r.ID] {
			t.Fatalf("limited result %q not in window", r.ID)
		}
	}
	// Newest first: win-2 then win-1
	if results2[0].ID != "win-2" || results2[1].ID != "win-1" {
		t.Fatalf("newest-first within window: got %v, want [win-2 win-1]", idsOf(results2))
	}
}

func TestListMemoryWithOptions_Ascending(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{ID: "mid", Tier: TierContextual, Timestamp: base.Add(time.Hour), Content: MemoryContent{Summary: "mid"}},
		{ID: "old", Tier: TierContextual, Timestamp: base, Content: MemoryContent{Summary: "old"}},
		{ID: "new", Tier: TierContextual, Timestamp: base.Add(2 * time.Hour), Content: MemoryContent{Summary: "new"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	newestFirst := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})
	if !sameIDsInOrder(newestFirst, []MemoryEntry{{ID: "new"}, {ID: "mid"}, {ID: "old"}}) {
		t.Fatalf("default newest-first: got %v, want [new mid old]", idsOf(newestFirst))
	}

	oldestFirst := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10, Ascending: true})
	if !sameIDsInOrder(oldestFirst, []MemoryEntry{{ID: "old"}, {ID: "mid"}, {ID: "new"}}) {
		t.Fatalf("Ascending: got %v, want [old mid new]", idsOf(oldestFirst))
	}
}

func TestListMemoryWithOptions_TagAndTagPrefix(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "t1", Tier: TierContextual, Timestamp: base,
			TemporalTags: []string{"cycle:1", "subject:auth"},
			Content:      MemoryContent{Summary: "temporal tag entry"},
		},
		{
			ID: "t2", Tier: TierContextual, Timestamp: base.Add(time.Hour),
			Content: MemoryContent{Summary: "content tag entry", Tags: []string{"subject:billing", "priority:high"}},
		},
		{
			ID: "t3", Tier: TierContextual, Timestamp: base.Add(2 * time.Hour),
			TemporalTags: []string{"session_seq:3"},
			Content:      MemoryContent{Summary: "other", Tags: []string{"misc"}},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Exact tag on TemporalTags
	byExact := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "subject:auth"})
	if len(byExact) != 1 || byExact[0].ID != "t1" {
		t.Fatalf("Tag TemporalTags: got %v, want [t1]", idsOf(byExact))
	}
	// Exact tag on Content.Tags
	byExact2 := store.ListMemoryWithOptions(ListMemoryOptions{Tag: "priority:high"})
	if len(byExact2) != 1 || byExact2[0].ID != "t2" {
		t.Fatalf("Tag Content.Tags: got %v, want [t2]", idsOf(byExact2))
	}

	// Prefix across both fields
	byPrefix := store.ListMemoryWithOptions(ListMemoryOptions{TagPrefix: "subject:"})
	got := map[string]bool{}
	for _, r := range byPrefix {
		got[r.ID] = true
	}
	if len(byPrefix) != 2 || !got["t1"] || !got["t2"] {
		t.Fatalf("TagPrefix subject:: got %v, want t1+t2", idsOf(byPrefix))
	}

	bySeq := store.ListMemoryWithOptions(ListMemoryOptions{TagPrefix: "session_seq:"})
	if len(bySeq) != 1 || bySeq[0].ID != "t3" {
		t.Fatalf("TagPrefix session_seq:: got %v, want [t3]", idsOf(bySeq))
	}

	// Helpers
	if !EntryHasTag(entries[0], "subject:auth") {
		t.Fatal("EntryHasTag expected true for temporal tag")
	}
	if !EntryHasTagPrefix(entries[1], "subject:") {
		t.Fatal("EntryHasTagPrefix expected true for content tag")
	}
	if EntryHasTag(entries[0], "missing") {
		t.Fatal("EntryHasTag expected false")
	}
}

func TestListMemoryWithOptions_QuerySubstring(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{ID: "sum", Tier: TierContextual, Timestamp: base, Content: MemoryContent{Summary: "Alpha Project Plan"}},
		{ID: "full", Tier: TierContextual, Timestamp: base.Add(time.Hour), Content: MemoryContent{Summary: "other", Full: "details about BETA release"}},
		{ID: "orig", Tier: TierContextual, Timestamp: base.Add(2 * time.Hour), OriginalText: "raw turn GAMMA notes", Content: MemoryContent{Summary: "turn"}},
		{ID: "none", Tier: TierContextual, Timestamp: base.Add(3 * time.Hour), Content: MemoryContent{Summary: "unrelated"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Case-insensitive on Summary
	r1 := store.ListMemoryWithOptions(ListMemoryOptions{Query: "alpha project"})
	if len(r1) != 1 || r1[0].ID != "sum" {
		t.Fatalf("Query summary: got %v, want [sum]", idsOf(r1))
	}
	// Full
	r2 := store.ListMemoryWithOptions(ListMemoryOptions{Query: "beta"})
	if len(r2) != 1 || r2[0].ID != "full" {
		t.Fatalf("Query full: got %v, want [full]", idsOf(r2))
	}
	// OriginalText
	r3 := store.ListMemoryWithOptions(ListMemoryOptions{Query: "gamma"})
	if len(r3) != 1 || r3[0].ID != "orig" {
		t.Fatalf("Query original: got %v, want [orig]", idsOf(r3))
	}
}

func TestListMemoryWithOptions_DefaultTiersExcludeArchival(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{ID: "w", Tier: TierWorking, Timestamp: base, Content: MemoryContent{Summary: "working"}},
		{ID: "c", Tier: TierContextual, Timestamp: base.Add(time.Hour), Content: MemoryContent{Summary: "contextual"}},
		{ID: "s", Tier: TierSemantic, Timestamp: base.Add(2 * time.Hour), Content: MemoryContent{Summary: "semantic"}},
		{ID: "a", Tier: TierArchival, Timestamp: base.Add(3 * time.Hour), Content: MemoryContent{Summary: "archival"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	def := store.ListMemoryWithOptions(ListMemoryOptions{})
	got := idsOf(def)
	for _, id := range got {
		if id == "a" {
			t.Fatalf("default tiers included archival: %v", got)
		}
	}
	if len(def) != 3 {
		t.Fatalf("default tiers len = %d, want 3 (W+C+S); got %v", len(def), got)
	}

	withArch := store.ListMemoryWithOptions(ListMemoryOptions{IncludeArchival: true})
	if len(withArch) != 4 {
		t.Fatalf("IncludeArchival len = %d, want 4; got %v", len(withArch), idsOf(withArch))
	}

	// Explicit Tier overrides default set
	arch := TierArchival
	onlyArch := store.ListMemoryWithOptions(ListMemoryOptions{Tier: &arch})
	if len(onlyArch) != 1 || onlyArch[0].ID != "a" {
		t.Fatalf("Tier=Archival: got %v, want [a]", idsOf(onlyArch))
	}
}

func TestListMemoryWithOptions_LimitDefault50(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	// Write 60 entries; default Limit should be 50.
	for i := 0; i < 60; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("e-%03d", i),
			Tier:      TierContextual,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Content:   MemoryContent{Summary: fmt.Sprintf("entry %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	def := store.ListMemoryWithOptions(ListMemoryOptions{})
	if len(def) != 50 {
		t.Fatalf("default Limit len = %d, want 50", len(def))
	}
	// Newest first: highest timestamps first → e-059 ..
	if def[0].ID != "e-059" {
		t.Fatalf("default newest first: first id = %q, want e-059", def[0].ID)
	}

	zero := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 0})
	if len(zero) != 50 {
		t.Fatalf("Limit=0 default len = %d, want 50", len(zero))
	}

	neg := store.ListMemoryWithOptions(ListMemoryOptions{Limit: -1})
	if len(neg) != 50 {
		t.Fatalf("Limit=-1 default len = %d, want 50", len(neg))
	}

	capped := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 7})
	if len(capped) != 7 {
		t.Fatalf("Limit=7 len = %d, want 7", len(capped))
	}
}
