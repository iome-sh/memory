# Memory Refactor & Improvements Features

**Document Purpose**: Comprehensive analysis of weaknesses/risks in the `internal/self` memory subsystem and concrete proposals for refactor, fixes, and new feature development. This builds directly on the deep review of `memory.go`, `compaction.go`, `ossa_self.go`, and `qdrant.go`.

**Scope**: Focuses on the Palace knowledge base (file-backed tiers + concept graph), compaction engine, simple embeddings, vector store integration, and related lifecycle methods (`updateKnowledgeBase`, `performAgentManagedCompaction`, etc.).

**Status**: Draft proposal for implementation planning. Target branch: `main`.

---

## 1. Executive Summary

The current memory layer implements a sophisticated three-tier hierarchical knowledge base (Working / Contextual / Archival) with:
- JSON file persistence under `.ossa/kb/palace/`
- Simple deterministic embeddings + Qdrant vector store
- LLM-driven agent-managed compaction
- Concept relationship graph
- Rich provenance, temporal tags, metrics, and versioning stubs

While architecturally sound and innovative (especially the self-compaction loop), several critical weaknesses limit reliability, portability, retrieval quality, and long-term maintainability.

This document catalogs those risks and provides prioritized, actionable improvement proposals with implementation considerations.

## 2. Identified Weaknesses & Risks

### 2.1 Non-Semantic / Low-Quality Embeddings

**Location**: `memory.go:generateSimpleEmbedding`

- Implementation: SHA-256 hash of lowercased text → deterministic seed → random float32 vector (normalized to unit length).
- Strengths: Reproducible, zero external deps, fast.
- **Major Risk**: Embeddings carry almost no semantic meaning. Cosine similarity will frequently return irrelevant or noisy results.
- Impact: `searchSimilarVectors` (qdrant.go) and any future retrieval/compaction ranking become unreliable. Agent decisions based on vector similarity degrade over time.
- Related: `cosineSimilarity` helper exists but is under-utilized in current flows.

### 2.2 Stubbed / Non-Functional Temporal Decay

**Location**: `memory.go:calculateTemporalDecay`

```go
func calculateTemporalDecay(entry MemoryEntry) float64 {
	return 1.0
}
```

- Always returns 1.0 (no-op).
- `calculateRecencyBoost` is implemented and useful, but decay is missing.
- **Risk**: Old entries in Contextual/Archival tiers never lose relevance scoring. Memory bloat and stale knowledge pollution over long agent runs (hundreds of cycles).
- Missing integration into scoring, compaction candidate selection, or `updateMemoryAccess`.

### 2.3 Brittle String-Based Compaction Action Parsing

**Location**: `compaction.go:parseCompactionActions`

- Relies on fragile `strings.HasPrefix` + manual line splitting for `ACTION:`, `TARGET:`, `REASON:` blocks.
- No schema validation, no handling of malformed LLM output, no JSON mode fallback.
- **Risk**: LLM output format drift (even small changes in prompting or model) silently breaks compaction or produces partial actions. Non-deterministic failures.
- Current prompt engineering tries to constrain output, but parsing remains the weak link.

### 2.4 Darwin-Only Build Constraint

All files in `internal/self/` start with:

```go
//go:build darwin
// +build darwin
```

- **Risks**:
  - Prevents building/running on Linux, Windows, or other Unix platforms.
  - Limits CI/CD, contributor access, and deployment flexibility.
  - Likely tied to llama.cpp Metal + podman assumptions, but many parts (memory, compaction, file I/O) are portable.
- `startSearxng` / `ensureQdrant` use podman + dynamic ports + tmpfs — these can be made cross-platform with fallbacks.

### 2.5 Insufficient Error Handling & Resilience

Throughout `memory.go` and `compaction.go`:
- Many functions print to stdout and return early on error (e.g., `writeMemoryEntry`, `updateKnowledgeBase`).
- JSON unmarshal errors are silently ignored in `loadConceptGraph` and `loadMemoryEntry`.
- No retries, no circuit breakers, no structured logging.
- **Risk**: Silent data loss or corruption during high-cycle agent runs. Hard to debug production issues.
- `ensurePalaceDirs` and atomic rename are good, but not consistently applied or monitored.

### 2.6 Under-Utilized Versioning System

