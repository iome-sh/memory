# TTFH (walking skeleton)

This page documents the first worked path for `github.com/iome-sh/memory`:
ingest three RCA-shaped turns, retrieve in the **same process**, list
facts-as-of, and print `source_hint`.

It is a local, single-process example. It is not a hosted service and not an
end-to-end production run.

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
this kernel path. They are not an end-to-end production run.

## Department overlay

Support-department kit: [`examples/dept-rca/support`](../examples/dept-rca/support)
(ticket export + policy + macro as **private overlay**). Temporal ask: what
was the refund rule as-of the ticket. Digest cite-both: **mesh miss is
success**. It does **not** replace the three-turn
technical skeleton above.

```bash
go run ./examples/dept-rca/support
```

## Host path (optional)

[iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) is the MCP
host over this kernel. It is a companion, not a kernel dependency.

> **Note:** The TUI is archived and no longer installable. To run the kernel
> directly, see the [README Quick start](../README.md#quick-start).

Host-side digests, miss classes and department tags are host formatting of
fields the kernel already stores (`source_hint`, `source_step`, `parent_ids`,
`id`, tags such as `dept:{id}`). The kernel does not classify miss classes,
does not emit a digest footer, and has no org IDs. A local-only palace has no
mesh-class receipts, so a cite-both check over it reports a mesh miss. Do not
stamp mesh on a local overlay to force cite-both.

Kernel facts a host can rely on:

- retrieve and facts-as-of take `SessionIDs` (any-of; also matches
  `conv:<id>` tags)
- an `IngestTurn` error means partial JSON may remain on disk (parent and
  earlier facts are kept; not silent success; not rollback)
- `PersistEmbeddings` defaults to off; hash vectors are never stored

Cost stays the same on the host path: hash embedder, no Qdrant, no cloud
palace.

## Notes

- inspectable filesystem palace remains the source of truth
- one process per palace root (multi-process writers unsupported)
- LongMemEval is a different eval — see [`LONGMEMEVAL.md`](LONGMEMEVAL.md)
