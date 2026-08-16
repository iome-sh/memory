# Changelog

All notable changes to this project (Palace memory kernel) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Go toolchain pin `go 1.26.6` so CI `govulncheck` is clean on stdlib GO-2026-5972 / GO-2026-5026 (fixed in go1.26.6).

### Added
- **K2 durable event-time snapshot (#44):** `indexes/event-time.json` best-effort on-disk meta index. Fresh processes skip re-parsing every tier JSON when the stamp (JSON count + max mtime) matches. FS Palace remains source of truth; dirty writes still rebuild from disk. `DisableDurableIndex` opt-out. Incremental/btree/tag secondary indexes remain residual. Kernel-only · not Memory GA · residual ≠ invent index green.

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

- Kernel-only · not product Memory GA · local-primary · dual_write OFF product path · hosted Palace sunset · future MCP host **iomesh-memory-mcp** · aion broker stays private · residual PASS ≠ public flip · M4 readiness ≠ invent public · s1473 final audit ≠ public flip.

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

[Unreleased]: https://github.com/iome-sh/memory/compare/v1.5.7...HEAD
[1.5.7]: https://github.com/iome-sh/memory/compare/v1.5.6...v1.5.7
[1.5.6]: https://github.com/iome-sh/memory/compare/v1.5.5...v1.5.6
[1.5.5]: https://github.com/iome-sh/memory/compare/v1.5.4...v1.5.5
[1.5.4]: https://github.com/iome-sh/memory/compare/v1.5.3...v1.5.4
[1.5.3]: https://github.com/iome-sh/memory/compare/v1.5.2...v1.5.3
[1.5.2]: https://github.com/iome-sh/memory/compare/v1.5.1...v1.5.2
[1.0.0]: https://github.com/iome-sh/memory/releases/tag/v1.0.0