- `archiveToVersions` function exists and creates `versions/memory-entries/<id>/vN.json`.
- However, it is **never called** from `updateKnowledgeBase`, compaction handlers, or `archiveEntry`.
- **Risk**: Loss of historical decision lineage. Harder to audit, rollback, or analyze agent evolution over time.
- Version field in `MemoryEntry` is set but not incremented on updates.

### 2.7 Missing Working Tier Management & Eviction Policy

- No logic to promote/demote between tiers automatically.
- No size limits, age-based eviction, or LRU-style cleanup for TierWorking.
- **Risk**: Working memory grows unbounded during active sessions → performance degradation and Palace directory bloat.
- Compaction currently only targets Contextual (every 12 cycles) and Archival (every 75 cycles).

### 2.8 Compaction Non-Determinism & Heavy LLM Dependency

- `performAgentManagedCompaction` + handlers rely entirely on LLM (`generateForCompaction` using gemma-2-2b or deepseek-r1-distill-1.5b).
- Prompt is long and complex; output quality varies.
- No fallback strategies, no verification of created entries, no scoring of compaction quality.
- `currentCompactionConfig` struct exists but is never updated or persisted with real metrics (`AvgScoreBefore/After`, `Improvement`).
- **Risk**: Inconsistent knowledge base evolution. Potential for hallucinated summaries or loss of important details.

### 2.9 Other Notable Gaps

- `truncate` helper lives in `ossa_self.go` but is used heavily by memory paths (should perhaps move or be shared cleanly).
- Limited use of `Relations` fields and `getRelatedMemoryIDs` in current flows.
- No memory query/search helper functions exposed for TUI or other packages.
- Payload indexes in Qdrant are created but search filters are rarely used.
- No compaction metrics or health reporting.

## 3. Proposed Fixes, Refactors & New Features

### 3.1 Semantic Embeddings (High Priority)

**Goal**: Replace or augment `generateSimpleEmbedding` with meaningful vectors.

**Options**:
1. Add embedding support to the existing llama router (many modern GGUF models support embeddings).
2. Introduce a dedicated small embedding model (e.g., `nomic-embed-text` or `bge-small`) loaded via the same router.
3. Hybrid: keep simple embedding for fallback + semantic for retrieval.

**Changes**:
- Extend `OssaSelf` / router with `GenerateEmbedding(text string) []float32`.
- Update `updateKnowledgeBase`, compaction handlers, and any future retrieval to use it.
- Store both legacy and new vectors during transition (or bump version).
- Add cosine similarity usage in compaction candidate ranking.

**Benefits**: Dramatically better retrieval, more intelligent compaction decisions, foundation for RAG-like features.

### 3.2 Implement Real Temporal Decay + Unified Scoring (High Priority)

**Goal**: Make old knowledge naturally lose influence.

**Proposal**:
- Implement `calculateTemporalDecay` using exponential decay based on `LastAccessed` or `UpdatedAt` vs now.
- Combine with existing `calculateRecencyBoost` into a single `calculateRelevanceScore(entry MemoryEntry) float64` helper.
- Use this score in:
  - `listEntriesInTier` sorting
  - Compaction candidate selection (beyond just ScoreImpact)
  - Future Working tier eviction
- Persist or compute on load; update `Metrics` when accessed.

**Additional**: Expose decay parameters in `CompactionConfig` or a new `MemoryConfig`.

### 3.3 Robust Compaction Action Handling (High Priority)

**Goal**: Make compaction reliable regardless of LLM output variance.

**Proposals**:
1. **Structured Output**: Switch compaction prompt to request JSON array of actions (or use constrained decoding if llama.cpp supports).
2. **Improved Parser**: Replace `parseCompactionActions` with a more tolerant parser + schema validation. Add unit tests.
3. **Verification Step**: After executing actions, re-load created entries and run basic sanity checks (non-empty content, valid parents, etc.).
4. **Fallbacks**: On parse failure or empty actions, fall back to simple heuristic compaction (e.g., archive lowest-score entries).
5. **Update `currentCompactionConfig`**: After each run, record before/after scores and persist the struct (or log it).

**New file idea**: `internal/self/compaction_test.go` with golden tests for parser and handlers.

### 3.4 Cross-Platform Portability (Medium-High Priority)

**Goal**: Remove hard darwin dependency for core memory features.

