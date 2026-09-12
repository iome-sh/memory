# memory

[![ci](https://github.com/iome-sh/memory/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/memory/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/iome-sh/memory.svg)](https://pkg.go.dev/github.com/iome-sh/memory)

**Hierarchical agent memory for Go** — a portable library for durable, searchable memory entries with optional vector search and temporal APIs.

This is a **library kernel** (posture: embeddable filesystem palace), not a memory SaaS and not an agent runtime. It is also **not** [MemPalace](https://github.com/MemPalace) / `mempalace` (an unrelated Python project). Module: [`github.com/iome-sh/memory`](https://pkg.go.dev/github.com/iome-sh/memory)

## Features

- **File-backed store** — atomic writes, tiers (working / contextual / semantic / archival), caller-managed best-effort version snapshots (overwrite does not auto-increment)
- **Hybrid search** — keyword + optional dense/sparse vectors (Qdrant) and multi-factor re-ranking
- **Temporal APIs** — session/time filters, event-time timelines, as-of fact listing, supersession helpers
- **Multi-hop retrieval** — lightweight entity-graph expansion with hop-distance ranking
- **Pluggable embeddings** — deterministic hash default for tests; production ONNX via [hugot](https://github.com/knights-analytics/hugot) (pure-Go GoMLX or optional ORT). `PersistEmbeddings` default **off**; hash vectors are never stored
- **Compaction hooks** — kernel primitives for recency/compaction pipelines
- **Benchmarks** — LongMemEval-oriented tooling under `cmd/` and `scripts/`

## Install

```bash
go get github.com/iome-sh/memory@latest
# or pin a release: go get github.com/iome-sh/memory@v1.5.11
```

Requires the Go version in [`go.mod`](go.mod). CI uses `GOTOOLCHAIN=auto`.

## Supported topology

**One process per palace root.** Multi-process writers on a shared `BaseDir` are **unsupported** — that is the product contract, not a defect to hide. In-process `writeMu` serializes the two shared files (`relations/entity-graph.json`, `indexes/event-time.json`). Per-entry JSON uses `CreateTemp` + `chmod 0600` + `Rename` (the rename is the ingest ack). Path isolation is not cloud tenancy. Operators can collect last-write-wins evidence with [`scripts/two_process_writer_probe.sh`](scripts/two_process_writer_probe.sh); flock is not shipped.

## Quick start (TTFH-shaped)

The first worked path is the walking skeleton: ingest **three RCA-shaped turns**, **retrieve in the same process**, **list facts-as-of**, and **print `source_hint`**. Hash embedder · no Qdrant · no cloud palace. This is not a chatbot “favourite colour” demo.

```bash
git clone https://github.com/iome-sh/memory.git
cd memory
go run ./examples/ttfh_rca
```

```go
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
// …two more RCA turns (HMAC 200 ≠ consume receipt; CreateConsumer mode NULL)…

hits := store.SearchMemoryWithOptions("hmac consume receipt", memory.SearchMemoryOptions{
	SessionID: session,
	Limit:     10,
})
for _, h := range hits {
	fmt.Println(h.Content.Summary, h.Provenance.SourceHint) // private
}

facts := store.ListFactsAsOf(memory.FactsAsOfOptions{
	SessionID: session,
	Limit:     10,
})
for _, f := range facts {
	fmt.Println(f.Content.Summary, f.Provenance.SourceHint)
}
```

Full program: [`examples/ttfh_rca`](examples/ttfh_rca). Operator page: [`docs/TTFH.md`](docs/TTFH.md). Host path (optional): [iomesh-tui](https://github.com/iome-sh/iomesh-tui) **v1.3.6** + [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) **v0.4.1** — `/memory ingest` three RCA turns, then `/memory digest --require-sources mesh,private` (cite-both or explicit miss). Cost-max: hash embedder, no Qdrant, no cloud palace, optional Ollama via the TUI. This page, the example, and a unit test document the walking skeleton; they are not a live laptop PULSE + three RCA + cite-both-or-miss run.

If `PalaceConfig.BaseDir` (or `NewPalaceStore`'s argument) is empty, the store uses **`.palace`** under the process working directory (`DefaultPalaceBaseDir`). Prefer an explicit path in applications. This is a local filesystem root — not a leftover `.ossa` product path and not a hosted palace.

This package is a **local filesystem library**, not a cloud multi-tenant service. It does not implement mesh `X-IOMesh-Org` headers. Isolation is the directory you pass as `BaseDir` (or OS isolation around that directory).

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

`PersistEmbeddings` defaults **off**. When on, only a non-hash `EmbeddingModel` (e.g. `bge-small-en-v1.5`) is stored on entry JSON. Hash embeddings (`GenerateSimpleEmbedding`, empty or `"hash"` model) are **never** persisted as stored vectors / `QueryVec`. Embed miss is not ingest failure; JSON rename remains the ack. usearch, ORT, and Qdrant stay optional.

Download helper (fail-soft; BGE is optional):

```bash
go run ./scripts/download_onnx_model.go
export MEMORY_ONNX_MODEL_PATH="$(go run ./scripts/download_onnx_model.go)"
```

Hugging Face **401/404** for `KnightsAnalytics/bge-small-en-v1.5` is **expected** — that repo is not published (valid token still **404**, not a login miss). The helper then downloads public `BAAI/bge-small-en-v1.5` (`onnx/model.onnx`) — **no Hugging Face login required**. `HF_TOKEN` is optional; on 401/403 with a token the helper retries **unauthenticated** so a bad token cannot block public BGE. Layout: `testdata/models/BAAI_bge-small-en-v1.5/` (gitignored; ~127 MB; not vendored). If that fetch fails, in-tree MiniLM is the local 384-d fallback — **not** the official V1 BGE pin. Hash-overlap unpublished. If MiniLM is missing too, stdout is empty and TTFH cost-max stays the **hash embedder**. `os.Exit(1)` only for mkdir failures. Generate/judge need `OPENAI_API_KEY` (unrelated to HF). Auth matrix: [`docs/LONGMEMEVAL_BASELINE.md`](docs/LONGMEMEVAL_BASELINE.md#auth-requirements-checked-2026-09-12). Locked mixed n=12: same page (`make longmemeval-baseline`) — **not a README number**, not official V1.

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

## When to use this kernel

Stars and vendor LongMemEval scores are a category error here.

| Job | Use |
|-----|-----|
| Inspectable local ops record (JSON files, `cat`/`diff`/cite) in Go, no required DB or extract LLM | **This kernel** |
| Chatbot personalization API / drop-in memory SaaS | Mem0 |
| Dual-clock temporal knowledge graph (Neo4j / FalkorDB / Neptune) | Graphiti / Zep |
| Agent runtime that edits its own memory blocks | Letta |
| Documents/tables → company knowledge graph | Cognee |
| Already on LangGraph, want a Python library | LangMem |
| Coding-agent session compressor / verbatim IDE store | claude-mem, **MemPalace** (unrelated Python project — name collision only) |

**Naming:** MemPalace / `mempalace` is a different project. Industry roundups that list “MemPalace” next to Mem0 are not describing this repository.

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
| `IncludeArchival` | When `Tier` is nil, also walk Archival (default retrieve skips it) |
| `QueryVec` | Dense re-rank when non-empty; keyword token hits stay ahead of `Limit` |
| `ReRankTemporal` | Sort by relevance after keyword/vector path; keyword hits stay ahead of `Limit` |

Default retrieve tiers are **Working + Contextual + Semantic** (Archival skipped), matching `ListMemoryWithOptions`. Archival is included when `IncludeArchival` is set, `Tier` is Archival, or the default-tier keyword hit set is empty (low-confidence fallback; not a numeric score cutoff).

### Timeline list

```go
timeline := store.ListMemoryWithOptions(memory.ListMemoryOptions{
	SessionID: "sess-abc",
	TimeFrom:  &from,
	TagPrefix: "subject:",
	Limit:     50,
})
```

Filters apply **before** `Limit`. Default limit is **50** when ≤ 0. Listing uses a best-effort in-memory meta index plus an optional durable snapshot (`indexes/event-time.json`). A clean index is patched on `Write` / unlink instead of walking every tier JSON; a new process skips re-parse when the stamp matches. FS Palace remains source of truth. `DisableMetaIndex` / `DisableDurableIndex` opt out. Btree/tag secondary indexes remain residual.

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

Optional last-write-wins evidence (not a lock): `make two-process-writer-probe` / [`scripts/two_process_writer_probe.sh`](scripts/two_process_writer_probe.sh). Multi-process writers remain **unsupported**. Probe ≠ flock; flock is not shipped. Not part of `make ci` / `make test`.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contributor guide and [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

### LongMemEval tooling (optional)

Methodology card: [`docs/LONGMEMEVAL.md`](docs/LONGMEMEVAL.md). **No official number is published.** Official V1 is upstream `evaluate_qa.py` + judge **`gpt-4o-2024-08-06`** against `longmemeval_oracle.json` (ONNX, mixed-type, `session_id` on retrieve). Makefile default judge `gpt-4o-mini` is a cheap local path — not official V1.

Offline overlap smoke/bench (no OpenAI). Printed `aggregate recall` is **top-k gold-answer string overlap** (judge-free). It is **not** official V1 and **not** V2 LAFS Gain. Hash embeddings are the no-dep default; do not publish hash overlap as a leaderboard number.

```bash
make longmemeval-smoke
make longmemeval-recall-gate
make longmemeval-bench
make longmemeval-v2-bench   # official V2 file layout; does not vendor the 7GB snapshot
make longmemeval-v1-card    # methodology card; SKIP if oracle missing (exit 0); not make ci
```

Official V1 scored QA: `make longmemeval-judge` (needs `OPENAI_API_KEY`). Official judge pin is **`gpt-4o-2024-08-06`**; Makefile default `gpt-4o-mini` is a cheap local path — not official V1. Methodology card (no published score): `make longmemeval-v1-card` — optional, not part of `make ci`; missing oracle is SKIP (exit 0). In-repo subset is 3 `single-session-user` items, not mixed official V1. Official V2 scored runs use the upstream harness with a fixed Qwen3.5-9B reader and GPT-5.2 judge — this kernel only loads V2 files and exposes Insert/Query. Full dataset / judge flows need extra deps and keys; see comments in `Makefile` and `scripts/`.

`--limit N` on `scripts/longmemeval_qa_generate.py` is **dataset prefix order**. Official V1 starts with `temporal-reasoning`, so a small n is not a mixed V1 score. Use `--sample mixed` (or `LONGMEMEVAL_QA_SAMPLE=mixed`) for a stratified slice and print the type histogram. Prefix-n is not overall V1. overlap ≠ gpt-4o ≠ V2 LAFS.

`/retrieve` accepts `session_id` (official generate passes `conv_id` / `question_id`). Shared-palace QA without it is other-session dominated. Hypothesis JSONL keeps `question_date`, retrieve snippets, and `embed_mode` for audit. Hash default. Not a leaderboard submit.

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
| [docs/TTFH.md](docs/TTFH.md) | Operator TTFH walking skeleton |
| [docs/LONGMEMEVAL.md](docs/LONGMEMEVAL.md) | LongMemEval methodology card (no published official number) |
| [docs/OPEN_SOURCE_AUDIT.md](docs/OPEN_SOURCE_AUDIT.md) | Maintainer OSS process residual (not a product spec) |

## Related projects

| Repository | Role |
|------------|------|
| [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) | Lean MCP host binary for this kernel |
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | Multi-provider agent TUI/CLI (optional mesh hooks) |
| [iomesh-client-sdk-go](https://github.com/iome-sh/iomesh-client-sdk-go) | Official Go client for I/O Mesh |
| [iomesh-client-sdk-python](https://github.com/iome-sh/iomesh-client-sdk-python) | Official Python client for I/O Mesh (**Beta** / pre-1.0) |

This module is a **library** (tags for `go get`). Binary packaging, SBOM, and cosign apply to host tools such as `iomesh-memory-mcp` — see [RELEASING.md](RELEASING.md).

## License

[MIT](LICENSE) · [NOTICE](NOTICE)
