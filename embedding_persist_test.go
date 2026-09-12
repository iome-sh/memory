package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

func knownVec(v ...float32) []float32 {
	out := make([]float32, len(v))
	copy(out, v)
	return out
}

func entryJSONHasEmbeddingKey(t *testing.T, baseDir, id string) bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(baseDir, "tier-2-contextual", id+".json"))
	if err != nil {
		t.Fatalf("read entry json: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal entry json: %v", err)
	}
	content, _ := raw["content"].(map[string]any)
	if content == nil {
		return false
	}
	_, hasEmb := content["embedding"]
	_, hasModel := content["embedding_model"]
	_, hasDim := content["embedding_dim"]
	return hasEmb || hasModel || hasDim
}

func TestNewPalaceStoreWithConfig_PersistEmbeddingsDefaultOff(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	if store.Config.PersistEmbeddings {
		t.Fatal("PersistEmbeddings default must be false")
	}
	if store.Config.EmbeddingModel != "" {
		t.Fatalf("EmbeddingModel = %q, want empty hash default", store.Config.EmbeddingModel)
	}
	if store.Config.EmbeddingDim != 0 {
		t.Fatalf("EmbeddingDim = %d, want 0 (infer)", store.Config.EmbeddingDim)
	}
}

func TestWrite_FlagOffStripsStuffedEmbedding(t *testing.T) {
	dir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: dir})
	err := store.Write(MemoryEntry{
		ID:   "e1",
		Tier: TierContextual,
		Content: MemoryContent{
			Summary:        "alpha notes",
			Embedding:      knownVec(0.1, 0.2, 0.3),
			EmbeddingModel: "bge-small-en-v1.5",
			EmbeddingDim:   3,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := store.Load("e1", TierContextual)
	if !ok {
		t.Fatal("load")
	}
	if len(got.Content.Embedding) != 0 || got.Content.EmbeddingModel != "" || got.Content.EmbeddingDim != 0 {
		t.Fatalf("flag off persisted embedding: %+v", got.Content)
	}
	if entryJSONHasEmbeddingKey(t, dir, "e1") {
		t.Fatal("flag off JSON still has embedding fields")
	}
}

func TestWrite_FlagOnHashModelDoesNotPersist(t *testing.T) {
	known := knownVec(1, 0, 0, 0)
	for _, model := range []string{"", "hash", "HASH"} {
		t.Run("model="+model, func(t *testing.T) {
			dir := t.TempDir()
			store := NewPalaceStoreWithConfig(PalaceConfig{
				BaseDir:           dir,
				PersistEmbeddings: true,
				EmbeddingModel:    model,
				EmbeddingFunc: func(string, int) []float32 {
					return known
				},
			})
			err := store.Write(MemoryEntry{
				ID:   "e1",
				Tier: TierContextual,
				Content: MemoryContent{
					Summary:        "alpha notes",
					Embedding:      knownVec(9, 9, 9),
					EmbeddingModel: "bge-small-en-v1.5",
					EmbeddingDim:   3,
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			got, ok := store.Load("e1", TierContextual)
			if !ok {
				t.Fatal("load")
			}
			if len(got.Content.Embedding) != 0 || got.Content.EmbeddingModel != "" || got.Content.EmbeddingDim != 0 {
				t.Fatalf("hash model persisted vector: %+v", got.Content)
			}
			if entryJSONHasEmbeddingKey(t, dir, "e1") {
				t.Fatal("hash model JSON still has embedding fields")
			}
		})
	}
}

func TestWrite_FlagOnNonHashPersistsAndSearchSkipsReembed(t *testing.T) {
	known := knownVec(0.1, 0.2, 0.3, 0.4)
	var calls atomic.Int64
	embedFn := func(text string, dim int) []float32 {
		calls.Add(1)
		out := make([]float32, len(known))
		copy(out, known)
		return out
	}
	dir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:           dir,
		PersistEmbeddings: true,
		EmbeddingModel:    "test-model",
		EmbeddingDim:      4,
		EmbeddingFunc:     embedFn,
	})
	if err := store.Write(MemoryEntry{
		ID:   "e1",
		Tier: TierContextual,
		Content: MemoryContent{
			Summary: "alpha project notes",
			Full:    "alpha details",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("write embed calls = %d, want 1", calls.Load())
	}
	got, ok := store.Load("e1", TierContextual)
	if !ok {
		t.Fatal("load")
	}
	if !slices.Equal(got.Content.Embedding, known) {
		t.Fatalf("persisted embedding = %v, want %v", got.Content.Embedding, known)
	}
	if got.Content.EmbeddingModel != "test-model" {
		t.Fatalf("EmbeddingModel = %q", got.Content.EmbeddingModel)
	}
	if got.Content.EmbeddingDim != 4 {
		t.Fatalf("EmbeddingDim = %d, want 4", got.Content.EmbeddingDim)
	}
	if !entryJSONHasEmbeddingKey(t, dir, "e1") {
		t.Fatal("expected embedding fields in JSON")
	}

	beforeSearch := calls.Load()
	hits := store.SearchMemoryWithOptions("alpha project notes", SearchMemoryOptions{
		Limit:    5,
		QueryVec: known,
	})
	if calls.Load() != beforeSearch {
		t.Fatalf("search re-called embed: before=%d after=%d", beforeSearch, calls.Load())
	}
	if len(hits) == 0 || hits[0].ID != "e1" {
		t.Fatalf("hits = %v", hits)
	}
}

func TestWrite_TextChangeDropsStaleVector(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*MemoryEntry)
	}{
		{name: "summary", mut: func(e *MemoryEntry) { e.Content.Summary = "updated summary" }},
		{name: "full", mut: func(e *MemoryEntry) { e.Content.Full = "updated full" }},
		{name: "original_text", mut: func(e *MemoryEntry) { e.OriginalText = "updated original" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int64
			embedFn := func(text string, dim int) []float32 {
				calls.Add(1)
				if strings.Contains(text, "updated") {
					return knownVec(0, 1, 0)
				}
				return knownVec(1, 0, 0)
			}
			store := NewPalaceStoreWithConfig(PalaceConfig{
				BaseDir:           t.TempDir(),
				PersistEmbeddings: true,
				EmbeddingModel:    "test-model",
				EmbeddingDim:      3,
				EmbeddingFunc:     embedFn,
			})
			seed := MemoryEntry{
				ID:           "e1",
				Tier:         TierContextual,
				OriginalText: "raw alpha",
				Content: MemoryContent{
					Summary: "alpha notes",
					Full:    "alpha full",
				},
			}
			if err := store.Write(seed); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 {
				t.Fatalf("seed embed calls = %d, want 1", calls.Load())
			}
			next := seed
			next.Content.Embedding = knownVec(9, 9, 9) // stuffed stale
			next.Content.EmbeddingModel = "test-model"
			next.Content.EmbeddingDim = 3
			tc.mut(&next)
			if err := store.Write(next); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 2 {
				t.Fatalf("text change embed calls = %d, want 2 (stale dropped, re-embed)", calls.Load())
			}
			got, ok := store.Load("e1", TierContextual)
			if !ok {
				t.Fatal("load")
			}
			if slices.Equal(got.Content.Embedding, knownVec(9, 9, 9)) {
				t.Fatal("stuffed stale vector kept after text change")
			}
			if tc.name == "original_text" {
				// OriginalText is not in entryEmbedText; drop + re-embed yields the seed vec.
				if !slices.Equal(got.Content.Embedding, knownVec(1, 0, 0)) {
					t.Fatalf("original_text re-embed = %v, want seed vec", got.Content.Embedding)
				}
				return
			}
			if !slices.Equal(got.Content.Embedding, knownVec(0, 1, 0)) {
				t.Fatalf("embedding = %v, want re-embed of new text", got.Content.Embedding)
			}
		})
	}
}

func TestWrite_EmbedPanicOrEmptyStillAcks(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		dir := t.TempDir()
		store := NewPalaceStoreWithConfig(PalaceConfig{
			BaseDir:           dir,
			PersistEmbeddings: true,
			EmbeddingModel:    "test-model",
			EmbeddingFunc: func(string, int) []float32 {
				panic("embed boom")
			},
		})
		if err := store.Write(MemoryEntry{
			ID:      "e1",
			Tier:    TierContextual,
			Content: MemoryContent{Summary: "alpha notes"},
		}); err != nil {
			t.Fatalf("Write after embed panic: %v", err)
		}
		got, ok := store.Load("e1", TierContextual)
		if !ok {
			t.Fatal("acked JSON missing after embed panic")
		}
		if got.Content.Summary != "alpha notes" {
			t.Fatalf("summary = %q", got.Content.Summary)
		}
		if len(got.Content.Embedding) != 0 {
			t.Fatalf("panic stored vector %v", got.Content.Embedding)
		}
		if entryJSONHasEmbeddingKey(t, dir, "e1") {
			t.Fatal("panic JSON still has embedding fields")
		}
	})
	t.Run("empty", func(t *testing.T) {
		dir := t.TempDir()
		store := NewPalaceStoreWithConfig(PalaceConfig{
			BaseDir:           dir,
			PersistEmbeddings: true,
			EmbeddingModel:    "test-model",
			EmbeddingFunc: func(string, int) []float32 {
				return nil
			},
		})
		if err := store.Write(MemoryEntry{
			ID:      "e1",
			Tier:    TierContextual,
			Content: MemoryContent{Summary: "alpha notes"},
		}); err != nil {
			t.Fatalf("Write after empty embed: %v", err)
		}
		got, ok := store.Load("e1", TierContextual)
		if !ok {
			t.Fatal("acked JSON missing after empty embed")
		}
		if len(got.Content.Embedding) != 0 {
			t.Fatalf("empty embed stored vector %v", got.Content.Embedding)
		}
	})
}

func TestSearchMemoryWithOptions_PersistedVecModelMismatchReembeds(t *testing.T) {
	var calls atomic.Int64
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:           t.TempDir(),
		PersistEmbeddings: true,
		EmbeddingModel:    "test-model",
		EmbeddingDim:      3,
		EmbeddingFunc: func(string, int) []float32 {
			calls.Add(1)
			return knownVec(0, 0, 1)
		},
	})
	if err := store.Write(MemoryEntry{
		ID:      "e1",
		Tier:    TierContextual,
		Content: MemoryContent{Summary: "alpha notes"},
	}); err != nil {
		t.Fatal(err)
	}
	store.Config.EmbeddingModel = "other-model"
	before := calls.Load()
	_ = store.SearchMemoryWithOptions("alpha notes", SearchMemoryOptions{
		Limit:    5,
		QueryVec: knownVec(0, 0, 1),
	})
	if calls.Load() <= before {
		t.Fatal("expected re-embed on model mismatch")
	}
}

