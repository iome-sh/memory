# Temporal Memory Kernel Roadmap

**Repository**: `github.com/iome-sh/memory`  
**Scope**: Temporal features **inside this package** (Palace kernel), not aion host product surfaces  
**Serial**: s587 (docs); K1 = s586 / v1.5.2; K2 first slice = s611 / v1.5.3; K4 first slice = s616 / v1.5.4; A2 first slice = s619 / v1.5.5; A3 first slice = s632 / v1.5.6; A2 hop ranking = s1067 / v1.5.7 continuum; hop ranking residual honesty = s1278  
**Last Updated**: 2026-08-05

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
| **K1** | **Shipped** (s586 / v1.5.2) | `SearchMemoryWithOptions` session/time filters + temporal re-rank |
| **K2** | **Partial shipped** (s611 / v1.5.3) | `ListMemoryWithOptions` timeline API + tag helpers; full FS event-time index residual |
| **K3** | Planned | Optional Qwen3-0.6B 1024-d embedding profile (dual-path with host workers) |
| **K4** | **Partial shipped** (s616 / v1.5.4; A3 supersession s632 / v1.5.6) | Facts-as-of / validity window (`ListFactsAsOf`, `EntryValidAt`) + entity-key supersession (`SupersedeEntityFacts`); not full temporal KG |
| **A2** | **Partial shipped** (s619 / v1.5.5; hop ranking s1067; residual honesty s1278) | Multi-hop / associative retrieval lite over EntityGraph + entry entity tags + hop-distance ranking lite; not full Zep KG; not product Memory GA |
| **A3** | **Partial shipped** (s632 / v1.5.6) | Fact supersession lite: close prior open validity windows for an entity key on write; not NLP contradiction / full KG |

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

## K1 — Search options — **Shipped** (s586 / v1.5.2)

**Goal**: First-class filtered temporal retrieval without callers reimplementing post-filters.

Shipped surface:

```go
type SearchMemoryOptions struct {
    SessionID      string
    TimeFrom       *time.Time // inclusive event time (entryEventTime)
    TimeTo         *time.Time // inclusive
    AsOf           *time.Time // optional; when set, EntryValidAt filter before Limit (s616)
    Limit          int        // default 10
    Tier           *MemoryTier
    QueryVec       []float32
    ReRankTemporal bool
}

func (ps *PalaceStore) SearchMemoryWithOptions(query string, opts SearchMemoryOptions) []MemoryEntry
```

### Acceptance (met)

- Filter by `SessionID` when set
- Filter by event-time window via `entryEventTime` (`Timestamp` else `CreatedAt` else `LastAccessed`); inclusive bounds
- Optional `ReRankTemporal` after vector/keyword scoring (`CalculateRelevanceScore`)
- Filters apply **before** Limit (underfill class)
- Backward compatible: `SearchMemory` remains a thin wrapper
- Tests: session isolation, window edges, re-rank order, wrapper parity

---

## K2 — Timeline list & tag helpers — **Partial shipped** (s611 / v1.5.3)

**Goal**: First-class event-time ordered listing (timeline) with session/time/tag/query filters applied before Limit; light tag helpers. Full FS event-time index remains residual.

### Shipped (s611)

```go
type ListMemoryOptions struct {
    SessionID       string
    TimeFrom        *time.Time // inclusive on entryEventTime
    TimeTo          *time.Time // inclusive
    Tag             string     // exact match TemporalTags or Content.Tags
    TagPrefix       string     // strings.HasPrefix on either tag set
    Query           string     // case-insensitive substring Summary/Full/OriginalText
    Limit           int        // default 50 when <= 0
    Tier            *MemoryTier
    IncludeArchival bool       // when Tier==nil, also include Archival
    Ascending       bool       // false = newest first (default)
}

func (ps *PalaceStore) ListMemoryWithOptions(opts ListMemoryOptions) []MemoryEntry
func EntryHasTag(e MemoryEntry, tag string) bool
func EntryHasTagPrefix(e MemoryEntry, prefix string) bool
```

Order of operations: collect candidates → session → time → tag filters → query → sort by `entryEventTime` → limit.

Default tiers when `Tier == nil`: Working + Contextual + Semantic (**exclude Archival** unless `IncludeArchival`).

### Residual / later within K2

- Full **event-time index** (avoid O(n) FS scan of tier JSON for every windowed list)
- Optional secondary indexes for tags if FS cost becomes the bottleneck

**Honesty**: FS Palace remains **O(n)** over tier files for list/search. Index is additive and optional; this slice does not invent multi-tenant or product Memory GA.

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

## K4 — Entity validity windows / temporal KG — **Partial shipped** (s616 / v1.5.4)

**Goal**: First-class **as-of validity** (bi-temporal lite) so hosts that write `valid_from:` / `valid_until:` TemporalTags can list and search facts valid at a point in time.

### Shipped (s616) — facts-as-of first slice

