package memory

import (
	"fmt"
	"testing"
	"time"
)

// listBenchN is large enough to see meta-index vs O(n) scan, small enough
// for a local `go test -bench` run. T2 btree / tag secondaries stay gated
// until this bench shows list rebuild cost.
const listBenchN = 200

var listBenchSessions = []string{"sess-A", "sess-B", "sess-C"}

func listBenchBaseTime() time.Time {
	return time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
}

func seedListBenchStore(b *testing.B, disableMetaIndex bool) *PalaceStore {
	return seedListBenchStoreN(b, listBenchN, disableMetaIndex)
}

func seedListBenchStoreN(b *testing.B, n int, disableMetaIndex bool) *PalaceStore {
	b.Helper()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:          b.TempDir(),
		DisableMetaIndex: disableMetaIndex,
	})
	base := listBenchBaseTime()
	for i := 0; i < n; i++ {
		sid := listBenchSessions[i%len(listBenchSessions)]
		e := MemoryEntry{
			ID:        fmt.Sprintf("e-%04d", i),
			Tier:      TierContextual,
			SessionID: sid,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Content: MemoryContent{
				Summary: fmt.Sprintf("timeline note %d %s", i, sid),
				Full:    "list-latency bench body",
			},
		}
		if err := store.Write(e); err != nil {
			b.Fatal(err)
		}
	}
	return store
}

func listBenchSessionTimeOpts() ListMemoryOptions {
	base := listBenchBaseTime()
	from := base
	to := base.Add(2 * time.Hour)
	return ListMemoryOptions{
		SessionID: "sess-A",
		TimeFrom:  &from,
		TimeTo:    &to,
		Limit:     50,
	}
}

func BenchmarkListMemoryWithOptions_SessionTimeLimit(b *testing.B) {
	opts := listBenchSessionTimeOpts()
	b.Run("MetaIndex", func(b *testing.B) {
		store := seedListBenchStore(b, false)
		// Warm so b.Loop measures the indexed list path, not the first O(n) rebuild.
		_ = store.ListMemoryWithOptions(opts)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.ListMemoryWithOptions(opts))
		}
		if n == 0 {
			b.Fatal("empty list")
		}
	})
	b.Run("DisableMetaIndex", func(b *testing.B) {
		store := seedListBenchStore(b, true)
		_ = store.ListMemoryWithOptions(opts)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.ListMemoryWithOptions(opts))
		}
		if n == 0 {
			b.Fatal("empty list")
		}
	})
}

func BenchmarkSearchMemoryWithOptions_CountQuery(b *testing.B) {
	store := seedListBenchStore(b, false)
	base := listBenchBaseTime()
	projects := []string{"atlas", "beacon", "cedar", "delta", "ember", "fjord"}
	for i, name := range projects {
		e := MemoryEntry{
			ID:        fmt.Sprintf("fact-%d", i),
			Type:      "turn_fact",
			Tier:      TierSemantic,
			SessionID: listBenchSessions[i%len(listBenchSessions)],
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			Content: MemoryContent{
				Summary: fmt.Sprintf("I led the %s project", name),
				Tags:    []string{"fact_augmented"},
			},
		}
		if err := store.Write(e); err != nil {
			b.Fatal(err)
		}
	}
	countQ := "How many projects have I led?"
	otherQ := "timeline note sess-A"
	countVec := GenerateSimpleEmbedding(countQ, 8)
	otherVec := GenerateSimpleEmbedding(otherQ, 8)
	// Warm so sub-benches measure search, not first-list meta rebuild.
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 1})

	b.Run("CountQuery_SkipVector", func(b *testing.B) {
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.SearchMemoryWithOptions(countQ, SearchMemoryOptions{
				Limit:    10,
				QueryVec: countVec,
			}))
		}
		if n == 0 {
			b.Fatal("empty search")
		}
	})
	b.Run("NonCount_WithQueryVec", func(b *testing.B) {
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.SearchMemoryWithOptions(otherQ, SearchMemoryOptions{
				Limit:    10,
				QueryVec: otherVec,
			}))
		}
		if n == 0 {
			b.Fatal("empty search")
		}
	})
	b.Run("CountQuery_SkipVector_SessionID", func(b *testing.B) {
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.SearchMemoryWithOptions(countQ, SearchMemoryOptions{
				SessionID: "sess-A",
				Limit:     10,
				QueryVec:  countVec,
			}))
		}
		if n == 0 {
			b.Fatal("empty search")
		}
	})
}