func TestApplyCompactionParentEmbedding_CopyOnlyIfCompatible(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:           t.TempDir(),
		PersistEmbeddings: true,
		EmbeddingModel:    "test-model",
		EmbeddingDim:      2,
		EmbeddingFunc:     func(string, int) []float32 { return knownVec(1, 0) },
	})
	parent := MemoryEntry{
		OriginalText: "same orig",
		Content: MemoryContent{
			Summary:        "same",
			Full:           "same full",
			Embedding:      knownVec(1, 0),
			EmbeddingModel: "test-model",
			EmbeddingDim:   2,
		},
	}

	copied := MemoryEntry{OriginalText: "same orig", Content: MemoryContent{Summary: "same", Full: "same full"}}
	store.applyCompactionParentEmbedding(&copied, parent)
	if !slices.Equal(copied.Content.Embedding, knownVec(1, 0)) {
		t.Fatalf("compatible copy = %v", copied.Content.Embedding)
	}

	changed := MemoryEntry{Content: MemoryContent{Summary: "different"}}
	store.applyCompactionParentEmbedding(&changed, parent)
	if len(changed.Content.Embedding) != 0 {
		t.Fatalf("text change should drop parent embedding, got %v", changed.Content.Embedding)
	}

	mismatch := parent
	mismatch.Content.EmbeddingModel = "other-model"
	sameText := MemoryEntry{OriginalText: "same orig", Content: MemoryContent{Summary: "same", Full: "same full"}}
	store.applyCompactionParentEmbedding(&sameText, mismatch)
	if len(sameText.Content.Embedding) != 0 {
		t.Fatalf("model mismatch should drop, got %v", sameText.Content.Embedding)
	}

	dimMismatch := parent
	dimMismatch.Content.EmbeddingDim = 99
	sameText = MemoryEntry{OriginalText: "same orig", Content: MemoryContent{Summary: "same", Full: "same full"}}
	store.applyCompactionParentEmbedding(&sameText, dimMismatch)
	if len(sameText.Content.Embedding) != 0 {
		t.Fatalf("dim mismatch should drop, got %v", sameText.Content.Embedding)
	}

	off := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	sameText = MemoryEntry{OriginalText: "same orig", Content: MemoryContent{Summary: "same", Full: "same full"}}
	off.applyCompactionParentEmbedding(&sameText, parent)
	if len(sameText.Content.Embedding) != 0 {
		t.Fatalf("flag off should drop, got %v", sameText.Content.Embedding)
	}
}

