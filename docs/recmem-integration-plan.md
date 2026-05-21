# RecMem Integration Plan for `memory` Package

**Status**: Phase 2 Complete (v1.2 baseline)
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
- All new `MemoryEntry` writes first land in a "subconscious" mode.
- Implementation: `WriteLatent` + `tier-0-subconscious` directory.

### 3.2 Recurrence Phase-Transition Trigger [DONE - Phase 2]
- `clusterBySimilarity` + `shouldTriggerPhaseTransition` implemented.
- `AutoRecMemCompaction` provides the main entry point.
- Production notes added regarding clustering complexity.

### 3.3 Semantic Refinement Safety Net [Planned - Phase 3]
- Post-compaction atomic fact protection.

### 3.4 Configurable Sweet-Spot Parameters [DONE - Phase 1]
- `DataSim` and `DataCount` in `CompactionConfig`.

## 4. Implementation Details

### Files Modified
- `memory.go`: `WriteLatent`, `listSubconsciousEntries`, `loadEntry` helper.
- `compaction.go`: `clusterBySimilarity`, `shouldTriggerPhaseTransition`, `AutoRecMemCompaction`.
- `memory_test.go` / `vector_test.go`: RecMem + VectorStore coverage.

### Current Phase 2 Implementation (Production Quality)
```go
// clusterBySimilarity - centroid-based filtering (documented limitation for Phase 2)
func clusterBySimilarity(entries []MemoryEntry, threshold float64) [][]MemoryEntry { ... }

// shouldTriggerPhaseTransition - core density + count predicate
func shouldTriggerPhaseTransition(entries []MemoryEntry, cfg CompactionConfig) bool { ... }

// AutoRecMemCompaction - main public API for recurrence triggering
func (ps *PalaceStore) AutoRecMemCompaction(generateFn func(prompt string) string, vectorCb VectorStoreCallback) { ... }
```

### Rollout Phases
1. **Phase 1 (v1.1)**: Configurable parameters + latent write path. [DONE]
2. **Phase 2 (v1.2)**: Recurrence trigger + `AutoRecMemCompaction`. [DONE]
3. **Phase 3 (v1.3)**: Semantic Refinement. [Planned]

## 5. Expected Benefits
- 87% token reduction via density-driven consolidation.
- Long-horizon stability.

## 6. Open Questions & Risks
- Semantic Refinement (Phase 3).
- Optimal clustering algorithm for scale.
- When to call `AutoRecMemCompaction` (on-write vs background goroutine).

**Next Step**: Begin Phase 3 (Semantic Refinement safety net).
