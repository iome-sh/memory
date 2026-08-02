package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExpandRelatedEntities_TwoHopChain(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	// A → B → C
	store.AddEntityRelationship("person:alice", "org:acme")
	store.AddEntityRelationship("org:acme", "project:widget")

	got := store.ExpandRelatedEntities("person:alice", 2)
	want := []string{"person:alice", "org:acme", "project:widget"}
	if len(got) != len(want) {
		t.Fatalf("ExpandRelatedEntities: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ExpandRelatedEntities[%d]=%q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}

	// MaxHops=1 stops at B
	one := store.ExpandRelatedEntities("person:alice", 1)
	if len(one) != 2 || one[0] != "person:alice" || one[1] != "org:acme" {
		t.Fatalf("MaxHops=1: got %v, want [person:alice org:acme]", one)
	}

	// Empty seed
	if out := store.ExpandRelatedEntities("", 2); out != nil {
		t.Fatalf("empty seed: got %v, want nil", out)
	}
}

func TestMultiHopRetrieve_Hop2EntityFromSeedA(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	// Graph: A → B → C
	store.AddEntityRelationship("person:alice", "org:acme")
	store.AddEntityRelationship("org:acme", "project:widget")

	base := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "seed-hit", Tier: TierContextual, Timestamp: base,
			TemporalTags: []string{"entity:person:alice"},
			Content:      MemoryContent{Summary: "alice profile"},
		},
		{
			ID: "hop1", Tier: TierContextual, Timestamp: base.Add(time.Hour),
			TemporalTags: []string{"entity:org:acme"},
			Content:      MemoryContent{Summary: "acme org notes"},
		},
		{
			ID: "hop2", Tier: TierSemantic, Timestamp: base.Add(2 * time.Hour),
			TemporalTags: []string{"entity:project:widget"},
			Content:      MemoryContent{Summary: "widget project plan"},
		},
		{
			ID: "unrelated", Tier: TierContextual, Timestamp: base.Add(3 * time.Hour),
			TemporalTags: []string{"entity:person:bob"},
			Content:      MemoryContent{Summary: "bob notes"},
		},
		// RelatedConcepts path
		{
			ID: "via-concepts", Tier: TierContextual, Timestamp: base.Add(4 * time.Hour),
			Content: MemoryContent{Summary: "concept-linked widget"},
			Relations: MemoryRelations{RelatedConcepts: []string{"project:widget"}},
		},
	}
	for _, e := range entries {
		store.Write(e)
	}

	results := store.MultiHopRetrieve(MultiHopOptions{
		SeedEntity: "person:alice",
		MaxHops:    2,
		Limit:      20,
	})
	ids := idsOf(results)
	// Should include seed, hop1, hop2, via-concepts — not bob
	wantIDs := map[string]bool{"seed-hit": true, "hop1": true, "hop2": true, "via-concepts": true}
	if len(results) != 4 {
		t.Fatalf("got %d results %v, want 4", len(results), ids)
	}
	for _, id := range ids {
		if !wantIDs[id] {
			t.Fatalf("unexpected id %q in %v", id, ids)
		}
	}
	// hop-2 entity entry must be present when seed is A
	foundHop2 := false
	for _, e := range results {
		if e.ID == "hop2" {
			foundHop2 = true
		}
	}
	if !foundHop2 {
		t.Fatalf("hop-2 entry missing: %v", ids)
	}

	// MaxHops=1 should exclude hop2 / via-concepts (only A and B)
	narrow := store.MultiHopRetrieve(MultiHopOptions{
		SeedEntity: "person:alice",
		MaxHops:    1,
		Limit:      20,
	})
	for _, e := range narrow {
		if e.ID == "hop2" || e.ID == "via-concepts" {
			t.Fatalf("MaxHops=1 should not include hop-2 entity entries, got %v", idsOf(narrow))
		}
	}
}