**Approach**:
- Move build tags to only the files that truly need them (`tui.go`, parts of `ossa_self.go` that use llama contexts).
- Make `memory.go` and `compaction.go` build on all platforms (they are pure Go + stdlib + cuid2).
- For `qdrant.go` and `searxng.go`:
  - Keep podman commands but add runtime detection + graceful fallback (e.g., assume external qdrant/searxng if podman unavailable).
  - Use `runtime.GOOS` checks where necessary.
- Update `NewOssaSelf` and `Start` to handle platform differences.
- Add Linux CI job (even if limited).

**Benefit**: Enables broader testing, Docker deployment, and contributor participation.

### 3.5 Strengthen Error Handling & Observability (Medium Priority)

- Introduce a simple structured logger (or use `log/slog`).
- Return errors up the call stack instead of fmt.Printf + silent return in most memory paths.
- Add retry logic (with backoff) for Qdrant HTTP calls and file writes.
- Expose memory health metrics (entry counts per tier, last compaction cycle, vector store status).
- Consider adding a `MemoryStats` struct returned by list/load functions.

### 3.6 Activate & Enhance Versioning (Medium Priority)

- Call `archiveToVersions(entry)` inside `writeMemoryEntry` (or before major mutations).
- Increment `Version` field on updates (currently always starts at 1).
- Expose version history loading function.
- Use versions during compaction audit or rollback scenarios.
- Update `Provenance` to track version lineage.

### 3.7 Working Tier Lifecycle & Automatic Promotion/Demotion (Medium Priority)

**New logic needed**:
- On access or score threshold, promote Working entries to Contextual.
- Implement size-based or age-based eviction from Working (move to Contextual or archive low-value items).
- Add `DemoteToWorking` / `PromoteToContextual` helpers.
- Integrate with TUI for manual overrides.
- Track Working tier pressure and trigger early compaction if needed.

### 3.8 Additional Feature Proposals

- **Memory Query API**: Add functions like `SearchMemory(query string, tier MemoryTier, limit int) []MemoryEntry` that combines vector + keyword + graph traversal.
- **Compaction Metrics Dashboard**: Persist and expose `currentCompactionConfig` improvements. Add to TUI.
- **Relation Graph Visualization / Traversal**: Expand `getRelatedMemoryIDs` and add graph algorithms (shortest path, influence scoring).
- **Configurable Compaction Strategies**: Make `Tier2Strategy` / `Tier3Strategy` runtime configurable (currently hardcoded in var).
- **Embedding Versioning & Migration**: Support multiple embedding generations and re-embedding on model upgrade.
- **Unit + Integration Tests**: Heavy investment here — file I/O, compaction parser, vector roundtrips.
- ** Palace Directory Size Monitoring & Cleanup**: Background goroutine or CLI command.

## 4. Implementation Roadmap (Suggested)

**Phase 1 (Quick Wins – 1-2 weeks)**
- Fix temporal decay + unified relevance scoring
- Improve compaction parser + add verification + config tracking
- Activate versioning in write paths
- Add basic error wrapping and logging

**Phase 2 (Core Quality – 2-3 weeks)**
- Semantic embeddings integration
- Working tier promotion/eviction logic
- Cross-platform build tag cleanup + basic Linux smoke tests

**Phase 3 (Advanced Features – Ongoing)**
- Memory query API + graph enhancements
- Observability, metrics, TUI integration
- Full test suite + compaction quality evaluation harness

## 5. Expected Benefits

- Higher quality, more reliable long-term agent memory
- Better retrieval and compaction decisions
- Portable codebase
- Easier debugging and auditing via versioning + logging
- Foundation for advanced features (RAG, self-reflection, multi-agent knowledge sharing)
- Reduced technical debt before scaling agent runs

## 6. Open Questions & Next Steps

- Should we keep the simple embedding as a fast path or deprecate it entirely?
- Preferred embedding model / integration approach with existing llama router?
- How aggressive should Working tier eviction be (risk of losing recent context)?
- Do we want to persist `CompactionConfig` to disk?
- Any constraints from TUI or other packages?

## 7. H-Mem Cross-Pollination Ideas

H-Mem (Hybrid Temporal-Semantic Memory) research proposes a dual-topology architecture coupling a **Temporal Semantic Tree** (hierarchical chronological + semantic consolidation) with a **Relational Knowledge Graph** (entity-centered multihop relationships). It uses hyperparameters α (similarity threshold) and β (temporal window size) to prevent temporal drift during progressive summarization, an agentic retrieval pipeline (planning → scope classification → hybrid graph+tree execution), and a multi-factor scoring function (s semantic + t temporal relevance + r robustness/forgetting curve).

