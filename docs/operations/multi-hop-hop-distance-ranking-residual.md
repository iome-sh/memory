# Multi-hop hop-distance ranking residual pin (s1278)

**Status:** residual pin closed (kernel docs + tests pin) · **2026-08-05**  
**Free eng residual pin:** **s1278** (memory serial; continuum alignment with host free eng floor **s1276+**)  
**Implementation pin:** **s1067** / A2 hop-distance ranking lite · **v1.5.7 continuum**  
**Scope:** kernel-only residual pin for `PreferShorterHops` / `ExpandRelatedEntitiesHops` / `MultiHopRetrieve` ranking

> **Non-claim (read first):** Closing this residual pin documents an already-shipped s1067 path-aware ranking lite slice. It does **not** promote multi-hop lite to full graph RAG, full Zep/Graphiti path scoring, typed-edge weights, or embedding-guided walks.

## Why this residual exists

s1067 landed hop-distance ranking on the multi-hop path (`ExpandRelatedEntitiesHops` + default `PreferShorterHops`). Host / product surfaces may still frame ranking quality loosely. This pin freezes residual documentation so free-eng continuum claims stay accurate:

| Surface | Truth |
|---------|--------|
| Kernel `MultiHopRetrieve` default sort | Min BFS hop ascending (seed = hop 0), then event time descending within hop |
| `PreferShorterHops` | Default **true** (`nil` or explicit true); `false` restores legacy seed-match-first then event time |
| `ExpandRelatedEntitiesHops` | Entity key → minimum hop distance from seed over `GetRelatedEntities` |
| Path scoring | **Lite only** — shorter hop preferred; not full path scores / edge weights / NLP |
| Product Memory | Library kernel package only |

## Shipped behavior (s1067; residual s1278)

```go
// PreferShorterHops ranks by minimum BFS hop distance from seed (lower first),
// then event time descending within the same hop. Default true (nil or true).
// Set to a false pointer to opt out and use legacy seed-match-first sort.
PreferShorterHops *bool

func (ps *PalaceStore) ExpandRelatedEntitiesHops(seed string, maxHops int) map[string]int
func (ps *PalaceStore) MultiHopRetrieve(opts MultiHopOptions) []MemoryEntry
```

Order of operations (ranking step only):

1. After expansion + entry collect + optional filters, each hit carries **min hop** among matched expanded entity keys.
2. **Default / `PreferShorterHops=true`:** sort hop ascending, then event time desc (stable ID tie-break).
3. **`PreferShorterHops=false`:** legacy seed-match first, then event time desc (does **not** prefer hop1 over hop2).

Tests: `TestMultiHopRetrieve_HopDistanceRanking` (+ residual strengthen for explicit true / hop0-before-hop2 when farther hops are newer).

## Non-goals

This remains **multi-hop lite**:

- BFS on a simple directed adjacency map (`AddEntityRelationship` / `GetRelatedEntities`)
- Entry collect via `TemporalTags` `entity:*`, `Content.Tags`, `Relations.RelatedConcepts`
- Hop-distance sort lite — **not** full Zep / Graphiti path scoring
- **Not** typed edges, community detection, embedding-guided graph walk
- **Not** full graph RAG
- **Not** multi-tenant product packaging

Residual A2 work still open (not claimed by s1278):

- Bidirectional / typed relation edges
- Full path scoring beyond min-hop preference
- Indexes if O(n) FS scans + graph BFS become the bottleneck

## Peer continuum (mention only)

| Peer | Role |
|------|------|
| **host s1277** | Host free-eng residual honesty peer (related routes / MCP wire — not this repo) |
| **TUI** | Related `hop_distance` display on multi-hop related path (host surface) |
| **SDK** | `HopDistance` on related hits when wired by host (not claimed here) |
| **s1297 inventory residual** | Kernel advanced agent inventory residual pin — see [`advanced-agent-inventory-residual.md`](advanced-agent-inventory-residual.md) (MultiHopRetrieve · PreferShorterHops · ListFactsAsOf · SupersedeEntityFacts · ListMemoryWithOptions; host/TUI peers mention only) |

Kernel pin closes residual documentation **inside** `github.com/iome-sh/memory`. Host SRED triad / dual-repo ledger lives in the private control-plane continuum when claimed.

## Gate (local)

```bash
go test ./... -count=1
# Focus:
go test . -count=1 -run 'TestMultiHopRetrieve_HopDistanceRanking|TestExpandRelatedEntitiesHops'
```

PASS proves tree SSOT for hop ranking defaults + tests + this residual doc — **not** full graph RAG.

## Pin summary

| Claim | Truth |
|-------|-------|
| Hop-distance ranking shipped | **Yes** — s1067 / v1.5.7 continuum (`PreferShorterHops` default true) |
| Residual pin | **s1278** closed (this doc + README/roadmap pin) |
| Multi-hop class | **Lite** — not full Zep/Graphiti path scoring |
| Full graph RAG | **No** |
| Package class | Library kernel |
| Free eng floor peer | host **s1276+** / peer **s1277**; TUI hop_distance display mention-only |
