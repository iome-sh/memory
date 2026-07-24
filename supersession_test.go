package memory

import (
	"strings"
	"testing"
	"time"
)

func TestSupersedeEntityFacts_TwoFactsSameEntity(t *testing.T) {
	// Two facts same entity; supersede prior open at asOf excluding the newer write
	// → ListFactsAsOf(now) returns only the second fact.
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	from1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	from2 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	first := MemoryEntry{
		ID:        "fact-1",
		Tier:      TierSemantic,
		Timestamp: from1,
		TemporalTags: []string{
			"entity:person:alice",
			"valid_from:" + from1.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "Alice lives in Boston"},
	}
	second := MemoryEntry{
		ID:        "fact-2",
		Tier:      TierSemantic,
		Timestamp: from2,
		TemporalTags: []string{
			"entity:person:alice",
			"valid_from:" + from2.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "Alice lives in Seattle"},
	}
	if err := store.Write(first); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(second); err != nil {
		t.Fatal(err)
	}

	// Before supersede both open at asOf (valid_from past, no until).
	before := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 10})
	if len(before) != 2 {
		t.Fatalf("before supersede: got %d facts %v, want 2", len(before), idsOf(before))
	}

	// Close prior open facts at from2, excluding the newer entry itself.
	n, err := store.supersedeEntityFactsExcluding("person:alice", from2, "fact-2")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("supersede count = %d, want 1", n)
	}

	// At asOf (>= from2): only fact-2 is valid (fact-1 closed exclusive-end at from2).
	after := store.ListFactsAsOf(FactsAsOfOptions{AsOf: asOf, Entity: "person:alice", Limit: 10})
	got := idsOf(after)
	if len(after) != 1 || after[0].ID != "fact-2" {
		t.Fatalf("after supersede ListFactsAsOf: got %v, want [fact-2]", got)
	}

	// Just before supersession instant: fact-1 still valid (exclusive end).
	hist := store.ListFactsAsOf(FactsAsOfOptions{
		AsOf:   from2.Add(-time.Second),
		Entity: "person:alice",
		Limit:  10,
	})
	if !entryIDsContain(hist, "fact-1") {
		t.Fatalf("historical as-of should include fact-1; got %v", idsOf(hist))
	}
}

func TestWriteAndSupersede_SameEntityOnlySecondVisible(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	from1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	first := MemoryEntry{
		ID:        "fact-1",
		Tier:      TierSemantic,
		Timestamp: from1,
		TemporalTags: []string{
			"entity:person:alice",
			"valid_from:" + from1.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "Alice lives in Boston"},
	}
	if err := store.Write(first); err != nil {
		t.Fatal(err)
	}

	// Second write supersedes person:alice (and entity:person:alice form).
	second := MemoryEntry{
		ID:   "fact-2",
		Tier: TierSemantic,
		TemporalTags: []string{
			"entity:person:alice",
		},
		Content: MemoryContent{Summary: "Alice lives in Seattle"},
	}
	if err := store.WriteAndSupersede(second, []string{"person:alice"}); err != nil {
		t.Fatal(err)
	}

	// After write, fact-1 should be closed; fact-2 open.
	loaded1, ok := store.Load("fact-1", TierSemantic)
	if !ok {
		t.Fatal("fact-1 missing")
	}
	_, until1 := ParseValidityWindow(loaded1)
	if until1 == nil {
		t.Fatal("fact-1 should have valid_until after supersede")
	}
	loaded2, ok := store.Load("fact-2", TierSemantic)
	if !ok {
		t.Fatal("fact-2 missing")
	}
	if !hasValidFromTag(loaded2) {
		t.Fatal("fact-2 should have valid_from stamped by WriteAndSupersede")
	}
	_, until2 := ParseValidityWindow(loaded2)
	if until2 != nil {
		t.Fatalf("fact-2 should remain open (no valid_until), got %v", until2)
	}

	// ListFactsAsOf(now) returns only second for this entity.
	now := time.Now().UTC()
	results := store.ListFactsAsOf(FactsAsOfOptions{AsOf: now, Entity: "person:alice", Limit: 10})
	got := idsOf(results)
	if len(results) != 1 || results[0].ID != "fact-2" {
		t.Fatalf("ListFactsAsOf now: got %v, want [fact-2]", got)
	}

	// Historical as-of before supersession still sees first (exclusive end).
	// fact-1 closed at WriteAndSupersede's now; query just before that until.
	hist := store.ListFactsAsOf(FactsAsOfOptions{
		AsOf:   until1.Add(-time.Second),
		Entity: "person:alice",
		Limit:  10,
	})
	histIDs := idsOf(hist)
	foundFirst := false
	for _, id := range histIDs {
		if id == "fact-1" {
			foundFirst = true
		}
	}
	if !foundFirst {
		t.Fatalf("historical ListFactsAsOf should include fact-1; got %v", histIDs)
	}
}