func TestExpandRelatedEntitiesHops_MinHop(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	// A → B → C; also A → C direct (shorter path to C should win)
	store.AddEntityRelationship("person:alice", "org:acme")
	store.AddEntityRelationship("org:acme", "project:widget")
	store.AddEntityRelationship("person:alice", "project:widget")

	hops := store.ExpandRelatedEntitiesHops("person:alice", 2)
	if hops == nil {
		t.Fatal("expected hop map")
	}
	if hops["person:alice"] != 0 {
		t.Fatalf("seed hop: got %d, want 0", hops["person:alice"])
	}
	if hops["org:acme"] != 1 {
		t.Fatalf("1-hop: got %d, want 1", hops["org:acme"])
	}
	if hops["project:widget"] != 1 {
		t.Fatalf("direct edge should yield hop 1, got %d", hops["project:widget"])
	}

	// Empty seed
	if out := store.ExpandRelatedEntitiesHops("", 2); out != nil {
		t.Fatalf("empty seed: got %v, want nil", out)
	}
}

func TestMultiHopRetrieve_HopDistanceRanking(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	// Graph: A → B → C
	store.AddEntityRelationship("person:alice", "org:acme")
	store.AddEntityRelationship("org:acme", "project:widget")

	base := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	// Deliberately make farther hops newer so legacy event-time sort would invert hop order.
	entries := []MemoryEntry{
		{
			ID: "hop2-newest", Tier: TierContextual, Timestamp: base.Add(3 * time.Hour),
			TemporalTags: []string{"entity:project:widget"},
			Content:      MemoryContent{Summary: "2-hop newest"},
		},
		{
			ID: "hop1-mid", Tier: TierContextual, Timestamp: base.Add(2 * time.Hour),
			TemporalTags: []string{"entity:org:acme"},
			Content:      MemoryContent{Summary: "1-hop mid"},
		},
		{
			ID: "seed-oldest", Tier: TierContextual, Timestamp: base,
			TemporalTags: []string{"entity:person:alice"},
			Content:      MemoryContent{Summary: "seed oldest"},
		},
		{
			ID: "hop1-older", Tier: TierContextual, Timestamp: base.Add(time.Hour),
			TemporalTags: []string{"entity:org:acme"},
			Content:      MemoryContent{Summary: "1-hop older"},
		},
	}
	for _, e := range entries {
		store.Write(e)
	}

	results := store.MultiHopRetrieve(MultiHopOptions{
		SeedEntity: "person:alice",
		MaxHops:    2,
		Limit:      20,
	})
	ids := idsOf(results)
	wantOrder := []string{"seed-oldest", "hop1-mid", "hop1-older", "hop2-newest"}
	if len(ids) != len(wantOrder) {
		t.Fatalf("got %v, want order %v", ids, wantOrder)
	}
	for i, want := range wantOrder {
		if ids[i] != want {
			t.Fatalf("hop ranking order[%d]=%q, want %q (full=%v)", i, ids[i], want, ids)
		}
	}

	// Within same hop, event time desc: hop1-mid (newer) before hop1-older.
	// PreferShorterHops=false: legacy seed-first then event time (ignores hop1 vs hop2).
	off := false
	legacy := store.MultiHopRetrieve(MultiHopOptions{
		SeedEntity:        "person:alice",
		MaxHops:           2,
		Limit:             20,
		PreferShorterHops: &off,
	})
	lids := idsOf(legacy)
	if lids[0] != "seed-oldest" {
		t.Fatalf("legacy seed-first: first=%q want seed-oldest; full=%v", lids[0], lids)
	}
	// Among non-seed, pure event time: hop2-newest, hop1-mid, hop1-older
	if lids[1] != "hop2-newest" || lids[2] != "hop1-mid" || lids[3] != "hop1-older" {
		t.Fatalf("legacy non-seed event-time order: got %v", lids)
	}
}

