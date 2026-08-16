# Memory Package - Improvements & Roadmap

**Repository**: `github.com/iome-sh/memory`
**Purpose**: Prioritized improvement plan for the standalone hierarchical memory package (Palace).
**Last Updated**: 2026-07-23

This document reorganizes improvements by **highest business/engineering impact first**, grouped into logical rollout categories suitable for a standalone, importable Go package.

**Temporal kernel track**: see the dedicated roadmap → [`docs/temporal-memory-kernel-roadmap.md`](./temporal-memory-kernel-roadmap.md) (K0 shipped … K4 later). Do not conflate this library with aion product Memory add-on GA.

---

## Recently Completed (as of 2026-07-23)

- **Vector Similarity Search** — `VectorStore.SearchSimilar`, `SearchByText`, and `SearchResult` with payload support implemented.
- **Vector Integration via Callbacks** — `VectorStoreCallback` added to `PerformCompaction`.
- **Temporal Window Filtering** in compaction (H-Mem inspired).
- **Real Temporal Decay** — `CalculateTemporalDecay` (exponential forgetting curve) + `CalculateRelevanceScore` / `CalculateRecencyBoost` used in tier listing.
- **Multi-Factor Scoring (s + t + r)** — `MultiFactorScore` combines semantic similarity, temporal decay, and usage robustness.
- **Hybrid `SearchMemory`** — keyword + vector re-rank via pluggable `EmbeddingFunc` / batch embed on `PalaceStore`.
- **Turn ingest** — `IngestTurn` with `Timestamp`, `SessionID`, fact-augmented child writes.
- **MemoryEntry temporal fields** — `Timestamp`, `SessionID`, `TemporalTags` (+ turn granularity fields).
- **Pluggable ONNX embeddings** — hugot GoMLX default; ORT optional (CUDA/CoreML); BGE-small 384-d production default.
- **PalaceConfig**, Working-tier eviction, versioning path, LongMemEval harness/gates.

---

## 1. High Impact: Core Stability & Reliability (Phase 1)

### 1.1 Real Temporal Decay + Unified Relevance Scoring **[Completed — K0]**

`CalculateTemporalDecay`, `CalculateRelevanceScore`, and related helpers are implemented and applied in `ListEntriesInTier`.

Further temporal **retrieval** work (session/time filters, temporal re-rank options) lives under **K1** in [`temporal-memory-kernel-roadmap.md`](./temporal-memory-kernel-roadmap.md).

### 1.2 Robust Error Handling & Observability

**Priority**: High

**Actions**:
- Consistent error returns from `PalaceStore`.
- Structured logging.
- Expose `MemoryStats`.

### 1.3 Activate Versioning **[Largely complete]**

Versioning and archive paths exist on the write path; keep hardening edge cases and observability.

---

## 2. High Impact: Vector & Retrieval Quality (Phase 1–2)

### 2.1 Robust Vector Similarity Search **[Completed]**

`VectorStore.SearchSimilar` + `SearchByText` with payload support now available.

### 2.2 Optional Vector Integration via Callbacks **[Completed]**

`VectorStoreCallback` supported in compaction; vector attach via `PalaceConfig` / store wiring as documented in README.

### 2.3 Multi-Factor Scoring (s + t + r) **[Completed — K0]**

`MultiFactorScore` implements combined semantic + temporal + robustness scoring (H-Mem style).

**Next (K1)**: wire multi-factor / event-time re-rank into an options-based search API (`SearchMemoryWithOptions`) — see temporal kernel roadmap.

### 2.4 Hybrid SearchMemory **[Completed — K0]**

`PalaceStore.SearchMemory` provides hybrid keyword + vector retrieval.

**Next (K1)**: `SearchMemoryWithOptions` for session/time filters + temporal re-rank.

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

### 4.1 Memory Query API **[Partially Complete — K0/K1]**

`SearchMemory` hybrid helper is shipped. Session/time-filtered options API is **K1** (s586 peer).

### 4.2 Working Tier Lifecycle **[Completed baseline]**

Promotion/eviction helpers for Working tier exist (`EvictWorkingTier` and related).

### 4.3 PalaceStore Configuration **[Completed]**

`PalaceConfig` is the supported setup path (base dir, embeddings, vector URL/collection).

---

## 5. Medium Impact: Embeddings & Semantic Quality (Phase 2–3)

### 5.1 Replace Simple Embedding **[Completed]**

`EmbeddingFunc` / batch embed are pluggable; production pure-Go ONNX via hugot (BGE-small 384-d default).

### 5.2 Embedding Versioning

Support multiple embedding generations.

### 5.3 Optional Qwen3-0.6B 1024-d preset **[Planned — K3]**

Optional local profile; aion host may prefer embed-worker path. See [`temporal-memory-kernel-roadmap.md`](./temporal-memory-kernel-roadmap.md) § K3 (dual-path honesty).

---

## 6. Foundation: Architecture, Portability & Testing (Ongoing)

- Cross-platform builds
- Comprehensive unit + integration tests
- H-Mem long-term ideas (entity graph, hybrid retrieval)
- **K2**: timeline helpers + tag query helpers shipped (s611); in-memory meta index (s1066); durable snapshot shipped (#44); incremental event-time index residual
- **K4 (later / non-goal now)**: entity validity windows / temporal KG — not scheduled on the current kernel track

---

## Recommended Rollout Order

| Phase | Focus | Key Items | Effort |
|-------|--------|-----------|--------|
| ~~1~~ | ~~Stability + Vector~~ | ~~Decay, multi-factor, vector search, SearchMemory~~ | **Done (K0)** |
| K1 | Temporal retrieval options | `SearchMemoryWithOptions`, session/time filters, temporal re-rank | s586 peer |
| K2 | Timeline / tags | `ListMemoryWithOptions` + meta index (s1066) + durable snapshot (#44); incremental index residual | Partial |
| 2–3 | Compaction polish + observability | Alpha compaction, structured output, MemoryStats | Ongoing |
| K3 | Optional dense embed preset | Qwen3-0.6B 1024-d opt-in (not default flip) | Planned |
| K4 | Temporal KG | Entity validity windows — **non-goal for now** | Later |

---

**Canonical improvement backlog for the `memory` package.**  
**Canonical temporal kernel phases:** [`docs/temporal-memory-kernel-roadmap.md`](./temporal-memory-kernel-roadmap.md).