func TestSupersedeEntityFacts_DifferentEntitiesUnaffected(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	alice := MemoryEntry{
		ID: "alice-fact", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{"entity:person:alice", "valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "Alice fact"},
	}
	bob := MemoryEntry{
		ID: "bob-fact", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{"entity:person:bob", "valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "Bob fact"},
	}
	if err := store.Write(alice); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(bob); err != nil {
		t.Fatal(err)
	}

	n, err := store.SupersedeEntityFacts("person:alice", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("updated count = %d, want 1", n)
	}

	loadedAlice, _ := store.Load("alice-fact", TierSemantic)
	_, untilAlice := ParseValidityWindow(loadedAlice)
	if untilAlice == nil || !untilAlice.Equal(asOf) {
		t.Fatalf("alice valid_until = %v, want %v", untilAlice, asOf)
	}

	loadedBob, _ := store.Load("bob-fact", TierSemantic)
	_, untilBob := ParseValidityWindow(loadedBob)
	if untilBob != nil {
		t.Fatalf("bob should be unaffected; valid_until = %v", untilBob)
	}
	if !EntryValidAt(loadedBob, asOf) {
		t.Fatal("bob should still be valid at asOf")
	}
}

func TestSupersedeEntityFacts_AlreadyClosedNotDoubleWritten(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	closedUntil := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	closed := MemoryEntry{
		ID: "closed", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{
			"entity:person:alice",
			"valid_from:" + from.Format(time.RFC3339),
			"valid_until:" + closedUntil.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "already closed"},
	}
	open := MemoryEntry{
		ID: "open", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{
			"entity:person:alice",
			"valid_from:" + from.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "still open"},
	}
	if err := store.Write(closed); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(open); err != nil {
		t.Fatal(err)
	}

	n, err := store.SupersedeEntityFacts("person:alice", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("updated count = %d, want 1 (only open)", n)
	}

	// Closed entry keeps original valid_until (not rewritten to asOf).
	loadedClosed, _ := store.Load("closed", TierSemantic)
	_, untilClosed := ParseValidityWindow(loadedClosed)
	if untilClosed == nil || !untilClosed.Equal(closedUntil) {
		t.Fatalf("closed valid_until = %v, want original %v", untilClosed, closedUntil)
	}
	// Count valid_until tags: exactly one.
	vuCount := 0
	for _, tag := range loadedClosed.TemporalTags {
		if strings.HasPrefix(tag, validUntilTagPrefix) {
			vuCount++
		}
	}
	if vuCount != 1 {
		t.Fatalf("closed should have exactly 1 valid_until tag, got %d in %v", vuCount, loadedClosed.TemporalTags)
	}

	// Open entry closed at asOf.
	loadedOpen, _ := store.Load("open", TierSemantic)
	_, untilOpen := ParseValidityWindow(loadedOpen)
	if untilOpen == nil || !untilOpen.Equal(asOf) {
		t.Fatalf("open valid_until = %v, want %v", untilOpen, asOf)
	}

	// Second supersede is a no-op (already closed at asOf → EntryValidAt false).
	n2, err := store.SupersedeEntityFacts("person:alice", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if n2 != 0 {
		t.Fatalf("second supersede count = %d, want 0", n2)
	}
}

func TestSupersedeEntityFacts_EmptyEntityKeyNoOp(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := MemoryEntry{
		ID: "x", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{"entity:person:alice", "valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "x"},
	}
	if err := store.Write(e); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"", "   ", "\t"} {
		n, err := store.SupersedeEntityFacts(key, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("key %q: %v", key, err)
		}
		if n != 0 {
			t.Fatalf("key %q: count = %d, want 0", key, n)
		}
	}

	loaded, _ := store.Load("x", TierSemantic)
	_, until := ParseValidityWindow(loaded)
	if until != nil {
		t.Fatalf("empty key should not close facts; valid_until = %v", until)
	}
}

func TestSupersedeEntityFacts_KeyNormalization(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := MemoryEntry{
		ID: "alice", Tier: TierSemantic, Timestamp: from,
		TemporalTags: []string{"entity:person:alice", "valid_from:" + from.Format(time.RFC3339)},
		Content:      MemoryContent{Summary: "Alice"},
	}
	if err := store.Write(e); err != nil {
		t.Fatal(err)
	}
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	// Mixed case / whitespace should match.
	n, err := store.SupersedeEntityFacts("  Person:Alice  ", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("normalized key count = %d, want 1", n)
	}
}

func TestSetValidUntilTag_PreservesValidFrom(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	tags := []string{
		"entity:person:alice",
		"valid_from:" + from.Format(time.RFC3339),
		"cycle:1",
	}
	out := setValidUntilTag(tags, until)
	if !containsTag(out, "entity:person:alice") || !containsTag(out, "cycle:1") {
		t.Fatalf("lost non-validity tags: %v", out)
	}
	if !containsTag(out, "valid_from:"+from.Format(time.RFC3339)) {
		t.Fatalf("lost valid_from: %v", out)
	}
	if !containsTag(out, "valid_until:"+until.Format(time.RFC3339)) {
		t.Fatalf("missing valid_until: %v", out)
	}

	// Replace existing valid_until without duplicating.
	tags2 := setValidUntilTag(out, until.Add(time.Hour))
	vu := 0
	for _, t := range tags2 {
		if strings.HasPrefix(t, validUntilTagPrefix) {
			vu++
		}
	}
	if vu != 1 {
		t.Fatalf("expected single valid_until, got %d in %v", vu, tags2)
	}
}

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func entryIDsContain(entries []MemoryEntry, id string) bool {
	for _, e := range entries {
		if e.ID == id {
			return true
		}
	}
	return false
}
