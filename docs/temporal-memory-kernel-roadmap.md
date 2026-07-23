# Temporal Memory Kernel Roadmap

**Repository**: `github.com/iome-sh/memory`  
**Scope**: Temporal features **inside this package** (Palace kernel), not aion host product surfaces  
**Serial**: s587 (docs); K1 peer work tracked as s586  
**Last Updated**: 2026-07-23

This is the standalone roadmap for temporal memory capabilities in the hierarchical agent memory library (Palace). It deliberately excludes aion Control Plane / mesh add-on GA claims, multi-tenant product packaging, and host MCP/sidecar surfaces.

---

## Honesty boundaries

| Concern | This package (`memory`) | Host (`aion` and product surfaces) |
|---------|-------------------------|-------------------------------------|
| Storage model | Single-tenant filesystem Palace (`PalaceStore` + tier dirs) | Multi-tenant isolation, org/agent paths, collection naming |
| API surface | Go types + `PalaceStore` methods | MCP, HTTP sidecar, mesh streams, console UX |
| Embeddings | Pluggable `EmbeddingFunc` / ONNX (local) | May prefer remote embed workers or fleet models |
| Product Memory add-on | **Not** claimed GA by this repo | Host owns packaging, quotas, and GA gates |

Do **not** treat kernel completeness as product Memory GA. Do **not** invent multi-tenant guarantees in this library: callers must enforce tenant boundaries above `PalaceStore`.

---

## Phase overview

| Phase | Status | Focus |
|-------|--------|--------|
| **K0** | **Shipped** | Temporal fields, decay, multi-factor score, `IngestTurn`, hybrid `SearchMemory` |
| **K1** | In flight (s586 peer) | `SearchMemoryWithOptions` session/time filters + temporal re-rank |
| **K2** | Planned | Explicit event-time index / timeline helper API; tag query helpers |
| **K3** | Planned | Optional Qwen3-0.6B 1024-d embedding profile (dual-path with host workers) |
| **K4** | Later / non-goal now | Entity validity windows / temporal knowledge graph |

---

## K0 — Shipped (kernel baseline)

Core temporal primitives are present and used by LongMemEval harnesses and production-style ingest.

### MemoryEntry temporal fields

On `MemoryEntry` (`memory.go`):

- `Timestamp` — event time for temporal reasoning (distinct from `CreatedAt` / `UpdatedAt`)
- `SessionID` — multi-session grouping
- `TemporalTags` — cycle / calendar-style tags (see `PopulateTemporalTags`)
- Related turn granularity: `TurnID`, `ExtractedFacts`, `Keyphrases`, `OriginalText`

### Scoring & decay

- `CalculateTemporalDecay(entry)` — H-Mem-style exponential forgetting (hours since last access)
- `CalculateRecencyBoost` / `CalculateRelevanceScore` — recency + decay + usage combined with score impact
- `MultiFactorScore(entry, queryVec)` — semantic + temporal + robustness blend (s + t + r)

`ListEntriesInTier` sorts by `CalculateRelevanceScore` (descending).

### Ingest & search

- `IngestTurn` — primary turn ingest path (defaults timestamp/IDs, fact-augmented child writes, session/timestamp propagation)
- `SearchMemory(query, tier, limit, vec)` — hybrid keyword + vector re-rank when a query embedding is supplied; embedding path uses configured `EmbeddingFunc` / batch embed

### Compaction touchpoints

- Temporal window filtering in compaction (`filterByTemporalWindow`) — H-Mem-inspired windowing on the compaction path

**K0 is not “temporal retrieval complete.”** Session/time **filters** and explicit temporal **re-rank options** are K1; a dedicated event-time **index** is K2.

---

## K1 — Search options (s586 peer)

**Goal**: First-class filtered temporal retrieval without callers reimplementing post-filters.

Planned surface (names indicative; implement on this package):

```go
// Illustrative — not yet the committed public API.
type SearchMemoryOptions struct {
    SessionID string
    Since     time.Time // inclusive event-time lower bound (Timestamp)
    Until     time.Time // exclusive or inclusive — document at implement time
    // Temporal re-rank: blend MultiFactorScore / event-time proximity with vector/keyword rank
}

func (ps *PalaceStore) SearchMemoryWithOptions(
    query string,
    tier *MemoryTier,
    limit int,
    vec []float32,
    opts SearchMemoryOptions,
) []MemoryEntry
```

