package memory

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParseValidityWindow(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	e := MemoryEntry{
		TemporalTags: []string{
			"valid_from:" + from.Format(time.RFC3339),
			"valid_until:" + until.Format(time.RFC3339),
			"entity:person:alice",
		},
	}
	gotFrom, gotUntil := ParseValidityWindow(e)
	if gotFrom == nil || !gotFrom.Equal(from) {
		t.Fatalf("from = %v, want %v", gotFrom, from)
	}
	if gotUntil == nil || !gotUntil.Equal(until) {
		t.Fatalf("until = %v, want %v", gotUntil, until)
	}

	// Missing bounds → open-ended nil
	open := MemoryEntry{TemporalTags: []string{"entity:x"}}
	f, u := ParseValidityWindow(open)
	if f != nil || u != nil {
		t.Fatalf("open-ended: got from=%v until=%v, want nil,nil", f, u)
	}
}

func TestEntryValidAt_NoValidityTags_KnownByAsOf(t *testing.T) {
	asOf := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	// Event before asOf → valid (known by asOf)
	before := MemoryEntry{
		ID: "before", Timestamp: asOf.Add(-time.Hour),
		Content: MemoryContent{Summary: "past"},
	}
	if !EntryValidAt(before, asOf) {
		t.Fatal("event before asOf should be valid (known by asOf)")
	}

	// Event after asOf → invalid (not yet known)
	after := MemoryEntry{
		ID: "after", Timestamp: asOf.Add(time.Hour),
		Content: MemoryContent{Summary: "future"},
	}
	if EntryValidAt(after, asOf) {
		t.Fatal("event after asOf should be invalid without validity tags")
	}

	// Zero event time → valid
	zero := MemoryEntry{ID: "zero", Content: MemoryContent{Summary: "no clock"}}
	if !EntryValidAt(zero, asOf) {
		t.Fatal("zero event time should be valid")
	}

	// Event equal to asOf → valid (!After)
	eq := MemoryEntry{
		ID: "eq", Timestamp: asOf,
		Content: MemoryContent{Summary: "exact"},
	}
	if !EntryValidAt(eq, asOf) {
		t.Fatal("event equal asOf should be valid")
	}
}