`ossa`'s current memory implementation in `internal/self` (3-tier Palace, LLM-driven compaction, temporal tags, basic concept graph, provenance, Qdrant vectors) has strong natural alignment with H-Mem concepts.

### Key Cross-Pollination Opportunities

**1. Temporal-Window Constrained Consolidation (α + β) — Highest immediate value**

Adopt H-Mem's anti-drift rule directly into compaction:
- Only consider entries for summarization/merge if they share the same β temporal window (leverage existing `TemporalTags` + `Cycle` fields).
- Introduce lightweight α threshold check using cosine similarity before consolidation.
- This turns ad-hoc LLM compaction into structured, progressive, drift-resistant summarization (Working leaves → Contextual summaries → Archival core principles).
- **Location to extend**: `memory.go` (new helper or inside `performAgentManagedCompaction` / `listEntriesInTier`) and `compaction.go` handlers.

**2. Multi-Factor Scoring Objective (s + t + r)**

- Implement real `calculateTemporalDecay` (currently a no-op returning 1.0) using exponential decay + reinforcement based on usage/score impact.
- Add temporal relevance (t) component that maps queries to time intervals using cycle/month tags and computes overlap.
- Combine into `calculateRelevanceScore(entry MemoryEntry)` and apply to compaction candidate sorting, future retrieval, and Working-tier management.
- Directly builds on existing `calculateRecencyBoost` and `Metrics`.

**3. Evolve Simple Concept Graph into Richer Relational KG**

- Add lightweight entity extraction (concepts, preferences, entities) during `writeMemoryEntry` or compaction using LLM.
- Apply normalization + merging to prevent graph bloat (H-Mem style).
- Use extracted entities as anchors for compaction candidate selection and future multihop retrieval.
- Expand `MemoryRelations` and provenance tracking.
- Leverages ossa's existing `addRelationship` / `getRelatedMemoryIDs` infrastructure.

**4. Hybrid Agentic Retrieval Workflow**

Future `SearchMemory` or TUI enhancements can adopt H-Mem's pipeline:
1. Retrieval planning / query decomposition.
2. Scope classification (Short=Working/raw, Long=Archival, Mixed).
3. Graph expansion for entity anchors → bottom-up tier traversal (raw + summaries).
- Combine Qdrant + graph + file tiers for low-noise context assembly.

**5. Stronger Tier-to-Hierarchy Mapping + Progressive Compactification**

Explicitly map:
- `TierWorking` ≈ Leaf / raw fragments
- `TierContextual` ≈ Mid-level semantic summaries
- `TierArchival` ≈ High-level abstractions / core principles
- Make compaction window-constrained and bottom-up.

**6. Additional Synergies**
- Use H-Mem robustness (r) ideas to improve `currentCompactionConfig` tracking and persistence of before/after metrics.
- Extend provenance with consolidation lineage.
- Enhance compaction prompts with temporal-window awareness and scope classification.
- Once semantic embeddings are added, α-threshold consolidation and s-component become highly effective.

### Prioritized Implementation for `ossa`

**Phase 1** (aligns with existing roadmap):
- Real temporal decay + multi-factor (s+t+r) scoring.
- Add temporal-window (β) constraint + optional α check to compaction.

**Phase 2**:
- Entity extraction + richer KG with normalization.
- Use graph entities to guide compaction.

**Phase 3**:
- Full hybrid retrieval workflow.
- Persistent compaction metrics + verification.

### Verdict

H-Mem is an excellent conceptual upgrade path for `ossa`. The Palace's tiered structure + LLM compaction already implements much of the hard scaffolding. H-Mem supplies the precise constraints, scoring mathematics, and hybrid topology thinking needed to make long-term memory more reliable, drift-resistant, and capable of true multihop reasoning across cycles. Adopting the top ideas will directly address documented weaknesses (stubbed decay, limited graph usage, temporal drift risk) while staying true to `ossa`'s practical local Go + file + vector design.

---

**Generated from deep analysis of current `internal/self` implementation.**
**Last reviewed**: May 2026
**Updated**: May 2026 — Added H-Mem Cross-Pollination section (Section 7)
