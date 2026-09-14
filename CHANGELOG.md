# Changelog

All notable changes to this project (Palace memory kernel) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **Restaurant stop-token `as`:** unique-entity restaurant clusters drop leftover `[restaurant:as]` (Wave L leftover cluster). Phrase stops include `as` and `is` (`at` already). Does not invent a numeric gold. Clothing N-header unchanged. Wave M hash n=12 kernel `7a9b956` still leftover `[restaurant:as]` is **before** this PR.
- **Restaurant tried-count:** unique-entity restaurant evidence lists latest-first “tried N” mentions; cuisine+BBQ dishes are not venues; stop-token `if` ([#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`). Does not invent a numeric gold. Wave J hash n=12 still 3 vs 4 is **before** this PR. Wave K hash n=12 `6aeb4375` **PASS**.
- **Latest-value clip:** long snippets keep the dollar amount in the visible window so a later `$400,000` is not truncated off ([#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5`). Does not NLP-supersede; stale amounts still listed. Remesure of KU `852ce960` **TBD** (n=60, not n=12).
- **`longmemeval-v1-card` scored run LIMIT:** with `LONGMEMEVAL_QA_LIMIT` unset/0, generate the full mixed oracle (**n=500**) and do not pass `--limit` (no longer coerced to n=12). `LONGMEMEVAL_QA_LIMIT=12` remains an explicit mixed sample, **not** official V1. Prefix-n is not V1. No README number. dual_write OFF. not Memory GA. Hash is not V1.

### Changed
- **ListFactsAsOf via meta index:** session/tier/query collect through the list meta index (no Limit); `EntryValidAt` / entity still after load. `DisableMetaIndex` keeps the JSON walk. T4 compaction contract unchanged. btree still gated.
- **README companion install:** optional TUI/MCP `go install` names Go **1.27+** (this module and both hosts) and prepends `$(go env GOPATH)/bin` to `PATH` so `iomesh` is discoverable after install.
- **Temporal roadmap:** [`docs/temporal-memory-kernel-roadmap.md`](docs/temporal-memory-kernel-roadmap.md) records restaurant tried-count **shipped** ([#146](https://github.com/iome-sh/memory/pull/146) `89c1dc0`): latest-first “tried N”; cuisine+BBQ not venues; stop-token `if`. Wave J hash n=12 still 3 vs 4 is **before** #146. Wave K hash-only n=12 remesure **11/12** (`6aeb4375` **PASS**, [#148](https://github.com/iome-sh/memory/pull/148) `fe1e66f`). **T6++** temporal-order extrema **shipped** ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`). Wave L hash-only n=12 remesure **11/12** (temporal `gpt4_2487a7cb` **PASS**, `text dates earliest: two months ago · latest: last Saturday`). Dated-span ago vs `question_date` **shipped** ([#150](https://github.com/iome-sh/memory/pull/150) `76175a0`). Wave M hash-only n=12 remesure **11/12** (kernel `7a9b956`; temporal `gpt4_2487a7cb` **PASS**, which-first not ago). Official V1 **run 2** recorded (same pin `9bee542`, **384/500**; first run **388/500**). Two runs not identical (reader/judge variance). INTERNAL unpublished. Not a README number. Not Memory GA. Do not publish a single %. T6+ **shipped** ([#141](https://github.com/iome-sh/memory/pull/141) `7deafa5`); latest-value clip **shipped** ([#143](https://github.com/iome-sh/memory/pull/143) `5d3aca5`). Clothes park; T2 btree / T3 edges / T5 Qwen3 / T8 dual-clock still parked. Shipped T1–T7 / T4-perf / T6 days-only / T7+ restaurants unchanged. btree / typed edges / Qwen3 default still gated.
- **README:** install pin **v1.5.12**, companion hosts TUI **v1.3.7** / MCP **v0.4.2**, table of contents, compiling quickstart, `SessionIDs` on the search table, LongMemEval details moved to `docs/`. Release badge. No published LongMemEval number.

### Added
- **Docs:** internal unpublished official V1 mixed n=500 BGE **run 2** (384/500, same pin `9bee542` as first 388/500, judge gpt-4o-2024-08-06). Two runs not identical (reader/judge variance, not a kernel change). Not a README number. Not Memory GA. Do not publish a single %. dual_write OFF.
- **Wave M locked mixed n=12 hash-only (unpublished):** kernel `7a9b956` after [#150](https://github.com/iome-sh/memory/pull/150) `76175a0` dated-span ago vs `question_date`. Hash **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered `1. 2. 3.` + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. Temporal `gpt4_2487a7cb` **PASS** (`text dates earliest: two months ago · latest: last Saturday`; n=12 temporal is which-first, not ago). KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no `[restaurant:korean-style-bbq]` / `[restaurant:if]`; leftover `[restaurant:as]` — **before** [#152](https://github.com/iome-sh/memory/pull/152)). MiniLM/BGE **not run**. **Not a README number.** **Not official V1.**
- **Dated-span ago vs question_date:** retrieve may pass `question_date`; temporal evidence appends text-date vs that instant for how-many-days/weeks/months-ago. Not a retrieve TimeTo filter. Not gold. Not ingest Timestamp.
- **Wave L locked mixed n=12 hash-only (unpublished):** kernel `d5bec89` after [#147](https://github.com/iome-sh/memory/pull/147) `298584c` T6++ temporal-order extrema. Hash **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered `1. 2. 3.` + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. Temporal `gpt4_2487a7cb` **PASS** (`text dates earliest: two months ago · latest: last Saturday`). KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no `[restaurant:korean-style-bbq]` / `[restaurant:if]`). MiniLM/BGE **not run**. **Not a README number.** **Not official V1.**
- **Wave K locked mixed n=12 hash-only (unpublished):** kernel `89c1dc0` after [#146](https://github.com/iome-sh/memory/pull/146) restaurant tried-count + cuisine-BBQ not venue. Hash **11/12** MS **1/2**. Clothes `0a995998` FAIL (numbered `1. 2. 3.` + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. Temporal `gpt4_2487a7cb` PASS. KU `6aeb4375` **PASS** (hyp four; latest-first tried-N; no `[restaurant:korean-style-bbq]` / `[restaurant:if]`). MiniLM/BGE **not run**. **Not a README number.** **Not official V1.**
- **Temporal-order extrema:** which-first evidence appends `text dates earliest: … · latest: …` from parsed text times ([#147](https://github.com/iome-sh/memory/pull/147) `298584c`). `how many days/weeks/months ago` is a dated-span query (no gold). Not ingest Timestamp.
- **Wave J locked mixed n=12 hash-only (unpublished):** kernel `5d3aca5` after [#136](https://github.com/iome-sh/memory/pull/136) restaurant clusters + [#143](https://github.com/iome-sh/memory/pull/143) latest-value clip. Hash **10/12** MS **1/2**. Clothes `0a995998` FAIL (numbered `1. 2. 3.` + N=3 in retrieve; reader summed 2). Projects `6d550036` PASS. Temporal `gpt4_2487a7cb` PASS. KU `6aeb4375` FAIL (3 vs gold 4; count-evidence clustered `[restaurant:korean-style-bbq]` / `[restaurant:if]`). MiniLM/BGE **not run**. **Not a README number.** **Not official V1.**
- **Dated-span weeks/months:** `AssembleTemporalEvidence` adds text-date week (floor days/7) and calendar-month deltas when the query asks how-many-weeks/months. Days line unchanged. Not ingest Timestamp. Does not invent gold.
- **Docs:** internal unpublished official V1 mixed n=500 BGE card (388/500, judge gpt-4o-2024-08-06, kernel `9bee542`). Not a README number. dual_write OFF. not Memory GA.
- **Locked mixed n=60 after #119+#122+#124+#129+#132 (unpublished):** kernel `5154a76` (`origin/main` at ingest; includes #134 T4-perf). Unlike v1.5.12 `e90a82d` (before #119). hash **49/60** MS **9/10**; MiniLM **51/60** MS **9/10**; BGE **47/60** MS **8/10**. Complete 60/60. Kits `gpt4_59c863d7` / hours `aae3761f` / plants `3a704032` pass all three. Days-between `2a1811e2` / `2c63a862` pass; `08f4fc43` miss all three. KU `852ce960` still $350k vs $400k all three; restaurants `6aeb4375` 3 vs 4 all three. Clothes `0a995998` MiniLM pass, hash/BGE miss. [#136](https://github.com/iome-sh/memory/pull/136) restaurant clusters **not** this remesure. **Not a README number.** **Not official V1.**
- **Unique-entity restaurant clusters:** `AssembleCountEvidence` clusters distinct restaurant names for how-many-restaurant questions; catalogs are aliases. Clothing N-header unchanged. Does not invent a numeric gold.
- **Wave I locked mixed n=12 (unpublished):** kernel `e094bec` after #122+#124+#127+#129+#132. hash/MiniLM/BGE **10/12** MS **1/2**. Clothes `0a995998` miss all three (numbered `1. 2. 3.` + N=3 in retrieve; reader summed 2). Projects `6d550036` pass all three (MiniLM recovered vs Wave H). Temporal `gpt4_2487a7cb` pass. KU `6aeb4375` miss all three (3 vs gold 4). `852ce960` not in n=12. #134 T4-perf not this remesure. **Not a README number.** **Not official V1.**
- **Dated-span text-date delta:** `AssembleTemporalEvidence` appends `text dates N days apart (phrase → phrase)` from parsed text times (not ingest Timestamp) when a how-many-days-between query has ≥2 dated bullets. Does not invent a gold answer.
- **T2 list-latency bench:** `BenchmarkListMemoryWithOptions_SessionTimeLimit` (meta index vs `DisableMetaIndex` O(n) fallback). Optional `BenchmarkSearchMemoryWithOptions_CountQuery` documents skip-vector (#122). btree / tag secondaries stay gated until this bench shows list rebuild cost.
- **Search/count candidates via meta index:** `collectSearchCandidates` applies session/time/tier through the existing list meta index and loads full JSON only for survivors. `unionCountQueryFacts` still unions that candidate slice (no second palace walk). `DisableMetaIndex` keeps the O(n) `ListEntriesInTier` path. Keyword haystack still includes Keyphrases / ExtractedFacts; AsOf still filters after load. btree / tag secondaries remain parked (warmed list at N=200 is not rebuild-bound).
- **T4 compaction products remain ListFactsAsOf-visible:** after MERGE/SUMMARIZE, `ListFactsAsOf({SessionID, AsOf: now})` still returns the product; sources move to archival without invented `valid_until`. ARCHIVE that only moves tiers stays valid at now.
- **Unique-entity count clusters:** `AssembleCountEvidence` clusters kits, plants, and hour+destination facts by distinctive object (repeated B-29 is one kit; two plants in one turn are two clusters; word numbers). Clothing action×object clustering is unchanged.
- **Temporal dated-event evidence:** `AssembleTemporalEvidence` lists unique dated bullets with the text date phrase labeled separately from ingest `Timestamp` (RFC3339) for which-first / how-many-days-between queries. Search promotes entries that mention either event name; LongMemEval retrieve prepends a synthetic `temporal_evidence` hit (not persisted). Dated-span how-many is still not a calendar window.
- **Latest-value evidence:** `AssembleLatestValueEvidence` lists dollar/scalar values matching the query entity, later `Timestamp` first, for amount / pre-approved / how-much-was-I questions. Search unions matching amount facts across sessions and ranks later sessions first. LongMemEval retrieve prepends a synthetic `latest_value_evidence` hit (not persisted). Does not NLP-supersede.

### Changed
- **Unique-entity clusters generalize:** hours extract dest after `N hours to/in/for/at`; kits extract `… kit` / `model kit` noun phrases; catalogs remain aliases. Clothing path unchanged. Does not invent a numeric gold.
- **Numbered clothing count-evidence bullets:** `AssembleCountEvidence` prefixes clothing action×object clusters with `1. ` `2. ` `3. ` in cluster order and keeps `(N distinct items)`. Unique-entity and exact-text paths stay unnumbered without N. Does not invent a numeric gold.
- **Count-evidence N is clothing-only:** `AssembleCountEvidence` prefixes `(N distinct items)` only for clothing action×object clusters (dry-clean/return/pick-up). Unique-entity and exact-text paths keep `Count evidence:` without N. `AssembleTemporalEvidence` still prefixes `N distinct events` when multiple bullets. Does not invent a numeric gold.
- **Skip vector on count/temporal retrieve:** `SearchMemoryWithOptions` does not call `scoreEntriesByVector` for count or temporal-order/dated-span queries even when `QueryVec` is set (keyword + evidence assembly). LongMemEval retrieve skips computing `QueryVec` for those classes.

## [1.5.12] — 2026-09-12

T1 multi-session retrieve and count-query assembly. Not official V1. Not a README number.

### Added
- **T1 multi-session retrieve:** `SearchMemoryOptions.SessionIDs` / `ListMemoryOptions.SessionIDs` / `FactsAsOfOptions.SessionIDs` / `MultiHopOptions.SessionIDs` (any-of). `conv:<id>` tags group inner haystack sessions so retrieve with the conversation id still finds them. Search diversifies distinct `SessionID`s before Limit. LongMemEval ingest keeps per-turn `session_id` and stamps `conv:<conv_id>`. Count questions (`how many` / `how much`) are not calendar windows.
- **T1 count-query assembly:** count queries union matching `turn_fact` / `fact_augmented` / `atomic_fact` children from the session/`conv:` set, rank named-pattern facts (including `led`/`leading` + project/team) above fallback chatter, then diversify+Limit. `AssembleCountEvidence` lists unique snippets and clusters clothing errands (dry-clean / return / pick-up). LongMemEval retrieve prepends a synthetic hit (not persisted).
- **LongMemEval batch retrieve scoring:** `cmd/longmemeval-server` sets `BatchEmbeddingFunc` like `cmd/longmemeval-bench`, so retrieve scores palace candidates in one ONNX forward pass when vectors are not persisted (library default). Optional `LONGMEMEVAL_PERSIST_EMBEDDINGS=1` persists ONNX vectors only (never hash).
- **Locked mixed LongMemEval harness (docs):** [`docs/LONGMEMEVAL_BASELINE.md`](docs/LONGMEMEVAL_BASELINE.md) + `make longmemeval-baseline` / `testdata/longmemeval_baseline_ids.json` (n=12) and `testdata/longmemeval_baseline_ids_n60.json` (n=60). Same IDs × hash vs MiniLM vs BAAI BGE. **Not official V1** (not mixed 500). **Not a README number.**
- **TTFH walking skeleton:** [`docs/TTFH.md`](docs/TTFH.md) + `go run ./examples/ttfh_rca`. Optional host path: iomesh-tui **v1.3.6** + iomesh-memory-mcp **v0.4.1** — `/memory ingest` then `/memory digest --require-sources mesh,private` (cite-both or explicit miss; miss is success). Cost-max: hash embedder, no Qdrant, no cloud palace. Do not stamp mesh on local overlay.
- **Two-process writer probe:** `scripts/two_process_writer_probe.sh` and `cmd/two-process-writer-probe`. Multi-process writers remain **unsupported**; flock is not shipped. Probe ≠ lock.
- **LongMemEval methodology runners:** `make longmemeval-v1-card` prints dataset/SHA/tag/histogram/embed/judge pin. Missing oracle is SKIP (exit 0). Unpublished MiniLM mixed-run card in [`docs/LONGMEMEVAL.md`](docs/LONGMEMEVAL.md). Official V1 pin stays BGE-small-en-v1.5 + judge `gpt-4o-2024-08-06`. No published score.

### Changed
- **ONNX download fail-soft:** Hugging Face 401/404 for `KnightsAnalytics/bge-small-en-v1.5` no longer `os.Exit(1)`. Public `BAAI/bge-small-en-v1.5` `onnx/model.onnx` reshapes into hugot layout `testdata/models/BAAI_bge-small-en-v1.5/` (gitignored). MiniLM remains the in-tree fallback. `HF_TOKEN` is optional; 401/403 retry without Authorization; 404 is not retried. TTFH cost-max is still hash.
- **QA reader uses full retrieve-k:** `LONGMEMEVAL_READER_K=0` feeds all retrieved snippets to the reader (was `memories[:15]` while retrieve k=40).
- **Temporal kernel roadmap:** [`docs/temporal-memory-kernel-roadmap.md`](docs/temporal-memory-kernel-roadmap.md) — T1 n=12 done-when met; next n=60 then T2 only if list latency. T3/T4/T5 parked.
- **Operator copy / public-docs hygiene:** README, SUPPORT, SECURITY, RELEASING, CONTRIBUTING, operator docs. Maintainer residuals marked; walking skeleton, LongMemEval methodology, single-writer, and inspectable FS facts unchanged.

## [1.5.11] — 2026-09-12

Kernel-only · not Memory GA · dual_write OFF.

### Added
- **Optional ONNX vector persist (default off):** `PalaceConfig.PersistEmbeddings` (default **false**) plus `EmbeddingModel` / `EmbeddingDim`. `MemoryContent` may store `embedding` / `embedding_model` / `embedding_dim` only when the flag is on and `EmbeddingModel` is a non-hash id (e.g. `bge-small-en-v1.5`). Hash / empty / `"hash"` models never persist `GenerateSimpleEmbedding` vectors as stored vectors / `QueryVec` (kernel #45). Write acks JSON first (`CreateTemp`+`chmod 0600`+`Rename`); embed miss, empty output, or panic is not an ingest failure. `SearchMemoryWithOptions` reuses a matching persisted vec instead of re-embedding; keyword hits still rank ahead of `Limit`. Compaction copies a parent embedding only when model+dim match the store config, else drops. Not a hot vector index. Inspectable FS palace remains source of truth. usearch / ORT / Qdrant stay optional. Kernel-only · not Memory GA · dual_write OFF.
- **TTFH-shaped quickstart (docs):** [`examples/ttfh_rca`](examples/ttfh_rca) ingests three RCA-shaped turns, retrieves in the same process, lists facts-as-of, and prints `provenance.source_hint`. README leads with that path (not a chatbot colour demo). Host companion pins: iomesh-tui **v1.3.3** + iomesh-memory-mcp **v0.3.2**.
- **LongMemEval methodology card (docs):** [`docs/LONGMEMEVAL.md`](docs/LONGMEMEVAL.md). Official V1 = upstream `evaluate_qa.py` + judge `gpt-4o-2024-08-06`. Hash overlap stays unpublished. No official number in this change.

### Changed
- **Default write path unchanged:** `PersistEmbeddings` defaults false; hash SearchMemory / ingest behavior is unchanged unless a caller opts into a non-hash model. Kernel-only · not Memory GA · dual_write OFF.
- **Buyer table + naming collision (docs):** README states when to use Palace versus Mem0 / Graphiti / Letta / Cognee / LangMem, and that MemPalace/`mempalace` is an unrelated Python project.
- **Single-writer contract (docs):** README + SECURITY.md name one process per palace root as the supported topology (multi-process writers unsupported — product contract).
- **gofmt / EditorConfig:** `.editorconfig` pins `*.go` to **tabs** (`indent_style=tab`, `indent_size=8`) — the `gofmt` standard. `embedding_persist.go` and the rest of the tree are `gofmt -w` / `make fmt-check` clean (tabs, not spaces). Kernel-only · not Memory GA · dual_write OFF.

## [1.5.10] — 2026-09-10

Kernel-only · not Memory GA · dual_write OFF.

### Fixed
- **Private ingest source class (#90):** `IngestTurn` stamps observable `provenance.source_hint=private` and tag `source_hint:private` on the parent and inherited `turn_fact` children when the caller does not already supply a classifiable mesh or private source. Host process labels (`mcp_memory_ingest_turn`, `source:iomesh-memory-mcp`) are not a cite-both class. Mesh-class hints (`source_hint:mesh`, `source:mesh`, …) stay distinct. `Write` persists caller `source_hint` / tags as-is (no default stamp). Kernel-only · not Memory GA · dual_write OFF.

## [1.5.9] — 2026-09-10

Kernel-only · not Memory GA · dual_write OFF.

### Changed
- **hugot 0.7.8 (#84):** bump `github.com/knights-analytics/hugot` 0.7.7 → 0.7.8. Kernel-only · not Memory GA · dual_write OFF.
- **Palace mode bits (#85):** `ensureDirs` / `MkdirAll` use `0700`; entry, version, `entity-graph.json`, and `event-time.json` writes use `0600`. Fresh-palace tests assert modes. Kernel-only · not Memory GA · dual_write OFF · not encryption at rest.
- **Shared graph/index writeMu (#86):** `relations/entity-graph.json` and `indexes/event-time.json` rewrite under `writeMu` via temp+rename (`0600`). Per-entry rename was already atomic; concurrent `Write` / `IngestTurn` / `AddEntityRelationship` keep those shared files readable JSON. Not async 500ms ingest. Kernel-only · not Memory GA · dual_write OFF.
- **Retrieve default tiers skip archival (#87):** `SearchMemory` / `SearchMemoryWithOptions` default to Working+Contextual+Semantic (same as list). Archival is included when `IncludeArchival` is set, an explicit `Tier` is Archival, or the default-tier keyword hit set is empty (low-confidence fallback — not a numeric cosine cutoff). Kernel-only · not Memory GA · dual_write OFF.
- **Public name hygiene:** drop product-plane names from user-facing docs, godoc, Makefile, and readiness-gate needles. Honesty stays: kernel-only · not Memory GA · dual_write OFF (host policy) · this package does not import private control-plane / broker packages · private control-plane / broker stays private.

## [1.5.8] — 2026-09-04

Kernel-only · not Memory GA · dual_write OFF.

### Added
- **K2 incremental meta-index patch (#63):** `Write` / unlink patches the in-memory (and optional durable) event-time meta index when it is already clean, so `ListMemoryWithOptions` does not rebuild-on-dirty from every tier JSON. FS Palace remains source of truth. First list / `InvalidateMetaIndex` / stamp mismatch still O(n). Btree/tag secondaries remain residual. Kernel-only · not Memory GA · residual ≠ invent index green.

### Changed
- **Public copy hygiene:** operator-facing README, RELEASING, SECURITY, and OPEN_SOURCE_AUDIT drop internal serials and private-plane names. Public MIT · local filesystem library · not Memory GA · not cloud multi-tenant.
- **`IngestTurn` persist honesty (#64):** godoc (and `TestIngestTurn_FactWriteError`) state the contract — parent and earlier facts remain after a child Write error. Partial persist · not rollback · not Memory GA · dual_write OFF.
- **Write versioning honesty (#65):** document `Write` / `WriteLatent` as caller-managed and best-effort. `Version==0` becomes 1; `archiveToVersions` errors are non-fatal; overwrite does not auto-increment. README no longer claims automatic overwrite versioning. Kernel-only · not Memory GA · not a hosted version store.
- **Empty `PalaceConfig.BaseDir` (#66):** `NewPalaceStoreWithConfig` / `NewPalaceStore("")` default to local `.palace` (`DefaultPalaceBaseDir`) instead of leftover `.ossa/kb/palace`. Callers that pass `BaseDir` are unchanged. Local-primary · not a hosted palace · not Memory GA.

### Fixed
- **govulncheck GO-2026-6355 / GO-2026-6354:** bump `golang.org/x/crypto` to v0.56.0 (ssh.Dial via hugot.NewPipeline). Kernel-only · not Memory GA · dual_write OFF.
- **`IngestTurn` fact-child tags (#78):** derived `turn_fact` children inherit the parent turn's `Content.Tags` and keep `fact_augmented` / `from_turn`. The kernel no longer stamps `longmemeval`; the LongMemEval harness (`cmd/longmemeval-*`, `internal/longmemeval`) supplies it on the parent. Kernel-only · not Memory GA · dual_write OFF.
- **Public-flip priced-rate strip (#75):** `docs/PUBLIC_FLIP_READINESS.md` and `docs/OPEN_SOURCE_AUDIT.md` drop priced product-surface figures. Palace sunset and mesh-optional honesty stay without SKU or dollar figures. Gate forbids those needles. Kernel-only · not Memory GA · dual_write OFF (host policy).
- **Internal close-token strip (#74):** CHANGELOG, RELEASING, recmem residual, and the longmemeval-server comment rephrase without the internal close-token. Judge-free overlap is not an official scored close; compaction stays advisory. Kernel-only · not Memory GA · dual_write OFF (host policy).
- **Tier-change persist (compaction / evict / promote):** ARCHIVE, summarize/merge/core-principle source archive, `EvictWorkingTier`, and `PromoteToContextual` unlink the source-tier JSON after a successful destination write (`unlinkEntry` patches the incremental meta-index). A tier change is a move, not a silent copy. `handleArchive` returns Write/unlink errors. Kernel-only · not Memory GA · dual_write OFF · not incremental/btree index green.
- **Public-flip residual docs (#62):** `docs/PUBLIC_FLIP_READINESS.md`, `docs/OPEN_SOURCE_AUDIT.md`, CONTRIBUTING public-repository policy, and the Makefile `public-flip-readiness-gate` comment treat flip-complete as current fact. Pre-flip “still private” language is historical. Public MIT ≠ Memory GA · kernel-only · gate PASS ≠ product GA · dual_write OFF (host policy, not kernel flag).
- **LongMemEval `/ingest` persist errors (#61):** failed `IngestTurn`/`Write` no longer increment ingested or return blanket `status: ok`. Local bench harness (not a production memory service). Qdrant opt-in via `LONGMEMEVAL_QDRANT_URL` (not implied live hybrid ingest). not Memory GA · not live ingest · dual_write OFF · judge-free overlap ≠ official scored close.
- **LongMemEval `/retrieve` session scope (#55):** official generate/orchestrator pass `session_id`. Shared-palace QA without it is other-session dominated. Kernel `SearchMemoryOptions.SessionID` already existed. Harness-only · not Memory GA.
- **Keyword OR-any-token flood (#56):** keyword hits still select on any token ≥3, but are ranked by distinct-token overlap before Limit so hash top-k cannot bury a unique gold phrase under incidental `when`/`did`/`last` matches. Kernel-only · not Memory GA.

### Added
- **Official QA `question_date` + retrieve snippets (#57):** generate prompt includes question date and per-hit timestamps/session ids. Hypothesis JSONL persists snippets + `embed_mode`. Not a substitute for session-scoped retrieve.
- **Mixed-type LongMemEval slice (#58):** `--sample mixed` / `LONGMEMEVAL_QA_SAMPLE=mixed` stratifies by `question_type`. Default `--limit N` warns that prefix-n is temporal-first, not overall V1. overlap ≠ gpt-4o ≠ V2 LAFS.

### Fixed
- **LongMemEval harness dates (#51):** `cmd/longmemeval-bench` and `cmd/longmemeval-server` parse official cleaned `haystack_dates` (`2006/01/02 (Mon) 15:04`) as well as RFC3339. In-repo subset includes one official-format date so smoke cannot regress to RFC3339-only. Harness-only · not PalaceStore API · not Memory GA.

### Added
- **LongMemEval-V2 loader + Insert/Query adapter (#52):** `internal/longmemeval` loads official `questions.jsonl` + `trajectories.jsonl` + `haystacks/lme_v2_<tier>.json` without vendoring the 7GB snapshot. `PalaceMemory` maps text steps into `IngestTurn` / `SearchMemory`. `cmd/longmemeval-v2-bench` is optional. Images later. Hash default. dual_write OFF. Not a leaderboard submit. Not Memory GA.

### Changed
- **Overlap vs official scores (#53):** README, `scripts/longmemeval_recall_bench.sh`, and bench stderr state that printed `recall` is judge-free top-k gold-answer string overlap — not official V1 gpt-4o QA and not V2 LAFS. Hash overlap is not a leaderboard number.

### Fixed
- **Compaction / SemanticRefine products (K4 leftover):** stamp `valid_from:<RFC3339>` when unset, copy `SessionID` and `Timestamp` (when set) from the first parent, and return product `Write` errors instead of discarding them. Ingest children stamped · compaction products now stamped · bi-temporal lite · not dual-clock KG · not NLP extract · not Memory GA · not incremental/btree index green.

### Changed
- **Compaction last-stamp + verify:** `PerformCompaction` sets `MemoryStats.LastCompaction` when a pass actually runs (non-empty target tier). `GetStats` reports the in-process stamp. `verifyAction` rejects unknown actions and missing/blank/unknown target IDs (allowlist: `SUMMARIZE`, `CREATE_CORE_PRINCIPLE`, `ARCHIVE`, `MERGE`; `MERGE` needs two IDs). Kernel-only · not Memory GA · host `memory_trigger_compact` stays advisory · RecMem leftover stays residual · not incremental/btree index green.

### Changed
- Go toolchain pin `go 1.26.6` so CI `govulncheck` is clean on stdlib GO-2026-5972 / GO-2026-5026 (fixed in go1.26.6).

### Added
- **K2 durable event-time snapshot (#44):** `indexes/event-time.json` best-effort on-disk meta index. Fresh processes skip re-parsing every tier JSON when the stamp (JSON count + max mtime) matches. FS Palace remains source of truth; dirty writes still rebuild from disk. `DisableDurableIndex` opt-out. Incremental/btree/tag secondary indexes remain residual. Kernel-only · not Memory GA · residual ≠ invent index green.

### Changed
- Go toolchain pin `go 1.26.6` so CI `govulncheck` is clean on stdlib GO-2026-5972 / GO-2026-5026 (fixed in go1.26.6).

### Fixed
- **`IngestTurn` fact children (K4 leftover):** stamp `valid_from:<RFC3339>` when unset and return child `Write` errors instead of discarding them. Bi-temporal lite · not dual-clock KG · not NLP extract · not Memory GA.

### Changed
- Go toolchain pin `go 1.26.6` so CI `govulncheck` is clean on stdlib GO-2026-5972 / GO-2026-5026 (fixed in go1.26.6).

### Fixed
- **`SearchMemoryWithOptions` hybrid recall (#45):** a non-empty `QueryVec` no longer skips the keyword path. Literal token hits stay ahead of cosine rank so hash embeddings (`GenerateSimpleEmbedding`) cannot drop an exact unique token past `Limit`. ONNX/semantic neighbors still fill remaining slots. Kernel-only · not Memory GA.
- **`ReRankTemporal` vs keyword `Limit`:** temporal re-rank no longer drops literal keyword hits past `Limit` (same underfill class as #45). Kernel-only · not Memory GA.

### Changed
- **Search keyword haystack:** `filterEntriesByKeywords` / `SearchMemoryWithOptions` now search `Summary` + `Full` + `OriginalText` + `Keyphrases` + `ExtractedFacts` (space-joined). Aligns keyword recall with list-path text plus fields `IngestTurn` already fills. `MultiFactorScore` uses the same haystack. Tokenizer unchanged (non-alnum split, length ≥ 3). Kernel-only · not Memory GA · dual_write N/A · not incremental/btree index green.
- **Public OSS:** repository is public MIT; docs drop still-private flip residual; `go get` without `GOPRIVATE`.


### Added

- Final private→public flip audit closeout (s1473): TUI-parity process bar for this Go **library** module — CONTRIBUTING public repository policy · Issues & discussions · CI/branch-protection table · `OPEN_SOURCE_AUDIT` final TUI-parity matrix · `PUBLIC_FLIP_READINESS` final pre-flight checklist · RELEASING library/no-GoReleaser note — **still private** · **residual PASS ≠ public flip** · does **not** flip visibility · kernel first then `iomesh-memory-mcp`.
- M4 public-flip **readiness** residual (s1467): `docs/PUBLIC_FLIP_READINESS.md` operator checklist · `scripts/public_flip_readiness_gate.sh` · `make public-flip-readiness-gate` — **still private** · **residual PASS ≠ public flip** · kernel first then `iomesh-memory-mcp` · does **not** flip visibility.
- TUI-grade open-source process bar (s1452): LICENSE (MIT), NOTICE, SECURITY, CONTRIBUTING, CODE_OF_CONDUCT, SUPPORT, RELEASING, OPEN_SOURCE_AUDIT, GitHub PR/issue templates, Dependabot, CI lint/govulncheck/ci-success, `make ci` — **repository remains private** until a deliberate visibility flip.

### Changed

- `docs/OPEN_SOURCE_AUDIT.md` continuum stamp to s1473 final TUI-parity audit (still private intentional Pass); links `PUBLIC_FLIP_READINESS.md`.
- Go toolchain pin `go 1.26.5`; transitive security bumps (`google.golang.org/grpc`, `golang.org/x/{crypto,text,image}`) so `govulncheck ./...` is clean on called symbols.
- gofmt across package for CI `gofmt -l` gate.

### Honesty

- Kernel-only · not product Memory GA · public MIT ≠ Memory GA · local-primary · dual_write OFF (host policy, not kernel flag) · hosted Palace sunset · future MCP host **iomesh-memory-mcp** · private control-plane / broker stays private · residual PASS ≠ public flip · gate PASS ≠ product GA · M4 readiness ≠ invent Memory GA.

## [1.5.7] — 2026-07

### Added

- **Hop-distance ranking** (s1067 / A2 residual): `MultiHopRetrieve` prefers shorter BFS hop distance from seed, then event time desc within hop. `ExpandRelatedEntitiesHops`; `PreferShorterHops` defaults true.
- **K2 event-time meta index** (s1066): in-memory entry meta index for `ListMemoryWithOptions` (residual: full durable event-time index still residual-honest).
- Residual honesty pins: s1278 hop-distance · s1297 advanced agent inventory · s1303 K2 event-time index · s1313 RecMem compaction (`make residual-gate`).

### Notes

- Kernel-only; not product Memory GA. BGE-small-en-v1.5 (384-d) default unchanged.

## [1.5.6] — 2026-07

### Added

- **`SupersedeEntityFacts`** (s632 / A3 first slice): close prior open validity windows for an entity key via `valid_until:<RFC3339>` (exclusive end).
- **`WriteAndSupersede`**: write entry first (stamps `valid_from=now` when unset), then supersede each key excluding the new entry ID.

### Notes

- Competitive lite supersession — not automatic NLP contradiction detection; not full Zep dual-clock KG.

## [1.5.5] — 2026-07

### Added

- **`MultiHopRetrieve` / `MultiHopOptions`** (s619 / A2 first slice): multi-hop lite associative retrieval over EntityGraph BFS.
- **`ExpandRelatedEntities`**, **`EntryEntityKeys`**, **`AddEntityRelationship`** helpers.

### Notes

- Multi-hop lite — not full Zep / Graphiti KG.

## [1.5.4] — 2026-07

### Added

- **`ListFactsAsOf` / `FactsAsOfOptions`** (s616 / K4 first slice): as-of validity listing over host-written validity tags.
- **`ParseValidityWindow` / `EntryValidAt`**; **`SearchMemoryOptions.AsOf`**.

### Notes

- Bi-temporal lite — not full Graphiti dual clocks + temporal KG.

## [1.5.3] — 2026-06

### Added

- **`ListMemoryWithOptions` / `ListMemoryOptions`** (s611 / K2 first slice): event-time ordered timeline listing; filters before Limit; default Limit 50.
- **`EntryHasTag` / `EntryHasTagPrefix`**.

## [1.5.2] — 2026-06

### Added

- **`SearchMemoryWithOptions` / `SearchMemoryOptions`** (s586): hybrid retrieval with session/time filters and optional `ReRankTemporal`.
- Temporal memory kernel roadmap (`docs/temporal-memory-kernel-roadmap.md`).
- GitHub Actions CI for unit tests on PR/push to `main`.

## [1.5.1] / [1.5.0] — 2026

### Changed

- Default ONNX export **BGE-small-en-v1.5** (384-d); no silent Qwen3 flip.

## [1.0.0] — 2026-05-20

### Added

- Stable hierarchical agent memory package (Palace): PalaceStore, compaction, hybrid SearchMemory, pluggable embeddings, optional Qdrant dense + sparse.

[Unreleased]: https://github.com/iome-sh/memory/compare/v1.5.12...HEAD
[1.5.12]: https://github.com/iome-sh/memory/compare/v1.5.11...v1.5.12
[1.5.11]: https://github.com/iome-sh/memory/compare/v1.5.10...v1.5.11
[1.5.10]: https://github.com/iome-sh/memory/compare/v1.5.9...v1.5.10
[1.5.9]: https://github.com/iome-sh/memory/compare/v1.5.8...v1.5.9
[1.5.8]: https://github.com/iome-sh/memory/compare/v1.5.7...v1.5.8
[1.5.7]: https://github.com/iome-sh/memory/compare/v1.5.6...v1.5.7
[1.5.6]: https://github.com/iome-sh/memory/compare/v1.5.5...v1.5.6
[1.5.5]: https://github.com/iome-sh/memory/compare/v1.5.4...v1.5.5
[1.5.4]: https://github.com/iome-sh/memory/compare/v1.5.3...v1.5.4
[1.5.3]: https://github.com/iome-sh/memory/compare/v1.5.2...v1.5.3
[1.5.2]: https://github.com/iome-sh/memory/compare/v1.5.1...v1.5.2
[1.0.0]: https://github.com/iome-sh/memory/releases/tag/v1.0.0
