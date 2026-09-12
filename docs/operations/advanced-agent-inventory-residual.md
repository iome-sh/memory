# Advanced agent inventory residual pin (s1297)

**Status:** residual pin closed (kernel docs + offline gate pin) · **2026-08-05**  
**Free eng residual pin:** **s1297**  
**Free eng concurrent:** **s1296+** after free-floor **s1294** · lag **s1295**  
**Scope:** kernel-only offline residual SSOT for **advanced agent inventory** surfaces already shipped in this package — **no invent GA**

> **Non-claim (read first):** This pin inventories kernel APIs that host/TUI agent MCP surfaces may wire. Closing the residual documents the inventory. It does **not** promote multi-hop lite to full graph RAG, bi-temporal lite to dual-clock Graphiti, or A3 lite to NLP contradiction detection. Residual PASS ≠ live dogfood.

## Why this residual exists

Free-eng continuum work on host/TUI agent MCP tools can over-read kernel slices as a hosted Memory product. This pin freezes the **kernel advanced agent inventory** so claims stay accurate:

| Kernel surface | Truth class |
|----------------|-------------|
| Multi-hop + PreferShorterHops | multi-hop lite ≠ full graph RAG |
| ListFactsAsOf / EntryValidAt | K4 bi-temporal lite ≠ dual-clock Graphiti |
| SupersedeEntityFacts | A3 lite ≠ NLP contradiction |
| ListMemoryWithOptions | K2 timeline filters before limit · full FS event-time index residual |

Host/TUI peers are **mention only** — not claimed as this repo's product packaging.

## Kernel inventory (SSOT)

| Kernel API | Serial / class | Class |
|------------|----------------|---------|
| **`MultiHopRetrieve` + `PreferShorterHops`** | s1067 / s1278 · A2 multi-hop lite | Shorter BFS hop then event time · multi-hop lite ≠ full graph RAG · not full Zep/Graphiti path scoring |
| **`ListFactsAsOf` / `EntryValidAt`** | s616 · K4 bi-temporal lite | Host-written `valid_from` / `valid_until` · bi-temporal lite ≠ dual-clock Graphiti |
| **`SupersedeEntityFacts`** | s632 · A3 lite | Closes `valid_until` for entity keys · A3 lite ≠ NLP contradiction |
| **`ListMemoryWithOptions`** | s611 / s1066 / #44 · K2 timeline | Filters apply **before** Limit · meta index + durable snapshot · incremental event-time index residual |

### Shipped signatures (reference)

```go
func (ps *PalaceStore) MultiHopRetrieve(opts MultiHopOptions) []MemoryEntry
// PreferShorterHops *bool — default true (nil or true); false = legacy seed-match-first

func (ps *PalaceStore) ListFactsAsOf(opts FactsAsOfOptions) []MemoryEntry
func EntryValidAt(e MemoryEntry, asOf time.Time) bool

func (ps *PalaceStore) SupersedeEntityFacts(entityKey string, asOf time.Time) (int, error)

func (ps *PalaceStore) ListMemoryWithOptions(opts ListMemoryOptions) []MemoryEntry
```

## Host / TUI peers (mention only)

These product surfaces wire the kernel inventory above. They are **peers mention only** in this residual — not owned here:

| Peer surface | Role (mention only) |
|--------------|---------------------|
| **`memory_timeline`** | Host MCP / agent timeline over ListMemoryWithOptions |
| **`memory_related`** | Host MCP multi-hop related · PreferShorterHops wire |
| **`memory_facts_as_of`** | Host MCP facts-as-of · ListFactsAsOf |
| **`memory_supersede_entity`** | Host MCP supersession · SupersedeEntityFacts |
| **TUI s1296** | timeline / compact-status (host UI peer) |

Kernel pin closes residual documentation **inside** `github.com/iome-sh/memory`. Host SRED triad / dual-repo ledger lives in the private control-plane continuum when claimed.

## Non-goals

| Claim people might over-read | Truth |
|------------------------------|-------|
| Kernel advanced agent inventory | Shipped lite slices only · library kernel APIs |
| Multi-hop | **Lite** · not full graph RAG · not full Zep/Graphiti path scoring |
| Facts as-of | **K4 bi-temporal lite** · not dual-clock Graphiti |
| Supersession | **A3 lite** · closes `valid_until` · not NLP contradiction |
| Timeline list | **K2** filters-before-limit · full FS event-time index residual |
| Residual / gate PASS | Offline tree SSOT only · **PASS ≠ live dogfood** · **no invent GA** |
| Free eng pin | **s1297** · concurrent **s1296+** after free-floor **s1294** · lag **s1295** |

## Related residuals

| Residual | Role |
|----------|------|
| [`k2-event-time-index-residual.md`](k2-event-time-index-residual.md) | s1303 K2 event-time index residual pin (peer · ListMemoryWithOptions O(n) / full event-time index residual) |
| [`recmem-compaction-residual.md`](recmem-compaction-residual.md) | s1313 RecMem / compaction residual pin (peer · AutoRecMemCompaction partial · trigger advisory · HITL · not invent GA token-reduction) |
| [`multi-hop-hop-distance-ranking-residual.md`](multi-hop-hop-distance-ranking-residual.md) | s1278 hop-distance ranking residual pin (peer) |
| Host advanced agent MCP residual | PreferShorterHops host wire / inventory (mention only · not this repo) |
| TUI s1296 timeline/compact-status | Host UI peer (mention only) |

## Gate (local)

```bash
make residual-gate
# equivalent:
./scripts/advanced_agent_inventory_residual_gate.sh

# Soft skip (CI nest / operator opt-out):
SKIP_ADVANCED_AGENT_INVENTORY=1 ./scripts/advanced_agent_inventory_residual_gate.sh

# Optional unit focus (cheap; not required for residual gate):
go test . -count=1 -run 'MultiHop|FactsAsOf|Supersede|ListMemory'
```

Offline PASS proves **tree SSOT for kernel advanced agent inventory + this residual doc** — **not** full graph RAG · **PASS ≠ live dogfood** · **no invent GA**.

## Pin summary

| Claim | Truth |
|-------|-------|
| Free eng residual pin | **s1297** closed (this doc + gate + README pin) |
| Free eng concurrent | **s1296+** after free-floor **s1294** · lag **s1295** |
| Kernel inventory | MultiHopRetrieve + PreferShorterHops · ListFactsAsOf · SupersedeEntityFacts · ListMemoryWithOptions |
| Multi-hop class | **Lite** ≠ full graph RAG |
| K4 class | bi-temporal **lite** ≠ dual-clock Graphiti |
| A3 class | **lite** ≠ NLP contradiction |
| K2 class | filters before limit · full FS event-time index residual |
| Gate result | **RESULT PASS** · residual PASS ≠ live dogfood · **no invent GA** |

*s1297 · 2026-08-05 · kernel advanced agent inventory residual free-eng pin · free eng concurrent s1296+ after free-floor s1294 · lag s1295 · MultiHopRetrieve PreferShorterHops multi-hop lite ≠ full graph RAG · ListFactsAsOf K4 bi-temporal lite ≠ dual-clock Graphiti · SupersedeEntityFacts A3 lite ≠ NLP contradiction · ListMemoryWithOptions K2 timeline · host/TUI peers memory_timeline · memory_related · memory_facts_as_of · memory_supersede_entity · TUI s1296 mention only · no invent GA · PASS ≠ live dogfood · RESULT PASS*
