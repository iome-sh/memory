# Changelog

All notable changes to this project (Palace memory kernel) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Temporal kernel roadmap:** [`docs/temporal-memory-kernel-roadmap.md`](docs/temporal-memory-kernel-roadmap.md) re-evaluates K0–K4 / A2–A3 against v1.5.11 and sequences next work as T1 multi-session retrieve → T2 event-time index → T3 temporal edges → T4 compaction/validity → T5 optional 1024-d preset. LongMemEval locked mixed n=12 `multi-session` 0/2 is the T1 trigger.
- **Operator copy polish:** README, SUPPORT, SECURITY, RELEASING, CONTRIBUTING, operator docs, exported comments, and residual-gate needles. Walking skeleton, LongMemEval methodology, single-writer, and inspectable FS facts unchanged.
- **HF auth: public BAAI must not depend on HF_TOKEN:** `HTTPGetToFile` treats `HF_TOKEN` / `HUGGING_FACE_HUB_TOKEN` as optional. BAAI/bge-small-en-v1.5 is **public (no auth)**. If a token is set and Hugging Face returns 401/403, retry once **without** Authorization so a bad/expired token cannot block the public ONNX. 404 is **not** retried (`KnightsAnalytics/bge-small-en-v1.5` is a missing repo, not an auth miss). `OPENAI_API_KEY` is required only for generate/judge.
- **QA reader uses full retrieve-k:** `longmemeval_qa_generate.py` default `LONGMEMEVAL_READER_K=0` feeds **all** retrieved snippets to the reader (was `memories[:15]` while retrieve k=40). Multi-session 0/2 on n=12 was retrieve/reader context, not embedder. Not official V1.
- **BGE ONNX from BAAI when KnightsAnalytics 404:** `KnightsAnalytics/bge-small-en-v1.5` is not a published HF repo (auth 404). The download helper reshapes public `BAAI/bge-small-en-v1.5` `onnx/model.onnx` into hugot layout `testdata/models/BAAI_bge-small-en-v1.5/` (gitignored; not vendored). MiniLM remains the local unpublished fallback. TTFH cost-max is still hash.
- **ONNX download helper fail-soft:** Hugging Face 401/404 for `KnightsAnalytics/bge-small-en-v1.5` no longer `os.Exit(1)`. Helper prints the error on stderr, then MiniLM fallback `testdata/models/KnightsAnalytics_all-MiniLM-L6-v2` when present (not the official V1 BGE pin; hash-overlap unpublished), else empty stdout and TTFH cost-max hash (no ONNX required). Mkdir failures still exit 1. Official V1 pin stays BGE-small-en-v1.5; this repo does not vendor BGE.
- **LongMemEval BGE ONNX retry (SKIP):** `KnightsAnalytics/bge-small-en-v1.5` hugot download still **401**; authenticated Hugging Face lookup **404**. Official V1 embed card not filled. MiniLM unpublished mixed card unchanged. **No README number.**
- **Public-docs hygiene:** `docs/PUBLIC_FLIP_READINESS.md` and `docs/OPEN_SOURCE_AUDIT.md` marked maintainer residuals (flip complete; not operator how-tos). Drop personal maintainer handle from flip checklist. CONTRIBUTING / SUPPORT / README / LONGMEMEVAL / RecMem plan: no “future public MCP”, no support-page ledger stamp, no fundraising exhibit wording, no product token-reduction SLA. Residual gates and `docs/operations/` pins unchanged.

### Added
- **Locked mixed n=60 ID list:** [`testdata/longmemeval_baseline_ids_n60.json`](testdata/longmemeval_baseline_ids_n60.json) (`sample_mixed` 10 of each type; first 12 IDs match the n=12 lock). `make longmemeval-baseline` still defaults to n=12. n=60 is the next scale, not official V1 (not 500). Not a README number.
- **Locked mixed n=12 embed baseline (docs):** [`docs/LONGMEMEVAL_BASELINE.md`](docs/LONGMEMEVAL_BASELINE.md) + `make longmemeval-baseline` / `testdata/longmemeval_baseline_ids.json`. Same 12 IDs × hash vs MiniLM vs BAAI BGE, reader `gpt-4o-mini`, judge `gpt-4o-2024-08-06`, isolated palace per mode. **Not official V1** (n=12, not 500). **Not a README number.** Isolated `LONGMEMEVAL_PALACE_ROOT`.
- **TTFH local-only cite-both miss:** [`docs/TTFH.md`](docs/TTFH.md) Host path shows `/memory digest --require-sources mesh,private` after three private RCA ingest turns printing an explicit miss (no mesh-class receipt). That miss is success for the flag. Do not stamp mesh on local overlay. Catalog/grant/external never satisfy cite-both.
- **Operator TTFH page:** [`docs/TTFH.md`](docs/TTFH.md) is the walking skeleton (three RCA-shaped turns, same-process retrieve, facts-as-of, print `source_hint`). Points at `go run ./examples/ttfh_rca`. Optional host path: iomesh-tui **v1.3.6** + iomesh-memory-mcp **v0.4.1** — `/memory ingest` + `/memory digest --require-sources mesh,private` (cite-both or explicit miss). Cost-max: hash embedder, no Qdrant, no cloud palace.
- **Two-process writer probe:** `scripts/two_process_writer_probe.sh` and `cmd/two-process-writer-probe` run two OS processes against one palace root (distinct entry IDs plus one colliding ID) and print last-write-wins / whether `entity-graph.json` and `event-time.json` remain valid JSON. `TestTwoProcessWriterProbe` is skipped unless `MEMORY_TWO_PROCESS_PROBE=1`. Multi-process writers remain **unsupported**; flock is not shipped. Probe ≠ lock.
- **LongMemEval unpublished MiniLM mixed-run card:** `docs/LONGMEMEVAL.md` Internal run log filled from a mixed-type V1-shaped generate+judge sample. Embed mode is in-tree ONNX MiniLM-L6-v2 (local path when BGE-small-en-v1.5 is unavailable; HF download may 401) — **not** the official V1 BGE pin. Official judge pin `gpt-4o-2024-08-06` (`scripts/longmemeval_judge.sh` maps that pin to upstream `evaluate_qa.py` zoo key `gpt-4o`). Makefile/script comments record the MiniLM fallback. **Unpublished · not a README number.**
- **LongMemEval V1 mixed-run methodology card runner:** `make longmemeval-v1-card` / `scripts/longmemeval_v1_card.sh` prints dataset, SHA, tag, mixed-sample histogram, embed mode, official judge pin (`gpt-4o-2024-08-06`), and date. Missing `data/longmemeval_oracle.json` is SKIP (exit 0), not a CI failure. In-repo subset is 3 `single-session-user` items — not mixed official V1. Makefile default `gpt-4o-mini` stays a cheap local path — not official V1. No published score.

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

[Unreleased]: https://github.com/iome-sh/memory/compare/v1.5.11...HEAD
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
