package memory

import (
	"sync/atomic"
	"testing"
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