func TestMultiHopRetrieve_LimitAfterExpansion_Underfill(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	store.AddEntityRelationship("seed:x", "related:y")

	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	// Many unrelated entries that would steal slots if Limit applied before filter.
	for i := 0; i < 30; i++ {
		store.Write(MemoryEntry{
			ID:           fmt.Sprintf("noise-%02d", i),
			Tier:         TierContextual,
			Timestamp:    base.Add(time.Duration(i) * time.Minute),
			TemporalTags: []string{"entity:noise:other"},
			Content:      MemoryContent{Summary: fmt.Sprintf("noise %d", i)},
		})
	}
	// Few related — newer first among related
	store.Write(MemoryEntry{
		ID: "rel-old", Tier: TierContextual, Timestamp: base.Add(-time.Hour),
		TemporalTags: []string{"entity:related:y"},
		Content:      MemoryContent{Summary: "related old"},
	})
	store.Write(MemoryEntry{
		ID: "rel-new", Tier: TierContextual, Timestamp: base.Add(time.Hour),
		TemporalTags: []string{"entity:related:y"},
		Content:      MemoryContent{Summary: "related new"},
	})
	store.Write(MemoryEntry{
		ID: "seed-entry", Tier: TierContextual, Timestamp: base,
		TemporalTags: []string{"entity:seed:x"},
		Content:      MemoryContent{Summary: "seed entry"},
	})

	results := store.MultiHopRetrieve(MultiHopOptions{
		SeedEntity: "seed:x",
		MaxHops:    1,
		Limit:      2,
	})
	if len(results) != 2 {
		t.Fatalf("Limit after expansion: got %d %v, want 2 related", len(results), idsOf(results))
	}
	for _, e := range results {
		if !entryMatchesExpandedEntity(e, "seed:x") && !entryMatchesExpandedEntity(e, "related:y") {
			t.Fatalf("underfill: unrelated entry leaked: %s tags=%v", e.ID, e.TemporalTags)
		}
	}
	// seed-entry preferred (seed match) over pure related; among equal class event time desc.
	// seed-entry matches seed → first; then related by event time (rel-new before rel-old).
	if results[0].ID != "seed-entry" {
		t.Fatalf("seed match should sort first: got %v", idsOf(results))
	}
}

func TestMultiHopRetrieve_MaxHopsClamp(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	// Chain of 6 hops: e0 → e1 → e2 → e3 → e4 → e5
	for i := 0; i < 5; i++ {
		store.AddEntityRelationship(fmt.Sprintf("e%d", i), fmt.Sprintf("e%d", i+1))
	}

	// MaxHops=0 → default 2 → e0,e1,e2
	def := store.ExpandRelatedEntities("e0", 0)
	// ExpandRelatedEntities clamps 0 to 1, but MultiHopRetrieve defaults 0 to 2.
	// Test MultiHopRetrieve clamp via expansion results through a tagged entry at e4.
	store.Write(MemoryEntry{
		ID: "at-e4", Tier: TierContextual,
		TemporalTags: []string{"entity:e4"},
		Content:      MemoryContent{Summary: "far"},
	})
	store.Write(MemoryEntry{
		ID: "at-e2", Tier: TierContextual,
		TemporalTags: []string{"entity:e2"},
		Content:      MemoryContent{Summary: "near"},
	})

	// MaxHops <= 0 defaults to 2
	got := store.MultiHopRetrieve(MultiHopOptions{SeedEntity: "e0", MaxHops: 0, Limit: 20})
	ids := idsOf(got)
	if !containsID(ids, "at-e2") {
		t.Fatalf("default MaxHops=2 should reach e2: %v", ids)
	}
	if containsID(ids, "at-e4") {
		t.Fatalf("default MaxHops=2 should not reach e4: %v", ids)
	}

	// MaxHops > 4 clamps to 4 → can reach e4, not e5 (if we had e5-only entry)
	store.Write(MemoryEntry{
		ID: "at-e5", Tier: TierContextual,
		TemporalTags: []string{"entity:e5"},
		Content:      MemoryContent{Summary: "farthest"},
	})
	clamped := store.MultiHopRetrieve(MultiHopOptions{SeedEntity: "e0", MaxHops: 99, Limit: 20})
	cids := idsOf(clamped)
	if !containsID(cids, "at-e4") {
		t.Fatalf("MaxHops clamp 4 should reach e4: %v", cids)
	}
	if containsID(cids, "at-e5") {
		t.Fatalf("MaxHops clamp 4 should not reach e5: %v", cids)
	}

	// ExpandRelatedEntities also clamps high values
	exp := store.ExpandRelatedEntities("e0", 100)
	if len(exp) != 5 { // e0..e4
		t.Fatalf("ExpandRelatedEntities clamp 4: got %v (len=%d), want 5 nodes e0..e4", exp, len(exp))
	}

	// ExpandRelatedEntities maxHops < 1 → 1
	if len(def) != 2 { // e0, e1 when maxHops clamped to 1
		// When called with 0, ExpandRelatedEntities clamps to 1
		t.Fatalf("ExpandRelatedEntities(0) clamp to 1: got %v", def)
	}
}

