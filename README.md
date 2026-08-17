# memory

[![ci](https://github.com/iome-sh/memory/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/memory/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/iome-sh/memory.svg)](https://pkg.go.dev/github.com/iome-sh/memory)

**Hierarchical agent memory for Go** — a portable library for durable, searchable memory entries with optional vector search and temporal APIs.

Module: [`github.com/iome-sh/memory`](https://pkg.go.dev/github.com/iome-sh/memory)

## Features

- **File-backed store** — atomic writes, tiers (working / contextual / semantic / archival), versioning
- **Hybrid search** — keyword + optional dense/sparse vectors (Qdrant) and multi-factor re-ranking
- **Temporal APIs** — session/time filters, event-time timelines, as-of fact listing, supersession helpers
- **Multi-hop retrieval** — lightweight entity-graph expansion with hop-distance ranking
- **Pluggable embeddings** — deterministic hash default for tests; production ONNX via [hugot](https://github.com/knights-analytics/hugot) (pure-Go GoMLX or optional ORT)
- **Compaction hooks** — kernel primitives for recency/compaction pipelines
- **Benchmarks** — LongMemEval-oriented tooling under `cmd/` and `scripts/`

## Install

```bash
go get github.com/iome-sh/memory@latest
# or pin a release: go get github.com/iome-sh/memory@v1.5.7
```

Requires the Go version in [`go.mod`](go.mod). CI uses `GOTOOLCHAIN=auto`.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/iome-sh/memory"
)

func main() {
	store := memory.NewPalaceStore("./data/palace")

	id := memory.GenerateMemoryID()
	_ = store.Write(memory.MemoryEntry{
		ID:   id,
		Tier: memory.TierContextual,
		Content: memory.MemoryContent{
			Summary: "Project alpha ships on Friday",
		},
	})

	hits := store.SearchMemory("project alpha", nil, 5, nil)
	for _, h := range hits {
		fmt.Println(h.Content.Summary)
	}
}
```

### Optional semantic embeddings

```go
embedFn, err := memory.NewGONNXEmbeddingFuncFromEnv()
if err != nil {
	panic(err)
}
store := memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
	BaseDir:       "./data/palace",
	EmbeddingFunc: embedFn,
})
```

| Variable | Purpose |
|----------|---------|
| `MEMORY_ONNX_MODEL_PATH` | Hugot model directory or `.onnx` file |
| `MEMORY_HUGOT_BACKEND` | `go` (default pure-Go), `ort`, or `auto` |
| `MEMORY_ORT_LIBRARY_DIR` | Directory containing ONNX Runtime shared library |
| `MEMORY_ORT_CUDA` | `1` to enable CUDA EP (Linux ORT builds) |
| `MEMORY_EMBEDDING_STRICT` | `true` to disable hash fallback on inference errors |

Default ONNX export is **BGE-small-en-v1.5** (**384** dimensions). When using Qdrant with that model, set collection `EmbeddingDim` to **384**.

Download helper:

```bash
go run ./scripts/download_onnx_model.go
```

### Optional Qdrant

```bash
podman run -d --name qdrant \
  -p 6333:6333 -p 6334:6334 \
  -v qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant
```

```go
store := memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
	BaseDir:          "./data/palace",
	VectorURL:        "http://localhost:6333",
	VectorCollection: "memory_collection",
	EmbeddingFunc:    embedFn, // recommended for semantic recall
})
```

Unit tests run without Podman/Qdrant. Integration helpers start a temporary container when available; set `PODMAN_QDRANT_SKIP=1` to force skip.

## API overview

| Area | Entry points |
|------|----------------|
| Store | `NewPalaceStore`, `NewPalaceStoreWithConfig`, `Write`, `Read`, … |
| Search | `SearchMemory`, `SearchMemoryWithOptions` |
| Timeline | `ListMemoryWithOptions` |
| As-of facts | `ListFactsAsOf`, `ParseValidityWindow`, `EntryValidAt` |
| Supersession | `SupersedeEntityFacts`, `WriteAndSupersede` |
| Multi-hop | `MultiHopRetrieve`, `ExpandRelatedEntities`, `ExpandRelatedEntitiesHops` |
| Vectors | `NewVectorStore`, collection create/upsert helpers |
| Embeddings | `GenerateSimpleEmbedding`, `NewGONNXEmbeddingFunc`, `NewGONNXEmbeddingFuncFromEnv` |

### Search options

```go
from := time.Now().Add(-24 * time.Hour)
results := store.SearchMemoryWithOptions("project goals", memory.SearchMemoryOptions{
	SessionID:      "sess-abc",
	TimeFrom:       &from,
	Limit:          10,
	ReRankTemporal: true,
})
```

| Field | Effect |
|-------|--------|
| `SessionID` | Keep entries with matching session |
| `TimeFrom` / `TimeTo` | Inclusive event-time window |
| `Limit` | Cap results (default 10 for search) |
| `Tier` | Optional tier filter |
| `QueryVec` | Dense re-rank when non-empty; keyword token hits stay ahead of `Limit` |
| `ReRankTemporal` | Sort by relevance after keyword/vector path; keyword hits stay ahead of `Limit` |

### Timeline list

```go
timeline := store.ListMemoryWithOptions(memory.ListMemoryOptions{
	SessionID: "sess-abc",
	TimeFrom:  &from,
	TagPrefix: "subject:",
	Limit:     50,
})
```

Filters apply **before** `Limit`. Default limit is **50** when ≤ 0. Listing uses a best-effort in-memory meta index plus an optional durable snapshot (`indexes/event-time.json`) so a new process does not re-parse every tier JSON when the stamp matches. FS Palace remains source of truth. `DisableMetaIndex` / `DisableDurableIndex` opt out. Incremental btree/tag indexes remain residual.

Full reference: [pkg.go.dev/github.com/iome-sh/memory](https://pkg.go.dev/github.com/iome-sh/memory).

## Development

```bash
git clone https://github.com/iome-sh/memory.git
cd memory
go mod download