// BenchmarkListMemoryWithOptions_RebuildAndN2000 is optional scale/rebuild
// evidence for the T2 btree gate. go test without -bench does not run it.
func BenchmarkListMemoryWithOptions_RebuildAndN2000(b *testing.B) {
	opts := listBenchSessionTimeOpts()
	b.Run("N200_RebuildEach", func(b *testing.B) {
		store := seedListBenchStoreN(b, listBenchN, false)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			store.InvalidateMetaIndex()
			n = len(store.ListMemoryWithOptions(opts))
		}
		if n == 0 {
			b.Fatal("empty list")
		}
	})
	b.Run("N2000_MetaIndex", func(b *testing.B) {
		store := seedListBenchStoreN(b, 2000, false)
		_ = store.ListMemoryWithOptions(opts)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.ListMemoryWithOptions(opts))
		}
		if n == 0 {
			b.Fatal("empty list")
		}
	})
	b.Run("N2000_DisableMetaIndex", func(b *testing.B) {
		store := seedListBenchStoreN(b, 2000, true)
		_ = store.ListMemoryWithOptions(opts)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.ListMemoryWithOptions(opts))
		}
		if n == 0 {
			b.Fatal("empty list")
		}
	})
	b.Run("N2000_RebuildEach", func(b *testing.B) {
		store := seedListBenchStoreN(b, 2000, false)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			store.InvalidateMetaIndex()
			n = len(store.ListMemoryWithOptions(opts))
		}
		if n == 0 {
			b.Fatal("empty list")
		}
	})
}

func seedFactsAsOfBenchStore(b *testing.B, disableMetaIndex bool) *PalaceStore {
	b.Helper()
	store := seedListBenchStore(b, disableMetaIndex)
	base := listBenchBaseTime()
	for i := 0; i < 6; i++ {
		sid := listBenchSessions[i%len(listBenchSessions)]
		e := MemoryEntry{
			ID:        fmt.Sprintf("turn-fact-%02d", i),
			Type:      "turn_fact",
			Tier:      TierSemantic,
			SessionID: sid,
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			Content: MemoryContent{
				Summary: fmt.Sprintf("as-of fact %d %s", i, sid),
				Full:    "facts-as-of bench body",
			},
		}
		if err := store.Write(e); err != nil {
			b.Fatal(err)
		}
	}
	return store
}

func BenchmarkListFactsAsOf_SessionAsOf(b *testing.B) {
	opts := FactsAsOfOptions{
		SessionID: "sess-A",
		AsOf:      time.Now().UTC(),
		Limit:     50,
	}
	b.Run("MetaIndex", func(b *testing.B) {
		store := seedFactsAsOfBenchStore(b, false)
		// Warm so b.Loop measures the indexed collect path, not the first O(n) rebuild.
		_ = store.ListFactsAsOf(opts)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.ListFactsAsOf(opts))
		}
		if n == 0 {
			b.Fatal("empty facts as-of")
		}
	})
	b.Run("DisableMetaIndex", func(b *testing.B) {
		store := seedFactsAsOfBenchStore(b, true)
		_ = store.ListFactsAsOf(opts)
		b.ReportAllocs()
		var n int
		for b.Loop() {
			n = len(store.ListFactsAsOf(opts))
		}
		if n == 0 {
			b.Fatal("empty facts as-of")
		}
	})
}