```go
func ParseValidityWindow(e MemoryEntry) (from, until *time.Time)
func EntryValidAt(e MemoryEntry, asOf time.Time) bool

type FactsAsOfOptions struct {
    AsOf            time.Time // zero = Now UTC
    Query           string
    SessionID       string
    Entity          string    // entity: TemporalTags filter
    Limit           int       // default 50
    Tier            *MemoryTier
    IncludeArchival bool
}

func (ps *PalaceStore) ListFactsAsOf(opts FactsAsOfOptions) []MemoryEntry
// SearchMemoryOptions.AsOf *time.Time — filter !EntryValidAt before Limit
```

#### Validity rules (`EntryValidAt`)

| Case | Rule |
|------|------|
| Zero `asOf` | Treated as `time.Now().UTC()` |
| `valid_from` set | Invalid if `asOf.Before(from)` (inclusive start) |
| `valid_until` set | Invalid if `!asOf.Before(until)` — **exclusive end** (`asOf == until` is invalid) |
| No validity tags | “Known by asOf”: valid if `entryEventTime` is zero **or** `!entryEventTime.After(asOf)` |

Tag format (host-written, e.g. aion `applyTemporalToEntry`): `valid_from:<RFC3339>`, `valid_until:<RFC3339>`.

#### ListFactsAsOf

- Default tiers: Working + Contextual + Semantic (+ Archival if `IncludeArchival`)
- Filters (session, entity, query, validity) apply **before** Limit (underfill class)
- Ordering: **Semantic first**, then event time descending within rank
- Entity filter: substring on `entity:` tags when value has no `:`; exact `entity:type:id` when value contains `:`

### Honesty / non-goals (still open)

This is **bi-temporal lite** (validity window on entries via tags), **not**:

- Full Graphiti-style dual clocks (transaction time + validity time as first-class stores)
- Temporal knowledge graph with edge validity
- Multi-tenant product Memory GA

### Shipped (s632) — A3 supersession first slice (K4 write path)

When a newer fact for the same **entity key** is written, close prior open validity windows by setting `valid_until` (exclusive end, same semantics as `EntryValidAt`). Entries are not deleted.

```go
// SupersedeEntityFacts finds entries matching entityKey (EntryEntityKeys /
// entity: tags) that are still EntryValidAt(asOf), and writes valid_until=asOf.
// Empty entityKey is a no-op. Returns count of updated entries.
func (ps *PalaceStore) SupersedeEntityFacts(entityKey string, asOf time.Time) (int, error)

// WriteAndSupersede writes entry first (stamps valid_from=now when unset), then
// supersedes each key excluding the new entry's ID.
func (ps *PalaceStore) WriteAndSupersede(entry MemoryEntry, supersedeKeys []string) error
```

#### Rules

| Concern | Behavior |
|---------|----------|
| Entity match | `EntryEntityKeys` contains key after lower-case trim normalization |
| Skip self | `WriteAndSupersede` excludes the newly written entry ID |
| Open only | Only entries currently `EntryValidAt(asOf)` are updated |
| Tag write | Add/replace `valid_until:<RFC3339>`; preserve `valid_from` and other tags |
| Empty key | No-op (0, nil) |

#### Honesty / non-goals (A3)

This is **competitive lite supersession** (explicit entity keys + validity tags), **not**:

- Automatic NLP contradiction detection
- Full Zep dual-clock knowledge graph
- Silent host-wide supersession without caller-supplied keys

Residual for later K4 / A3 slices (when product demand is explicit):

- Temporal edges on relations for KG-style recall
- Consistency with compaction and fact-augmented ingest beyond host tags
- Optional indexes if O(n) FS scans become the bottleneck
- Auto-extract entity keys from new writes (still explicit keys in this slice)

---

## A2 — Multi-hop / associative retrieval — **Partial shipped** (s619 / v1.5.5; hop ranking s1067; residual honesty s1278)

**Goal**: Competitive multi-hop lite over the existing EntityGraph (`AddEntityRelationship` / `GetRelatedEntities`) and entry Relations / entity tags — not a full Zep / Graphiti knowledge graph.

### Shipped (s619) — multi-hop first slice

```go
type MultiHopOptions struct {
    SeedEntity        string     // starting entity key (prefer exact graph node keys)
    SeedQuery         string     // optional: derive seeds from SearchMemoryWithOptions hits
    MaxHops           int        // default 2; clamp 1..4 (hop 0 = seed)
    Limit             int        // default 20; AFTER expansion + collect + filters
    SessionID         string
    AsOf              *time.Time // optional EntryValidAt filter
    Tier              *MemoryTier
    IncludeArchival   bool
    QueryVec          []float32  // pass-through when seeding via SeedQuery
    PreferShorterHops *bool      // default true (nil); false = legacy seed-match-first
}

func (ps *PalaceStore) ExpandRelatedEntities(seed string, maxHops int) []string
func (ps *PalaceStore) ExpandRelatedEntitiesHops(seed string, maxHops int) map[string]int // entity → min hop
func (ps *PalaceStore) MultiHopRetrieve(opts MultiHopOptions) []MemoryEntry
func EntryEntityKeys(e MemoryEntry) []string // entity: / subject: / RelatedConcepts
```

#### Order of operations (`MultiHopRetrieve`)