make check   # fmt-check + vet + test
make ci      # + govulncheck + build
make test
make test-race   # optional
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contributor guide and [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

### LongMemEval tooling (optional)

Offline overlap smoke/bench (no OpenAI). Printed `aggregate recall` is **top-k gold-answer string overlap** (judge-free). It is **not** official V1 `evaluate_qa.py` + gpt-4o accuracy and **not** V2 LAFS Gain. Hash embeddings are the no-dep default; do not publish hash overlap as a leaderboard number. dual_write stays OFF. Not Memory GA.

```bash
make longmemeval-smoke
make longmemeval-recall-gate
make longmemeval-bench
make longmemeval-v2-bench   # official V2 file layout; does not vendor the 7GB snapshot
```

Official V1 scored QA: `make longmemeval-judge` (needs `OPENAI_API_KEY`). Official V2 scored runs use the upstream harness with a fixed Qwen3.5-9B reader and GPT-5.2 judge — this kernel only loads V2 files and exposes Insert/Query. Full dataset / judge flows need extra deps and keys; see comments in `Makefile` and `scripts/`.

`--limit N` on `scripts/longmemeval_qa_generate.py` is **dataset prefix order**. Official V1 starts with `temporal-reasoning`, so a small n is not a mixed V1 score. Use `--sample mixed` (or `LONGMEMEVAL_QA_SAMPLE=mixed`) for a stratified slice and print the type histogram. Prefix-n is not overall V1. overlap ≠ gpt-4o ≠ V2 LAFS.

`/retrieve` accepts `session_id` (official generate passes `conv_id` / `question_id`). Shared-palace QA without it is other-session dominated. Hypothesis JSONL keeps `question_date`, retrieve snippets, and `embed_mode` for audit. Hash default. Not a leaderboard submit. Not Memory GA.

Haystack dates accept official cleaned `2006/01/02 (Mon) 15:04` as well as RFC3339.

## Documentation

| Document | Description |
|----------|-------------|
| [CHANGELOG.md](CHANGELOG.md) | Release notes |
| [RELEASING.md](RELEASING.md) | How maintainers tag module versions; **support / version policy** for consumers |
| [SECURITY.md](SECURITY.md) | Security policy and supported-versions table |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Development workflow |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |
| [SUPPORT.md](SUPPORT.md) | How to get help; scope (library kernel) and related host |
| [docs/temporal-memory-kernel-roadmap.md](docs/temporal-memory-kernel-roadmap.md) | Temporal API roadmap (K0–K4 style) |
| [docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md) | OSS process checklist |

## Related projects

| Repository | Role |
|------------|------|
| [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) | Lean MCP host binary for this kernel |
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | Multi-provider agent TUI/CLI (optional mesh hooks) |
| [iomesh-client-sdk-go](https://github.com/iome-sh/iomesh-client-sdk-go) | Official Go client for I/O Mesh |
| [iomesh-client-sdk-python](https://github.com/iome-sh/iomesh-client-sdk-python) | Official Python client for I/O Mesh (**Beta** / pre-1.0 — not invent 1.0 / live PyPI GA) |

This module is a **library** (tags for `go get`). Binary packaging, SBOM, and cosign apply to host tools such as `iomesh-memory-mcp` — see [RELEASING.md](RELEASING.md).

## License

[MIT](LICENSE) · [NOTICE](NOTICE)
