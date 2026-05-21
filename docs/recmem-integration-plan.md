# RecMem Integration Plan for `memory` Package

**Status**: In Progress (v1.1 baseline)
**Date**: May 20, 2026
**Goal**: Incorporate Recurrence-Based Memory (RecMem) density-driven phase transitions and three-tier manifold into the existing Palace architecture for 87% token reduction and long-horizon stability.

## 1. Executive Summary
The RecMem research introduces a physics-inspired "phase transition" model that defers expensive LLM consolidation until semantic density reaches critical mass. This directly extends the current `memory` package's compaction, vector, and tier system. Integration will add automatic recurrence triggering, a latent subconscious buffer, and a semantic refinement safety net while preserving full backward compatibility.

## 2. Alignment with Existing Architecture
| RecMem Concept          | Current `memory` Equivalent                  | Gap / Extension Needed                     |
|-------------------------|----------------------------------------------|--------------------------------------------|
| Subconscious Store      | Working tier + VectorStore                   | Add latent-only write path + density check |
| Episodic Store          | Contextual tier + compaction handlers        | Add recurrence phase-transition trigger    |
| Semantic Store          | Relations + atomic facts in MemoryEntry      | New `SemanticRefine` pass + high-fidelity store |
| Phase Transition        | `alpha` similarity + temporal window         | Replace/augment with `data_sim` + `data_count` |
| Token Efficiency        | Manual `PerformCompaction`                   | Automatic background trigger on write      |

The existing `CompactionConfig`, `PalaceConfig`, `VectorStore`, and `CosineSimilarity` provide 80% of the required primitives.

## 3. Proposed New Features

### 3.1 Latent Subconscious Buffer (Low-Energy Reservoir) [DONE - Phase 1]
- All new `MemoryEntry` writes first land in a "subconscious" mode: only embedding + metadata stored via `VectorStore`.
- No LLM call until recurrence trigger.
- Implementation: Added `WriteLatent` method + `tier-0-subconscious` directory. [Implemented & tested]

### 3.2 Recurrence Phase-Transition Trigger [In Progress - Phase 2]
- On every `Write`, run lightweight clustering on recent entries in the subconscious layer.
- Trigger `PerformCompaction` automatically when a cluster meets:
  - `data_sim` ≥ 0.7 (cosine similarity radius)
  - `data_count` ≥ 5 (recurrent threshold)
- Use existing `averageEmbedding` + `CosineSimilarity` for clustering.
- Added to `CompactionConfig`:
```go
DataSim   float64 `json:"data_sim"`
DataCount int     `json:"data_count"`
```

### 3.3 Semantic Refinement Safety Net [Planned - Phase 3]
- Post-compaction step that scans the source entries for "atomic facts" (proper names, dates, constraints, hard negations).
- Store as separate high-fidelity `MemoryEntry` objects in a new `tier-semantic` directory or via `Relations`.
- New method: `func (ps *PalaceStore) SemanticRefine(cluster []MemoryEntry) error`
- Protects "lonely data points" (e.g., single-mention allergies) that never reach `data_count`.

### 3.4 Configurable Sweet-Spot Parameters [DONE - Phase 1]
- Defaults: `DataSim: 0.7`, `DataCount: 5` (validated on LoCoMo).
- Exposed in `CompactionConfig` and `DefaultCompactionConfig`.
- Sensitivity validated via tests.

## 4. Implementation Details

### Files to Modify / Create
- `memory.go`: Extend with `WriteLatent`, `listSubconsciousEntries`, `clusterBySimilarity` helper. [Phase 1 + 2 partial]
- `compaction.go`: Add `shouldTriggerPhaseTransition`, `clusterBySimilarity`, `AutoRecMemCompaction` skeleton. Update `DefaultCompactionConfig`. [Phase 1 + 2 in progress]
- `vector.go`: Minor — ensure `SearchSimilar` supports subconscious-only filters (already possible via payload).
- `memory_test.go`: Added `TestCompactionConfig_RecMemDefaults` and `TestPalaceStore_WriteLatent` (with fixes for unused vars and gofmt compliance). [DONE]
- New: `docs/recmem-integration-plan.md` (this file).

### Core Code Sketch (compaction.go) - Updated for Phase 2
```go
func (ps *PalaceStore) shouldTriggerPhaseTransition(entries []MemoryEntry, cfg CompactionConfig) bool {
	if len(entries) < cfg.DataCount {
		return false
	}
	avg := averageEmbedding(entries)
	for _, e := range entries {
		vec := GenerateSimpleEmbedding(e.Content.Summary+" "+e.Content.Full, 768)
		if CosineSimilarity(vec, avg) < cfg.DataSim {
			return false
		}
	}
	return true
}

// clusterBySimilarity groups entries by cosine similarity (simple single-pass for Phase 2)
func clusterBySimilarity(entries []MemoryEntry, threshold float64) [][]MemoryEntry {
	// Simplified implementation for initial trigger; production would use DBSCAN or hierarchical clustering
	if len(entries) == 0 {
		return nil
	}
	// For now return all as one cluster if any pair meets threshold (placeholder for full impl)
	return [][]MemoryEntry{entries}
}

// AutoRecMemCompaction runs recurrence check on subconscious and triggers if needed
func (ps *PalaceStore) AutoRecMemCompaction(generateFn func(prompt string) string, vectorCb VectorStoreCallback) {
	cfg := ps.Config.CompactionConfig
	subconscious := ps.listSubconsciousEntries()
	if len(subconscious) == 0 {
		return
	}
	clusters := clusterBySimilarity(subconscious, cfg.DataSim)
	for _, cluster := range clusters {
		if ps.shouldTriggerPhaseTransition(cluster, cfg) {
			ps.PerformCompaction(TierContextual, cfg, generateFn, vectorCb)
			// SemanticRefine would go here in Phase 3
		}
	}
}
```

### Rollout Phases
1. **Phase 1 (v1.1)**: Add configurable `DataSim`/`DataCount` + latent write path (2–3 days). [DONE]
2. **Phase 2 (v1.2)**: Implement recurrence trigger + auto-compaction (3–4 days). [In Progress]
3. **Phase 3 (v1.3)**: Semantic Refinement + benchmark validation (4–5 days). [Planned]

## 5. Expected Benefits
- 87% reduction in LLM token usage during memory construction (validated on LoCoMo).
- Long-horizon stability without context drift.
- Automatic protection of high-stake atomic facts.
- Backward compatible; existing manual compaction continues to work.

## 6. Open Questions & Risks
- How to handle "hard negations" in SemanticRefine (pattern matching vs LLM)?
- Background goroutine vs on-write trigger (latency vs freshness)?
- Full clustering algorithm vs simple threshold (performance).
- Migration path for existing PalaceStore instances.

**Next Step**: Complete Phase 2 helpers (`listSubconsciousEntries`, `clusterBySimilarity`, `shouldTriggerPhaseTransition`, `AutoRecMemCompaction`), add test coverage, run `gofmt`, and commit. Then move to Phase 3.