### Acceptance intent

- Filter by `SessionID` when set
- Filter by event-time window on `Timestamp` (with clear zero-value semantics)
- Optional temporal re-rank after vector/keyword scoring (prefer event-time + decay over pure access-time-only decay when `Timestamp` is set)
- Backward compatible: existing `SearchMemory` remains the simple path
- Tests under package unit tests (session isolation, window edges, re-rank order)

**Peer serial**: s586 tracks implementation; this document (s587) is the kernel roadmap anchor.

---

## K2 — Event-time index & tag helpers (planned)

**Goal**: Make timeline queries efficient and tag queries first-class without scanning all tier JSON files for every windowed read.

### Planned work

- Explicit **event-time index** or timeline helper API on `PalaceStore` (e.g. list/scan by `Timestamp` range, optional session scope)
- **Tag query helpers** for `TemporalTags` / content tags (exact or prefix; document FS cost vs future index)
- Document complexity: FS Palace remains O(n) without index; index is additive and optional

### Non-goals for K2

- Cross-tenant indexes
- Distributed timeline stores
- Full temporal KG (see K4)

---

## K3 — Qwen3-0.6B 1024-d embedding profile (planned)

**Goal**: Optional embedding **preset** for denser local vectors, without forcing host architecture.

### Dual-path honesty

| Path | Who owns it | Notes |
|------|-------------|--------|
| **Library preset** | `memory` package | Optional ONNX/hugot profile: Qwen3-0.6B → **1024-d**; caller sets `EmbeddingFunc` + collection dim |
| **Host worker path** | aion (and fleet) | Host may prefer remote/embed-worker pipelines, model routing, or different dims; **must not** assume library default is Qwen3 |

Default production ONNX path today is **BGE-small-en-v1.5 (384-d)** (see README / `BGESmallEmbeddingDim`). Hash fallback remains **768-d** when ONNX is unset.

K3 must:

- Ship as **opt-in** preset (env or constructor), not silent default flip that breaks existing 384-d collections
- Document dimension mismatches and re-index requirements
- State clearly that aion hosts may ignore this preset and inject their own `EmbeddingFunc`

---

## K4 — Entity validity windows / temporal KG (later)

**Status**: Non-goal for the current kernel track.

Possible future direction (research / product-dependent):

- Entity validity intervals (valid-from / valid-to)
- Temporal edges on relations for knowledge-graph style recall
- Consistency with compaction and fact-augmented ingest

Do not schedule K4 against current s586/s587 delivery. Host product “Memory” branding must not claim K4 until separately chartered.

---

## How this relates to other docs

| Doc | Role |
|-----|------|
| [memory-refactor-improvements.md](./memory-refactor-improvements.md) | Broader package improvement backlog (stability, vectors, embeddings, RecMem-adjacent) |
| [recmem-integration-plan.md](./recmem-integration-plan.md) | RecMem density / phase-transition integration |
| `README.md` | Public usage, ONNX backends, LongMemEval harness |

**This file** is the canonical **temporal kernel** roadmap for `github.com/iome-sh/memory`.

---

## Suggested implementation order

1. Finish **K1** (`SearchMemoryWithOptions` + tests) — unblocks host temporal retrieve filters without host-side hacks  
2. **K2** timeline/index helpers when FS scans become the bottleneck in real sessions  
3. **K3** only when a concrete consumer needs 1024-d Qwen3 locally (keep BGE-small default until then)  
4. Defer **K4** until entity-graph product demand is explicit  

---

## Versioning notes

- Kernel temporal APIs should remain backward compatible within major module versions where practical  
- New options structs and methods are preferred over breaking `SearchMemory` signatures  
- Embedding dimension changes require collection recreation when using Qdrant; document in release notes  

---

*s587 — docs-only kernel roadmap. Implementation ships under subsequent serials (K1 = s586 peer).*

<!-- ci poke 20260723T073104Z -->