func TestMultiHopRetrieve_AsOfFilter(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	store.AddEntityRelationship("person:alice", "org:acme")

	asOf := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	fromPast := asOf.Add(-48 * time.Hour)
	fromFuture := asOf.Add(48 * time.Hour)

	store.Write(MemoryEntry{
		ID: "valid-now", Tier: TierSemantic, Timestamp: asOf.Add(-time.Hour),
		TemporalTags: []string{
			"entity:org:acme",
			"valid_from:" + fromPast.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "acme fact valid"},
	})
	store.Write(MemoryEntry{
		ID: "valid-later", Tier: TierSemantic, Timestamp: asOf.Add(-time.Hour),
		TemporalTags: []string{
			"entity:org:acme",
			"valid_from:" + fromFuture.Format(time.RFC3339),
		},
		Content: MemoryContent{Summary: "acme fact future"},
	})
	// Untagged future event → invalid under known-by-asOf
	store.Write(MemoryEntry{
		ID: "future-event", Tier: TierContextual, Timestamp: asOf.Add(time.Hour),
		TemporalTags: []string{"entity:org:acme"},
		Content:      MemoryContent{Summary: "future event"},
	})

	results := store.MultiHopRetrieve(MultiHopOptions{
		SeedEntity: "person:alice",
		MaxHops:    1,
		Limit:      20,
		AsOf:       &asOf,
	})
	ids := idsOf(results)
	if !containsID(ids, "valid-now") {
		t.Fatalf("AsOf should keep valid-now: %v", ids)
	}
	if containsID(ids, "valid-later") {
		t.Fatalf("AsOf should drop valid-later: %v", ids)
	}
	if containsID(ids, "future-event") {
		t.Fatalf("AsOf should drop future-event: %v", ids)
	}
}

func TestEntryEntityKeys(t *testing.T) {
	e := MemoryEntry{
		TemporalTags: []string{"entity:person:alice", "subject:auth", "cycle:1"},
		Content:      MemoryContent{Tags: []string{"entity:org:acme", "misc"}},
		Relations:    MemoryRelations{RelatedConcepts: []string{"project:widget", " person:alice "}},
	}
	keys := EntryEntityKeys(e)
	want := map[string]bool{
		"person:alice":      true,
		"entity:person:alice": true,
		"subject:auth":      true,
		"org:acme":          true,
		"entity:org:acme":   true,
		"project:widget":    true,
	}
	for _, k := range keys {
		if !want[k] {
			t.Fatalf("unexpected key %q in %v", k, keys)
		}
	}
	for k := range want {
		found := false
		for _, got := range keys {
			if got == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing key %q in %v", k, keys)
		}
	}
}

func TestAddEntityRelationship_CreatesRelationsDir(t *testing.T) {
	// Simulate missing relations dir (e.g. deleted after ensureDirs).
	base := t.TempDir()
	store := NewPalaceStore(base)
	relDir := filepath.Join(base, "relations")
	if err := os.RemoveAll(relDir); err != nil {
		t.Fatal(err)
	}
	store.AddEntityRelationship("a", "b")
	if _, err := os.Stat(filepath.Join(relDir, "entity-graph.json")); err != nil {
		t.Fatalf("entity-graph.json not written after relations dir recreate: %v", err)
	}
	got := store.GetRelatedEntities("a")
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("GetRelatedEntities: got %v, want [b]", got)
	}
}

func containsID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