func TestHandleArchive_KeepsCompatibleParentEmbedding(t *testing.T) {
	known := knownVec(0.2, 0.3, 0.4)
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:           t.TempDir(),
		PersistEmbeddings: true,
		EmbeddingModel:    "test-model",
		EmbeddingDim:      3,
		EmbeddingFunc:     func(string, int) []float32 { return knownVec(known...) },
	})
	if err := store.Write(MemoryEntry{
		ID:      "keep-1",
		Type:    "note",
		Tier:    TierContextual,
		Content: MemoryContent{Summary: "sprint note keep-1", Full: "sprint planning notes keep-1"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.handleArchive([]string{"keep-1"}, TierContextual, DefaultCompactionConfig); err != nil {
		t.Fatal(err)
	}
	got, ok := store.Load("keep-1", TierArchival)
	if !ok {
		t.Fatal("expected archival copy")
	}
	if !slices.Equal(got.Content.Embedding, known) {
		t.Fatalf("archive embedding = %v, want %v", got.Content.Embedding, known)
	}
}

func TestHandleSummarize_DropsParentEmbeddingOnNewText(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:           t.TempDir(),
		PersistEmbeddings: true,
		EmbeddingModel:    "test-model",
		EmbeddingDim:      4,
		EmbeddingFunc: func(text string, dim int) []float32 {
			return GenerateSimpleEmbedding(text, 4)
		},
	})
	if err := store.Write(MemoryEntry{
		ID:      "p1",
		Type:    "note",
		Tier:    TierContextual,
		Content: MemoryContent{Summary: "sprint note p1", Full: "sprint planning notes p1 unique parent body"},
	}); err != nil {
		t.Fatal(err)
	}
	writeCompactionNote(t, store, "p2")
	if err := store.handleSummarize([]string{"p1", "p2"}, TierContextual, DefaultCompactionConfig, nil); err != nil {
		t.Fatal(err)
	}
	parent, ok := store.Load("p1", TierArchival)
	if !ok {
		t.Fatal("parent archival")
	}
	var product MemoryEntry
	found := false
	for _, e := range store.ListEntriesInTier(TierContextual) {
		if e.Type == "summary" {
			product = e
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected summary product")
	}
	if len(parent.Content.Embedding) == 0 {
		t.Fatal("parent should have persisted embedding")
	}
	if slices.Equal(product.Content.Embedding, parent.Content.Embedding) {
		t.Fatal("summarize product copied parent embedding despite new text")
	}
}
