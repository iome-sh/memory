package memory

import (
	"strings"
)

// isHashEmbeddingModel reports whether model is the hash/default embedder.
// Empty and "hash" (any case, trimmed) are hash. Detection is by this string
// only — EmbeddingFunc pointer identity is not compared (kernel #45).
func isHashEmbeddingModel(model string) bool {
	m := strings.TrimSpace(model)
	return m == "" || strings.EqualFold(m, "hash")
}

func stripEntryEmbedding(e *MemoryEntry) {
	if e == nil {
		return
	}
	e.Content.Embedding = nil
	e.Content.EmbeddingModel = ""
	e.Content.EmbeddingDim = 0
}

func cloneFloat32(v []float32) []float32 {
	if len(v) == 0 {
		return nil
	}
	out := make([]float32, len(v))
	copy(out, v)
	return out
}

func embeddingSourceChanged(prev, next MemoryEntry) bool {
	return prev.Content.Summary != next.Content.Summary ||
		prev.Content.Full != next.Content.Full ||
		prev.OriginalText != next.OriginalText
}

// canPersistONNXEmbeddings is the write-side gate: flag on, non-hash model,
// EmbeddingFunc set. Hash vectors are never persisted even if the flag is true.
func (ps *PalaceStore) canPersistONNXEmbeddings() bool {
	if ps == nil || !ps.Config.PersistEmbeddings {
		return false
	}
	if isHashEmbeddingModel(ps.Config.EmbeddingModel) {
		return false
	}
	return ps.Config.EmbeddingFunc != nil
}

func (ps *PalaceStore) shouldComputePersistedEmbedding(entry MemoryEntry) bool {
	return ps.canPersistONNXEmbeddings() && len(entry.Content.Embedding) == 0
}

func persistedVectorForStore(e MemoryEntry, cfg PalaceConfig) []float32 {
	if isHashEmbeddingModel(cfg.EmbeddingModel) || isHashEmbeddingModel(e.Content.EmbeddingModel) {
		return nil
	}
	if strings.TrimSpace(e.Content.EmbeddingModel) != strings.TrimSpace(cfg.EmbeddingModel) {
		return nil
	}
	if len(e.Content.Embedding) == 0 {
		return nil
	}
	if e.Content.EmbeddingDim != 0 && e.Content.EmbeddingDim != len(e.Content.Embedding) {
		return nil
	}
	if cfg.EmbeddingDim > 0 && cfg.EmbeddingDim != len(e.Content.Embedding) {
		return nil
	}
	return e.Content.Embedding
}

func persistedVectorForQuery(e MemoryEntry, queryVec []float32, cfg PalaceConfig) []float32 {
	v := persistedVectorForStore(e, cfg)
	if len(v) == 0 || len(v) != len(queryVec) {
		return nil
	}
	return v
}

func (ps *PalaceStore) loadSameID(id string, prefer MemoryTier) (MemoryEntry, bool) {
	if id == "" || ps == nil {
		return MemoryEntry{}, false
	}
	if e, ok := ps.Load(id, prefer); ok {
		return e, true
	}
	for _, t := range []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic} {
		if t == prefer {
			continue
		}
		if e, ok := ps.Load(id, t); ok {
			return e, true
		}
	}
	return MemoryEntry{}, false
}

// preparePersistedEntry strips hash / stale / incompatible vectors before the
// ingest-ack write. A matching previously stored ONNX vec is kept when
// Summary/Full/OriginalText are unchanged.
func (ps *PalaceStore) preparePersistedEntry(entry MemoryEntry, prev MemoryEntry, hasPrev bool) MemoryEntry {
	if !ps.canPersistONNXEmbeddings() {
		stripEntryEmbedding(&entry)
		return entry
	}
	if hasPrev && !embeddingSourceChanged(prev, entry) {
		if vec := persistedVectorForStore(prev, ps.Config); len(vec) > 0 {
			entry.Content.Embedding = cloneFloat32(vec)
			entry.Content.EmbeddingModel = strings.TrimSpace(ps.Config.EmbeddingModel)
			entry.Content.EmbeddingDim = len(vec)
			return entry
		}
	}
	stripEntryEmbedding(&entry)
	return entry
}

func (ps *PalaceStore) safeEmbedEntry(entry MemoryEntry) (vec []float32, ok bool) {
	defer func() {
		if recover() != nil {
			vec = nil
			ok = false
		}
	}()
	if ps == nil || ps.Config.EmbeddingFunc == nil {
		return nil, false
	}
	out := ps.Config.EmbeddingFunc(entryEmbedText(entry), ps.Config.EmbeddingDim)
	if len(out) == 0 {
		return nil, false
	}
	return cloneFloat32(out), true
}

// applyCompactionParentEmbedding copies Content.Embedding from parent only when
// the store can persist ONNX vectors, model+dim match config, and embed source
// text is unchanged. Otherwise embedding fields stay empty (drop). Compaction
// stays HITL — this is not a hot vector index.
func (ps *PalaceStore) applyCompactionParentEmbedding(product *MemoryEntry, parent MemoryEntry) {
	if product == nil {
		return
	}
	stripEntryEmbedding(product)
	if ps == nil || !ps.canPersistONNXEmbeddings() {
		return
	}
	if embeddingSourceChanged(parent, *product) {
		return
	}
	vec := persistedVectorForStore(parent, ps.Config)
	if len(vec) == 0 {
		return
	}
	product.Content.Embedding = cloneFloat32(vec)
	product.Content.EmbeddingModel = strings.TrimSpace(ps.Config.EmbeddingModel)
	product.Content.EmbeddingDim = len(vec)
}
