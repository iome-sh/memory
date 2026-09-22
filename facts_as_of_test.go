package memory

import (
	"fmt"
	"path/filepath"
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

func TestMetaValidAt_MatchesEntryValidAt(t *testing.T) {
	asOf := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	from := asOf.Add(-time.Hour)
	until := asOf
	cases := []MemoryEntry{
		{ID: "untagged-before", Timestamp: asOf.Add(-time.Hour)},
		{ID: "untagged-after", Timestamp: asOf.Add(time.Hour)},
		{ID: "untagged-zero"},
		{ID: "untagged-eq", Timestamp: asOf},
		{
			ID: "from-open", Timestamp: asOf.Add(48 * time.Hour),
			TemporalTags: []string{"valid_from:" + from.Format(time.RFC3339)},
		},
		{
			ID: "from-future",
			TemporalTags: []string{"valid_from:" + asOf.Add(time.Hour).Format(time.RFC3339)},
		},
		{
			ID: "until-exclusive",
			TemporalTags: []string{
				"valid_from:" + from.Format(time.RFC3339),
				"valid_until:" + until.Format(time.RFC3339),
			},
		},
		{
			ID: "until-later",
			TemporalTags: []string{"valid_until:" + asOf.Add(time.Hour).Format(time.RFC3339)},
		},
		{
			ID: "bad-tag", Timestamp: asOf.Add(time.Hour),
			TemporalTags: []string{"valid_from:not-a-time"},
		},
		{
			ID: "last-wins",
			TemporalTags: []string{
				"valid_from:" + asOf.Add(2*time.Hour).Format(time.RFC3339),
				"valid_from:" + from.Format(time.RFC3339),
			},
		},
		{
			ID: "content-until-ignored", Timestamp: asOf.Add(-time.Hour),
			Content: MemoryContent{Tags: []string{"valid_until:" + from.Format(time.RFC3339)}},
		},
		{
			ID: "content-entity-ignored", Timestamp: asOf.Add(-time.Hour),
			Content: MemoryContent{Tags: []string{"entity:person:mallory"}},
		},
		{ID: "created-at-only", CreatedAt: asOf.Add(-time.Minute)},
		{ID: "last-access-future", LastAccessed: asOf.Add(time.Minute)},
	}
	for _, e := range cases {
		m := metaFromEntry(e, "x")
		if got, want := metaValidAt(m, asOf), EntryValidAt(e, asOf); got != want {
			t.Fatalf("%s metaValidAt=%v EntryValidAt=%v tags=%v content=%v", e.ID, got, want, e.TemporalTags, e.Content.Tags)
		}
		if e.ID == "content-until-ignored" && (m.hasValidity || m.validUntil != nil) {
			t.Fatalf("content valid_until stored on meta: %+v", m.validUntil)
		}
		if e.ID == "content-entity-ignored" && len(m.entityTags) != 0 {
			t.Fatalf("content entity tag stored on meta: %v", m.entityTags)
		}
		if e.ID == "content-entity-ignored" && metaMatchesEntity(m, "mallory") {
			t.Fatal("content entity: tag widened meta entity match")
		}
	}
}

// v1 rows have no has_validity. Decoding them as v2 would treat a closed fact
// as known-by event time. toMeta must not recover validity from the mixed Tags slice.
func TestDurableMetaV1Shape_DoesNotInferValidityOrEntity(t *testing.T) {
	asOf := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	past := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	m := (durableEntryMeta{
		ID:        "closed",
		EventTime: asOf.Add(-time.Hour),
		Tags: []string{
			"entity:person:alice",
			"valid_until:" + past.Format(time.RFC3339),
		},
	}).toMeta(t.TempDir())
	if !metaValidAt(m, asOf) {
		t.Fatal("v1-shaped row (has_validity false) must stay known-by, not parse Tags")
	}
	if metaMatchesEntity(m, "person:alice") {
		t.Fatal("mixed Tags must not satisfy facts-as-of entity match")
	}
	if m.entityKeysKnown {
		t.Fatal("v1-shaped row must not claim known entity keys")
	}
	if !metaCouldMatchEntityKey(m, "person:alice") {
		t.Fatal("unknown entity keys cannot prove a non-match")
	}
}

func writeClosedOpenAsOfFixture(t *testing.T, store *PalaceStore) time.Time {
	t.Helper()
	asOf := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vf := "valid_from:" + from.Format(time.RFC3339)
	untilPast := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	vuPast := "valid_until:" + untilPast.Format(time.RFC3339)
	vuAsOf := "valid_until:" + asOf.Format(time.RFC3339)
	entries := []MemoryEntry{
		{
			ID: "alice-closed-a", Tier: TierContextual, Timestamp: asOf.Add(-10 * time.Hour),
			TemporalTags: []string{vf, vuPast, "entity:person:alice"},
			Content:      MemoryContent{Summary: "alice closed a"},
		},
		{
			ID: "alice-closed-b", Tier: TierContextual, Timestamp: asOf.Add(-9 * time.Hour),
			TemporalTags: []string{vf, vuAsOf, "entity:person:alice"},
			Content:      MemoryContent{Summary: "alice closed at asOf exclusive"},
		},
		{
			ID: "alice-closed-sem", Tier: TierSemantic, Timestamp: asOf.Add(-time.Hour),
			TemporalTags: []string{vf, vuPast, "entity:person:alice"},
			Content:      MemoryContent{Summary: "alice closed semantic"},
		},
		{
			ID: "alice-open", Tier: TierContextual, Timestamp: asOf.Add(-6 * time.Hour),
			TemporalTags: []string{vf, "entity:person:alice"},
			Content:      MemoryContent{Summary: "alice open"},
		},
		{
			ID: "bob-open", Tier: TierSemantic, Timestamp: asOf.Add(-8 * time.Hour),
			TemporalTags: []string{vf, "entity:person:bob"},
			Content:      MemoryContent{Summary: "bob open"},
		},
		{
			ID: "future-untagged", Tier: TierContextual, Timestamp: asOf.Add(2 * time.Hour),
			Content: MemoryContent{Summary: "event after asOf"},
		},
		{
			ID: "past-untagged", Tier: TierContextual, Timestamp: asOf.Add(-2 * time.Hour),
			Content: MemoryContent{Summary: "known by event time"},
		},
		{
			ID: "content-entity-decoy", Tier: TierContextual, Timestamp: asOf.Add(-4 * time.Hour),
			TemporalTags: []string{vf},
			Content: MemoryContent{
				Summary: "content entity tag",
				Tags:    []string{"entity:person:alice"},
			},
		},
		{
			ID: "content-until-decoy", Tier: TierContextual, Timestamp: asOf.Add(-3 * time.Hour),
			Content: MemoryContent{
				Summary: "content valid_until is not a window",
				Tags:    []string{vuPast},
			},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	return asOf
}

func TestListFactsAsOf_IndexMatchesScanWithClosedFacts(t *testing.T) {
	baseDir := t.TempDir()
	idx := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	asOf := writeClosedOpenAsOfFixture(t, idx)
	scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir, DisableMetaIndex: true})

	unfiltered := []string{
		"bob-open",
		"past-untagged",
		"content-until-decoy",
		"content-entity-decoy",
		"alice-open",
	}
	absent := []string{
		"alice-closed-a", "alice-closed-b", "alice-closed-sem", "future-untagged",
	}
	assertSameAsOf := func(t *testing.T, opts FactsAsOfOptions, want []string) {
		t.Helper()
		gotIdx := idx.ListFactsAsOf(opts)
		gotScan := scan.ListFactsAsOf(opts)
		if !sameIDsInOrder(gotIdx, gotScan) {
			t.Fatalf("index vs scan: index=%v scan=%v", idsOf(gotIdx), idsOf(gotScan))
		}
		if got := idsOf(gotIdx); !stringSliceEq(got, want) {
			t.Fatalf("ids = %v, want %v", got, want)
		}
		for _, id := range absent {
			if entryIDsContain(gotIdx, id) {
				t.Fatalf("absent id %s in %v", id, idsOf(gotIdx))
			}
		}
	}
	assertSameAsOf(t, FactsAsOfOptions{AsOf: asOf, Limit: 20}, unfiltered)
	assertSameAsOf(t, FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 20}, []string{"alice-open"})
	assertSameAsOf(t, FactsAsOfOptions{AsOf: asOf, Entity: "alice", Limit: 20}, []string{"alice-open"})
}

