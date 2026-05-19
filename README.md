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

Initial port in progress. Core memory models and storage from ossa.

## Planned Structure

```
memory/
├── memory.go          # Core types, Palace storage, embeddings, relations, temporal tags
├── compaction.go      # LLM-driven compaction (agent-managed)
├── vector.go          # Optional Qdrant client (or interface)
├── README.md
└── go.mod
```

## Usage (Planned)

```go
import (
	"github.com/sudo-jin/memory"
)

// Example future API
store := memory.NewPalaceStore(".ossa/kb/palace")
entry := memory.MemoryEntry{ ... }
store.Write(entry)
```

## Cross-Pollination from H-Mem

See ossa `docs/memory-refactor-improvements.md` Section 7 for detailed ideas (temporal-window consolidation, s+t+r scoring, richer KG).

## License
Private (same as ossa).
