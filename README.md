# memory

[![ci](https://github.com/iome-sh/memory/actions/workflows/ci.yml/badge.svg)](https://github.com/iome-sh/memory/actions/workflows/ci.yml)

## v1.5.7 continuum (in progress)

A2 residual partial advance on the multi-hop path (not a full release tag yet):

- **Hop-distance ranking** (s1067 / A2 residual): `MultiHopRetrieve` prefers **shorter BFS hop distance** from seed, then event time descending within the same hop. Seed matches remain hop 0. Implemented via `ExpandRelatedEntitiesHops` (entity → min hop). Optional `PreferShorterHops *bool` defaults **true** (set false to restore legacy seed-match-first sort).
- **Honesty**: still multi-hop lite — not full Zep/Graphiti path scoring, typed edges, or NLP; not product Memory GA.

## v1.5.6 Release

Module: `github.com/iome-sh/memory@v1.5.6` (install: `go get github.com/iome-sh/memory@v1.5.6`).

Includes commits on `main` since `v1.5.5`:

- **`SupersedeEntityFacts`** (s632 / A3 first slice): close prior open validity windows for an entity key by writing `valid_until:<RFC3339>` (exclusive end, same as `EntryValidAt`). Matches via `EntryEntityKeys` (lower-case trim). Empty key is a no-op. Does **not** delete entries.
- **`WriteAndSupersede`**: write entry first (stamps `valid_from=now` when unset), then supersede each key **excluding** the new entry ID.
- **Roadmap**: A3 / K4 Partial shipped (competitive lite supersession — **not** automatic NLP contradiction detection, **not** full Zep dual-clock KG).

Kernel-only; not product Memory GA. BGE-small-en-v1.5 (384-d) default unchanged (no silent Qwen3 flip).

## v1.5.5 Release

Module: `github.com/iome-sh/memory@v1.5.5` (install: `go get github.com/iome-sh/memory@v1.5.5`).

Includes commits on `main` since `v1.5.4`:

- **`MultiHopRetrieve` / `MultiHopOptions`** (s619 / A2 first slice; hop ranking s1067): multi-hop lite associative retrieval over EntityGraph BFS (`ExpandRelatedEntities` / `ExpandRelatedEntitiesHops`) + entry collect via `TemporalTags` `entity:*`, `Content.Tags`, and `Relations.RelatedConcepts`. Seeds from `SeedEntity` and/or `SeedQuery` search hits. `MaxHops` default **2** (clamp 1..4); `Limit` default **20** after expansion + filters (underfill class). Optional `SessionID` / `AsOf` / tier opts. Default sort: **min hop ascending**, then event time desc (`PreferShorterHops` default true).
- **`ExpandRelatedEntities`**: BFS from seed over `GetRelatedEntities` (includes seed at hop 0).
- **`ExpandRelatedEntitiesHops`**: same BFS; returns entity → minimum hop distance (seed = 0).
- **`EntryEntityKeys`**: parse entity keys from `entity:` / `subject:` tags and RelatedConcepts.
- **`AddEntityRelationship`**: ensures `relations/` dir exists before graph write.
- **Roadmap**: A2 Partial shipped (multi-hop lite — not full Zep / Graphiti KG); hop-distance ranking residual advanced s1067.

Kernel-only; not product Memory GA. BGE-small-en-v1.5 (384-d) default unchanged (no silent Qwen3 flip).

## v1.5.4 Release

Module: `github.com/iome-sh/memory@v1.5.4` (install: `go get github.com/iome-sh/memory@v1.5.4`).

Includes commits on `main` since `v1.5.3`:

- **`ListFactsAsOf` / `FactsAsOfOptions`** (s616 / K4 first slice): as-of validity listing over host-written `valid_from:<RFC3339>` / `valid_until:<RFC3339>` TemporalTags. Filters (session, entity, query, `EntryValidAt`) apply **before** Limit. Semantic-first ordering; default Limit **50**.
- **`ParseValidityWindow` / `EntryValidAt`**: validity window parse + rules (`valid_until` exclusive end; untagged entries fall back to “known by asOf” via event time).
- **`SearchMemoryOptions.AsOf`**: optional as-of filter on hybrid search (after time-window, before Limit).
- **Roadmap**: K4 Partial shipped (bi-temporal lite — not full Graphiti dual clocks + temporal KG).

Kernel-only; not product Memory GA. BGE-small-en-v1.5 (384-d) default unchanged (no silent Qwen3 flip).

## v1.5.3 Release

Module: `github.com/iome-sh/memory@v1.5.3` (install: `go get github.com/iome-sh/memory@v1.5.3`).

Includes commits on `main` since `v1.5.2`:

- **`ListMemoryWithOptions` / `ListMemoryOptions`** (s611 / K2 first slice): event-time ordered timeline listing with optional `SessionID`, `TimeFrom` / `TimeTo`, `Tag` / `TagPrefix`, substring `Query`, tier defaults (Working+Contextual+Semantic; Archival opt-in), and `Ascending`. Filters apply **before** Limit (underfill class, same as K1 search). Default `Limit` is **50**. FS scan remains O(n); full event-time index is residual.
- **`EntryHasTag` / `EntryHasTagPrefix`**: light helpers over `TemporalTags` and `Content.Tags`.
- **Roadmap**: K1 marked Shipped; K2 Partial shipped in `docs/temporal-memory-kernel-roadmap.md`.

Kernel-only; not product Memory GA. BGE-small-en-v1.5 (384-d) default unchanged (no silent Qwen3 flip).

## v1.5.2 Release

Module: `github.com/iome-sh/memory@v1.5.2` (install: `go get github.com/iome-sh/memory@v1.5.2`).

Includes commits on `main` since `v1.5.1`:

- **`SearchMemoryWithOptions` / `SearchMemoryOptions`** (s586): hybrid retrieval with optional `SessionID`, `TimeFrom` / `TimeTo` inclusive event-time filters, and `ReRankTemporal` (sort by `CalculateRelevanceScore` after keyword/vector path). `SearchMemory` remains a thin behavior-preserving wrapper.
- **Temporal memory kernel docs** (s587): `docs/temporal-memory-kernel-roadmap.md` (K0–K4 fields, decay, multi-factor score, ingest/search phases).
- **GitHub Actions CI** (s588): unit tests on PR and push to `main` (`.github/workflows/ci.yml`).

Also inherits v1.5.0/v1.5.1 defaults (BGE-small-en-v1.5 ONNX, 384-d). See [Usage](#usage) for API examples.

## v1.0.0 Release (May 20, 2026)

**Stable release** of the hierarchical agent memory package (Palace).

- Production vector search (dense + sparse via Qdrant Query API)
- Full PalaceStore with atomic writes, versioning, temporal decay, multi-factor scoring
- Compaction with temporal windows and alpha constraints
- Hybrid SearchMemory + **pluggable pure-Go ONNX embeddings** (new in this refactor)
- All core features complete, tested, and hardened

**Module**: `github.com/iome-sh/memory@v1.0.0`

Standalone, portable Go package for hierarchical agent memory (Palace).

Extracted and refactored from `github.com/sudo-jin/ossa/internal/self`.

## Goals

- Portable (remove darwin-only where possible)
- Importable as a clean package
- Core types: MemoryEntry, MemoryTier, compaction, **real semantic embeddings via pure-Go ONNX**, relations, file-backed Palace
- Optional Qdrant vector integration
- Incorporate H-Mem inspired improvements (temporal windows, multi-factor scoring, entity-aware graph)
- Clean API for other agents/repos to import

## Current Status

**v1.0.0 released and stable.** Core memory models, PalaceStore, compaction, versioning, temporal decay, multi-factor scoring, pluggable embeddings (**production pure-Go ONNX via hugot**), Working tier lifecycle, PalaceConfig, hybrid SearchMemory, and full Qdrant vector integration (dense + sparse) are complete, tested, and production-ready.

## Embedding Options (New)

The `EmbeddingFunc` in `PalaceConfig` is fully pluggable.

### Default (fast, deterministic, zero deps)
`GenerateSimpleEmbedding` — hash-seeded random unit vectors (good for structure/tests, zero semantic value).

### Hugot ONNX embeddings (GoMLX default; ORT optional for speed)

**Default (dev/CI):** pure-Go GoMLX — `MEMORY_HUGOT_BACKEND=go` (or unset). No CGO, no dylibs.

**Linux prod (NVIDIA):** ONNX Runtime + CUDA — build with `-tags ORT`, set:

```bash
export MEMORY_HUGOT_BACKEND=ort
export MEMORY_ORT_CUDA=1
export MEMORY_ORT_LIBRARY_DIR=/usr/lib   # libonnxruntime.so directory
```

**macOS dev:** ORT + CoreML — one-time native deps, then build with `-tags ORT`:

```bash
make download-ort-deps          # testdata/ort-deps/lib/{libtokenizers.a,libonnxruntime.dylib}
eval "$(./scripts/ort_cgo_env.sh)"
export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5
export MEMORY_HUGOT_BACKEND=auto   # CoreML on darwin; GoMLX fallback if ORT init fails
go build -tags ORT -o bin/longmemeval-bench-ort ./cmd/longmemeval-bench
# or: make build-ort-bench
```

**Auto:** `MEMORY_HUGOT_BACKEND=auto` tries ORT (CoreML on Mac, CUDA when `MEMORY_ORT_CUDA=1` on Linux), then falls back to GoMLX.

### Semantic recall (all ONNX backends)
```go
import "github.com/iome-sh/memory"

// Download once (or set MEMORY_ONNX_MODEL_PATH to an existing hugot model dir)
embedFn, err := memory.NewGONNXEmbeddingFuncFromEnv()
if err != nil { log.Fatal(err) }

cfg := memory.PalaceConfig{
	BaseDir:       ".ossa/kb/palace",
	EmbeddingFunc: embedFn,
}
store := memory.NewPalaceStoreWithConfig(cfg)
```

`NewGONNXEmbeddingFunc` uses [`hugot`](https://github.com/knights-analytics/hugot) (pure Go backend, no ORT dylibs). Pass a **hugot model directory** (tokenizer + `model.onnx`) or a direct `.onnx` file path. Default export is `KnightsAnalytics/bge-small-en-v1.5` (**384** dimensions).

Environment:

| Variable | Purpose |
|----------|---------|
| `MEMORY_ONNX_MODEL_PATH` | Hugot model directory or `.onnx` file |
| `MEMORY_HUGOT_BACKEND` | `go` (default), `ort`, or `auto` |
| `MEMORY_ORT_LIBRARY_DIR` | Directory with `libonnxruntime.so` / `.dylib` (ORT builds) |
| `MEMORY_ORT_CUDA` | `1` to enable CUDA EP (Linux ORT+GPU) |
| `MEMORY_ORT_COREML` | `1` to enable CoreML EP (macOS ORT) |
| `MEMORY_ORT_CUDA_DEVICE_ID` | CUDA device index (default `0`) |
| `MEMORY_EMBEDDING_STRICT=true` | Disable silent hash fallback on inference errors |

**Qdrant / sidecar:** set collection `EmbeddingDim` to **384** when using the default BGE-small ONNX export (not 768 hash default).

Download helper:

```bash
go run ./scripts/download_onnx_model.go
```

## Planned Structure

```
memory/
├── memory.go          # Core types, Palace storage, embeddings (now with NewGONNXEmbeddingFunc), relations, temporal tags, PalaceConfig
├── compaction.go      # LLM-driven compaction (agent-managed) with VectorStoreCallback
├── vector.go          # Official Qdrant Go client (dense + sparse vectors, batch upsert, worker-pool batch creation, retries, context, partial results)
├── memory_test.go     # Unit tests for PalaceStore, embeddings, compaction, vector
├── vector_test.go     # VectorStore tests + temporary Qdrant via Podman helper
├── README.md
├── go.mod
└── docs/
    └── memory-refactor-improvements.md
```

## Usage

```go
import (
	"context"
	"time"

	"github.com/iome-sh/memory"
)

// Basic PalaceStore
store := memory.NewPalaceStore(".ossa/kb/palace")
entry := memory.MemoryEntry{
	ID:      memory.GenerateMemoryID(),
	Tier:    memory.TierContextual,
	Content:  memory.MemoryContent{Summary: "Important fact"},
	// ...
}
store.Write(entry)

// With Qdrant
cfg := memory.PalaceConfig{
	BaseDir:          ".ossa/kb/palace",
	VectorURL:        "http://localhost:6333",
	VectorCollection: "memory_collection",
}
store = memory.NewPalaceStoreWithConfig(cfg)

// Hybrid search (keyword + optional vector re-rank)
results := store.SearchMemory("project goals", nil, 10, nil)

// Temporal / session filters + optional relevance re-rank (s586)
from := time.Now().Add(-24 * time.Hour)
results = store.SearchMemoryWithOptions("project goals", memory.SearchMemoryOptions{
	SessionID:      "sess-abc",
	TimeFrom:       &from,
	Limit:          10,
	ReRankTemporal: true, // sort by CalculateRelevanceScore after keyword/vector path
})

// Event-time timeline list (s611 / K2) — filters before Limit; default Limit 50
timeline := store.ListMemoryWithOptions(memory.ListMemoryOptions{
	SessionID: "sess-abc",
	TimeFrom:  &from,
	TagPrefix: "subject:",
	Limit:     50, // <=0 also defaults to 50
})
```

### SearchMemory / temporal options

`SearchMemory(query, tier, limit, vec)` is a thin wrapper over `SearchMemoryWithOptions` with no session/time filters and `ReRankTemporal=false` (behavior-preserving).

`SearchMemoryOptions` supports:

| Field | Effect |
|-------|--------|
| `SessionID` | Keep only entries with matching `SessionID` |
| `TimeFrom` / `TimeTo` | Inclusive filter on event time: `Timestamp` if set, else `CreatedAt`, else `LastAccessed` |
| `Limit` | Result cap (default 10) |
| `Tier` | Optional tier restriction |
| `QueryVec` | Vector semantic ranking when non-empty |
| `ReRankTemporal` | After keyword/vector path, sort by `CalculateRelevanceScore` descending |

### ListMemoryWithOptions / timeline (s611)

Event-time ordered listing for session timelines. Filters run **before** Limit (same underfill class as search). FS scan is **O(n)**; a full event-time index is residual/later.

| Field | Effect |
|-------|--------|
| `SessionID` | Keep only matching `SessionID` |
| `TimeFrom` / `TimeTo` | Inclusive filter on `entryEventTime` |
| `Tag` / `TagPrefix` | Exact or prefix match on `TemporalTags` or `Content.Tags` |
| `Query` | Case-insensitive substring on Summary / Full / OriginalText |
| `Limit` | Result cap (default **50** when ≤ 0) |
| `Tier` | When set, only that tier |
| `IncludeArchival` | When `Tier` is nil, also include Archival (default tiers: W+C+S) |
| `Ascending` | `false` = newest first (default) |

## Qdrant Integration (Vector Database)

### Podman Example (Recommended for local dev)

```bash
podman run -d --name qdrant \
  -p 6333:6333 -p 6334:6334 \
  -v qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant
```

Then initialize:

```go
vs := memory.NewVectorStore("http://localhost:6333", "memory_collection")
_ = vs.CreateCollection(768) // or CreateSparseCollection()
```

## Development & Testing

### Prerequisites

- Go 1.22 or later
- **Podman** (or Docker) — required for full Qdrant integration testing
  - The test suite includes `startTemporaryQdrant()` which automatically starts a temporary Qdrant container via Podman when available.
  - Unit tests run without Podman/Qdrant (they test the disabled graceful path).
- Git

### Setup

```bash
git clone https://github.com/iome-sh/memory.git
cd memory
go mod download
```

### Running Tests

```bash
# Run all tests (unit + optional Qdrant integration if Podman is present)
go test -v ./...

# Run with race detector
go test -race ./...

# Specific package
go test -v ./... -run TestPalaceStore

# Force skip Podman-based Qdrant tests
PODMAN_QDRANT_SKIP=1 go test -v ./... -run TestVectorStore_WithTemporaryQdrant
```

### Local Qdrant Development

```bash
# Start Qdrant in background
podman run -d --rm \
  -p 6333:6333 -p 6334:6334 \
  -v $(pwd)/data/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant

# Verify
curl http://localhost:6333/collections
```

### Building & Linting

```bash
go build ./...
go fmt ./...
go vet ./...
```

## Benchmarking with LongMemEval

We provide first-class support for the official [LongMemEval](https://github.com/xiaowu0162/LongMemEval) benchmark (ICLR 2025) so you can measure recall accuracy, temporal reasoning, and token efficiency of the RecMem features.

### Quick Start

**Production ONNX smoke gate** (no Qdrant; ingest + retrieve recall on golden-retriever fixture):

```bash
make longmemeval-smoke
```

**Offline ONNX recall gate** (four semantic personal-fact sessions; no OpenAI; requires ≥3/4 top-hit recall):

```bash
make longmemeval-recall-gate
```

**Offline LongMemEval recall benchmark** (official JSON schema subset in `testdata/`; no OpenAI):

```bash
# Committed 3-example oracle subset (CI default)
make longmemeval-bench

# Full official oracle split (download first)
make download-dataset
make longmemeval-bench-full
```

The Go CLI (`cmd/longmemeval-bench`) ingests `haystack_sessions` via `IngestTurn`, retrieves with ONNX `SearchMemory`, and scores recall when the oracle answer (or significant answer tokens) appears in top-k memory text. Exit code 1 when recall falls below `-min-recall` (default 0.6).

`SearchMemory` / `SearchMemoryWithOptions` precomputes one embedding per candidate (batch ONNX when `BatchEmbeddingFunc` is set on `PalaceConfig`), avoiding O(n log n) re-embeds inside the sort comparator. Session and time-window filters run before ranking; optional `ReRankTemporal` re-sorts by multi-factor relevance after the keyword/vector path.

Bench flags:

```bash
go run ./cmd/longmemeval-bench \
  -dataset testdata/longmemeval_oracle_subset.json \
  -quiet \
  -json-report /tmp/bench-report.json
```

Optional env overrides for `scripts/longmemeval_recall_bench.sh`:

| Variable | Default |
|----------|---------|
| `LONGMEMEVAL_DATASET` | `testdata/longmemeval_oracle_subset.json` |
| `LONGMEMEVAL_TOPK` | `5` |
| `LONGMEMEVAL_MIN_RECALL` | `0.6` |
| `LONGMEMEVAL_QUIET` | unset (`1` = aggregate line only) |
| `LONGMEMEVAL_JSON_REPORT` | unset (path for JSON aggregate report) |
| `MEMORY_ONNX_MODEL_PATH` | `testdata/models/KnightsAnalytics_bge-small-en-v1.5` when present |

**Full 500-question eval + OpenAI judge** (offline recall first, then QA generation + official scorer):

```bash
pip install -r requirements-bench.txt
make download-dataset

# Phase 1–2: offline recall on full oracle split (no API key)
make longmemeval-full-eval
# Or quiet 500-q bench only:
LONGMEMEVAL_DATASET=data/longmemeval_oracle.json LONGMEMEVAL_QUIET=1 \
  make longmemeval-bench-full

# Phase 3: QA + judge (requires OPENAI_API_KEY and running server)
export OPENAI_API_KEY=sk-...
export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5
go run cmd/longmemeval-server/main.go &

make longmemeval-qa-generate LONGMEMEVAL_QA_LIMIT=500 LONGMEMEVAL_QA_WORKERS=4
make longmemeval-judge LONGMEMEVAL_JUDGE_MODEL=gpt-4o-mini
```

`scripts/longmemeval_qa_generate.py` uses `requests.Session` + `ThreadPoolExecutor` (default 4 workers) and **one OpenAI call per question** (combined extract+answer prompt). `scripts/longmemeval_judge.sh` shallow-clones the official LongMemEval eval scripts via `scripts/clone_longmemeval_eval.sh` and runs `evaluate_qa.py`; it skips gracefully when `OPENAI_API_KEY` is unset.

Python orchestrator recall-only mode (server must be running; no API key):

```bash
export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5
go run cmd/longmemeval-server/main.go &
python scripts/longmemeval_orchestrator.py \
  --dataset testdata/longmemeval_oracle_subset.json \
  --recall-only --topk 5
```

1. **Start the benchmark server** (uses your PalaceStore + SearchMemory):
   ```bash
   # Optional: production ONNX embeddings (384-d BGE-small; dramatically better recall)
   export MEMORY_ONNX_MODEL_PATH="$(go run ./scripts/download_onnx_model.go)"
   # Or point at cached testdata: export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5

   go run cmd/longmemeval-server/main.go
   ```
   The server listens on `http://localhost:8765` and exposes `/ingest`, `/retrieve`, and `/synthesize`.

   On startup the server logs `embedding mode=onnx` or `embedding mode=hash`. When `MEMORY_ONNX_MODEL_PATH` is unset, hash embeddings use dimension **768**; with ONNX, Qdrant collections are created at **384** (`memory.BGESmallEmbeddingDim`). Init failures fall back to hash with a log line.

2. **Download the official dataset** (small split for testing):
   ```bash
   mkdir -p data/
   wget https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/resolve/main/longmemeval_s_cleaned.json -O data/longmemeval_s_cleaned.json
   # For full eval also download longmemeval_oracle.json
   ```

3. **Run the Python orchestrator** (requires `openai`, `tqdm`, `requests`):
   ```bash
   pip install openai tqdm requests
   python scripts/longmemeval_orchestrator.py \
       --dataset data/longmemeval_s_cleaned.json \
       --output hypotheses.jsonl
   ```

4. **Run the official judge** (from the LongMemEval repo):
   ```bash
   git clone https://github.com/xiaowu0162/LongMemEval.git
   cd LongMemEval/src/evaluation
   export OPENAI_API_KEY=sk-...
   python evaluate_qa.py gpt-4o ../../../hypotheses.jsonl ../../../data/longmemeval_oracle.json
   python print_qa_metrics.py gpt-4o ../../../hypotheses.jsonl.log ../../../data/longmemeval_oracle.json
   ```

The harness wires `memory.NewGONNXEmbeddingFuncFromEnv()` automatically when `MEMORY_ONNX_MODEL_PATH` is set, exercising `Write` + hybrid `SearchMemory` with semantic ONNX re-ranking. Without ONNX, recall on personal-fact questions is limited by hash embeddings.

See `scripts/longmemeval_orchestrator.py` for the full mapping logic and easy customization.

### Updating Documentation

When rolling out new features, always update:
- `docs/memory-refactor-improvements.md` (mark completed items)
- This README.md if public API or setup changes

Then commit with a clear message, e.g.:
```bash
git commit -m "feat(benchmark): add LongMemEval HTTP server and Python orchestrator harness"
```

## Cross-Pollination from H-Mem

See `docs/memory-refactor-improvements.md` for the full prioritized roadmap (H-Mem temporal windows, s+t+r scoring, entity graph, etc.).

**Temporal memory kernel** (fields, decay, multi-factor score, ingest/search phases K0–K4): see [`docs/temporal-memory-kernel-roadmap.md`](docs/temporal-memory-kernel-roadmap.md). This package is a single-tenant FS Palace; multi-tenant isolation and product Memory GA live in the host (aion), not here.

## License
Private (same as ossa).
