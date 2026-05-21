# memory

## v1.0.0 Release (May 20, 2026)

**Stable release** of the hierarchical agent memory package (Palace).

- Production vector search (dense + sparse via Qdrant Query API)
- Full PalaceStore with atomic writes, versioning, temporal decay, multi-factor scoring
- Compaction with temporal windows and alpha constraints
- Hybrid SearchMemory + pluggable embeddings
- All core features complete, tested, and hardened

**Module**: `github.com/sudo-jin/memory@v1.0.0`

Standalone, portable Go package for hierarchical agent memory (Palace).

Extracted and refactored from `github.com/sudo-jin/ossa/internal/self`.

## Goals

- Portable (remove darwin-only where possible)
- Importable as a clean package
- Core types: MemoryEntry, MemoryTier, compaction, simple embeddings, relations, file-backed Palace
- Optional Qdrant vector integration
- Incorporate H-Mem inspired improvements (temporal windows, multi-factor scoring, entity-aware graph)
- Clean API for other agents/repos to import

## Current Status

**v1.0.0 released and stable.** Core memory models, PalaceStore, compaction, versioning, temporal decay, multi-factor scoring, pluggable embeddings, Working tier lifecycle, PalaceConfig, hybrid SearchMemory, and full Qdrant vector integration (dense + sparse) are complete, tested, and production-ready.

## Planned Structure

```
memory/
├── memory.go          # Core types, Palace storage, embeddings, relations, temporal tags, PalaceConfig
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
	"github.com/sudo-jin/memory"
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

// Hybrid search
results := store.SearchMemory("project goals", nil, 10, nil)
```

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
git clone https://github.com/sudo-jin/memory.git
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
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
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

1. **Start the benchmark server** (uses your PalaceStore + SearchMemory):
   ```bash
   go run cmd/longmemeval-server/main.go
   ```
   The server listens on `http://localhost:8765` and exposes `/ingest`, `/retrieve`, and `/compact`.

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

The harness exercises `Write` + hybrid `SearchMemory`. You can easily extend `handleIngest` to call `WriteLatent` + `AutoRecMemCompaction` (with a real `generateFn`) and `SemanticRefine` to quantify the RecMem density-driven gains.

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

## License
Private (same as ossa).
