# Memory Package - Improvements & Roadmap

**Repository**: `github.com/sudo-jin/memory`
**Purpose**: Prioritized improvement plan for the standalone hierarchical memory package (Palace).
**Last Updated**: May 2026

This document reorganizes improvements by **highest business/engineering impact first**, grouped into logical rollout categories suitable for a standalone, importable Go package.

---

## 1. High Impact: Core Stability & Reliability (Phase 1)

### 1.1 Real Temporal Decay + Unified Relevance Scoring

**Priority**: Critical

Currently `calculateTemporalDecay` is a no-op. Old memories never lose relevance.

**Actions**:
- Implement exponential decay based on `LastAccessed` / `UpdatedAt`.
- Combine with recency boost into `CalculateRelevanceScore(entry MemoryEntry)`.
- Use this score in `PalaceStore` listing, compaction candidate selection, and future eviction.

**Impact**: Prevents memory bloat and stale knowledge pollution in long-running agents.

### 1.2 Robust Error Handling & Observability

**Priority**: High

Many operations silently fail or only print to stdout.

**Actions**:
- Return errors consistently from `PalaceStore` methods.
- Add structured logging (or `log/slog`).
- Add retry logic for vector operations.
- Expose `MemoryStats` (entry counts per tier, last compaction, etc.).

### 1.3 Activate Versioning

**Priority**: High

`archiveToVersions` exists but is never called.

**Actions**:
- Call versioning inside `Write` or before major mutations.
- Increment `Version` field properly.
- Add helper to load version history.

**Impact**: Enables auditability and safe evolution of the knowledge base.

---

## 2. High Impact: Vector & Retrieval Quality (Phase 1–2)

### 2.1 Robust Vector Similarity Search (Implemented)

`VectorStore.SearchSimilar` and `SearchByText` now exist with payload support and HNSW tuning.

**Next Steps**:
- Add `SearchMemory` helper on `PalaceStore` that combines vector results + file-backed entries.
- Support hybrid search (vector + keyword + temporal filter).

### 2.2 Optional Vector Integration via Callbacks (Implemented)

`VectorStoreCallback` is now accepted in `PerformCompaction`.

**Next Steps**:
- Make vector storage automatic when a `VectorStore` is attached to `PalaceStore`.
- Add configuration for when to embed (on write, on compaction, both).

### 2.3 Multi-Factor Scoring (s + t + r)

**Priority**: High

Combine semantic similarity, temporal relevance, and robustness (forgetting curve) — inspired by H-Mem.

**Actions**:
- Implement full `s + t + r` scoring.
- Use in both file-based listing and vector search re-ranking.

---

## 3. High Impact: Compaction & Knowledge Evolution (Phase 1–2)

### 3.1 Temporal Window + Alpha-Constrained Compaction (Partially Implemented)

`PerformCompaction` now filters by temporal window (`TemporalWindowSize`).

**Next Steps**:
- Add optional α (similarity threshold) check before merging/summarizing within the window.
- Make compaction respect H-Mem-style constraints to avoid temporal drift.

### 3.2 Structured Compaction Output + Verification

**Priority**: High

Current action parsing is brittle.

**Actions**:
- Support JSON output mode from the LLM.
- Add verification step after executing compaction actions.
- Add fallback heuristic compaction when LLM output is invalid.

### 3.3 Persist Compaction Metrics

Track and persist `AvgScoreBefore/After` and `Improvement` in `CompactionConfig`.

---

## 4. Medium Impact: API & Usability (Phase 2)

### 4.1 Memory Query API

Add high-level search methods:

```go
func (ps *PalaceStore) Search(query string, opts SearchOptions) ([]MemoryEntry, error)
```

Support hybrid (vector + keyword + tier + temporal) queries.

### 4.2 Working Tier Lifecycle

- Automatic promotion from Working → Contextual on high access/score.
- Size or age-based eviction from Working tier.

### 4.3 PalaceStore Configuration

Introduce `PalaceConfig` struct for:
- Base directory
- Default compaction config
- Vector store attachment
- Embedding function injection

---

## 5. Medium Impact: Embeddings & Semantic Quality (Phase 2–3)

### 5.1 Replace Simple Embedding

Current `generateSimpleEmbedding` is deterministic but non-semantic.

**Recommended Path**:
- Make embedding generation injectable (`EmbeddingFunc`).
- Support pluggable backends (local model, external service, or future llama.cpp embeddings).

### 5.2 Embedding Versioning & Migration

Support multiple embedding generations and re-embedding when upgrading the embedding model.

---

## 6. Foundation: Architecture, Portability & Testing (Ongoing)

### 6.1 Cross-Platform Support

Remove any remaining platform-specific assumptions. Ensure `memory`, `compaction`, and `vector` packages build cleanly on Linux/macOS/Windows.

### 6.2 Comprehensive Testing

- Unit tests for `PalaceStore`, compaction logic, temporal window filtering.
- Integration tests with `VectorStore`.
- Golden tests for compaction action parsing.

### 6.3 H-Mem Inspired Features (Long-term)

- Entity extraction + richer relational graph during write/compaction.
- Hybrid retrieval (graph expansion + bottom-up tier search).
- Progressive hierarchical compaction aligned with temporal windows (α/β).

---

## Recommended Rollout Order

| Phase | Focus Area                          | Key Deliverables                              | Estimated Effort |
|-------|-------------------------------------|-----------------------------------------------|------------------|
| 1     | Core Stability + Vector Basics      | Temporal decay, error handling, versioning, vector search + callbacks | 1–2 weeks       |
| 2     | Compaction Quality + API            | Windowed + alpha compaction, structured output, Memory Query API     | 2 weeks         |
| 3     | Semantic Embeddings + Polish        | Injectable embeddings, Working tier lifecycle, full tests            | 2–3 weeks       |
| 4     | Advanced (H-Mem)                    | Entity KG, hybrid retrieval, progressive compaction                  | Ongoing         |

---

**This document lives in the `memory` package** and should be the single source of truth for prioritized improvements.
