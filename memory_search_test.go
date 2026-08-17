package memory

import (
	"sync/atomic"
	"testing"
	"time"
)

type countingEmbeddingFunc struct {
	inner EmbeddingFunc
	calls atomic.Int64
}

func (c *countingEmbeddingFunc) Func() EmbeddingFunc {
	return func(text string, dim int) []float32 {
		c.calls.Add(1)
		return c.inner(text, dim)
	}
}

func TestSearchMemory_PrecomputesEmbeddings(t *testing.T) {
	counter := &countingEmbeddingFunc{inner: GenerateSimpleEmbedding}

	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       t.TempDir(),
		EmbeddingFunc: counter.Func(),
	})

	entries := []MemoryEntry{
		{ID: "a", Tier: TierContextual, Content: MemoryContent{Summary: "alpha project notes", Full: "alpha details"}},
		{ID: "b", Tier: TierContextual, Content: MemoryContent{Summary: "beta release checklist", Full: "beta details"}},
		{ID: "c", Tier: TierContextual, Content: MemoryContent{Summary: "gamma incident timeline", Full: "gamma details"}},
		{ID: "d", Tier: TierContextual, Content: MemoryContent{Summary: "delta customer feedback", Full: "delta details"}},
		{ID: "e", Tier: TierContextual, Content: MemoryContent{Summary: "epsilon roadmap draft", Full: "epsilon details"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	queryVec := GenerateSimpleEmbedding("project alpha notes", 768)
	_ = store.SearchMemory("project alpha notes", nil, 3, queryVec)

	got := counter.calls.Load()
	want := int64(len(entries))
	if got != want {
		t.Fatalf("embed calls = %d, want %d (precompute once per entry, not O(n log n) sort compares)", got, want)
	}
}

func TestSearchMemoryWithOptions_SessionFilter(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	entries := []MemoryEntry{
		{ID: "s1a", Tier: TierContextual, SessionID: "sess-A", Content: MemoryContent{Summary: "alpha project notes session A"}},
		{ID: "s1b", Tier: TierContextual, SessionID: "sess-A", Content: MemoryContent{Summary: "alpha follow-up notes session A"}},
		{ID: "s2a", Tier: TierContextual, SessionID: "sess-B", Content: MemoryContent{Summary: "alpha project notes session B"}},
		{ID: "s0", Tier: TierContextual, SessionID: "", Content: MemoryContent{Summary: "alpha project notes no session"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	results := store.SearchMemoryWithOptions("alpha project notes", SearchMemoryOptions{
		SessionID: "sess-A",
		Limit:     10,
	})
	if len(results) != 2 {
		t.Fatalf("session filter len = %d, want 2; got %+v", len(results), idsOf(results))
	}
	for _, r := range results {
		if r.SessionID != "sess-A" {
			t.Fatalf("got SessionID %q, want sess-A", r.SessionID)
		}
	}
}

func TestSearchMemoryWithOptions_TimeWindow(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		// Timestamp preferred over CreatedAt
		{
			ID: "ts-old", Tier: TierContextual,
			Timestamp: base.Add(-48 * time.Hour),
			CreatedAt: base, // would pass window if used; Timestamp must win
			Content:   MemoryContent{Summary: "alpha event old timestamp"},
		},
		{
			ID: "ts-mid", Tier: TierContextual,
			Timestamp: base,
			Content:   MemoryContent{Summary: "alpha event mid timestamp"},
		},
		{
			ID: "ts-new", Tier: TierContextual,
			Timestamp: base.Add(48 * time.Hour),
			Content:   MemoryContent{Summary: "alpha event new timestamp"},
		},
		// CreatedAt fallback when Timestamp zero
		{
			ID: "ca-mid", Tier: TierContextual,
			CreatedAt: base.Add(1 * time.Hour),
			Content:   MemoryContent{Summary: "alpha event mid created_at"},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	from := base.Add(-1 * time.Hour)
	to := base.Add(2 * time.Hour)
	results := store.SearchMemoryWithOptions("alpha event", SearchMemoryOptions{
		TimeFrom: &from,
		TimeTo:   &to,
		Limit:    10,
	})
	got := idsOf(results)
	wantIDs := map[string]bool{"ts-mid": true, "ca-mid": true}
	if len(results) != 2 {
		t.Fatalf("time window len = %d, want 2; got %v", len(results), got)
	}
	for _, id := range got {
		if !wantIDs[id] {
			t.Fatalf("unexpected id %q in results %v", id, got)
		}
	}
}

func TestSearchMemoryWithOptions_ReRankTemporal(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	now := time.Now()
	// Keyword path returns insertion/list order; temporal re-rank should order by relevance.
	// CalculateRelevanceScore needs ScoreImpact > 0; recency uses LastAccessed.
	entries := []MemoryEntry{
		{
			ID: "low", Tier: TierContextual, LastAccessed: now.Add(-200 * time.Hour),
			Content: MemoryContent{Summary: "alpha relevance low"},
			Metrics: MemoryMetrics{ScoreImpact: 0.2, UsageCount: 0},
		},
		{
			ID: "high", Tier: TierContextual, LastAccessed: now,
			Content: MemoryContent{Summary: "alpha relevance high"},
			Metrics: MemoryMetrics{ScoreImpact: 0.9, UsageCount: 5},
		},
		{
			ID: "mid", Tier: TierContextual, LastAccessed: now.Add(-24 * time.Hour),
			Content: MemoryContent{Summary: "alpha relevance mid"},
			Metrics: MemoryMetrics{ScoreImpact: 0.5, UsageCount: 1},
		},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Without re-rank: order is whatever keyword path leaves (not guaranteed by score).
	plain := store.SearchMemoryWithOptions("alpha relevance", SearchMemoryOptions{Limit: 10})
	if len(plain) != 3 {
		t.Fatalf("plain len = %d, want 3", len(plain))
	}

	ranked := store.SearchMemoryWithOptions("alpha relevance", SearchMemoryOptions{
		Limit:          10,
		ReRankTemporal: true,
	})
	if len(ranked) != 3 {
		t.Fatalf("ranked len = %d, want 3", len(ranked))
	}
	// Stable descending by CalculateRelevanceScore
	for i := 0; i < len(ranked)-1; i++ {
		si := CalculateRelevanceScore(ranked[i])
		sj := CalculateRelevanceScore(ranked[i+1])
		if si < sj {
			t.Fatalf("re-rank not descending at %d: %f < %f (ids %v)", i, si, sj, idsOf(ranked))
		}
	}
	if ranked[0].ID != "high" {
		t.Fatalf("top id = %q, want high (scores high=%f mid=%f low=%f)",
			ranked[0].ID,
			CalculateRelevanceScore(entries[1]),
			CalculateRelevanceScore(entries[2]),
			CalculateRelevanceScore(entries[0]),
		)
	}

	// Stability: same opts → same order
	ranked2 := store.SearchMemoryWithOptions("alpha relevance", SearchMemoryOptions{
		Limit:          10,
		ReRankTemporal: true,
	})
	if !sameIDsInOrder(ranked, ranked2) {
		t.Fatalf("re-rank unstable: %v vs %v", idsOf(ranked), idsOf(ranked2))
	}
}

func TestSearchMemory_WrapperParityEmptyOpts(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	entries := []MemoryEntry{
		{ID: "a", Tier: TierContextual, Content: MemoryContent{Summary: "alpha project notes", Full: "alpha details"}},
		{ID: "b", Tier: TierContextual, Content: MemoryContent{Summary: "beta release checklist", Full: "beta details"}},
		{ID: "c", Tier: TierWorking, Content: MemoryContent{Summary: "alpha working notes", Full: "working details"}},
	}
	for _, e := range entries {
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Keyword path, no tier
	legacy := store.SearchMemory("alpha project notes", nil, 10, nil)
	opts := store.SearchMemoryWithOptions("alpha project notes", SearchMemoryOptions{Limit: 10})
	if !sameIDsInOrder(legacy, opts) {
		t.Fatalf("wrapper keyword parity: legacy %v vs opts %v", idsOf(legacy), idsOf(opts))
	}

	// Vector path with tier filter
	tier := TierContextual
	queryVec := GenerateSimpleEmbedding("alpha project notes", 768)
	legacyVec := store.SearchMemory("alpha project notes", &tier, 2, queryVec)
	optsVec := store.SearchMemoryWithOptions("alpha project notes", SearchMemoryOptions{
		Tier:     &tier,
		Limit:    2,
		QueryVec: queryVec,
	})
	if !sameIDsInOrder(legacyVec, optsVec) {
		t.Fatalf("wrapper vector parity: legacy %v vs opts %v", idsOf(legacyVec), idsOf(optsVec))
	}
}

func TestSearchMemoryWithOptions_HashQueryVecKeepsKeywordHit(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	const needle = "zircon-lantern-4829"
	if err := store.Write(MemoryEntry{
		ID:   "hit",
		Tier: TierContextual,
		Content: MemoryContent{
			Summary: "lab note " + needle,
			Full:    "synthetic unique token for hash-embedding recall",
		},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		e := MemoryEntry{
			ID:   "d" + string(rune('a'+i)),
			Tier: TierContextual,
			Content: MemoryContent{
				Summary: "unrelated distractor checklist " + string(rune('a'+i)),
				Full:    "no overlap with the needle token",
			},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	// Host lean hash path uses DefaultHashEmbeddingDim (768), not 384.
	vec := GenerateSimpleEmbedding(needle, DefaultHashEmbeddingDim)
	got := store.SearchMemoryWithOptions(needle, SearchMemoryOptions{
		Limit:    5,
		QueryVec: vec,
	})
	if len(got) > 5 {
		t.Fatalf("Limit 5 not applied; ids=%v", idsOf(got))
	}
	if !entryHasID(got, "hit") {
		t.Fatalf("hash QueryVec dropped keyword hit; ids=%v", idsOf(got))
	}
	if got[0].ID != "hit" {
		t.Fatalf("keyword hit should rank first under hash QueryVec, got %v", idsOf(got))
	}

	keywordOnly := store.SearchMemoryWithOptions(needle, SearchMemoryOptions{Limit: 5})
	if !entryHasID(keywordOnly, "hit") {
		t.Fatalf("keyword path missed hit; ids=%v", idsOf(keywordOnly))
	}
}

func TestSearchMemoryWithOptions_HaystackOriginalTextOnly(t *testing.T) {
	assertHaystackFieldHit(t, "obsidian-cinder-7714", MemoryEntry{
		ID:           "hit",
		Tier:         TierContextual,
		OriginalText: "raw turn obsidian-cinder-7714",
		Content:      MemoryContent{Summary: "lab turn", Full: "no unique token here"},
	})
}

func TestSearchMemoryWithOptions_HaystackKeyphrasesOnly(t *testing.T) {
	assertHaystackFieldHit(t, "quartz-harbor-3391", MemoryEntry{
		ID:         "hit",
		Tier:       TierContextual,
		Keyphrases: []string{"quartz-harbor-3391", "session note"},
		Content:    MemoryContent{Summary: "lab turn", Full: "no unique token here"},
	})
}

func TestSearchMemoryWithOptions_HaystackExtractedFactsOnly(t *testing.T) {
	assertHaystackFieldHit(t, "basalt-xenon-5502", MemoryEntry{
		ID:             "hit",
		Tier:           TierContextual,
		ExtractedFacts: []string{"user mentioned basalt-xenon-5502"},
		Content:        MemoryContent{Summary: "lab turn", Full: "no unique token here"},
	})
}

func assertHaystackFieldHit(t *testing.T, needle string, hit MemoryEntry) {
	t.Helper()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	if err := store.Write(hit); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		e := MemoryEntry{
			ID:   "d" + string(rune('a'+i)),
			Tier: TierContextual,
			Content: MemoryContent{
				Summary: "unrelated distractor checklist " + string(rune('a'+i)),
				Full:    "no overlap with the needle token",
			},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	keywordOnly := store.SearchMemoryWithOptions(needle, SearchMemoryOptions{Limit: 5})
	if !entryHasID(keywordOnly, "hit") {
		t.Fatalf("keyword path missed haystack-only hit; ids=%v", idsOf(keywordOnly))
	}

	vec := GenerateSimpleEmbedding(needle, DefaultHashEmbeddingDim)
	got := store.SearchMemoryWithOptions(needle, SearchMemoryOptions{
		Limit:    5,
		QueryVec: vec,
	})
	if len(got) > 5 {
		t.Fatalf("Limit 5 not applied; ids=%v", idsOf(got))
	}
	if !entryHasID(got, "hit") {
		t.Fatalf("hash QueryVec dropped haystack-only keyword hit; ids=%v", idsOf(got))
	}
	if got[0].ID != "hit" {
		t.Fatalf("keyword hit should rank first under hash QueryVec, got %v", idsOf(got))
	}
}

func TestSearchMemoryWithOptions_ReRankTemporalKeepsKeywordHit(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	now := time.Now()
	const needle = "zircon-lantern-4829"
	if err := store.Write(MemoryEntry{
		ID:           "hit",
		Tier:         TierContextual,
		LastAccessed: now.Add(-200 * time.Hour),
		OriginalText: "raw turn " + needle,
		Content:      MemoryContent{Summary: "lab turn", Full: "no unique token here"},
		Metrics:      MemoryMetrics{ScoreImpact: 0.05, UsageCount: 0},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		e := MemoryEntry{
			ID:           "d" + string(rune('a'+i)),
			Tier:         TierContextual,
			LastAccessed: now,
			Content: MemoryContent{
				Summary: "unrelated distractor checklist " + string(rune('a'+i)),
				Full:    "no overlap with the needle token",
			},
			Metrics: MemoryMetrics{ScoreImpact: 0.99, UsageCount: 8},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	vec := GenerateSimpleEmbedding(needle, DefaultHashEmbeddingDim)
	got := store.SearchMemoryWithOptions(needle, SearchMemoryOptions{
		Limit:          5,
		QueryVec:       vec,
		ReRankTemporal: true,
	})
	if len(got) > 5 {
		t.Fatalf("Limit 5 not applied; ids=%v", idsOf(got))
	}
	if !entryHasID(got, "hit") {
		t.Fatalf("ReRankTemporal dropped keyword hit past Limit; ids=%v", idsOf(got))
	}
	if got[0].ID != "hit" {
		t.Fatalf("keyword hit should stay first under ReRankTemporal, got %v", idsOf(got))
	}
}

func TestSearchMemoryWithOptions_KeywordOverlapOutranksIncidentalOR(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	if err := store.Write(MemoryEntry{
		ID:        "gold",
		Tier:      TierContextual,
		SessionID: "sess-gold",
		Content: MemoryContent{
			Summary: "cleaned white Adidas sneakers",
			Full:    "I last cleaned my white Adidas sneakers on Sunday.",
		},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 24; i++ {
		e := MemoryEntry{
			ID:        "d" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
			Tier:      TierContextual,
			SessionID: "sess-noise",
			Content: MemoryContent{
				Summary: "when did I last check the inbox",
				Full:    "when did I last update the weekly notes " + string(rune('a'+i%26)),
			},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	q := "when did I last clean my white Adidas sneakers"
	vec := GenerateSimpleEmbedding(q, DefaultHashEmbeddingDim)
	got := store.SearchMemoryWithOptions(q, SearchMemoryOptions{
		Limit:    15,
		QueryVec: vec,
	})
	if !entryHasID(got, "gold") {
		t.Fatalf("OR-any-token + hash top-k buried gold; ids=%v", idsOf(got))
	}
	if got[0].ID != "gold" {
		t.Fatalf("gold phrase should outrank incidental when/did/last hits, got %v", idsOf(got))
	}

	scoped := store.SearchMemoryWithOptions(q, SearchMemoryOptions{
		SessionID: "sess-gold",
		Limit:     15,
		QueryVec:  vec,
	})
	if !entryHasID(scoped, "gold") {
		t.Fatalf("session-scoped retrieve missed gold; ids=%v", idsOf(scoped))
	}
}

func TestKeywordTokens_HyphenNeedle(t *testing.T) {
	got := keywordTokens("zircon-lantern-4829")
	want := []string{"zircon", "lantern", "4829"}
	if len(got) != len(want) {
		t.Fatalf("tokens=%v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tokens=%v want %v", got, want)
		}
	}
}

func entryHasID(entries []MemoryEntry, id string) bool {
	for _, e := range entries {
		if e.ID == id {
			return true
		}
	}
	return false
}

func idsOf(entries []MemoryEntry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}

func sameIDsInOrder(a, b []MemoryEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}
