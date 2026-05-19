# memory

Standalone, portable Go package for hierarchical agent memory (Palace).

Extracted and refactored from `github.com/sudo-jin/ossa/internal/self`.

**Module**: `github.com/sudo-jin/memory`

## Goals

- Portable (remove darwin-only where possible)
- Importable as a clean package
- Core types: MemoryEntry, MemoryTier, compaction, simple embeddings, relations, file-backed Palace
- Optional Qdrant vector integration
- Incorporate H-Mem inspired improvements (temporal windows, multi-factor scoring, entity-aware graph)
- Clean API for other agents/repos to import

## Current Status

Core memory models, PalaceStore, compaction, versioning, temporal decay, multi-factor scoring, pluggable embeddings, Working tier lifecycle, PalaceConfig, hybrid SearchMemory, and full Qdrant vector integration (dense + sparse) are complete and tested.

## Planned Structure

```
memory/
├── memory.go          # Core types, Palace storage, embeddings, relations, temporal tags, PalaceConfig
├── compaction.go      # LLM-driven compaction (agent-managed) with VectorStoreCallback
├── vector.go          # Official Qdrant Go client (dense + sparse vectors, batch upsert, worker-pool batch creation, retries, context, partial results)
├── memory_test.go     # Unit tests for PalaceStore, embeddings, compaction, vector
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
- Podman (or Docker) for local Qdrant
- Git

### Setup

```bash
git clone https://github.com/sudo-jin/memory.git
cd memory
go mod download
```

### Running Tests

```bash
# Run all tests with verbose output
go test -v ./...

# Run with race detector
go test -race ./...

# Specific package
go test -v ./... -run TestPalaceStore
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

### Updating Documentation

When rolling out new features, always update:
- `docs/memory-refactor-improvements.md` (mark completed items)
- This README.md if public API or setup changes

Then commit with a clear message, e.g.:
```bash
git commit -m "feat(vector): add worker-pool batch sparse collection creation with retry"
```

## Cross-Pollination from H-Mem

See `docs/memory-refactor-improvements.md` for the full prioritized roadmap (H-Mem temporal windows, s+t+r scoring, entity graph, etc.).

## License
Private (same as ossa).
