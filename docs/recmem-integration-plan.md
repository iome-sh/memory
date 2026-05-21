# RecMem Integration Plan for `memory` Package

**Status**: Draft (v1.0.0 baseline)
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

### 3.1 Latent Subconscious Buffer (Low-Energy Reservoir)
- All new `MemoryEntry` writes first land in a "subconscious" mode: only embedding + metadata stored via `VectorStore`.
- No LLM call until recurrence trigger.
- Implementation: Add `SubconsciousMode` flag to `MemoryEntry` or new `WriteLatent` method.

### 3.2 Recurrence Phase-Transition Trigger
- On every `Write`, run lightweight clustering on recent entries in the subconscious layer.
- Trigger `PerformCompaction` automatically when a cluster meets:
  - `data_sim` ≥ 0.7 (cosine similarity radius)
  - `data_count` ≥ 5 (recurrent threshold)
- Use existing `averageEmbedding` + `CosineSimilarity` for clustering.
- Add to `CompactionConfig`:
  ```go
  DataSim   float64 `json:"data_sim"`
  DataCount int     `json:"data_count"`
  ```

### 3.3 Semantic Refinement Safety Net
- Post-compaction step that scans the source entries for "atomic facts" (proper names, dates, constraints, hard negations).
- Store as separate high-fidelity `MemoryEntry` objects in a new `tier-semantic` directory or via `Relations`.
- New method: `func (ps *PalaceStore) SemanticRefine(cluster []MemoryEntry) error`
- Protects "lonely data points" (e.g., single-mention allergies) that never reach `data_count`.

### 3.4 Configurable Sweet-Spot Parameters
- Defaults: `DataSim: 0.7`, `DataCount: 5` (validated on LoCoMo).
- Expose in `PalaceConfig` and `NewPalaceStoreWithConfig`.
- Add sensitivity helper: `func ValidateRecMemParams(sim float64, count int) error`.

## 4. Implementation Details

### Files to Modify / Create
- `memory.go`: Extend `PalaceConfig`, add `WriteLatent`, `listSubconsciousEntries`, `clusterRecentEntries`.
- `compaction.go`: Add `RecurrenceTrigger` logic inside `PerformCompaction` (or new `AutoCompaction` wrapper). Update `DefaultCompactionConfig`.
- `vector.go`: Minor — ensure `SearchSimilar` supports subconscious-only filters (already possible via payload).
- New: `docs/recmem-integration-plan.md` (this file).

### Core Code Sketch (compaction.go)
```go
func (ps *PalaceStore) shouldTriggerPhaseTransition(entries []MemoryEntry, cfg CompactionConfig) bool {
    if len(entries) < cfg.DataCount { return false }
    avg := averageEmbedding(entries)
    for _, e := range entries {
        if CosineSimilarity(GenerateSimpleEmbedding(e.Content.Summary, 768), avg) < cfg.DataSim {
            return false
        }
    }
    return true
}

// Called from Write or background goroutine
func (ps *PalaceStore) AutoRecMemCompaction() {
    subconscious := ps.listSubconsciousEntries()
    clusters := clusterBySimilarity(subconscious, cfg.DataSim)
    for _, cluster := range clusters {
        if ps.shouldTriggerPhaseTransition(cluster, cfg) {
            ps.PerformCompaction(TierContextual, cfg, generateFn, vectorCb)
            ps.SemanticRefine(cluster)
        }
    }
}
```

### Rollout Phases
1. **Phase 1 (v1.1)**: Add configurable `DataSim`/`DataCount` + latent write path (2–3 days).
2. **Phase 2 (v1.2)**: Implement recurrence trigger + auto-compaction (3–4 days).
3. **Phase 3 (v1.3)**: Semantic Refinement + benchmark validation (4–5 days).

## 5. Expected Benefits
- 87% reduction in LLM token usage during memory construction (validated on LoCoMo).
- Long-horizon stability without context drift.
- Automatic protection of high-stake atomic facts.
- Backward compatible; existing manual compaction continues to work.

## 6. Open Questions & Risks
- How to handle "hard negations" in SemanticRefine (pattern matching vs LLM)?
- Background goroutine vs on-write trigger (latency vs freshness)?
- Migration path for existing PalaceStore instances.

**Next Step**: Implement Phase 1 and update `memory_test.go` with RecMem-specific unit tests.
