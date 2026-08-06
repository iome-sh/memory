# RecMem / compaction residual honesty (s1313)

**Status:** residual-honest closed (kernel docs + offline gate pin) · **2026-08-06**  
**Free eng residual pin:** **s1313**  
**Free eng concurrent:** **s1311+** after free-floor **s1309** · lag **s1310**  
**Scope:** kernel-only offline residual SSOT for **RecMem / compaction** primitives already present in this package — **not** invent product Memory GA or GA token-reduction marketing claims

> **Non-claim (read first):** This pin freezes residual honesty for kernel compaction / RecMem paths. `AutoRecMemCompaction` / `PerformCompaction` / `CompactionConfig` exist as **kernel primitives** (partial shipped). Host MCP `memory_trigger_compact` is **advisory** (publish trigger only). TUI requires **HITL** before compact apply (peer **s1311** · mention only). Closing residual honesty does **not** invent product Memory GA, does **not** invent a GA token-reduction product claim from research plan numbers, and does **not** claim dual_write is a kernel product flag (**dual_write OFF** is a host concern). Residual PASS ≠ live dogfood · Compaction PASS ≠ invent Memory GA token-reduction · **no invent GA**.

## Why this residual exists

Host/TUI surfaces (`memory_trigger_compact`, TUI compact-status / HITL) rest on kernel compaction primitives. Free-eng continuum work can over-read RecMem Phase 1–2 slices or historical plan benefits (e.g. research token-reduction language in the integration plan) as product Memory GA. This pin freezes residual honesty so claims stay accurate:

| Surface | Truth class |
|---------|-------------|
| `AutoRecMemCompaction` / `PerformCompaction` / `CompactionConfig` | Kernel primitives **exist** · **AutoRecMemCompaction shipped partial** |
| Host `memory_trigger_compact` | **Advisory** (publish trigger) · not silent apply · host concern |
| TUI compact path | **HITL required** (peer **s1311**) · mention only · not kernel GA |
| Token-reduction product claim | **Not invent GA** · Compaction PASS ≠ invent Memory GA token-reduction |
| dual_write | Host concern · **dual_write OFF** |
| Phase 3 semantic refine | First-slice `SemanticRefine` may exist · full product-grade refine **residual** if still planned |

## Truth table (SSOT)

| Claim | Status | Notes |
|-------|--------|-------|
| `CompactionConfig` + `DefaultCompactionConfig` | **Shipped** (kernel) | Includes RecMem `DataSim` / `DataCount` + fact-protection fields |
| `PerformCompaction` | **Shipped** (kernel primitive) | Agent-managed compaction with temporal window / alpha constraints |
| `AutoRecMemCompaction` | **Shipped partial** | Density / count phase-transition over subconscious buffer → may call `PerformCompaction` |
| `WriteLatent` / subconscious path | **Shipped partial** (Phase 1 class) | Latent buffer for RecMem-style recurrence |
| `SemanticRefine` | **Partial / residual-honest** | Phase 3 first-slice may exist; full product-grade semantic refine residual if still planned |
| Host `memory_trigger_compact` | **Advisory** (peer · mention only) | Publish trigger only · not silent compact APPLY |
| TUI HITL on compact | **Peer s1311** (mention only) | TUI requires HITL · no `trigger_compact` without HITL |
| Product Memory GA | **No** | Kernel ≠ product Memory GA · **not Memory GA** |
| GA token-reduction product claim | **No** | Compaction PASS ≠ invent Memory GA token-reduction · plan research language is historical |
| dual_write | **OFF** | Host concern · not a kernel product flag |

### Shipped signatures (reference)

```go
// CompactionConfig (extended with RecMem phase-transition parameters)
type CompactionConfig struct { /* DataSim, DataCount, strategies, fact protection, … */ }

// PerformCompaction runs agent-managed compaction with H-Mem temporal window + alpha constraints.
func (ps *PalaceStore) PerformCompaction(
	targetTier MemoryTier,
	cfg CompactionConfig,
	generateFn func(prompt string) string,
	vectorCallback VectorStoreCallback,
)

// AutoRecMemCompaction is the production entry point for automatic RecMem formation (partial).
func (ps *PalaceStore) AutoRecMemCompaction(generateFn func(prompt string) string, vectorCb VectorStoreCallback)
```

## Historical / plan SSOT

| Doc | Role |
|-----|------|
| [`docs/recmem-integration-plan.md`](../recmem-integration-plan.md) | **Historical / plan** SSOT for RecMem phases · research goals · Phase 1–3 framing. **Not** product Memory GA packaging. Do **not** promote plan benefit language (e.g. research token-reduction percentages) to invent GA product claims. |

Phase framing (plan-aligned residual honesty):

| Phase | Class | Honesty |
|-------|-------|---------|
| Phase 1 | Latent buffer + configurable `DataSim` / `DataCount` | Partial shipped · kernel primitive class |
| Phase 2 | `AutoRecMemCompaction` + phase-transition trigger | **AutoRecMemCompaction shipped partial** |
| Phase 3 | Semantic refinement safety net | Residual / partial · full product-grade refine residual if still planned |