1. Resolve seeds (`SeedEntity` and/or entity keys from `SeedQuery` search hits via `EntryEntityKeys`)
2. `ExpandRelatedEntitiesHops` BFS for each seed over `GetRelatedEntities` up to `MaxHops` (min hop across seeds)
3. Collect entries from default tiers matching any expanded entity (`TemporalTags` `entity:*`, `Content.Tags`, `Relations.RelatedConcepts`); assign **min hop** among matched entity keys
4. Optional `SessionID` + `AsOf` filters **before** Limit (underfill class)
5. Sort (default): **lower hop first** (seed = hop 0), then event time descending within hop; opt-out via `PreferShorterHops=false` → legacy seed-match first then event time
6. Limit

Default tiers when `Tier == nil`: Working + Contextual + Semantic (+ Archival if `IncludeArchival`).

`AddEntityRelationship` ensures `BaseDir/relations` exists before writing the graph file.

### Shipped (s1067 / v1.5.7 continuum) — hop-distance ranking lite

Path-aware ranking lite: prefer shorter BFS hop distance from seed when ordering multi-hop hits. Improves host TUI / MCP multi-hop recall quality without a full path-scoring graph.

- `ExpandRelatedEntitiesHops` returns entity → min hop (seed at 0)
- Entry hop = min hop among matched expanded entity keys
- Still **not** typed-edge weights, embedding-guided walks, or Zep/Graphiti path scores

### Residual honesty pin (s1278) — closed residual-honest

Free eng residual pin for A2 hop-distance ranking honesty (memory serial **s1278**; continuum with aion free eng floor **s1276+** / peer **s1277**). Documents:

- `PreferShorterHops` default **true**; explicit false = legacy seed-match-first (does not prefer shorter hops)
- multi-hop lite · not full Zep/Graphiti path scoring · not full graph RAG · not product Memory GA · kernel-only
- TUI related `hop_distance` display: host surface mention only

Canonical residual SSOT: [`operations/multi-hop-hop-distance-ranking-residual.md`](./operations/multi-hop-hop-distance-ranking-residual.md).

### Honesty / non-goals (still open)

This is **multi-hop lite** (BFS on a simple directed adjacency map + tag collect + hop-distance sort), **not**:

- Full Zep / Graphiti temporal knowledge graph with typed edges and edge validity
- Community detection, full path scoring, or embedding-guided graph walk
- Multi-tenant product Memory GA

Residual for later A2 slices:

- Bidirectional / typed relation edges
- ~~Path-aware ranking (prefer shorter hops)~~ — **done** s1067 (hop-distance ranking lite; not full path scoring); residual honesty **s1278**
- Indexes if O(n) FS scans + graph BFS become the bottleneck

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

1. ~~Finish **K1** (`SearchMemoryWithOptions` + tests)~~ — **done** s586 / v1.5.2  
2. **K2** first slice (`ListMemoryWithOptions` + tag helpers) — **done** s611 / v1.5.3; residual: full FS event-time index when scans become the bottleneck  
3. **K4** first slice (facts-as-of / validity window) — **done** s616 / v1.5.4; residual: temporal edges / full dual-clock KG  
4. **A2** first slice (multi-hop / associative retrieval) — **done** s619 / v1.5.5; hop-distance ranking lite **done** s1067; residual honesty pin **s1278**; residual: typed edges / full path scoring / full Zep KG  
5. **A3** first slice (entity-key fact supersession) — **done** s632 / v1.5.6; residual: auto entity extract / NLP contradiction / full dual-clock KG  
6. **K3** only when a concrete consumer needs 1024-d Qwen3 locally (keep BGE-small default until then; no silent flip)

---

## Versioning notes

- Kernel temporal APIs should remain backward compatible within major module versions where practical  
- New options structs and methods are preferred over breaking `SearchMemory` signatures  
- Embedding dimension changes require collection recreation when using Qdrant; document in release notes  
- **v1.5.2**: K1 `SearchMemoryWithOptions`  
- **v1.5.3**: K2 partial `ListMemoryWithOptions` + `EntryHasTag` / `EntryHasTagPrefix`  
- **v1.5.4**: K4 partial facts-as-of (`ListFactsAsOf`, `EntryValidAt`, `SearchMemoryOptions.AsOf`)  
- **v1.5.5**: A2 partial multi-hop / associative (`MultiHopRetrieve`, `ExpandRelatedEntities`, `EntryEntityKeys`)  
- **v1.5.6**: A3 / K4 partial fact supersession (`SupersedeEntityFacts`, `WriteAndSupersede`)  
- **v1.5.7 continuum**: A2 residual hop-distance ranking (`ExpandRelatedEntitiesHops`, `PreferShorterHops` default true) — s1067; residual honesty pin **s1278**  
- BGE-small-en-v1.5 (384-d) remains the default ONNX profile; Qwen3 1024-d is K3 residual  

---

*s587 roadmap anchor; K1 shipped s586/v1.5.2; K2 partial shipped s611/v1.5.3; K4 partial shipped s616/v1.5.4; A2 partial shipped s619/v1.5.5; A3 partial shipped s632/v1.5.6; A2 hop ranking s1067 (v1.5.7 continuum); hop ranking residual honesty s1278.*

