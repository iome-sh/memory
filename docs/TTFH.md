# TTFH (walking skeleton)

This is the operator-facing TTFH page for `github.com/iome-sh/memory`. It
documents the **walking skeleton** only: ingest three RCA-shaped turns,
retrieve in the **same process**, list facts-as-of, and print `source_hint`.

## Walking skeleton

The first worked path is RCA-shaped, not a chatbot “favourite colour” demo.

1. Ingest **three RCA-shaped turns**
2. **Retrieve in the same process**
3. **List facts-as-of**
4. **Print `source_hint`**

Cost-max for this path: **hash embedder**, **no Qdrant**, **no cloud palace**.

Worked program: [`examples/ttfh_rca`](../examples/ttfh_rca).

```bash
go run ./examples/ttfh_rca
```

Optional palace root: `PALACE_ROOT` (otherwise a temp directory). Default
`IngestTurn` stamps observable `provenance.source_hint=private` (and tag
`source_hint:private`) on the parent and inherited fact children when the
caller does not already supply a classifiable mesh or private source. Host
process labels (`mcp_memory_ingest_turn`, `source:iomesh-memory-mcp`) are
**not** a cite-both class.

The example turns are:

| Turn | What it records |
|------|-----------------|
| 1 | PagerDuty page: webhook ingress 5xx |
| 2 | HMAC-verified delivery HTTP 200 is **not** a consume receipt |
| 3 | `CreateConsumer` 500 when `consumers.mode` is NULL |

Retrieve query: `hmac consume receipt` (session `inc-webhook-5xx`). Then
`ListFactsAsOf` for the same session. Both print `source_hint`.

A green `go run` and a green unit test
(`TestIngestTurn_TTFHShapedWalkingSkeleton`) lock retrieve-after-ingest for
this kernel path. They are the walking skeleton, not a live laptop PULSE run.

## Host path (optional)

Companion pins, not a kernel dependency:

| Piece | Pin | Role |
|-------|-----|------|
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | **v1.3.6** | Agent TUI/CLI |
| [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) | **v0.4.1** | MCP host over this kernel |

In the TUI, with the memory host attached:

- `/memory ingest` — three RCA-shaped turns (local overlay stays **private**)
- `/memory digest --require-sources mesh,private` — **cite-both or explicit miss**

Cite-both needs a mesh-class receipt **and** a private-class receipt in the
digest window. Catalog list is not consume. Grant-only is not cite-both. An
explicit miss is success for this flag; inventing mesh is not.

### Air-gap / no-PULSE worked miss

This transcript is **docs**, not a laptop PULSE run.

Air-gap: no broker, no PULSE consume, local overlay only. After three private
RCA ingest turns (same content as `examples/ttfh_rca`), cite-both must miss
mesh. Do **not** stamp mesh on the local overlay to force cite-both.
Catalog / grant / `source=external` never satisfy cite-both.

```text
/memory ingest
# turn 1: PagerDuty page: webhook ingress 5xx
#         → provenance.source_hint=private  tag=source_hint:private
# turn 2: HMAC-verified delivery HTTP 200 is not a consume receipt
#         → provenance.source_hint=private  tag=source_hint:private
# turn 3: CreateConsumer 500 when consumers.mode is NULL
#         → provenance.source_hint=private  tag=source_hint:private

/memory digest --require-sources mesh,private
require-sources: miss · required=mesh,private · cited=private · missing=mesh · receipt window newest-first · n=3 · mesh not in this receipt set · local palace on disk
```

That miss is **success** for the flag (no mesh-class receipt in an air-gap).
A green unit test and this page still do **not** prove a live laptop PULSE run.

Cost-max stays the same on the host path: hash embedder, no Qdrant, no cloud
palace. Optional Ollama is a TUI pin, not a kernel requirement.

This optional host path is the walking skeleton on the TUI, not a live
PULSE + three RCA + cite-both-or-miss laptop clock. Writing the commands here
does not prove a laptop ran them against live PULSE.

## What this page is not

A **real laptop** run is **PULSE + 3 RCA + cite-both-or-miss**.

This page does **not** substitute for that. A unit test does **not**.
`go run ./examples/ttfh_rca` does **not**. Optional TUI/MCP
pins and slash-command names do **not**. The air-gap /
no-PULSE worked miss transcript above is **docs**, not a laptop PULSE run.

Do not treat a docs PR, a README table row, or CI green as that laptop clock.

## Notes

- inspectable filesystem palace remains the source of truth
- one process per palace root (multi-process writers unsupported)
- LongMemEval is a different clock — see [`LONGMEMEVAL.md`](LONGMEMEVAL.md);
  a methodology card does not move TTFH / cite-both
