# K2 event-time index residual honesty (s1303)

**Status:** residual-honest closed (kernel docs + offline gate pin) · **2026-08-05**  
**Free eng residual pin:** **s1303**  
**Free eng concurrent:** **s1301+** after free-floor **s1299** · lag **s1300**  
**Scope:** kernel-only offline residual SSOT for **K2 `ListMemoryWithOptions` / timeline path** — full **event-time index residual**; does **not** invent index green or product Memory GA

> **Non-claim (read first):** This pin freezes residual honesty for the K2 timeline list path. Closing residual honesty documents that **filters before Limit** is shipped and that a **full event-time index is still residual** (O(n) FS scan class remains). It does **not** invent index green, claim a durable on-disk event-time index, promote the best-effort meta index to product GA, or invent **product Memory GA**. Host dual_write is a host concern (**OFF**). Residual PASS ≠ live dogfood · residual ≠ invent index green · **no invent GA**.

## Why this residual exists

Host/TUI agent MCP timeline surfaces (`memory_timeline`, TUI `/memory timeline`) rest on kernel `ListMemoryWithOptions`. Free-eng continuum work can over-read s611/s1066 slices as a shipped full event-time index or Memory GA. This pin freezes residual honesty so claims stay accurate:

| Surface | Truth class |
|---------|-------------|
| `ListMemoryWithOptions` filters **before** Limit | **Shipped** (underfill class; s611 / K2) |
| FS scan **O(n)** | Residual class — full event-time index **not shipped** |
| Best-effort meta index (s1066) | Kernel optional cache only · **not** full durable event-time index · **not** invent index green |
| Host `memory_timeline` | Uses list path · filters-before-limit residual-honest · mention only |

## Truth table (SSOT)

| Claim | Status | Notes |
|-------|--------|-------|
| Filters apply **before** Limit (underfill class) | **Shipped** | Same underfill class as K1 `SearchMemoryWithOptions` |
| Event-time ordered listing (`ListMemoryOptions`) | **Shipped** | Session / time / tag / query / tier / ascending |
| Default Limit 50 | **Shipped** | Timeline-friendly |
| FS Palace source of truth | **Yes** | Index paths are best-effort only when present |
| Full **event-time index** (durable / product-grade) | **Residual** · **not shipped** | Residual honesty pin **s1303** |
| O(n) FS scan class | **Residual-honest** | Scan / rebuild walk remains O(n); do **not** invent index green |
| Host `memory_timeline` list path | **Peer** (mention only) | Filters before limit residual-honest when wired to kernel |
| Product Memory GA | **No** | Kernel ≠ product Memory GA |

### Shipped signature (reference)

```go
// ListMemoryOptions configures event-time ordered listing with optional filters (K2 / s611).
// Filters are applied before Limit so windowed/session timelines do not underfill.
func (ps *PalaceStore) ListMemoryWithOptions(opts ListMemoryOptions) []MemoryEntry
```

Order of operations (list path): collect candidates → session → time → tag filters → query substring → sort by `entryEventTime` → **Limit**.

## Kernel peers (mention)

| Peer | Role |
|------|------|
| **s611** | `ListMemoryWithOptions` / `ListMemoryOptions` first slice (K2 timeline) |
| **s1066** | Optional best-effort meta index (partial · not full event-time index · not invent index green) |
| **s1297** | Advanced agent inventory residual (includes K2 timeline row) — see [`advanced-agent-inventory-residual.md`](advanced-agent-inventory-residual.md) |

## Host / TUI peers (mention only)

| Peer surface | Role (mention only) |
|--------------|---------------------|
| **`memory_timeline`** | Host MCP / agent timeline over `ListMemoryWithOptions` |
| **TUI s1296** | `/memory timeline` + compact-status (host UI peer · not kernel GA) |
| **Concurrent s1301** | Host/TUI free-eng concurrent peer (mention only · not this repo) |

Kernel pin closes residual honesty **inside** `github.com/iome-sh/memory`. Host SRED triad / dual-repo ledger lives in aion continuum when claimed.

## Honesty / non-goals

| Claim people might over-read | Truth |
|------------------------------|-------|
| Full event-time index shipped | **No** — **event-time index residual** · residual ≠ invent index green |
| O(n) FS scan gone | **No** — O(n) class residual-honest |
| Filters after limit (underfill) | **No** — filters **before** limit **shipped** |
| Meta index = product index GA | **No** — best-effort kernel cache only when present |
| Product Memory GA | **No** — kernel ≠ product Memory GA · **not Memory GA** |
| dual_write | Host concern · **dual_write OFF** (not a kernel product flag) |
| Residual / gate PASS | Offline tree SSOT only · **PASS ≠ live dogfood** · **no invent GA** |
| Free eng pin | **s1303** · concurrent **s1301+** after free-floor **s1299** · lag **s1300** |

## Related residuals

| Residual | Role |
|----------|------|
| [`advanced-agent-inventory-residual.md`](advanced-agent-inventory-residual.md) | s1297 kernel advanced agent inventory honesty pin (peer · includes ListMemoryWithOptions row) |
| [`recmem-compaction-residual.md`](recmem-compaction-residual.md) | s1313 RecMem / compaction residual honesty pin (peer · AutoRecMemCompaction partial · trigger advisory · HITL) |
| [`multi-hop-hop-distance-ranking-residual.md`](multi-hop-hop-distance-ranking-residual.md) | s1278 hop-distance ranking honesty pin (peer) |
| Host aion dual agent MCP residual | timeline / inventory host wire (mention only · not this repo) |
| TUI s1296 timeline/compact-status | Host UI peer (mention only) |

## Gate (local)

```bash
make residual-gate
# or targeted:
make k2-event-time-index-residual-gate
# equivalent:
./scripts/k2_event_time_index_residual_gate.sh

# Soft skip (CI nest / operator opt-out):
SKIP_K2_EVENT_TIME_INDEX=1 ./scripts/k2_event_time_index_residual_gate.sh

# Optional unit honesty (cheap focus; not required for residual gate):
go test . -count=1 -run 'ListMemory'
```

Offline PASS proves **tree SSOT for K2 ListMemoryWithOptions residual honesty + this residual doc** — **not** product Memory GA · **not** invent index green · full event-time index residual · **PASS ≠ live dogfood**.

## Honesty footer

| Claim | Truth |
|-------|-------|
| Free eng residual pin | **s1303** closed residual-honest (this doc + gate + README pin) |
| Free eng concurrent | **s1301+** after free-floor **s1299** · lag **s1300** |
| Kernel surface | `ListMemoryWithOptions` / `ListMemoryOptions` (s611 / K2) |
| Filters before limit | **Shipped** (underfill class) |
| Full event-time index | **Residual** · **not shipped** · residual ≠ invent index green |
| O(n) FS scan | Residual-honest class remains |
| Host peer | `memory_timeline` · TUI s1296 / concurrent s1301 mention only |
| Product Memory GA | **No** — kernel ≠ product Memory GA · **not Memory GA** |
| dual_write | **OFF** (host concern) · **no invent GA** |
| Gate result | **RESULT PASS** / honesty chain · residual PASS ≠ live dogfood |

*s1303 · 2026-08-05 · K2 event-time index residual free-eng pin · free eng concurrent s1301+ after free-floor s1299 · lag s1300 · ListMemoryWithOptions filters before limit shipped · O(n) FS scan · full event-time index residual not shipped · residual ≠ invent index green · host memory_timeline / TUI s1296 / concurrent s1301 mention only · kernel ≠ product Memory GA · dual_write OFF · no invent GA · PASS ≠ live dogfood · RESULT PASS*