func TestEntryValidAt_ValidFrom(t *testing.T) {
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	e := MemoryEntry{
		ID: "vf",
		// Event far in past — validity tags govern, not known-by fallback.
		Timestamp:    from.Add(-365 * 24 * time.Hour),
		TemporalTags: []string{"valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "starts April"},
	}

	if EntryValidAt(e, from.Add(-time.Second)) {
		t.Fatal("asOf before valid_from should be invalid")
	}
	if !EntryValidAt(e, from) {
		t.Fatal("asOf == valid_from should be valid (inclusive start)")
	}
	if !EntryValidAt(e, from.Add(time.Hour)) {
		t.Fatal("asOf after valid_from (open until) should be valid")
	}
}

func TestEntryValidAt_ValidUntilExclusive(t *testing.T) {
	// Documented rule: valid_until is exclusive end — asOf in [from, until).
	until := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := MemoryEntry{
		ID: "vu",
		TemporalTags: []string{
			"valid_from:" + from.Format(time.RFC3339),
			"valid_until:" + until.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "half year fact"},
	}

	if !EntryValidAt(e, from) {
		t.Fatal("asOf == valid_from should be valid")
	}
	if !EntryValidAt(e, until.Add(-time.Second)) {
		t.Fatal("asOf just before valid_until should be valid")
	}
	if EntryValidAt(e, until) {
		t.Fatal("asOf == valid_until should be invalid (exclusive end)")
	}
	if EntryValidAt(e, until.Add(time.Hour)) {
		t.Fatal("asOf after valid_until should be invalid")
	}
}

func TestEntryValidAt_ZeroAsOfUsesNow(t *testing.T) {
	// Open-ended valid_from in the past → valid at Now.
	from := time.Now().UTC().Add(-24 * time.Hour)
	e := MemoryEntry{
		ID:           "now",
		TemporalTags: []string{"valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "current fact"},
	}
	if !EntryValidAt(e, time.Time{}) {
		t.Fatal("zero asOf should use Now; past valid_from open until should be valid")
	}

	// Future valid_from only → invalid at Now.
	future := MemoryEntry{
		ID:           "future",
		TemporalTags: []string{"valid_from:" + time.Now().UTC().Add(48*time.Hour).Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "not yet"},
	}
	if EntryValidAt(future, time.Time{}) {
		t.Fatal("future valid_from should be invalid at Now")
	}
}

func TestListFactsAsOf_Underfill(t *testing.T) {
	// Many invalid at asOf + few valid; Limit must return only valid (underfill class).
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	asOf := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	fromOK := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	untilOK := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	// Invalid: expired before asOf
	untilExpired := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	// Invalid: not yet valid
	fromFuture := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 40; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("expired-%02d", i),
			Tier:      TierContextual,
			Timestamp: asOf.Add(-time.Duration(i+1) * time.Hour),
			TemporalTags: []string{
				"valid_from:" + fromOK.Format(time.RFC3339),
				"valid_until:" + untilExpired.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: fmt.Sprintf("expired fact %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 20; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("future-%02d", i),
			Tier:      TierContextual,
			Timestamp: asOf.Add(-time.Duration(i+1) * time.Hour),
			TemporalTags: []string{
				"valid_from:" + fromFuture.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: fmt.Sprintf("future fact %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	validIDs := []string{"ok-0", "ok-1", "ok-2"}
	for i, id := range validIDs {
		e := MemoryEntry{
			ID:        id,
			Tier:      TierContextual,
			Timestamp: asOf.Add(-time.Duration(i) * time.Hour),
			TemporalTags: []string{
				"valid_from:" + fromOK.Format(time.RFC3339),
				"valid_until:" + untilOK.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: fmt.Sprintf("valid fact %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Limit=10 > 3 valid: must return all 3 valid, not pad with invalid.
	results := store.ListFactsAsOf(FactsAsOfOptions{
		AsOf:  asOf,
		Limit: 10,
	})
	got := idsOf(results)
	if len(results) != 3 {
		t.Fatalf("underfill len = %d, want 3; got %v", len(results), got)
	}
	want := map[string]bool{"ok-0": true, "ok-1": true, "ok-2": true}
	for _, id := range got {
		if !want[id] {
			t.Fatalf("unexpected id %q in results %v", id, got)
		}
	}

	// Limit=2 of valid: return 2 (newest event time first among same non-Semantic rank).
	results2 := store.ListFactsAsOf(FactsAsOfOptions{
		AsOf:  asOf,
		Limit: 2,
	})
	if len(results2) != 2 {
		t.Fatalf("limit=2 len = %d, want 2; got %v", len(results2), idsOf(results2))
	}
	for _, r := range results2 {
		if !want[r.ID] {
			t.Fatalf("limit=2 unexpected %q", r.ID)
		}
	}
}

func TestListFactsAsOf_EntityFilter(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	asOf := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vf := "valid_from:" + from.Format(time.RFC3339)

	entries := []MemoryEntry{
		{
			ID: "alice", Tier: TierSemantic, Timestamp: asOf.Add(-time.Hour),
			TemporalTags: []string{vf, "entity:person:alice"},
			Content:      MemoryContent{Summary: "Alice works at Acme"},
		},
		{
			ID: "bob", Tier: TierSemantic, Timestamp: asOf.Add(-2 * time.Hour),
			TemporalTags: []string{vf, "entity:person:bob"},
			Content:      MemoryContent{Summary: "Bob works at Beta"},
		},
		{
			ID: "acme", Tier: TierContextual, Timestamp: asOf.Add(-3 * time.Hour),
			TemporalTags: []string{vf, "entity:org:acme"},
			Content:      MemoryContent{Summary: "Acme is a company"},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Substring match within entity: tags
	byAlice := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "alice"})
	if len(byAlice) != 1 || byAlice[0].ID != "alice" {
		t.Fatalf("Entity alice: got %v", idsOf(byAlice))
	}

	// Exact entity:type:id when Entity contains ':'
	byExact := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person:bob"})
	if len(byExact) != 1 || byExact[0].ID != "bob" {
		t.Fatalf("Entity person:bob: got %v", idsOf(byExact))
	}

	// Full tag form
	byFull := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "entity:org:acme"})
	if len(byFull) != 1 || byFull[0].ID != "acme" {
		t.Fatalf("Entity entity:org:acme: got %v", idsOf(byFull))
	}

	// person: substring would match person:alice and person:bob if we used contains
	// without requiring ':', but "person" has no ':' so contains on entity: tags.
	byPerson := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person"})
	if len(byPerson) != 2 {
		t.Fatalf("Entity person: got %v, want 2", idsOf(byPerson))
	}
}

func TestListFactsAsOf_SemanticFirstOrdering(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	asOf := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	// Untagged: known-by-asOf via event time
	entries := []MemoryEntry{
		{ID: "ctx-new", Tier: TierContextual, Timestamp: asOf.Add(-time.Minute), Content: MemoryContent{Summary: "ctx new"}},
		{ID: "sem-old", Tier: TierSemantic, Timestamp: asOf.Add(-48 * time.Hour), Content: MemoryContent{Summary: "sem old"}},
		{ID: "wrk", Tier: TierWorking, Timestamp: asOf.Add(-time.Hour), Content: MemoryContent{Summary: "working"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	results := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Limit: 10})
	if len(results) != 3 {
		t.Fatalf("len = %d, want 3; got %v", len(results), idsOf(results))
	}
	if results[0].ID != "sem-old" {
		t.Fatalf("Semantic should sort first, got %v", idsOf(results))
	}
}

func TestListFactsAsOf_DoesNotStarveOnListLimit(t *testing.T) {
	// Newer-invalid + older-valid in one session: ListMemoryWithOptions
	// Limit=50 (newest first) would keep 40 invalid + 10 valid and starve
	// as-of. collect must not pass Limit into listMemoryViaIndex.
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	asOf := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	fromOK := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fromFuture := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 40; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("invalid-%02d", i),
			Tier:      TierContextual,
			SessionID: "sess-A",
			Timestamp: asOf.Add(-time.Duration(i) * time.Minute),
			TemporalTags: []string{
				"valid_from:" + fromFuture.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: "not yet valid"},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 40; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("valid-%02d", i),
			Tier:      TierContextual,
			SessionID: "sess-A",
			Timestamp: asOf.Add(-time.Duration(60+i) * time.Minute),
			TemporalTags: []string{
				"valid_from:" + fromOK.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: "valid now"},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	results := store.ListFactsAsOf(FactsAsOfOptions{SessionID: "sess-A", AsOf: asOf, Limit: 50})
	if len(results) != 40 {
		t.Fatalf("starved as-of: len=%d want 40; got %v", len(results), idsOf(results))
	}
	for _, e := range results {
		if !strings.HasPrefix(e.ID, "valid-") {
			t.Fatalf("unexpected id %q in %v", e.ID, idsOf(results))
		}
	}
}

func TestListFactsAsOf_MetaIndexMatchesScan(t *testing.T) {
	baseDir := t.TempDir()
	idx := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	asOf := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	fromOK := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vf := "valid_from:" + fromOK.Format(time.RFC3339)
	conv := "qa-conv"
	sem := TierSemantic
	entries := []MemoryEntry{
		{
			ID: "a1", Tier: TierContextual, SessionID: "sess-A", Timestamp: asOf.Add(-time.Hour),
			TemporalTags: []string{vf, "entity:person:alice"},
			Content:      MemoryContent{Summary: "Alice alpha notes", Tags: []string{ConvTag(conv)}},
		},
		{
			ID: "b1", Tier: TierContextual, SessionID: "sess-B", Timestamp: asOf.Add(-2 * time.Hour),
			TemporalTags: []string{vf, "entity:person:bob"},
			Content:      MemoryContent{Summary: "Bob alpha notes"},
		},
		{
			ID: "a-sem", Type: "turn_fact", Tier: TierSemantic, SessionID: "sess-A", Timestamp: asOf.Add(-3 * time.Hour),
			TemporalTags: []string{vf, "entity:person:alice"},
			Content:      MemoryContent{Summary: "Alice semantic fact"},
		},
		{
			ID: "arch", Tier: TierArchival, SessionID: "sess-A", Timestamp: asOf.Add(-4 * time.Hour),
			TemporalTags: []string{vf},
			Content:      MemoryContent{Summary: "archived alpha"},
		},
	}
	for _, e := range entries {
		if err := idx.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir, DisableMetaIndex: true})
	cases := []struct {
		name string
		opts FactsAsOfOptions
	}{
		{"session", FactsAsOfOptions{SessionID: "sess-A", AsOf: asOf, Limit: 10}},
		{"conv_tag", FactsAsOfOptions{SessionID: conv, AsOf: asOf, Limit: 10}},
		{"query", FactsAsOfOptions{SessionID: "sess-A", Query: "alpha", AsOf: asOf, Limit: 10}},
		{"entity", FactsAsOfOptions{Entity: "alice", AsOf: asOf, Limit: 10}},
		{"tier", FactsAsOfOptions{Tier: &sem, AsOf: asOf, Limit: 10}},
		{"include_archival", FactsAsOfOptions{SessionID: "sess-A", AsOf: asOf, IncludeArchival: true, Limit: 10}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotIdx := idx.ListFactsAsOf(tc.opts)
			gotScan := scan.ListFactsAsOf(tc.opts)
			if !sameIDsInOrder(gotIdx, gotScan) {
				t.Fatalf("index vs scan: index=%v scan=%v", idsOf(gotIdx), idsOf(gotScan))
			}
			if len(gotIdx) == 0 {
				t.Fatal("empty facts as-of")
			}
		})
	}
}

func TestSearchMemoryWithOptions_AsOfFilter(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	asOf := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	fromOK := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	untilOK := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	untilExpired := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	entries := []MemoryEntry{
		{
			ID: "valid-alpha", Tier: TierContextual, Timestamp: asOf.Add(-time.Hour),
			TemporalTags: []string{
				"valid_from:" + fromOK.Format(time.RFC3339),
				"valid_until:" + untilOK.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: "alpha project valid"},
		},
		{
			ID: "expired-alpha", Tier: TierContextual, Timestamp: asOf.Add(-2 * time.Hour),
			TemporalTags: []string{
				"valid_from:" + fromOK.Format(time.RFC3339),
				"valid_until:" + untilExpired.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: "alpha project expired"},
		},
		{
			// No validity tags: known-by — event after asOf → drop
			ID: "future-alpha", Tier: TierContextual, Timestamp: asOf.Add(24 * time.Hour),
			Content: MemoryContent{Summary: "alpha project future event"},
		},
		{
			// No validity tags: known-by — event before asOf → keep
			ID: "history-alpha", Tier: TierContextual, Timestamp: asOf.Add(-48 * time.Hour),
			Content: MemoryContent{Summary: "alpha project history"},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Without AsOf: keyword path matches all four "alpha" entries.
	plain := store.SearchMemoryWithOptions("alpha project", SearchMemoryOptions{Limit: 10})
	if len(plain) != 4 {
		t.Fatalf("without AsOf len = %d, want 4; got %v", len(plain), idsOf(plain))
	}

	results := store.SearchMemoryWithOptions("alpha project", SearchMemoryOptions{
		AsOf:  &asOf,
		Limit: 10,
	})
	got := idsOf(results)
	want := map[string]bool{"valid-alpha": true, "history-alpha": true}
	if len(results) != 2 {
		t.Fatalf("AsOf filter len = %d, want 2; got %v", len(results), got)
	}
	for _, id := range got {
		if !want[id] {
			t.Fatalf("unexpected id %q in AsOf results %v", id, got)
		}
	}

	// Underfill: many expired + 1 valid; Limit must not return expired.
	for i := 0; i < 30; i++ {
		e := MemoryEntry{
			ID:        fmt.Sprintf("noise-%02d", i),
			Tier:      TierContextual,
			Timestamp: asOf.Add(-time.Duration(i+3) * time.Hour),
			TemporalTags: []string{
				"valid_from:" + fromOK.Format(time.RFC3339),
				"valid_until:" + untilExpired.Format(time.RFC3339),
			},
			Content: MemoryContent{Summary: fmt.Sprintf("alpha noise %d", i)},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	under := store.SearchMemoryWithOptions("alpha", SearchMemoryOptions{
		AsOf:  &asOf,
		Limit: 5,
	})
	// Only valid-alpha + history-alpha remain valid among alpha matches.
	if len(under) != 2 {
		t.Fatalf("AsOf underfill len = %d, want 2; got %v", len(under), idsOf(under))
	}
	for _, r := range under {
		if !want[r.ID] {
			t.Fatalf("AsOf underfill unexpected %q in %v", r.ID, idsOf(under))
		}
	}
}
