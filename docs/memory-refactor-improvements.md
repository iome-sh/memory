# Memory Package - Improvements & Roadmap

**Repository**: `github.com/iome-sh/memory`
**Purpose**: Prioritized improvement plan for the standalone hierarchical memory package (Palace).
**Last Updated**: May 2026

This document reorganizes improvements by **highest business/engineering impact first**, grouped into logical rollout categories suitable for a standalone, importable Go package.

---

## Recently Completed (as of latest update)

- **Vector Similarity Search** — `VectorStore.SearchSimilar`, `SearchByText`, and `SearchResult` with payload support implemented.
- **Vector Integration via Callbacks** — `VectorStoreCallback` added to `PerformCompaction`.
- **Temporal Window Filtering** in compaction (H-Mem inspired).

---

## 1. High Impact: Core Stability & Reliability (Phase 1)

### 1.1 Real Temporal Decay + Unified Relevance Scoring

**Priority**: Critical

`calculateTemporalDecay` is still a no-op.

**Actions**:
- Implement exponential decay + combine with recency into `CalculateRelevanceScore()`.
- Apply score in listing and compaction candidate selection.

### 1.2 Robust Error Handling & Observability

**Priority**: High

**Actions**:
- Consistent error returns from `PalaceStore`.
- Structured logging.
- Expose `MemoryStats`.

### 1.3 Activate Versioning

**Priority**: High

**Actions**:
- Wire up `archiveToVersions` in `Write` path.
- Proper version incrementing.

---

## 2. High Impact: Vector & Retrieval Quality (Phase 1–2)

### 2.1 Robust Vector Similarity Search **[Completed]**

`VectorStore.SearchSimilar` + `SearchByText` with payload support now available.

**Next**:
- Add `SearchMemory` helper on `PalaceStore` for hybrid search.

### 2.2 Optional Vector Integration via Callbacks **[Completed]**

`VectorStoreCallback` supported in compaction.

**Next**:
- Auto-attach vector storage when `VectorStore` is provided to `PalaceStore`.

### 2.3 Multi-Factor Scoring (s + t + r)

**Priority**: High

Implement combined semantic + temporal + robustness scoring (H-Mem style).

---

## 3. High Impact: Compaction & Knowledge Evolution (Phase 1–2)

### 3.1 Temporal Window + Alpha-Constrained Compaction **[Partially Complete]**

Temporal window filtering is implemented.

**Next**:
- Add optional α similarity threshold before consolidation.

### 3.2 Structured Compaction Output + Verification

**Priority**: High

Improve action parsing robustness + add verification step.

### 3.3 Persist Compaction Metrics

Track before/after scores in `CompactionConfig`.

---

## 4. Medium Impact: API & Usability (Phase 2)

### 4.1 Memory Query API

Add `Search(...)` method with hybrid support.

### 4.2 Working Tier Lifecycle

Promotion and eviction logic for Working tier.

### 4.3 PalaceStore Configuration

Introduce `PalaceConfig` for cleaner setup.

---

## 5. Medium Impact: Embeddings & Semantic Quality (Phase 2–3)

### 5.1 Replace Simple Embedding

Make embedding generation pluggable via `EmbeddingFunc`.

### 5.2 Embedding Versioning

Support multiple embedding generations.

---

## 6. Foundation: Architecture, Portability & Testing (Ongoing)

- Cross-platform builds
- Comprehensive unit + integration tests
- H-Mem long-term ideas (entity graph, hybrid retrieval)

---

## Recommended Rollout Order

| Phase | Focus                              | Key Items                                      | Effort    |
|-------|------------------------------------|------------------------------------------------|-----------|
| 1     | Stability + Vector                 | Decay, errors, versioning, vector search       | 1–2 weeks |
| 2     | Compaction Quality + API           | Alpha compaction, structured output, Search API| 2 weeks   |
| 3     | Semantic + Polish                  | Pluggable embeddings, Working tier, tests      | 2–3 weeks |

---

**This is the canonical improvement roadmap for the `memory` package.**
