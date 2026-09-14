# Temporal Memory Kernel Roadmap

**Repository:** [`github.com/iome-sh/memory`](https://github.com/iome-sh/memory)  
**Scope:** Temporal features **inside this package** (`PalaceStore`), not MCP/TUI hosts.  
**As of:** 2026-09-14 · tagged **v1.5.12** · unreleased on `main`: unique-entity / dated-event / latest-value / skip-vector / clothing-only N · [#126](https://github.com/iome-sh/memory/pull/126) T4 `ListFactsAsOf` tests · [#127](https://github.com/iome-sh/memory/pull/127) numbered clothes · [#128](https://github.com/iome-sh/memory/pull/128) T2 list-latency bench · [#129](https://github.com/iome-sh/memory/pull/129) T6 text-date delta · [#131](https://github.com/iome-sh/memory/pull/131) search/count via meta index · [#134](https://github.com/iome-sh/memory/pull/134) `2a3257c` T4-perf `ListFactsAsOf` via meta index · [#132](https://github.com/iome-sh/memory/pull/132) T7 generalized unique-entity · [#135](https://github.com/iome-sh/memory/pull/135) Wave I n=12 · [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a` unique-entity restaurant clusters · [#138](https://github.com/iome-sh/memory/pull/138) n=60 remesure · [#140](https://github.com/iome-sh/memory/pull/140) `865994b` official V1 **first** run unpublished (BGE mixed n=500 **388/500**) · [#141](https://github.com/iome-sh/memory/pull/141) `7deafa5` T6+ weeks/months text-date delta · [#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5` latest-value clip keeps amount · Wave J n=12 hash-only restaurant remesure · [#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0` restaurant tried-count latest-first · [#147](https://github.com/iome-sh/memory/pull/147) `298584c` T6++ temporal-order extrema · [#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f` Wave K n=12 hash-only remesure · Wave L n=12 hash-only T6++ remesure (`d5bec89`) · [#150](https://github.com/iome-sh/memory/pull/150) `76175a0` dated-span ago vs `question_date` · Wave M n=12 hash-only remesure (`7a9b956`) · official V1 **run 2** recorded (same pin `9bee542`, BGE mixed n=500 **384/500**; not identical to first **388/500**; unpublished; do not publish a single %) · [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8` restaurant stop-token `as` · Wave N n=12 hash-only remesure (`77b2839`)

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

**Eval evidence (unpublished, not a README number, not Memory GA):** locked mixed LongMemEval, judge `gpt-4o-2024-08-06`, reader `gpt-4o-mini`, isolated palace per embed. n=12 / n=60 are an **improvement baseline**, not official V1. Official V1 mixed n=500 BGE on pin `9bee542`: **first** run ([#140](https://github.com/iome-sh/memory/pull/140) `865994b`) **388/500**; **run 2** **384/500**. Two runs are **not identical**. INTERNAL unpublished. Do not publish a single %.

| Wave | Kernel | hash | MiniLM | BGE | Notes |
|------|--------|------|--------|-----|-------|
| A pre-T1 | `cb93b08` | 9/12 MS 0/2 | 10/12 0/2 | 10/12 0/2 | Single `SessionID` |
| E T1 done-when | `f99c140` [#111](https://github.com/iome-sh/memory/pull/111) | 11/12 MS 1/2 | **12/12 2/2** | 11/12 1/2 | MiniLM+BGE left 0/2 |
| F clothing clusters | `aa64dbc` [#115](https://github.com/iome-sh/memory/pull/115) | 11/12 **2/2** | **12/12 2/2** | **12/12 2/2** | Clothes pass; hash temporal miss |
| G dated events | `b02abaf` [#119](https://github.com/iome-sh/memory/pull/119) | 11/12 1/2 | 11/12 1/2 | 11/12 1/2 | Temporal first **pass all 3**; clothes reader summed 2 |
| H N-header | `895f255` [#120](https://github.com/iome-sh/memory/pull/120) | 11/12 1/2 | 10/12 0/2 | 11/12 1/2 | Clothes still 2; MiniLM projects overcount (`8 distinct`) |
| I #122+#124+#127+#129+#132 | `e094bec` | 10/12 1/2 | 10/12 1/2 | 10/12 1/2 | Numbered clothes still reader 2; MiniLM projects recovered (no N); KU `6aeb4375` 3 vs 4 all three. [#134](https://github.com/iome-sh/memory/pull/134) **not** this remesure. |
| J #136+#143 hash-only | `5d3aca5` | 10/12 1/2 | SKIP | SKIP | After restaurant clusters [#136](https://github.com/iome-sh/memory/pull/136) + clip [#143](https://github.com/iome-sh/memory/pull/143); **before** tried-count [#146](https://github.com/iome-sh/memory/pull/146). Clothes `0a995998` FAIL (reader 2); projects `6d550036` PASS; KU `6aeb4375` still 3 vs 4. MiniLM/BGE **not run**. |
| K #146 hash-only | `89c1dc0` | 11/12 1/2 | SKIP | SKIP | After restaurant tried-count [#146](https://github.com/iome-sh/memory/pull/146). Clothes `0a995998` FAIL (reader 2); projects `6d550036` PASS; KU `6aeb4375` **PASS** (hyp four). MiniLM/BGE **not run**. |
| L #147 hash-only | `d5bec89` | 11/12 1/2 | SKIP | SKIP | After T6++ order extrema [#147](https://github.com/iome-sh/memory/pull/147) `298584c`. Clothes `0a995998` FAIL (reader 2); projects `6d550036` PASS; KU `6aeb4375` **PASS** (hyp four); temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`). MiniLM/BGE **not run**. |
| M #150 hash-only | `7a9b956` | 11/12 1/2 | SKIP | SKIP | After dated-span ago vs `question_date` [#150](https://github.com/iome-sh/memory/pull/150) `76175a0`. Clothes `0a995998` FAIL (reader 2); projects `6d550036` PASS; KU `6aeb4375` **PASS** (hyp four); temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`; which-first, not ago). MiniLM/BGE **not run**. |
| N #152 hash-only | `77b2839` | 11/12 1/2 | SKIP | SKIP | After restaurant stop-token `as` [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8` (on Wave M / #150; includes [#153](https://github.com/iome-sh/memory/pull/153)+[#154](https://github.com/iome-sh/memory/pull/154)). Clothes `0a995998` FAIL (reader 2); projects `6d550036` PASS; KU `6aeb4375` **PASS** (hyp four; no `[restaurant:as]`). Temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`; which-first, not ago). MiniLM/BGE **not run**. |
| n=60 v1.5.12 | `e90a82d` | 47/60 MS 8/10 | 48/60 MS 7/10 | **46/58** MS 6/8 | **Before [#119](https://github.com/iome-sh/memory/pull/119).** Clothes pass; kits/hours miss. BGE **incomplete** (timeouts `gpt4_59c863d7`, `e831120c`; 58/60 IDs). Do not treat 0.793 as comparable /60. |
| n=60 after #119+#122+#124+#129+#132 | `5154a76` | 49/60 MS 9/10 | 51/60 MS 9/10 | 47/60 MS 8/10 | **After [#119](https://github.com/iome-sh/memory/pull/119).** Complete 60/60. Kits/hours/plants pass all 3. Days `2a1811e2`/`2c63a862` pass; `08f4fc43` miss. KU `852ce960` $350k all 3; restaurants `6aeb4375` 3 vs 4 all 3. [#136](https://github.com/iome-sh/memory/pull/136) **not** this remesure. |
| official V1 mixed 500 (first) | `9bee542` [#140](https://github.com/iome-sh/memory/pull/140) `865994b` | — | — | **388/500** TR **83/133** | BGE-small-en-v1.5 ONNX only (hash/MiniLM **n/a**). Judge `gpt-4o-2024-08-06` · reader gpt-4o-mini · session_id=conv_id. Task-averaged 0.802 · abstention 0.633 (30). ss-user 68/70 · ss-asst 54/56 · KU 63/78 · MS 99/133 · pref 21/30 · TR 83/133 (0.624). INTERNAL unpublished ≠ n=12/n=60. Not Memory GA. Do not publish a single %. |
| official V1 mixed 500 (run 2) | `9bee542` (same pin as first) | — | — | **384/500** TR **84/133** | Same kernel/embed/judge/reader as first. Isolated palace `:8782`. Task-averaged 0.7915 · abstention 0.6333 (30). ss-user 67/70 · ss-asst 54/56 · KU 63/78 · MS 96/133 · pref 20/30 · TR 84/133 (0.632). Δ vs first: −4 overall (TR +1, MS −3, pref −1, ss-user −1). Reader/judge variance, not a kernel change. Pin is **before** #143/#146/#147/#150. INTERNAL unpublished ≠ n=12/n=60. Not Memory GA. Two runs not identical; do not publish a single %. |

Clothes remesure after numbered bullets ([#127](https://github.com/iome-sh/memory/pull/127)): Wave I **miss all three** (retrieve `1. 2. 3.` + N=3; reader summed 2). n=12 after [#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#127](https://github.com/iome-sh/memory/pull/127)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132): Wave I hash/MiniLM/BGE **10/12** MS **1/2** (kernel `e094bec`). Projects pass all three. Wave J hash-only after [#136](https://github.com/iome-sh/memory/pull/136)+[#143](https://github.com/iome-sh/memory/pull/143): kernel `5d3aca5` hash **10/12** MS **1/2**; clothes `0a995998` FAIL; projects PASS; KU `6aeb4375` still 3 vs 4. Wave K hash-only after [#146](https://github.com/iome-sh/memory/pull/146): kernel `89c1dc0` hash **11/12** MS **1/2**; clothes `0a995998` FAIL; projects PASS; KU `6aeb4375` **PASS** (hyp four) ([#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`). Wave L hash-only after [#147](https://github.com/iome-sh/memory/pull/147) `298584c`: kernel `d5bec89` hash **11/12** MS **1/2**; clothes `0a995998` FAIL; projects PASS; KU `6aeb4375` **PASS** (hyp four); temporal `gpt4_2487a7cb` PASS (`text dates earliest: two months ago · latest: last Saturday`). Wave M hash-only after [#150](https://github.com/iome-sh/memory/pull/150) `76175a0`: kernel `7a9b956` hash **11/12** MS **1/2**; clothes `0a995998` FAIL; projects PASS; KU `6aeb4375` **PASS** (hyp four); temporal `gpt4_2487a7cb` PASS (which-first, not ago). Wave N hash-only after [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`: kernel `77b2839` hash **11/12** MS **1/2**; clothes `0a995998` FAIL; projects PASS; KU `6aeb4375` **PASS** (hyp four; no `[restaurant:as]`); temporal `gpt4_2487a7cb` PASS (which-first, not ago). MiniLM/BGE **not this remesure**. n=60 after [#119](https://github.com/iome-sh/memory/pull/119)+[#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#129](https://github.com/iome-sh/memory/pull/129)+[#132](https://github.com/iome-sh/memory/pull/132): kernel `5154a76` hash **49/60** MS **9/10** / MiniLM **51/60** MS **9/10** / BGE **47/60** MS **8/10** (complete 60/60). Official V1 first run is BGE mixed n=500 **388/500** TR **83/133** (hash/MiniLM **n/a**). Official V1 run 2 (same pin `9bee542`) is **384/500** TR **84/133**. Two runs are **not identical** (reader/judge variance, not a kernel change). Unpublished, not Memory GA, ≠ n=12/n=60. Do not publish a single %. Do not invent MiniLM/BGE Wave J/K/L/M/N scores.

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

`Supersede*` closes prior open windows for an **explicit** entity key (no NLP). `AssembleLatestValueEvidence` ranks dollar/scalar facts later-`Timestamp` first for amount / pre-approved queries ([#122](https://github.com/iome-sh/memory/pull/122)). Long snippets keep extracted dollar amounts in the 280-char window (`clipKeepingAmount`, [#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5`); stale amounts stay listed. Does **not** NLP-supersede. LongMemEval retrieve may prepend a synthetic `latest_value_evidence` hit (not persisted).

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

Count queries are **not** calendar windows. They union matching `turn_fact` children across `conv:` sessions (`unionCountQueryFacts` over `collectSearchCandidates`), rank named-pattern facts, diversify by session, then Limit. Clothing clusters are action×object (dry-clean / return / pick-up); N-header is clothing-only ([#124](https://github.com/iome-sh/memory/pull/124)); bullets are `1. 2. 3.` in cluster order ([#127](https://github.com/iome-sh/memory/pull/127) `a4c0445`). That does **not** invent “the answer is 3”. Unique-entity clusters cover kits, plants, hour+destination, and restaurants (catalogs as aliases; noun-phrase / dest extract in [#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`; restaurant clusters in [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a`). Restaurant how-many prepends latest-first “tried N” self-reports; cuisine+BBQ dishes are not venues; stop-token `if` ([#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`). Does not invent gold 4. Temporal evidence lists **text date phrases** separately from ingest `Timestamp`; which-first sorts by parsed text time. Dated-span how-many with ≥2 parsed text times appends `text dates N days apart (phrase → phrase)` ([#129](https://github.com/iome-sh/memory/pull/129) `625a772`); how-many-weeks/months also append week (floor days/7) and calendar-month lines ([#141](https://github.com/iome-sh/memory/pull/141) `7deafa5`). Temporal-order queries with ≥2 dated bullets also append `text dates earliest: PHRASE · latest: PHRASE` ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`) — not ingest `Timestamp`, not a gold answer.

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

Do not start Qwen3 or a dual-clock KG to chase n=12 clothes (that miss is reader assembly). btree / typed edges stay gated on measured limiters, not on eval hunger. **T6+** weeks/months text-date delta is **shipped** ([#141](https://github.com/iome-sh/memory/pull/141) `7deafa5`). Latest-value clip is **shipped** ([#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5`). Restaurant tried-count is **shipped** ([#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`). Wave J hash n=12 still 3 vs 4 is **before** #146. Wave K hash n=12 **PASS** (`6aeb4375` hyp four) ([#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`). **T6++** temporal-order extrema is **shipped** ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`). Wave L hash n=12 **11/12** (temporal `gpt4_2487a7cb` **PASS**). Dated-span ago vs `question_date` is **shipped** ([#150](https://github.com/iome-sh/memory/pull/150) `76175a0`). Wave M hash n=12 **11/12** (temporal `gpt4_2487a7cb` **PASS**, which-first not ago). Restaurant stop-token `as` is **shipped** ([#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`). Wave N hash n=12 **11/12** (`6aeb4375` **PASS**, no `[restaurant:as]`). Official V1 **run 2** recorded (same pin `9bee542`, **384/500**; first run **388/500**). Two runs not identical (reader/judge variance). INTERNAL unpublished. Not Memory GA. Not a README number. Do not publish a single %. Clothes park. T2 / T3 / T5 / T8 parked.

### T1 residuals (after v1.5.12)

**Shipped** (see helpers above). **Done-when on n=12 MiniLM+BGE multi-session 0/2** was met at Wave E/F.

**Still open**

| Residual | Evidence | Next slice |
|----------|----------|------------|
| Clothes gold 3, reader sums 2 | Waves G–I: 3 numbered clusters in retrieve | **Numbered `1. 2. 3.` bullets shipped** ([#127](https://github.com/iome-sh/memory/pull/127) `a4c0445`). Wave I: **miss all three**. Wave J hash: **FAIL** (reader summed 2). Wave K hash: **FAIL** (reader summed 2). Wave L hash: **FAIL** (reader summed 2). Wave M hash: **FAIL** (reader summed 2). Wave N hash: **FAIL** (reader summed 2). n=60 `5154a76`: MiniLM **PASS**; hash/BGE still summed 2. Official V1 run 2 (pin `9bee542`, before #143/#146/#147/#150): `0a995998` **FAIL** reader 2. Does not invent “the answer is 3”. |
| Projects N=8 overcount | Wave H MiniLM | Clothing-only N header ([#124](https://github.com/iome-sh/memory/pull/124)). Wave I: projects **pass all three** (`Count evidence:` without N). MiniLM recovered. |
| Unique-entity n=60 (kits 5, hours 15, plants) | n=60 v1.5.12 **before** [#119](https://github.com/iome-sh/memory/pull/119) | **T7 shipped** ([#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`). n=60 `5154a76`: kits `gpt4_59c863d7` / hours `aae3761f` / plants `3a704032` **pass all three**. |
| Days-between TR | n=60 `08f4fc43` / `2a1811e2` / `2c63a862` | **T6 shipped** ([#129](https://github.com/iome-sh/memory/pull/129) `625a772`). n=60 `5154a76`: `2a1811e2` / `2c63a862` **pass all three**; `08f4fc43` **miss all three**. Official V1 run 2 (pin `9bee542`): `08f4fc43` **FAIL** despite hyp 30 days (judge). Days-only. Does not invent gold. |
| Official V1 TR weeks/months | First run TR **83/133 (0.624)** (`9bee542`) | **T6+ shipped** ([#141](https://github.com/iome-sh/memory/pull/141) `7deafa5`). Week (floor days/7) + calendar-month text-date deltas. T6 [#129](https://github.com/iome-sh/memory/pull/129) days line kept. Parsed text times, not ingest `Timestamp`. Official V1 remesure **TBD**. Do not invent gold. |
| Official V1 TR order extrema | First run TR **83/133 (0.624)** (`9bee542`) | **T6++ shipped** ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`). `text dates earliest: PHRASE · latest: PHRASE` from parsed text times. Not ingest `Timestamp`. Wave L hash n=12 `gpt4_2487a7cb` **PASS** (`text dates earliest: two months ago · latest: last Saturday`). Wave M hash n=12 `gpt4_2487a7cb` **PASS** (which-first, not ago). Wave N hash n=12 `gpt4_2487a7cb` **PASS** (which-first, not ago). Official V1 remesure **TBD**. Do not invent gold. |
| KU stale amount | `852ce960` $350k vs gold $400k | **Latest-value clip shipped** ([#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5`): `clipKeepingAmount` keeps `$400,000` in long snippets (Nov 30 turn was prefix-clipped at 280 chars). Stale `$350,000` still listed. Does **not** NLP-supersede. Last measured n=60 `5154a76` and official V1 first and run 2 on pin `9bee542` (before #143) still $350k. Remesure **TBD** (`852ce960` is n=60, not n=12). |
| KU restaurants 3 vs 4 | Wave I `6aeb4375` miss all three (gold **four**, hyp **three**) | **T7+ restaurant clusters shipped** ([#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a`). **Tried-count shipped** ([#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`): latest-first “tried N” self-reports; cuisine+BBQ dishes are not venues; stop-token `if`. **Stop-token `as` shipped** ([#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`). Does not invent gold 4. Wave J hash `5d3aca5` is **before** #146: **FAIL** (hyp three; `[restaurant:korean-style-bbq]` / `[restaurant:if]`). Wave K hash `89c1dc0` ([#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`): **PASS** (hyp four; latest-first tried-N; no korean-style-bbq/if). Wave L hash `d5bec89`: **PASS** (hyp four). Wave M hash `7a9b956`: **PASS** (hyp four). Leftover `[restaurant:as]` (**before** [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`). Wave N hash `77b2839`: **PASS** (hyp four; no `[restaurant:as]`). MiniLM/BGE **not this remesure**. n=60 `5154a76` is **before** #136: still 3 vs 4 all three. Official V1 run 2 (pin `9bee542`, before #146): `6aeb4375` **FAIL** three. Clothing N-header unchanged. |
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

Days-only T6. How-many-weeks / how-many-months dated-span is **T6+ shipped**.

### T6+ — Weeks/months text-date delta (**shipped**)

**Why:** Official V1 **first** run (BGE mixed n=500, kernel `9bee542`, [#140](https://github.com/iome-sh/memory/pull/140) `865994b`) scores TR **83/133 (0.624)**. T6 ([#129](https://github.com/iome-sh/memory/pull/129) `625a772`) already appends `text dates N days apart` from parsed text times. How-many-**weeks** / how-many-**months** dated-span questions still fail on that run — evidence reports days while the question asks weeks/months. Using ingest `Timestamp` remains the wrong clock.

**Shipped** ([#141](https://github.com/iome-sh/memory/pull/141) `7deafa5`):

- How-many-weeks / how-many-months still keep the T6 **days** line (`text dates N days apart (phrase → phrase)`)
- Weeks queries also append `text dates N weeks apart (floor days/7; remainder R days)` (remainder omitted when R=0)
- Months queries also append `text dates M calendar months apart (phrase → phrase)` (UTC `y*12+m` difference)
- Arithmetic uses parsed **text** times only, **not** ingest `Timestamp`
- Does **not** invent gold (“the answer is N”)
- Does **not** treat how-many-weeks/months as a calendar-window filter
- Days-only queries omit the week/month lines

Official V1 remesure after #141 is **TBD**. Do not invent a score.

**Out of scope:** NLP supersede; dual-clock store; publishing a README number; inventing remesure scores.

Do not start T2 btree / T3 edges / T5 Qwen3 / T8 dual-clock to chase this TR residual.

### T6++ — Temporal-order extrema (**shipped**)

**Why:** T6/T6+ ship dated-span **arithmetic** (days/weeks/months from parsed text times). Which-first / before / after still only listed dated bullets. Official V1 first-run TR **83/133 (0.624)** (`9bee542`). Official V1 remesure after T6+/T6++ is **TBD**. Do not invent a score.

**Shipped** ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`):

- `AssembleTemporalEvidence` appends `text dates earliest: PHRASE · latest: PHRASE` from parsed **text** times when the query is temporal-order and ≥2 dated bullets exist
- Dated-span queries keep T6/T6+ delta lines; extrema emit only when the query is also an order query
- `how many days/weeks/months ago` is a dated-span query (not a calendar window)
- Arithmetic only; **not** ingest `Timestamp`; does **not** invent gold

**Out of scope:** NLP supersede; dual-clock store; publishing a README number; inventing remesure scores.

Wave L hash n=12 kernel `d5bec89`: temporal `gpt4_2487a7cb` **PASS** (`text dates earliest: two months ago · latest: last Saturday`). Wave N hash n=12 kernel `77b2839`: temporal `gpt4_2487a7cb` **PASS** (which-first, not ago). Official V1 remesure **TBD**. Do not invent a score.

Do not start T2 btree / T3 edges / T5 Qwen3 / T8 dual-clock to chase this TR residual.

### T6 ago vs question_date (**shipped**)

**Why:** Official V1 first-run TR includes how-many-days/weeks/months-**ago**. T6/T6+ dated-span arithmetic uses parsed text times; T6++ extrema is which-first. Ago vs “now” needs the retrieve `question_date`, not ingest `Timestamp` and not a `TimeTo` filter.

**Shipped** ([#150](https://github.com/iome-sh/memory/pull/150) `76175a0`):

- Retrieve may pass `question_date` (RFC3339 or `YYYY-MM-DD`; invalid/empty is skipped, not a retrieve error)
- Dated-span **ago** evidence appends text-date vs that instant (`N days/weeks/months before question_date YYYY-MM-DD`, or `after` if the event is later)
- Uses the latest parsed **text** time among bullets, never ingest `Timestamp`
- Not a retrieve `TimeTo` / `AsOf` ranking filter. Not gold.

Wave M hash n=12 kernel `7a9b956`: n=12 temporal `gpt4_2487a7cb` is **which-first**, not ago (no `question_date` evidence line). Hash **11/12**. Wave N hash n=12 kernel `77b2839`: same which-first PASS; hash **11/12**. Official V1 remesure **TBD**. Do not invent a score.

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

**Tried-count shipped** ([#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`):

- Unique-entity restaurant evidence prepends latest-first “tried N” self-reports (stale “three” still listed; not NLP-supersede)
- Cuisine+BBQ dishes (`Korean-style BBQ`, `Korean BBQ`) are **not** venues
- Stop-token `if` is not a restaurant name
- Stop-token `as`/`is` shipped ([#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`); leftover `[restaurant:as]` is not a venue
- Does **not** invent “the answer is 4”; no clothing-only `(N distinct items)` on unique-entity restaurant

**Out of scope (still):** clothing path changes; `(N distinct items)` on unique-entity; inventing “the answer is N”.

n=60 `5154a76` remesure: kits/hours/plants **pass all three**. Restaurant remesure after [#146](https://github.com/iome-sh/memory/pull/146): Wave J hash n=12 kernel `5d3aca5` `6aeb4375` still **FAIL** (3 vs gold 4) is **before** #146. Wave K hash n=12 kernel `89c1dc0` `6aeb4375` **PASS** (hyp four) ([#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`). Wave L hash n=12 kernel `d5bec89` `6aeb4375` **PASS** (hyp four). Wave M hash n=12 kernel `7a9b956` `6aeb4375` **PASS** (hyp four). Wave N hash n=12 kernel `77b2839` `6aeb4375` **PASS** (hyp four; no `[restaurant:as]` after [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`). MiniLM/BGE **not this remesure**. n=60 `5154a76` is **before** #136.

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

1. **T1–T7 / T4-perf / T6 / T7+ restaurants — shipped.** T1 SessionIDs / skip-vector / latest-value / clothing-only N / numbered clothes ([#122](https://github.com/iome-sh/memory/pull/122)+[#124](https://github.com/iome-sh/memory/pull/124)+[#127](https://github.com/iome-sh/memory/pull/127)); T6 days-only text-date delta [#129](https://github.com/iome-sh/memory/pull/129) `625a772`; T7 unique-entity [#132](https://github.com/iome-sh/memory/pull/132) `e4dcb73`; T4-perf [#134](https://github.com/iome-sh/memory/pull/134) `2a3257c`; T7+ restaurants [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a`; Wave I n=12 [#135](https://github.com/iome-sh/memory/pull/135); n=60 [#138](https://github.com/iome-sh/memory/pull/138). Official V1 first run unpublished [#140](https://github.com/iome-sh/memory/pull/140) `865994b`. Latest-value clip [#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5`.
2. **T6+ weeks/months text-date delta — shipped.** [#141](https://github.com/iome-sh/memory/pull/141) `7deafa5`. Official V1 first-run TR **83/133 (0.624)** weeks/months still fail on kernel `9bee542` (T6 was days-only). Week (floor days/7) + calendar-month text-date deltas. Parsed text times, not ingest `Timestamp`. Do not invent gold. Official V1 remesure **TBD**.
3. **Restaurant tried-count — shipped.** [#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`: latest-first “tried N” self-reports; cuisine+BBQ dishes are not venues; stop-token `if`. Stop-token `as` [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`. Does not invent gold 4. Wave J hash n=12 kernel `5d3aca5` `6aeb4375` still **FAIL** (3 vs 4) is **before** this PR. Wave K hash n=12 kernel `89c1dc0` `6aeb4375` **PASS** (hyp four) ([#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`). Wave L hash n=12 kernel `d5bec89` `6aeb4375` **PASS** (hyp four). Wave M hash n=12 kernel `7a9b956` `6aeb4375` **PASS** (hyp four). Wave N hash n=12 kernel `77b2839` `6aeb4375` **PASS** (hyp four; no `[restaurant:as]`). MiniLM/BGE **not** Wave J/K/L/M/N — do not invent those scores. n=60 `5154a76` is before #136.
4. **T6++ temporal-order extrema — shipped.** [#147](https://github.com/iome-sh/memory/pull/147) `298584c`: `text dates earliest: PHRASE · latest: PHRASE` from parsed text times. Dated-span keeps T6/T6+ deltas. Not ingest `Timestamp`. Wave L hash n=12 kernel `d5bec89` `gpt4_2487a7cb` **PASS** (`text dates earliest: two months ago · latest: last Saturday`). Wave M hash n=12 kernel `7a9b956` `gpt4_2487a7cb` **PASS** (which-first, not ago). Wave N hash n=12 kernel `77b2839` `gpt4_2487a7cb` **PASS** (which-first, not ago). Do not invent gold. Official V1 remesure **TBD**.
5. **Dated-span ago vs question_date — shipped.** [#150](https://github.com/iome-sh/memory/pull/150) `76175a0`: retrieve may pass `question_date`; ago evidence uses latest parsed text time vs that instant. Not TimeTo. Not gold. Not ingest Timestamp. Wave M hash n=12 kernel `7a9b956` **11/12** (n=12 temporal `gpt4_2487a7cb` is which-first, not ago). Wave N hash n=12 kernel `77b2839` **11/12** (which-first, not ago). Official V1 remesure **TBD**.
6. **Official V1 run 2 recorded.** Same pin `9bee542`, BGE-small-en-v1.5 ONNX, mixed n=500, judge `gpt-4o-2024-08-06`, reader gpt-4o-mini, session_id=conv_id. Isolated palace `:8782`. Overall **384/500** (task-averaged 0.7915; abstention 0.6333). First run **388/500**. Two runs **not identical** (reader/judge variance, not a kernel change). INTERNAL unpublished, not Memory GA, not a README number. Do not publish a single %. Do not claim “reproduced twice.”
7. **Clothes reader residual — park more assembly.** Wave I numbered `1. 2. 3.` + N=3 in retrieve; reader summed 2. Wave J/K/L/M/N hash `0a995998` still **FAIL**. Do not start Qwen3 or dual-clock to chase it.
8. **Latest-value clip shipped; KU `852ce960` remesure TBD.** [#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5` `clipKeepingAmount` keeps `$400,000` in long snippets (Nov 30 turn was prefix-clipped at 280 chars). Stale `$350,000` still listed. Does **not** NLP-supersede. Last measured n=60 `5154a76` / official V1 first and run 2 on pin `9bee542` (before #143) still $350k. `852ce960` is n=60, not n=12. Remesure **TBD**.
9. **T2 btree / T3 edges / T8 dual-clock / T5 Qwen3 — still parked.** btree only if rebuild is the limiter ([#131](https://github.com/iome-sh/memory/pull/131) says it is not). T3 only if gold is an expired relation. T8 only if an expired-window miss shows up. T5 Qwen3 last, consumer-driven.

---

## Versioning

- Prefer new options fields and methods over breaking `SearchMemory` signatures
- Embedding dimension changes require Qdrant collection recreation; note in the release
- v1.5.2 K1 · v1.5.3 K2 list · v1.5.4 K4 as-of · v1.5.5 A2 multi-hop · v1.5.6 A3 supersession · v1.5.7 hop ranking · v1.5.8 meta-index patch · v1.5.11 persist-onnx-vec opt-in, TTFH, LongMemEval card · v1.5.12 T1 SessionIDs / conv tags / count assembly
- Unreleased on `main` after v1.5.12: [#119](https://github.com/iome-sh/memory/pull/119) unique-entity + dated evidence · [#122](https://github.com/iome-sh/memory/pull/122) latest-value + skip-vector · [#124](https://github.com/iome-sh/memory/pull/124) clothing-only N · [#126](https://github.com/iome-sh/memory/pull/126) T4 `ListFactsAsOf` tests · [#127](https://github.com/iome-sh/memory/pull/127) numbered clothes · [#128](https://github.com/iome-sh/memory/pull/128) T2 list-latency bench · [#129](https://github.com/iome-sh/memory/pull/129) T6 text-date delta · [#131](https://github.com/iome-sh/memory/pull/131) search/count via meta index (btree parked) · [#134](https://github.com/iome-sh/memory/pull/134) `2a3257c` T4-perf `ListFactsAsOf` via meta index · [#132](https://github.com/iome-sh/memory/pull/132) T7 generalized unique-entity · [#135](https://github.com/iome-sh/memory/pull/135) Wave I n=12 · [#136](https://github.com/iome-sh/memory/pull/136) `0a40b0a` unique-entity restaurant clusters · [#138](https://github.com/iome-sh/memory/pull/138) n=60 remesure · [#140](https://github.com/iome-sh/memory/pull/140) `865994b` official V1 first run unpublished (BGE mixed n=500 **388/500**) · [#141](https://github.com/iome-sh/memory/pull/141) `7deafa5` T6+ weeks/months text-date delta · [#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5` latest-value clip keeps amount · Wave J n=12 hash-only restaurant remesure (`6aeb4375` still 3 vs 4, **before** #146) · [#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0` restaurant tried-count latest-first · [#147](https://github.com/iome-sh/memory/pull/147) `298584c` T6++ temporal-order extrema · [#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f` Wave K n=12 hash-only remesure (`6aeb4375` PASS) · Wave L n=12 hash-only T6++ remesure (`d5bec89`, 11/12; `gpt4_2487a7cb` PASS) · [#150](https://github.com/iome-sh/memory/pull/150) `76175a0` dated-span ago vs `question_date` · Wave M n=12 hash-only remesure (`7a9b956`, 11/12; `gpt4_2487a7cb` PASS, which-first not ago) · official V1 run 2 recorded (same pin `9bee542`, BGE mixed n=500 **384/500**; not identical to first **388/500**; unpublished; do not publish a single %) · [#152](https://github.com/iome-sh/memory/pull/152) `85d44c8` restaurant stop-token `as` · Wave N n=12 hash-only remesure (`77b2839`, 11/12; `6aeb4375` PASS, no `[restaurant:as]`)

---

## Honesty

- Eval numbers in this document are **unpublished**. They are not a README number and **not Memory GA**. Official V1 **first** run is BGE mixed n=500 **388/500** (kernel `9bee542`, [#140](https://github.com/iome-sh/memory/pull/140) `865994b`). Official V1 **run 2** (same pin) is **384/500**. Two runs are **not identical** (reader/judge variance). INTERNAL unpublished. Do not publish a single %. n=12 / n=60 are an improvement baseline ≠ these V1 runs.
- Hash-overlap is unpublished. It is not official V1.
- TTFH / cite-both is a different clock from LongMemEval (see [TTFH.md](./TTFH.md)).
- `PersistEmbeddings` defaults **off**. Hash / empty / `"hash"` models **never** persist `GenerateSimpleEmbedding` vectors as stored vectors or as `QueryVec`.
- Flock is **not** shipped. Supported topology is **one process per palace root**.
- Do not start Qwen3 or a dual-clock knowledge graph to chase n=12 clothes. That miss is reader assembly (clusters are in retrieve; the reader still summed 2). Numbered bullets ([#127](https://github.com/iome-sh/memory/pull/127)) are a kernel nudge, not a gold answer.
- btree / typed edges / Qwen3 default stay gated as written above. T6+ is **shipped**; restaurant tried-count is **shipped**; T6++ order extrema is **shipped** ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`). Dated-span ago vs `question_date` is **shipped** ([#150](https://github.com/iome-sh/memory/pull/150) `76175a0`). Restaurant stop-token `as` is **shipped** ([#152](https://github.com/iome-sh/memory/pull/152) `85d44c8`). Wave K remesure is **recorded** ([#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`). Wave L hash n=12 remesure is **recorded** (kernel `d5bec89`, 11/12). Wave M hash n=12 remesure is **recorded** (kernel `7a9b956`, 11/12). Wave N hash n=12 remesure is **recorded** (kernel `77b2839`, 11/12; no `[restaurant:as]`). V1 run 2 is **recorded** (384/500, same pin as 388/500; not identical; unpublished). T2 / T3 / T5 / T8 stay parked.
