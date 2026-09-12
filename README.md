# memory

[![ci](https://github.com/iome-sh/memory/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/memory/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/iome-sh/memory)](https://github.com/iome-sh/memory/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/iome-sh/memory.svg)](https://pkg.go.dev/github.com/iome-sh/memory)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Hierarchical agent memory for Go** — an embeddable, file-backed palace you can `cat`, `diff`, and search. Hybrid keyword + optional vectors, session/time filters, and temporal helpers. No required database.

Module: [`github.com/iome-sh/memory`](https://pkg.go.dev/github.com/iome-sh/memory) · latest tag **[v1.5.12](https://github.com/iome-sh/memory/releases/tag/v1.5.12)**

This is a **library**, not a memory SaaS and not an agent runtime. It is also **not** [MemPalace](https://github.com/MemPalace) / `mempalace` (an unrelated Python project).

## Table of contents

- [Install](#install)
- [Quick start](#quick-start)
- [Features](#features)
- [When to use this kernel](#when-to-use-this-kernel)
- [Topology](#topology)
- [Optional embeddings](#optional-embeddings)
- [Optional Qdrant](#optional-qdrant)
- [API overview](#api-overview)
- [Development](#development)
- [Documentation](#documentation)
- [Related projects](#related-projects)
- [License](#license)

## Install

```bash
go get github.com/iome-sh/memory@v1.5.12
# or follow the latest tagged release:
# go get github.com/iome-sh/memory@latest
```

Requires the Go version in [`go.mod`](go.mod) (currently **1.27**). CI uses `GOTOOLCHAIN=auto`.

Optional hosts that already pin this module:

```bash
go install github.com/iome-sh/iomesh-memory-mcp/cmd/iomesh-memory-mcp@v0.4.2
go install github.com/iome-sh/iomesh-tui/cmd/iomesh@v1.3.7
```

## Quick start

Ingest turns, search, and list facts-as-of in one process. Default embedder is a deterministic hash (no ONNX, no Qdrant).

```bash
git clone https://github.com/iome-sh/memory.git
cd memory
go run ./examples/ttfh_rca
```

```go
package main

import (
	"fmt"

	"github.com/iome-sh/memory"
)

func main() {
	store := memory.NewPalaceStore("./data/ttfh-palace")
	session := "inc-webhook-5xx"

	_ = store.IngestTurn(memory.MemoryEntry{
		SessionID: session,
		Content: memory.MemoryContent{
			Summary: "PagerDuty page: webhook ingress 5xx",
			Full:    "On-call: webhook ingress returned 5xx. Start RCA from the signed delivery, not the dashboard chrome.",
			Tags:    []string{"pagerduty"},
		},
		ExtractedFacts: []string{"PagerDuty page fired for webhook ingress 5xx"},
	})

	hits := store.SearchMemoryWithOptions("hmac consume receipt", memory.SearchMemoryOptions{
		SessionID: session,
		Limit:     10,
	})
	for _, h := range hits {
		fmt.Println(h.Content.Summary, h.Provenance.SourceHint)
	}

	facts := store.ListFactsAsOf(memory.FactsAsOfOptions{
		SessionID: session,
		Limit:     10,
	})
	for _, f := range facts {
		fmt.Println(f.Content.Summary, f.Provenance.SourceHint)
	}
}
```

Worked example: [`examples/ttfh_rca`](examples/ttfh_rca) (three RCA-shaped turns, same-process retrieve, facts-as-of, print `source_hint`). Operator notes: [`docs/TTFH.md`](docs/TTFH.md).

If `PalaceConfig.BaseDir` (or `NewPalaceStore`'s argument) is empty, the store uses **`.palace`** under the process working directory (`DefaultPalaceBaseDir`). Prefer an explicit path in applications.

## Features

- **File-backed store** — atomic JSON writes (`CreateTemp` + `chmod 0600` + `Rename`); tiers working / contextual / semantic / archival
- **Hybrid search** — keyword first, optional dense re-rank; count and temporal-order queries skip vector scoring
- **Temporal APIs** — `SessionID` / `SessionIDs`, `conv:` grouping, event-time timelines, as-of facts, supersession, dated-event and latest-value evidence helpers
- **Multi-hop retrieval** — lightweight entity-graph expansion with hop-distance ranking
- **Pluggable embeddings** — hash default for tests; production ONNX via [hugot](https://github.com/knights-analytics/hugot) (pure-Go GoMLX or optional ORT). `PersistEmbeddings` default **off**; hash vectors are never stored
- **Compaction hooks** — kernel primitives for recency/compaction pipelines
- **Eval harness** — LongMemEval-oriented tooling under `cmd/` and `scripts/` (optional; no published leaderboard number)

## When to use this kernel

| Job | Use |
|-----|-----|
| Inspectable local ops record (JSON files, `cat`/`diff`/cite) in Go, no required DB | **This kernel** |
| Chatbot personalization API / drop-in memory SaaS | Mem0 |
| Dual-clock temporal knowledge graph (Neo4j / FalkorDB / Neptune) | Graphiti / Zep |
| Agent runtime that edits its own memory blocks | Letta |
| Documents/tables → company knowledge graph | Cognee |
| Already on LangGraph, want a Python library | LangMem |
| Coding-agent session compressor / verbatim IDE store | claude-mem, **MemPalace** (unrelated Python project — name collision only) |

**Naming:** MemPalace / `mempalace` is a different project. Roundups that list “MemPalace” next to Mem0 are not describing this repository.

## Topology

**One process per palace root.** Multi-process writers on a shared `BaseDir` are unsupported. In-process `writeMu` serializes `relations/entity-graph.json` and `indexes/event-time.json`. Isolation is the directory you pass as `BaseDir` (this library does not implement mesh `X-IOMesh-Org`).

Last-write-wins evidence (not a lock): [`scripts/two_process_writer_probe.sh`](scripts/two_process_writer_probe.sh). Flock is not shipped.

## Optional embeddings

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

Default ONNX export is **BGE-small-en-v1.5** (**384-d**). `PersistEmbeddings` defaults **off**. Hash embeddings are **never** stored. usearch, ORT, and Qdrant stay optional.

```bash
go run ./scripts/download_onnx_model.go
export MEMORY_ONNX_MODEL_PATH="$(go run ./scripts/download_onnx_model.go)"
```

`KnightsAnalytics/bge-small-en-v1.5` is not a published Hugging Face repo (404). The helper downloads public `BAAI/bge-small-en-v1.5` (`onnx/model.onnx`) with no login required. Layout: `testdata/models/BAAI_bge-small-en-v1.5/` (gitignored). MiniLM is the in-tree 384-d fallback. Details: [`docs/LONGMEMEVAL_BASELINE.md`](docs/LONGMEMEVAL_BASELINE.md).

## Optional Qdrant

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
	EmbeddingFunc:    embedFn,
})
```

Unit tests run without Podman/Qdrant. Set `PODMAN_QDRANT_SKIP=1` to skip integration helpers.

## API overview

| Area | Entry points |
|------|----------------|
| Store | `NewPalaceStore`, `NewPalaceStoreWithConfig`, `Write`, `IngestTurn`, `Load` |
| Search | `SearchMemory`, `SearchMemoryWithOptions` |
| Timeline | `ListMemoryWithOptions` |
| As-of facts | `ListFactsAsOf`, `ParseValidityWindow`, `EntryValidAt` |
| Evidence helpers | `AssembleCountEvidence`, `AssembleTemporalEvidence`, `AssembleLatestValueEvidence` |
| Supersession | `SupersedeEntityFacts`, `WriteAndSupersede` |
| Multi-hop | `MultiHopRetrieve`, `ExpandRelatedEntities`, `ExpandRelatedEntitiesHops` |
| Vectors | `NewVectorStore`, collection create/upsert helpers |
| Embeddings | `GenerateSimpleEmbedding`, `NewGONNXEmbeddingFunc`, `NewGONNXEmbeddingFuncFromEnv` |

### Search options

```go
from := time.Now().Add(-24 * time.Hour)
results := store.SearchMemoryWithOptions("project goals", memory.SearchMemoryOptions{
	SessionID:      "sess-abc",
	SessionIDs:     []string{"sess-abc", "sess-def"}, // any-of; also matches conv:<id> tags
	TimeFrom:       &from,
	Limit:          10,
	ReRankTemporal: true,
})
```

| Field | Effect |
|-------|--------|
| `SessionID` | Keep entries with matching session (or `conv:<id>` tag) |
| `SessionIDs` | Any-of session / conv-tag match |
| `TimeFrom` / `TimeTo` | Inclusive event-time window |
| `Limit` | Cap results (default 10 for search) |
| `Tier` | Optional tier filter |
| `IncludeArchival` | When `Tier` is nil, also walk Archival (default retrieve skips it) |
| `QueryVec` | Dense re-rank when non-empty; skipped for count and temporal-order queries |
| `ReRankTemporal` | Sort by relevance after keyword/vector path; keyword hits stay ahead of `Limit` |

Default retrieve tiers: **Working + Contextual + Semantic**. Archival is included when `IncludeArchival` is set, `Tier` is Archival, or default-tier keyword hits are empty.

### Timeline list

```go
timeline := store.ListMemoryWithOptions(memory.ListMemoryOptions{
	SessionID: "sess-abc",
	TimeFrom:  &from,
	TagPrefix: "subject:",
	Limit:     50,
})
```

Filters apply **before** `Limit` (default 50). Listing uses a best-effort in-memory meta index plus optional `indexes/event-time.json`. The filesystem palace remains the source of truth.

Full reference: [pkg.go.dev/github.com/iome-sh/memory](https://pkg.go.dev/github.com/iome-sh/memory).

## Development

```bash
git clone https://github.com/iome-sh/memory.git
cd memory
go mod download
make check    # fmt-check + vet + test
make ci       # + govulncheck + build
make test
make test-race
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

### LongMemEval (optional)

Methodology: [`docs/LONGMEMEVAL.md`](docs/LONGMEMEVAL.md). **No official number is published** in this README. Official V1 is upstream `evaluate_qa.py` + judge `gpt-4o-2024-08-06`. Makefile default `gpt-4o-mini` is a cheap local path.

```bash
make longmemeval-smoke
make longmemeval-v1-card    # SKIP (exit 0) if the oracle file is missing; not part of make ci
```

Locked mixed slices and auth notes: [`docs/LONGMEMEVAL_BASELINE.md`](docs/LONGMEMEVAL_BASELINE.md).

## Documentation

| Document | Description |
|----------|-------------|
| [CHANGELOG.md](CHANGELOG.md) | Release notes |
| [RELEASING.md](RELEASING.md) | How maintainers tag module versions |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting and supported versions |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Development workflow |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |
| [SUPPORT.md](SUPPORT.md) | How to get help |
| [docs/temporal-memory-kernel-roadmap.md](docs/temporal-memory-kernel-roadmap.md) | Temporal API roadmap |
| [docs/TTFH.md](docs/TTFH.md) | Walking-skeleton operator notes |
| [docs/LONGMEMEVAL.md](docs/LONGMEMEVAL.md) | LongMemEval methodology (no published official number) |

## Related projects

| Repository | Role |
|------------|------|
| [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) | MCP host binary for this kernel (**v0.4.2**) |
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | Multi-provider agent TUI/CLI (**v1.3.7**) |
| [iomesh-client-sdk-go](https://github.com/iome-sh/iomesh-client-sdk-go) | Official Go client for I/O Mesh |
| [iomesh-client-sdk-python](https://github.com/iome-sh/iomesh-client-sdk-python) | Official Python client for I/O Mesh (Beta / pre-1.0) |

This module is a **library** (`go get` tags). Binary packaging, SBOM, and cosign apply to host tools such as `iomesh-memory-mcp` — see [RELEASING.md](RELEASING.md).

## License

[MIT](LICENSE) · [NOTICE](NOTICE)
