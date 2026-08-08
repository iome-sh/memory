# Changelog

All notable changes to this project (Palace memory kernel) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- TUI-grade open-source process bar (s1452): LICENSE (MIT), NOTICE, SECURITY, CONTRIBUTING, CODE_OF_CONDUCT, SUPPORT, RELEASING, OPEN_SOURCE_AUDIT, GitHub PR/issue templates, Dependabot, CI lint/govulncheck/ci-success, `make ci` — **repository remains private** until a deliberate visibility flip.

### Changed

- Go toolchain pin `go 1.26.5`; transitive security bumps (`google.golang.org/grpc`, `golang.org/x/{crypto,text,image}`) so `govulncheck ./...` is clean on called symbols.
- gofmt across package for CI `gofmt -l` gate.

### Honesty

- Kernel-only · not product Memory GA · local-primary · dual_write OFF product path · hosted Palace sunset · future MCP host **iomesh-memory-mcp** · aion broker stays private.

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
