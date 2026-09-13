# Temporal Memory Kernel Roadmap

**Repository:** [`github.com/iome-sh/memory`](https://github.com/iome-sh/memory)  
**Scope:** Temporal features **inside this package** (`PalaceStore`), not MCP/TUI hosts.  
**As of:** 2026-09-13 · tagged **v1.5.12** · unreleased on `main`: unique-entity / dated-event / latest-value / skip-vector / clothing-only N · [#126](https://github.com/iome-sh/memory/pull/126) T4 `ListFactsAsOf` tests · [#127](https://github.com/iome-sh/memory/pull/127) numbered clothes · [#128](https://github.com/iome-sh/memory/pull/128) T2 list-latency bench · [#129](https://github.com/iome-sh/memory/pull/129) T6 text-date delta · [#131](https://github.com/iome-sh/memory/pull/131) search/count via meta index · [#134](https://github.com/iome-sh/memory/pull/134) `2a3257c` T4-perf `ListFactsAsOf` via meta index · [#132](https://github.com/iome-sh/memory/pull/132) T7 generalized unique-entity · [#135](https://github.com/iome-sh/memory/pull/135) Wave I n=12 · [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a` unique-entity restaurant clusters

This is the canonical temporal plan for the hierarchical agent memory library. Callers own tenancy above `BaseDir`. Companion hosts ([iomesh-tui](https://github.com/iome-sh/iomesh-tui) **v1.3.7**, [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) **v0.4.2**) are optional.

Related: [TTFH walking skeleton](./TTFH.md) · [LongMemEval methodology](./LONGMEMEVAL.md) · [locked mixed baseline](./LONGMEMEVAL_BASELINE.md) · [package improvements](./memory-refactor-improvements.md)

---

## Re-evaluation (original plan vs v1.5.12)

The original document (last updated 2026-08-05) sequenced **K0–K4** plus **A2/A3**. That surface is in the tree. After v1.5.7: index patching, palace file modes, retrieve default tiers, private `source_hint`, optional ONNX persist, TTFH, locked LongMemEval slices, then **T1** multi-session retrieve (v1.5.12).

| Original phase | Original intent | Status | Still open |
|----------------|-----------------|--------|------------|
| **K0** | Event time, decay, `IngestTurn`, hybrid search | **Shipped** | — |
| **K1** | Session/time search + temporal re-rank | **Shipped** (v1.5.2) + T1 `SessionIDs` / skip-vector on count/temporal ([#122](https://github.com/iome-sh/memory/pull/122)) | — |
| **K2** | Event-time timeline list + FS index | **Mostly shipped** — list API v1.5.3; durable `indexes/event-time.json`; in-memory patch on Write/unlink (v1.5.8). First list / stamp mismatch still rebuilds. T2 **bench** shipped ([#128](https://github.com/iome-sh/memory/pull/128) `376dd69`). | btree / tag secondaries **parked** — [#131](https://github.com/iome-sh/memory/pull/131) `ea94317` measure: rebuild is not the limiter |
| **K3** | Optional Qwen3-0.6B **1024-d** | **Not started** (not blocking). Default ONNX **BGE-small 384-d**; MiniLM in-tree fallback; `PersistEmbeddings` default off. | Consumer-driven (**T5**) |
| **K4** | Facts-as-of / validity windows | **Shipped lite** (v1.5.4) — `ListFactsAsOf`, `EntryValidAt`, `SearchMemoryOptions.AsOf`. Compaction SUMMARIZE/MERGE stamp `valid_from`. **T4 tests shipped** ([#126](https://github.com/iome-sh/memory/pull/126) `3fcbe4d`): after MERGE/SUMMARIZE the product is still `ListFactsAsOf`-visible; ARCHIVE tier-move stays valid at now; no invented `valid_until`. **T4-perf shipped** ([#134](https://github.com/iome-sh/memory/pull/134) `2a3257c`): `ListFactsAsOf` via list meta index (no Limit); `EntryValidAt` / entity still after load. | Temporal **edges** (**T3**); dual-clock **store** (**T8**, parked). btree still gated. |
| **A2** | Multi-hop retrieve | **Shipped lite** (v1.5.5–1.5.7) | Typed / bidirectional edges (**T3**) |
| **A3** | Fact supersession | **Shipped lite** (v1.5.6) + latest-value **retrieve** evidence ([#122](https://github.com/iome-sh/memory/pull/122)) | Auto entity extract; NLP contradiction |

**Eval evidence (unpublished, not a README number, not official V1):** locked mixed LongMemEval, judge `gpt-4o-2024-08-06`, reader `gpt-4o-mini`, isolated palace per embed.

| Wave | Kernel | hash | MiniLM | BGE | Notes |
|------|--------|------|--------|-----|-------|
| A pre-T1 | `cb93b08` | 9/12 MS 0/2 | 10/12 0/2 | 10/12 0/2 | Single `SessionID` |
| E T1 done-when | `f99c140` [#111](https://github.com/iome-sh/memory/pull/111) | 11/12 MS 1/2 | **12/12 2/2** | 11/12 1/2 | MiniLM+BGE left 0/2 |
| F clothing clusters | `aa64dbc` [#115](https://github.com/iome-sh/memory/pull/115) | 11/12 **2/2** | **12/12 2/2** | **12/12 2/2** | Clothes pass; hash temporal miss |
| G dated events | `b02abaf` [#119](https://github.com/iome-sh/memory/pull/119) | 11/12 1/2 | 11/12 1/2 | 11/12 1/2 | Temporal first **pass all 3**; clothes reader summed 2 |
| H N-header | `895f255` [#120](https://github.com/iome-sh/memory/pull/120) | 11/12 1/2 | 10/12 0/2 | 11/12 1/2 | Clothes still 2; MiniLM projects overcount (`8 distinct`) |
| I #122+#124+#127+#129+#132 | `e094bec` | 10/12 1/2 | 10/12 1/2 | 10/12 1/2 | Numbered clothes still reader 2; MiniLM projects recovered (no N); KU `6aeb4375` 3 vs 4 all three. [#134](https://github.com/iome-sh/memory/pull/134) **not** this remesure. |
| n=60 v1.5.12 | `e90a82d` | 47/60 MS 8/10 | 48/60 MS 7/10 | **46/58** MS 6/8 | **Before [#119](https://github.com/iome-sh/memory/pull/119).** Clothes pass; kits/hours miss. BGE **incomplete** (timeouts `gpt4_59c863d7`, `e831120c`; 58/60 IDs). Do not treat 0.793 as comparable /60. |
| n=60 after #119+#122+#124+#129+#132 | `5154a76` | 49/60 MS 9/10 | 51/60 MS 9/10 | 47/60 MS 8/10 | **After [#119](https://github.com/iome-sh/memory/pull/119).** Complete 60/60. Kits/hours/plants pass all 3. Days `2a1811e2`/`2c63a862` pass; `08f4fc43` miss. KU `852ce960` $350k all 3; restaurants `6aeb4375` 3 vs 4 all 3. [#136](https://github.com/iome-sh/memory/pull/136) **not** this remesure. |

Clothes remesure after numbered bullets ([#127](https://github.com/iome-sh/memory/pull/127)): Wave I **miss all three** (retrieve `1. 2. 3.` + N=3; reader summed 2). n=12 after [#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#127](https://github.com/iome-sh/memory/pull/127)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132): Wave I hash/MiniLM/BGE **10/12** MS **1/2** (kernel `e094bec`). Projects pass all three. n=60 after [#119](https://github.com/iome-sh/memory/pull/119)+[#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132): kernel `5154a76` hash **49/60** MS **9/10** / MiniLM **51/60** MS **9/10** / BGE **47/60** MS **8/10** (complete 60/60). Do not invent further remesure scores.

**Walking skeleton:** `go run ./examples/ttfh_rca` — K0 + K1 session retrieve + K4 as-of. It does not exercise multi-session counts.

---

## Competitive landscape

Public positioning only. No invented latency, accuracy, or revenue figures. This kernel is **not** [MemPalace](https://github.com/MemPalace) / `mempalace` (an unrelated Python project; already noted in the README).

Palace write path is **raw ingest + heuristic facts** (`IngestTurn` children, named-pattern / clothing / unique-entity / dated-phrase helpers). There is **no LLM on write**. Several peers extract facts with an LLM on every `add` / episode.

| Project | Write path | Store | Temporal | Retrieve | Deploy |
|---------|------------|-------|----------|----------|--------|
| **Palace (this kernel)** | Raw ingest + heuristic facts; **no LLM on write** | Inspectable FS JSON palace (`cat` / `diff` / cite) | Entry `valid_from` / `valid_until` tags + event `Timestamp` (not bi-temporal edges) | Keyword-first; skip-vector on count / temporal-order; optional `QueryVec` otherwise | Local embeddable **MIT Go**. One process per palace root. |
| **Mem0** | LLM extract on `add()` (`infer=True` default: ADD / UPDATE / DELETE) | Vector store; optional graph on Platform | Update pipeline / recency, not entry validity tags | Vector (graph ranking on Platform) | Hosted platform + OSS |
| **Graphiti** | LLM extract entities / relations from episodes | Graph DB (Neo4j / FalkorDB / Neptune, …) | **Bi-temporal edges** (`valid_at` / `invalid_at` + transaction time) | Hybrid semantic + BM25 + graph walk | OSS; needs a graph database |
| **Zep** | Graphiti-style episodes (hosted Context Lake) | Hosted graph service | Bi-temporal (Graphiti model) | Hybrid (hosted) | Hosted |
| **Letta** | Agent **self-edits** memory blocks | Core blocks always in-context + archival | Runtime memory tiers, not a dual-clock KG | Agentic tool search | Agent **runtime**, not an embeddable Go library |
| **LangMem** | LLM memory-manager extract | LangGraph store / collections | Semantic / episodic / procedural types | Vector / collection lookup | Python library (LangGraph-native) |
| **Cognee** | Documents / tables / code → entities (typically LLM extract) | Graph + vector + relational | Graph RAG, not palace validity tags | Graph / vector / auto-route | Python OSS + cloud |

Use Palace when the job is an inspectable local ops record in Go with no required database. Use the others when you want LLM extraction on write, a hosted memory SaaS, a bi-temporal knowledge graph, or an agent runtime that edits its own blocks. See also the README “When to use this kernel” table.

---

## Shipped surface (keep these contracts)

### K0 — Baseline

On `MemoryEntry`: `Timestamp` (event time), `SessionID`, `TemporalTags`, turn fields (`TurnID`, `ExtractedFacts`, `Keyphrases`, `OriginalText`). `IngestTurn` stamps `source_hint=private` when the caller does not already supply a classifiable mesh or private source.

Scoring: `CalculateTemporalDecay`, `CalculateRecencyBoost`, `CalculateRelevanceScore`, `MultiFactorScore`.

Search: hybrid keyword-first + optional `QueryVec`. **Count and temporal-order queries skip vector scoring** (keyword + evidence assembly). Hash embeddings never persist.

Topology: **one process per palace root**. `writeMu` serializes `relations/entity-graph.json` and `indexes/event-time.json`. Flock is not shipped.

### K1 — Filtered search

```go
type SearchMemoryOptions struct {
    SessionID        string
    SessionIDs       []string   // any-of; also matches conv:<id> tags (T1)
    TimeFrom, TimeTo *time.Time
    AsOf             *time.Time // EntryValidAt before Limit
    Limit            int        // default 10
    Tier             *MemoryTier
    QueryVec         []float32  // skipped for count / temporal-order queries
    ReRankTemporal   bool
    IncludeArchival  bool       // default tiers: Working+Contextual+Semantic
}
func (ps *PalaceStore) SearchMemoryWithOptions(query string, opts SearchMemoryOptions) []MemoryEntry
```

Filters apply **before** Limit. `SearchMemory` remains a thin wrapper.

### K2 — Timeline list + meta index

```go
type ListMemoryOptions struct {
    SessionID        string
    SessionIDs       []string // any-of + conv: tags
    TimeFrom, TimeTo *time.Time
    Tag, TagPrefix, Query string
    Limit            int  // default 50
    Tier             *MemoryTier
    IncludeArchival, Ascending bool
}
func (ps *PalaceStore) ListMemoryWithOptions(opts ListMemoryOptions) []MemoryEntry
```

`Write` / unlink **patch** a clean in-memory meta index (optional durable snapshot). Dirty/missing index rebuilds lazily from tier JSON (O(n)). FS files remain source of truth.

### K4 lite — Validity windows

```go
func ParseValidityWindow(e MemoryEntry) (from, until *time.Time)
func EntryValidAt(e MemoryEntry, asOf time.Time) bool
func (ps *PalaceStore) ListFactsAsOf(opts FactsAsOfOptions) []MemoryEntry
```

Tags: `valid_from:<RFC3339>` inclusive start; `valid_until:<RFC3339>` **exclusive** end. No tags → valid if event time is zero or `!eventTime.After(asOf)`. Compaction SUMMARIZE / MERGE / CREATE_CORE_PRINCIPLE stamp `valid_from` from the parent.

T4 tests ([#126](https://github.com/iome-sh/memory/pull/126) `3fcbe4d`): `ListFactsAsOf` after MERGE/SUMMARIZE still returns the product; sources move to archival without invented `valid_until`; ARCHIVE that only moves tiers stays valid at now. Dual-clock store is **not** shipped. **T4-perf shipped** ([#134](https://github.com/iome-sh/memory/pull/134) `2a3257c`): session/tier/query collect through the list meta index (no Limit); `EntryValidAt` / entity still after load. btree still parked.

### A3 lite — Supersession + latest-value retrieve

```go
func (ps *PalaceStore) SupersedeEntityFacts(entityKey string, asOf time.Time) (int, error)
func (ps *PalaceStore) WriteAndSupersede(entry MemoryEntry, supersedeKeys []string) error
func AssembleLatestValueEvidence(query string, facts []MemoryEntry) string
```

`Supersede*` closes prior open windows for an **explicit** entity key (no NLP). `AssembleLatestValueEvidence` ranks dollar/scalar facts later-`Timestamp` first for amount / pre-approved queries ([#122](https://github.com/iome-sh/memory/pull/122)). LongMemEval retrieve may prepend a synthetic `latest_value_evidence` hit (not persisted).

### A2 lite — Multi-hop

```go
func (ps *PalaceStore) MultiHopRetrieve(opts MultiHopOptions) []MemoryEntry
func (ps *PalaceStore) ExpandRelatedEntitiesHops(seed string, maxHops int) map[string]int
```

BFS on `GetRelatedEntities`, collect by `entity:` tags, default **shorter hop first**. `MultiHopOptions.SessionIDs` is any-of. Not typed-edge weights.

### T1 helpers (v1.5.12 + unreleased)

```go
func ConvTag(id string) string
func AssembleCountEvidence(query string, facts []MemoryEntry) string
func AssembleTemporalEvidence(query string, facts []MemoryEntry) string
```

Count queries are **not** calendar windows. They union matching `turn_fact` children across `conv:` sessions (`unionCountQueryFacts` over `collectSearchCandidates`), rank named-pattern facts, diversify by session, then Limit. Clothing clusters are action×object (dry-clean / return / pick-up); N-header is clothing-only ([#124](https://github.com/iome-sh/memory/pull/124)); bullets are `1. 2. 3.` in cluster order ([#127](https://github.com/iome-sh/memory/pull/127) `a4c0445`). That does **not** invent “the answer is 3”. Unique-entity clusters cover kits, plants, hour+destination, and restaurants (catalogs as aliases; noun-phrase / dest extract in [#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`; restaurant clusters in [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a`). Temporal evidence lists **text date phrases** separately from ingest `Timestamp`; which-first sorts by parsed text time. Dated-span how-many with ≥2 parsed text times appends `text dates N days apart (phrase → phrase)` ([#129](https://github.com/iome-sh/memory/pull/129) `625a772`) — not ingest `Timestamp`, not a gold answer.

---

## Performance map (shipped vs next)

| Lever | Status | Notes |
|-------|--------|-------|
| Keyword-first retrieve | **Shipped** | Vector cannot bury a token hit past Limit |
| Skip `QueryVec` on count / temporal-order | **Shipped** ([#122](https://github.com/iome-sh/memory/pull/122)) | Fake embedder must not run on those queries |
| Batch ONNX scoring on LME retrieve | **Shipped** ([#114](https://github.com/iome-sh/memory/pull/114)) | Harness path; library `PersistEmbeddings` still default **off** |
| Meta-index patch on Write | **Shipped** (v1.5.8) | First list after process start may rebuild |
| Durable `indexes/event-time.json` | **Shipped** | Stamp mismatch → O(n) JSON walk |
| T2 list-latency bench | **Shipped** ([#128](https://github.com/iome-sh/memory/pull/128) `376dd69`) | `BenchmarkListMemoryWithOptions_SessionTimeLimit` MetaIndex vs `DisableMetaIndex`; `BenchmarkSearchMemoryWithOptions_CountQuery` skip-vector vs `QueryVec`. Measured ([#131](https://github.com/iome-sh/memory/pull/131) `ea94317`, Apple M4, N=200, `-benchtime=500ms -count=1`): warmed MetaIndex **792540 ns/op**; DisableMetaIndex **5125908 ns/op**; CountQuery_SkipVector **5510644 ns/op**; NonCount_WithQueryVec **6502154 ns/op**. Rebuild is **not** the limiter at N=200 (filter/load of survivors dominates). btree stays **parked**. |
| Count-union on search candidates | **Shipped** ([#131](https://github.com/iome-sh/memory/pull/131) `ea94317`) | `collectSearchCandidates` applies session/time/tier through the list meta index and loads JSON only for survivors; `unionCountQueryFacts` unions that slice (no second palace walk). `DisableMetaIndex` keeps the O(n) path. Do **not** claim a hot vector index. |
| btree / tag secondaries | **T2**, parked | #131 measure: DisableMetaIndex gap is full JSON scan vs index filter, not a btree range. Warmed MetaIndex ~1ms at N=200 and N=2000. |
| `ListFactsAsOf` via meta index | **T4-perf shipped** ([#134](https://github.com/iome-sh/memory/pull/134) `2a3257c`) | Session/tier/query collect through `listMemoryViaIndex` (no Limit). `EntryValidAt` / entity still after load. `DisableMetaIndex` keeps the JSON walk. T4 compaction contract unchanged. Apple M4, `BenchmarkListFactsAsOf_SessionAsOf`, `-benchtime=300ms -count=1`: warmed MetaIndex **1361143 ns/op**; DisableMetaIndex **5605138 ns/op**. btree still parked. |
| Persist ONNX vectors | Opt-in | Default **off**. Hash never stored. |
| usearch / ORT / Qdrant | Optional, off default path | Must not become required. Hash SearchMemory stays the zero-dep path. |

---

## Future phases

Do not start Qwen3 or a dual-clock KG to chase n=12 clothes (that miss is reader assembly). btree / typed edges stay gated on measured limiters, not on eval hunger.

### T1 residuals (after v1.5.12)

**Shipped** (see helpers above). **Done-when on n=12 MiniLM+BGE multi-session 0/2** was met at Wave E/F.

**Still open**

| Residual | Evidence | Next slice |
|----------|----------|------------|
| Clothes gold 3, reader sums 2 | Waves G–I: 3 numbered clusters in retrieve | **Numbered `1. 2. 3.` bullets shipped** ([#127](https://github.com/iome-sh/memory/pull/127) `a4c0445`). Wave I: **miss all three**. n=60 `5154a76`: MiniLM **PASS**; hash/BGE still summed 2. Does not invent “the answer is 3”. |
| Projects N=8 overcount | Wave H MiniLM | Clothing-only N header ([#124](https://github.com/iome-sh/memory/pull/124)). Wave I: projects **pass all three** (`Count evidence:` without N). MiniLM recovered. |
| Unique-entity n=60 (kits 5, hours 15, plants) | n=60 v1.5.12 **before** [#119](https://github.com/iome-sh/memory/pull/119) | **T7 shipped** ([#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`). n=60 `5154a76`: kits `gpt4_59c863d7` / hours `aae3761f` / plants `3a704032` **pass all three**. |
| Days-between TR | n=60 `08f4fc43` / `2a1811e2` / `2c63a862` | **T6 shipped** ([#129](https://github.com/iome-sh/memory/pull/129) `625a772`). n=60 `5154a76`: `2a1811e2` / `2c63a862` **pass all three**; `08f4fc43` **miss all three**. Does not invent gold. |
| KU stale amount | `852ce960` $350k vs gold $400k | Latest-value evidence ([#122](https://github.com/iome-sh/memory/pull/122)). n=60 `5154a76`: **miss all three** (reader $350k). |
| KU restaurants 3 vs 4 | Wave I `6aeb4375` miss all three (gold **four**, hyp **three**) | **T7+ restaurant clusters shipped** ([#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a`). n=60 `5154a76` is **before** #136: still 3 vs 4 all three. Remesure after #136 **TBD**. Clothing N-header unchanged. Does not invent gold 4. |
| Skip-vector + #124 + #127 n=12 | Wave I `e094bec` | hash/MiniLM/BGE **10/12** MS **1/2**. Clothes still reader-side. KU `6aeb4375` 3 vs 4. [#134](https://github.com/iome-sh/memory/pull/134) not this remesure. |

**Out of scope:** publishing a LongMemEval leaderboard number; default embedder Qwen3; flock.

### T2 — Event-time index beyond patch

**Why:** First list after process start still walks JSON when the durable stamp mismatches.

**Bench shipped:** [#128](https://github.com/iome-sh/memory/pull/128) `376dd69` — `BenchmarkListMemoryWithOptions_SessionTimeLimit` (MetaIndex vs `DisableMetaIndex`) and `BenchmarkSearchMemoryWithOptions_CountQuery` (skip-vector vs `QueryVec`).

**Start btree / tag secondaries when:** that bench (or `MetaIndexRebuilds` in a laptop-palace dogfood) shows **rebuild**, not filter/Limit, as the limiter.

**Measure shipped** ([#131](https://github.com/iome-sh/memory/pull/131) `ea94317`, Apple M4): warmed MetaIndex **792540 ns/op** vs DisableMetaIndex **5125908 ns/op** at N=200; first-list rebuild is **not** the limiter vs filter/load of survivors. N=2000 warmed stays ~1ms (same session+time survivor set). btree / tag secondaries **parked**. Search/count candidates reuse the list meta index for session/time/tier (`DisableMetaIndex` opt-out). Laptop palaces of a few thousand entries stay on the patch.

**In scope:** optional btree / tag secondary; keep FS as source of truth; opt-out flags stay.

**Out of scope:** flock; cross-process writers; distributed timelines.

### T3 — Temporal relation edges

**Why:** `AddEntityRelationship` is untimed adjacency. As-of graph walk needs `valid_from` / `valid_until` on **edges**, not only entries.

**Sketch (do not ship until gated):**

```text
edge: { from, to, rel, valid_from?, valid_until? }
walk: skip edges where !validAt(asOf)
```

**Start only if** a measured miss is an **expired relation** (gold lives on an edge that should have closed). n=12/n=60 misses so far are counts, dated events, and latest-value — **not** expired edges. Park.

### T4 — Compaction vs validity (tests shipped)

**Why:** Ingest children stamp `valid_from`; MERGE / SUMMARIZE / ARCHIVE must not drop or invent validity windows.

**Shipped:** `applyParentSessionAndValidFrom` on SUMMARIZE / MERGE / CREATE_CORE_PRINCIPLE products; unit tests that the product has `valid_from` and `EntryValidAt(now)`.

**T4 tests shipped** ([#126](https://github.com/iome-sh/memory/pull/126) `3fcbe4d`):

- `ListFactsAsOf` after MERGE/SUMMARIZE still returns the product
- sources archived, not current-tier
- no invented `valid_until`
- ARCHIVE tier-move stays valid at now (`IncludeArchival` to see it)

Dual-clock store is **not** this slice (see **T8**). **T4-perf shipped** ([#134](https://github.com/iome-sh/memory/pull/134) `2a3257c`): `ListFactsAsOf` collects session/tier/query through the list meta index (no Limit); `EntryValidAt` / entity still after load. `DisableMetaIndex` keeps the JSON walk. Apple M4 MetaIndex **1361143 ns/op** vs DisableMetaIndex **5605138 ns/op**. btree still parked.

### T5 — Embedding profiles (original K3)

Keep **BGE-small-en-v1.5 384-d** as the documented ONNX default. MiniLM is the in-tree fallback. `PersistEmbeddings` stays default **off**.

Qwen3-0.6B **1024-d** only as an **opt-in** constructor/env preset when a concrete consumer needs it. Document re-index if Qdrant collection dim changes. No silent default flip. Last, consumer-driven.

### T6 — Dated-span text-date delta (**shipped**)

**Why:** n=60 TR class includes how-many-days-between. Dated bullets already exist (`AssembleTemporalEvidence` labels the **text** date phrase separately from ingest `Timestamp`). The reader still had to do arithmetic of **parsed text dates**. Using ingest `Timestamp` as the delta is the wrong clock.

**Shipped** ([#129](https://github.com/iome-sh/memory/pull/129) `625a772`):

- Dated-span queries with **≥2 parsed text times** append `text dates N days apart (phrase → phrase)`
- N is the UTC calendar-day difference of those parsed text times, **not** of ingest `Timestamp`
- One dated bullet → no delta line; which-first stays bullets-only
- Does **not** invent gold (“the answer is N”)
- Does **not** treat how-many-days-between as a calendar-window filter

n=60 `5154a76` remesure: `2a1811e2` / `2c63a862` pass all three; `08f4fc43` miss all three.

### T7 — Generalized unique-entity clusters (**shipped**)

**Why:** n=60 kits / hours miss. The unique-entity path used small catalogs (B-29, Spitfire, Outer Banks, …). That overfit the locked slice.

**Shipped** ([#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`):

- Hours: destination after `N hours to/in/for/at/toward` (catalog dests remain aliases)
- Kits: `… kit` / `model kit` noun phrases (repeated identity still one cluster)
- Plants: `<name> plant(s)` plus catalog aliases (cheap exclude for `power plant` / verb `plant a`)
- Existing catalogs stay as **aliases**, not the only matcher

**T7+ restaurants shipped** ([#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a`):

- How-many-restaurant queries cluster distinct restaurant identities (`… restaurant` / `… Korean restaurant` noun phrases, quoted names, dining-verb proper nouns, Kitchen/House/Grill-style suffixes)
- Optional catalog is **aliases**, not the only matcher; no LongMemEval gold names hardcoded as the extractor
- Repeated identity is one cluster; unique-entity evidence stays unnumbered dashes without `(N distinct items)` (clothing-only)

**Out of scope (still):** clothing path changes; `(N distinct items)` on unique-entity; inventing “the answer is N”.

n=60 `5154a76` remesure: kits/hours/plants **pass all three**. Restaurant remesure after [#136](https://github.com/iome-sh/memory/pull/136) is **TBD** (this n=60 is **before** #136; `6aeb4375` still 3 vs 4 all three).

### T8 — Dual-clock store (parked)

Three clocks are easy to conflate:

| Clock | What it is in this kernel today |
|-------|----------------------------------|
| Event time | `Timestamp` (when the turn happened) |
| Validity window | `valid_from` / `valid_until` tags (`EntryValidAt`) |
| Transaction time | **Not stored.** Write time is not a first-class as-of axis. |

A dual-clock store would record **when the row was written** separately from event time and from the validity window, so “what did the palace believe on date T?” can differ from “what was true in the world on date T?”.

**Start only if** a measured miss is an **expired window** / compaction that ate gold because transaction time and validity were the same field. T4 tests already lock `ListFactsAsOf` after MERGE/SUMMARIZE. n=12/n=60 misses so far are not that class. Park.

---

## Suggested implementation order

1. **Remesure n=12** after [#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#127](https://github.com/iome-sh/memory/pull/127)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132) — **done (Wave I)** [#135](https://github.com/iome-sh/memory/pull/135) kernel `e094bec` hash/MiniLM/BGE **10/12** MS **1/2** KU **1/2** TR **2/2** (do not invent further). Clothes residual still reader.
2. **Remesure n=60** after [#119](https://github.com/iome-sh/memory/pull/119)+[#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132) — **done** kernel `5154a76` hash **49/60** MS **9/10** / MiniLM **51/60** MS **9/10** / BGE **47/60** MS **8/10** (complete 60/60; unlike v1.5.12 `e90a82d` **before** #119). [#136](https://github.com/iome-sh/memory/pull/136) restaurant clusters **not** this remesure.
3. **T6** dated-span text-date delta — **shipped** [#129](https://github.com/iome-sh/memory/pull/129)
4. **T7** generalized unique-entity — **shipped** [#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`; **T7+ restaurants shipped** [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a` — restaurant remesure **TBD**
5. **T4-perf** `ListFactsAsOf` via meta index — **shipped** [#134](https://github.com/iome-sh/memory/pull/134) `2a3257c`. **T2 btree** only if rebuild is the limiter — still **parked**
6. **T3** typed edges only if gold is an expired relation — **parked**
7. **T8** dual-clock only if an expired-window miss shows up — **parked**
8. **T5** Qwen3 last, consumer-driven — **parked**

---

## Versioning

- Prefer new options fields and methods over breaking `SearchMemory` signatures
- Embedding dimension changes require Qdrant collection recreation; note in the release
- v1.5.2 K1 · v1.5.3 K2 list · v1.5.4 K4 as-of · v1.5.5 A2 multi-hop · v1.5.6 A3 supersession · v1.5.7 hop ranking · v1.5.8 meta-index patch · v1.5.11 persist-onnx-vec opt-in, TTFH, LongMemEval card · v1.5.12 T1 SessionIDs / conv tags / count assembly
- Unreleased on `main` after v1.5.12: [#119](https://github.com/iome-sh/memory/pull/119) unique-entity + dated evidence · [#122](https://github.com/iome-sh/memory/pull/122) latest-value + skip-vector · [#124](https://github.com/iome-sh/memory/pull/124) clothing-only N · [#126](https://github.com/iome-sh/memory/pull/126) T4 `ListFactsAsOf` tests · [#127](https://github.com/iome-sh/memory/pull/127) numbered clothes · [#128](https://github.com/iome-sh/memory/pull/128) T2 list-latency bench · [#129](https://github.com/iome-sh/memory/pull/129) T6 text-date delta · [#131](https://github.com/iome-sh/memory/pull/131) search/count via meta index (btree parked) · [#134](https://github.com/iome-sh/memory/pull/134) `2a3257c` T4-perf `ListFactsAsOf` via meta index · [#132](https://github.com/iome-sh/memory/pull/132) T7 generalized unique-entity · [#135](https://github.com/iome-sh/memory/pull/135) Wave I n=12 · [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a` unique-entity restaurant clusters

---

## Honesty

- Eval numbers in this document are **unpublished**. They are not a README number and not official V1 (official V1 remains mixed n=500 + BGE-small-en-v1.5 + judge `gpt-4o-2024-08-06`, reproduced twice — see [LONGMEMEVAL.md](./LONGMEMEVAL.md)).
- `PersistEmbeddings` defaults **off**. Hash / empty / `"hash"` models **never** persist `GenerateSimpleEmbedding` vectors as stored vectors or as `QueryVec`.
- Flock is **not** shipped. Supported topology is **one process per palace root**.
- Do not start Qwen3 or a dual-clock knowledge graph to chase n=12 clothes. That miss is reader assembly (clusters are in retrieve; the reader still summed 2). Numbered bullets ([#127](https://github.com/iome-sh/memory/pull/127)) are a kernel nudge, not a gold answer.
- btree / typed edges / Qwen3 default stay gated as written above.