func TestListFactsAsOf_IndexSkipsClosedBodies(t *testing.T) {
	baseDir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	asOf := writeClosedOpenAsOfFixture(t, store)
	// Warm so the measured list uses the index, not the rebuild walk.
	_ = store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Limit: 20})

	opts := FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 20}
	reads := trackEntryJSONReads(store)
	got := store.ListFactsAsOf(opts)
	if ids := idsOf(got); !stringSliceEq(ids, []string{"alice-open"}) {
		t.Fatalf("indexed ids = %v", ids)
	}
	for _, id := range []string{"alice-closed-a", "alice-closed-b", "alice-closed-sem", "future-untagged", "bob-open", "content-entity-decoy"} {
		if idsInclude(*reads, id) {
			t.Fatalf("indexed as-of read %s: %v", id, *reads)
		}
	}
	if !idsInclude(*reads, "alice-open") {
		t.Fatalf("open fact body was not loaded: %v", *reads)
	}

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir})
	rreads := trackEntryJSONReads(reopened)
	rgot := reopened.ListFactsAsOf(opts)
	if !sameIDsInOrder(got, rgot) {
		t.Fatalf("reopen ids = %v, want %v", idsOf(rgot), idsOf(got))
	}
	if reopened.MetaIndexRebuilds() != 0 {
		t.Fatalf("v2 snapshot should load without rebuild, rebuilds=%d", reopened.MetaIndexRebuilds())
	}
	for _, id := range []string{"alice-closed-a", "alice-closed-b", "alice-closed-sem", "future-untagged"} {
		if idsInclude(*rreads, id) {
			t.Fatalf("durable as-of read %s: %v", id, *rreads)
		}
	}
	if !idsInclude(*rreads, "alice-open") {
		t.Fatalf("reopen did not load open fact: %v", *rreads)
	}

	scan := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: baseDir, DisableMetaIndex: true})
	sreads := trackEntryJSONReads(scan)
	sgot := scan.ListFactsAsOf(opts)
	if !sameIDsInOrder(got, sgot) {
		t.Fatalf("scan ids = %v, want %v", idsOf(sgot), idsOf(got))
	}
	if !idsInclude(*sreads, "alice-closed-a") || !idsInclude(*sreads, "future-untagged") {
		t.Fatalf("DisableMetaIndex should still read non-survivors: %v", *sreads)
	}
}

func trackEntryJSONReads(ps *PalaceStore) *[]string {
	got := &[]string{}
	ps.entryJSONRead = func(path string) {
		id := strings.TrimSuffix(filepath.Base(path), ".json")
		*got = append(*got, id)
	}
	return got
}

func idsInclude(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func stringSliceEq(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
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
