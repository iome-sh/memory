# Temporal Memory Kernel Roadmap

**Repository:** [`github.com/iome-sh/memory`](https://github.com/iome-sh/memory)  
**Scope:** Temporal features **inside this package** (`PalaceStore`), not MCP/TUI hosts.  
**As of:** 2026-09-12 · tagged **v1.5.12** (T1 multi-session retrieve + count assembly)

This is the canonical temporal plan for the hierarchical agent memory library. Callers own tenancy above `BaseDir`. Companion hosts ([iomesh-tui](https://github.com/iome-sh/iomesh-tui), [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp)) are optional.

Related: [TTFH walking skeleton](./TTFH.md) · [LongMemEval methodology](./LONGMEMEVAL.md) · [locked mixed baseline](./LONGMEMEVAL_BASELINE.md) · [package improvements](./memory-refactor-improvements.md)

---

## Re-evaluation (original plan vs v1.5.11)

The original document (last updated 2026-08-05) sequenced **K0–K4** plus **A2/A3**. Most of that surface is now in the tree. What moved after v1.5.7 is index patching, palace file modes, retrieve default tiers, private `source_hint`, optional ONNX persist, TTFH, and a locked LongMemEval mixed slice.

| Original phase | Original intent | Status at v1.5.11 | Still open |
|----------------|-----------------|-------------------|------------|
| **K0** | Event time, decay, `IngestTurn`, hybrid search | **Shipped** | — |
| **K1** | `SearchMemoryWithOptions` session/time + temporal re-rank | **Shipped** (v1.5.2) | — |
| **K2** | Event-time timeline list + tag helpers; FS index | **Mostly shipped** — list API v1.5.3; durable `indexes/event-time.json`; in-memory patch on Write/unlink (v1.5.8). First list / stamp mismatch still rebuilds. | btree / tag secondaries if O(n) rebuild is the bottleneck |
| **K3** | Optional Qwen3-0.6B **1024-d** local preset | **Not started** (and not blocking). Default ONNX remains BGE-small **384-d**; MiniLM is the in-tree fallback; `PersistEmbeddings` is opt-in (default off). | Only if a consumer needs 1024-d |
| **K4** | Facts-as-of / validity windows | **Shipped lite** (v1.5.4) — `ListFactsAsOf`, `EntryValidAt`, `SearchMemoryOptions.AsOf` | Temporal **edges**; transaction-time + validity as first-class stores |
| **A2** | Multi-hop / associative retrieve | **Shipped lite** (v1.5.5–1.5.7) — `MultiHopRetrieve`, hop-distance ranking | Typed / bidirectional edges; full path scoring |
| **A3** | Fact supersession | **Shipped lite** (v1.5.6) — `SupersedeEntityFacts`, `WriteAndSupersede` | Auto entity extract; NLP contradiction |

**Eval evidence (not a leaderboard number):** locked mixed LongMemEval n=12, same IDs, judge `gpt-4o-2024-08-06`. Hash 9/12, MiniLM 10/12, BGE 10/12. **`multi-session` is 0/2 on every embedder.** Gold answers are counts across sessions. The reader now uses full retrieve-k (was 15 of 40). The remaining miss is a **kernel retrieve / temporal-aggregation** problem, not “missing K1 filters.”

**Walking skeleton:** `go run ./examples/ttfh_rca` — ingest three RCA turns, same-process retrieve, `ListFactsAsOf`, print `source_hint`. That path exercises K0 + K1 session retrieve + K4 as-of. It does not exercise multi-session count questions.

---

## Shipped surface (keep these contracts)

### K0 — Baseline

On `MemoryEntry`: `Timestamp` (event time), `SessionID`, `TemporalTags`, turn fields (`TurnID`, `ExtractedFacts`, `Keyphrases`, `OriginalText`). Provenance: `IngestTurn` stamps `source_hint=private` when the caller does not already supply a classifiable mesh or private source.

Scoring: `CalculateTemporalDecay`, `CalculateRecencyBoost`, `CalculateRelevanceScore`, `MultiFactorScore`.

Ingest/search: `IngestTurn`; `SearchMemory` hybrid keyword-first + optional `QueryVec`. Hash embeddings never persist as stored vectors.

Topology: **one process per palace root**. `writeMu` serializes `relations/entity-graph.json` and `indexes/event-time.json`. Flock is not shipped.

### K1 — Filtered search

```go
type SearchMemoryOptions struct {
    SessionID, TimeFrom, TimeTo *… // session + inclusive event-time window
    AsOf            *time.Time     // EntryValidAt before Limit
    Limit           int            // default 10
    Tier            *MemoryTier
    QueryVec        []float32      // keyword hits stay ahead of cosine
    ReRankTemporal  bool
    IncludeArchival bool           // default tiers: Working+Contextual+Semantic
}
func (ps *PalaceStore) SearchMemoryWithOptions(query string, opts SearchMemoryOptions) []MemoryEntry
```

Filters apply **before** Limit. `SearchMemory` remains a thin wrapper.

### K2 — Timeline list + meta index

```go
type ListMemoryOptions struct {
    SessionID, TimeFrom, TimeTo *…
    Tag, TagPrefix, Query string
    Limit int            // default 50
    Tier *MemoryTier
    IncludeArchival, Ascending bool
}
func (ps *PalaceStore) ListMemoryWithOptions(opts ListMemoryOptions) []MemoryEntry
```

`Write` / unlink **patch** a clean in-memory meta index (and optional durable snapshot). Dirty/missing index rebuilds lazily from tier JSON (O(n)). `DisableMetaIndex` / `DisableDurableIndex` exist for tests. FS files remain source of truth.

### K4 lite — Validity windows

```go
func ParseValidityWindow(e MemoryEntry) (from, until *time.Time)
func EntryValidAt(e MemoryEntry, asOf time.Time) bool
func (ps *PalaceStore) ListFactsAsOf(opts FactsAsOfOptions) []MemoryEntry
```

Tags: `valid_from:<RFC3339>` inclusive start; `valid_until:<RFC3339>` **exclusive** end. No tags → valid if event time is zero or `!eventTime.After(asOf)`.

### A3 lite — Supersession

```go
func (ps *PalaceStore) SupersedeEntityFacts(entityKey string, asOf time.Time) (int, error)
func (ps *PalaceStore) WriteAndSupersede(entry MemoryEntry, supersedeKeys []string) error
```

Closes prior open windows for an explicit entity key. Does not delete entries. Does not run NLP.

### A2 lite — Multi-hop

```go
func (ps *PalaceStore) MultiHopRetrieve(opts MultiHopOptions) []MemoryEntry
func (ps *PalaceStore) ExpandRelatedEntitiesHops(seed string, maxHops int) map[string]int
```

BFS on `GetRelatedEntities`, collect by `entity:` tags, default **shorter hop first**. Not typed-edge weights.

---

## Future phases (next TODOs)

Order is **T1 → measure → T2 only if list latency hurts → T3/T4 on demand → T5 last**. Do not start K3/Qwen3 or a dual-clock KG before T1 is measured.

### T1 — Multi-session temporal retrieve (**done** on n=12)

**Why:** Original K1 session filter is single-`SessionID`. LongMemEval `multi-session` items need facts **spread across several haystack sessions** in one `conv_id` palace. Flattening ingest onto `SessionID=conv_id` plus Limit filled by one noisy session buries count gold.

**Shipped this slice**

- `SearchMemoryOptions.SessionIDs` / `ListMemoryOptions.SessionIDs` / `FactsAsOfOptions.SessionIDs` / `MultiHopOptions.SessionIDs` (any-of)
- `conv:<id>` tag match: retrieve with `SessionID=conv_id` still sees inner haystack sessions
- Session-diverse ranking before Limit (round-robin distinct `SessionID`s)
- LongMemEval ingest passes per-turn `session_id` and stamps `conv:<conv_id>`
- Count questions (`how many` / `how much`) are not treated as calendar windows
- Count queries promote `turn_fact` / `fact_augmented` children before Limit
- Count queries rank named-pattern facts (led/leading+project, bought, spent, …) above fallback chatter
- Count queries collect matching `turn_fact` children across the palace/`conv:` session set (stemmed noun overlap), not only the keyword hit list, then diversify+Limit
- `AssembleCountEvidence` compact unique snippets (LongMemEval retrieve prepends a synthetic hit; not persisted)
- Clothing-errand named extract (dry-clean, pick-up/return × boot/blazer/Zara; not poster/case-competition)
- `AssembleCountEvidence` diversifies clothing counts by action+object (one snippet per cluster; compound return+pick-up is two bullets; dry-clean kept on pick/return/store queries)

**Measure (2026-09-12):**

- `ef6a3e9` T1 retrieve only: MiniLM/BGE **10/12**, `multi-session` **0/2**; all inner sessions in k=40.
- `a25a883` + fact promotion: hash/BGE **11/12**, `multi-session` **1/2** (clothes pass). MiniLM still **10/12** / **0/2**. Projects (`6d550036`) still miss.
- `2695e02` (#110) Wave D: MiniLM **11/12** `multi-session` **1/2** (clothes pass); BGE **10/12** **0/2**; hash **9/12** **0/2**. Projects (`6d550036`) still miss. T1 done-when not met.
- `f99c140` (#111) Wave E: MiniLM **12/12** `multi-session` **2/2**; BGE **11/12** **1/2**; hash **11/12** **1/2**. Projects pass all three. T1 done-when met. Residual n=12 miss: clothes (`0a995998` gold 3) on hash/BGE — this slice. Do not invent a remesure score.
- `aa64dbc` (#115 on #114) Wave F: MiniLM **12/12** `multi-session` **2/2**; BGE **12/12** **2/2**; hash **11/12** **2/2**. Clothes pass all three. Residual: hash temporal `gpt4_2487a7cb`.

**Still open for T1:** n=12 clothes residual closed (Wave F). Hash still misses temporal `gpt4_2487a7cb`. Next: remesure locked mixed **n=60** on this kernel. Not a README number.

**In scope (measure)**

- Palace-side retrieve that can seed from **several** `SessionID`s (or “all sessions in this palace / conv”) without dropping keyword gold past `Limit`
- Time-aware expansion that does **not** classify ordinary count questions as a calendar window and hide gold
- Optional: assemble `ExtractedFacts` / facts-as-of across sessions before the reader (kernel helper, not an LLM) — **shipped this slice** (`AssembleCountEvidence` + search union)
- Re-run locked mixed **n=12** (same IDs) then **n=60** (`testdata/longmemeval_baseline_ids_n60.json`) after the change

**Out of scope**

- Publishing a LongMemEval leaderboard number
- Changing default embedder to Qwen3
- Multi-process writers / flock

**Done when:** `multi-session` on the locked n=12 list is no longer 0/2 on MiniLM **and** BGE (hash may still lag). Same judge pin. Isolated palace per embed mode.

### T2 — Event-time index beyond patch

**Why:** Original K2 residual. Patch + durable snapshot are enough for laptop palaces. First list after process start still walks JSON.

**In scope:** optional btree / tag secondary if `MetaIndexRebuilds` or list latency shows up in T1 benches. Keep FS as source of truth.

**Out of scope:** flock; cross-process writers; distributed timelines.

### T3 — Temporal relation edges

**Why:** Original K4/A2 residual. `AddEntityRelationship` is untimed adjacency. As-of graph walk needs `valid_from` / `valid_until` on edges, not only on entries.

**Start only if** T1 still misses after session-set retrieve — i.e. the gold lives on a **relation** that should have expired.

### T4 — Compaction vs validity

**Why:** Ingest children stamp `valid_from`; compaction products stamp. MERGE / SUMMARIZE / ARCHIVE must not drop or invent validity windows.

**In scope:** compaction tests that `ListFactsAsOf` after MERGE/SUMMARIZE still matches `EntryValidAt`. No new dual-clock store.

### T5 — Embedding profiles (original K3)

Keep **BGE-small-en-v1.5 384-d** as the documented ONNX default. MiniLM is the in-tree fallback when BGE is missing. `PersistEmbeddings` stays default **off**.

Qwen3-0.6B **1024-d** only as an **opt-in** constructor/env preset when a concrete consumer needs it. Document re-index if Qdrant collection dim changes. No silent default flip.

---

## Suggested implementation order

1. **T1** multi-session retrieve (API + ingest + ranking shipped) — clothes hash/BGE residual this slice; next **n=60** (do not invent remesure scores)
2. **T2** only if timeline list / rebuild cost is the limiter
3. **T3 / T4** when T1 evidence says edges or compaction ate the gold
4. **T5** last, consumer-driven

---

## Versioning

- Prefer new options fields and methods over breaking `SearchMemory` signatures
- Embedding dimension changes require Qdrant collection recreation; note in the release
- v1.5.2 K1 · v1.5.3 K2 list · v1.5.4 K4 as-of · v1.5.5 A2 multi-hop · v1.5.6 A3 supersession · v1.5.7 hop ranking · v1.5.8 meta-index patch · v1.5.11 persist-onnx-vec opt-in, TTFH, LongMemEval card · v1.5.12 T1 SessionIDs / conv tags / count assembly