## Host / TUI peers (mention only)

| Peer surface | Role (mention only) |
|--------------|---------------------|
| **`memory_trigger_compact`** | Host MCP advisory publish trigger · **not** silent apply |
| **TUI s1311** | HITL on compact path · free eng concurrent peer · not kernel GA |
| **TUI s1296** | `/memory timeline` + compact-status (status UI · not invent compaction green) |
| **Concurrent s1311+** | Host/TUI free-eng concurrent peer after free-floor **s1309** · lag **s1310** |

Kernel pin closes residual honesty **inside** `github.com/iome-sh/memory`. Host SRED triad / dual-repo ledger lives in aion continuum when claimed.

## Honesty / non-goals

| Claim people might over-read | Truth |
|------------------------------|-------|
| RecMem = product Memory GA | **No** — kernel primitives only · **kernel ≠ product Memory GA** · **not Memory GA** |
| Compaction gate PASS = GA token-reduction | **No** — Compaction PASS ≠ invent Memory GA token-reduction · **no invent GA** |
| `memory_trigger_compact` auto-applies | **No** — host path is **advisory** (publish trigger) |
| TUI compact without confirmation | **No** — TUI requires **HITL** (peer s1311) |
| Phase 3 full semantic product GA | **No** — Phase 3 semantic refine residual if still planned · partial ≠ invent GA |
| dual_write ON by kernel | **No** — **dual_write OFF** is host concern |
| Residual / gate PASS | Offline tree SSOT only · **PASS ≠ live dogfood** |
| Free eng pin | **s1313** · concurrent **s1311+** after free-floor **s1309** · lag **s1310** |

## Related residuals

| Residual | Role |
|----------|------|
| [`advanced-agent-inventory-residual.md`](advanced-agent-inventory-residual.md) | s1297 kernel advanced agent inventory honesty pin (peer) |
| [`k2-event-time-index-residual.md`](k2-event-time-index-residual.md) | s1303 K2 event-time index residual honesty pin (peer) |
| [`multi-hop-hop-distance-ranking-residual.md`](multi-hop-hop-distance-ranking-residual.md) | s1278 hop-distance ranking honesty pin (peer) |
| [`docs/recmem-integration-plan.md`](../recmem-integration-plan.md) | Historical RecMem plan (not invent GA) |
| Host aion / TUI compact residual | `memory_trigger_compact` advisory · TUI s1311 HITL (mention only · not this repo) |

## Gate (local)

```bash
make residual-gate
# or targeted:
make recmem-compaction-residual-gate
# equivalent:
./scripts/recmem_compaction_residual_gate.sh

# Soft skip (CI nest / operator opt-out):
SKIP_RECMEM_COMPACTION=1 ./scripts/recmem_compaction_residual_gate.sh

# Optional unit honesty (cheap focus; not required for residual gate):
go test . -count=1 -run 'Compaction|RecMem|SemanticRefine|WriteLatent'
```

Offline PASS proves **tree SSOT for RecMem / compaction residual honesty + this residual doc** — **not** product Memory GA · Compaction PASS ≠ invent Memory GA token-reduction · **PASS ≠ live dogfood**.

## Honesty footer

| Claim | Truth |
|-------|-------|
| Free eng residual pin | **s1313** closed residual-honest (this doc + gate + README pin) |
| Free eng concurrent | **s1311+** after free-floor **s1309** · lag **s1310** |
| Kernel surfaces | `AutoRecMemCompaction` · `PerformCompaction` · `CompactionConfig` |
| AutoRecMemCompaction class | **Shipped partial** · not invent product Memory GA |
| Host `memory_trigger_compact` | **Advisory** (publish trigger) · mention only |
| TUI compact | **HITL** required · peer **s1311** mention only |
| Phase 3 semantic refine | Residual / partial · residual if still planned |
| Token-reduction product claim | **No invent GA** · Compaction PASS ≠ invent Memory GA token-reduction |
| Product Memory GA | **No** — kernel ≠ product Memory GA · **not Memory GA** |
| dual_write | **OFF** (host concern) · **no invent GA** |
| Gate result | **RESULT PASS** / honesty chain · residual PASS ≠ live dogfood |

*s1313 · 2026-08-06 · RecMem compaction residual free-eng pin · free eng concurrent s1311+ after free-floor s1309 · lag s1310 · AutoRecMemCompaction shipped partial · PerformCompaction · CompactionConfig kernel primitives · host memory_trigger_compact advisory · TUI s1311 HITL · Phase 3 semantic refine residual if still planned · Compaction PASS ≠ invent Memory GA token-reduction · docs/recmem-integration-plan.md historical · kernel ≠ product Memory GA · dual_write OFF · no invent GA · PASS ≠ live dogfood · RESULT PASS*
