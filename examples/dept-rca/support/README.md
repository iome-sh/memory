# Support department RCA overlay kit (V1.6 Wave 1)

Private overlay fixture: Zendesk-**shaped** ticket export + unused-seat refund
policy + agent macro. Temporal ask: **what was the refund rule as-of the
ticket?**

This is **not** Memory GA. It is **not E-G1**. dual_write **OFF**. Catalog list
is **not** Connected. Event prefix `dept.support.events.*` is **routing**, not
Connected.

Default ingest class: `source_hint=private`. Tags: `dept:support`
`scenario:support`. Digest cite-both (`mesh,private`): **mesh miss is success**.
Do **not** stamp mesh on this overlay to force cite-both.

Cost-max: hash embedder, no Qdrant, no cloud palace. `PersistEmbeddings`
stays default **off**; hash vectors are never stored.

## Run

From the module root:

```bash
go run ./examples/dept-rca/support
```

Optional palace root: `PALACE_ROOT` (otherwise a temp directory).

Session: `dept-support-zd-1001`. Ticket `ZD-1001` created
`2026-06-15T14:22:00Z`. Policy `valid_from 2026-01-01` (unused seats refundable
within 14 days of invoice). Requester is **Example workspace** (no live
customer).

The program ingests each markdown body, searches the refund policy, lists
facts-as-of the ticket created time, and prints `source_hint`. Exit 0 on
success.

A green `go run` and `TestDeptRCASupportKit_WalkingSkeleton` lock this overlay
path. They are **not** E-G1. They do **not** replace the technical three-turn
skeleton in [`examples/ttfh_rca`](../../ttfh_rca).
