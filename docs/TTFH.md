# TTFH (walking skeleton)

This page documents the first worked path for `github.com/iome-sh/memory`:
ingest three RCA-shaped turns, retrieve in the **same process**, list
facts-as-of, and print `source_hint`.

It does **not** declare Memory GA. It does **not** close **E-G1**. dual_write
**OFF**.

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
this kernel path. They are **not** E-G1.

## Department overlay (V1.6)

Support-department kit: [`examples/dept-rca/support`](../examples/dept-rca/support)
(ticket export + policy + macro as **private overlay**). Temporal ask: what
was the refund rule as-of the ticket. Digest cite-both: **mesh miss is
success**. Overlay, **not E-G1**. It does **not** replace the three-turn
technical skeleton above.

```bash
go run ./examples/dept-rca/support
```

Host TUI `iomesh memory ingest-dir` (V1.6 **D2**): default **128 files ×
64 KiB**; `--source-hint private` (never mesh); `--department` /
`--scenario` tags; skip `.pdf` with export-text-first (**no OCR**); skip
report. Kernel example [`examples/dept-rca/support`](../examples/dept-rca/support)
is **unchanged**. Host `--department` maps to palace `Tag` `dept:{id}` on
search and facts-as-of (exact `EntryHasTag`); the kernel has no org IDs.
**not** Memory GA. **not E-G1.** dual_write **OFF**.

Host TUI one-tenant two-kit (V1.6 **D5**): ingest TUI kits
[`examples/dept-rca/support`](../examples/dept-rca/support) and
`examples/dept-rca/ops` into **one** palace (one-tenant composition).
facts-as-of `--department support|ops` is host Tag `dept:{id}` — **not**
a kernel org filter. **≠ two-org**. leftover_is_bind **OPEN**. **not** an
org-wide RCA engine. Mesh miss is **success**; do **not** stamp mesh on
the files. **D5b** two-org **Parked**. **D6** parked (T2/T3/T5/T8, LME %).
**not** Memory GA. **not E-G1.** dual_write **OFF**. Kernel has no org IDs.

Host TUI sales kit (V1.6 **D5c**): ingest TUI kit `examples/dept-rca/sales`
(call-notes + qbr + list-price UTF-8) with support+ops into **one** palace
(still one-tenant composition). facts-as-of `--department sales` is host
Tag `dept:{id}` — **not** a kernel org filter; as-of **before** list-price
change `2026-03-01` (`2026-02-28T18:00:00Z`). **≠ two-org**. leftover_is_bind
**OPEN**. Overlay does **not** GET Salesforce/CRM. Mesh miss is **success**;
do **not** stamp mesh; no invented ARR; no customer names. **D5b** two-org
**Parked**. **D6** parked. **not** Memory GA. **not E-G1.** dual_write **OFF**.
Kernel has no org IDs.

Host TUI customer_success kit (V1.6 **D5d**): ingest TUI kit
`examples/dept-rca/customer_success` (health-note + renewal + playbook
UTF-8) with support+ops+sales into **one** palace (still one-tenant
composition). facts-as-of `--department customer_success` is host Tag
`dept:{id}` — **not** a kernel org filter; as-of **before** renewal
`2026-09-01` (`2026-08-31T18:00:00Z`). **≠ two-org**. leftover_is_bind
**OPEN**. Overlay does **not** GET Salesforce/CRM. Mesh miss is **success**;
do **not** stamp mesh; no invented ARR; no customer names. **D5b** two-org
**Parked**. **D6** parked. **not** Memory GA. **not E-G1.** dual_write **OFF**.
Kernel has no org IDs.

## Host path (optional)

Companion pins, not a kernel dependency. Published tags — do not invent a
newer tag.

| Piece | Pin | Role |
|-------|-----|------|
| [iomesh-tui](https://github.com/iome-sh/iomesh-tui) | **v1.3.7** | Agent TUI/CLI |
| [iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp) | **v0.4.2** | MCP host over this kernel |

In the TUI, with the memory host attached. Same walk as the bullets
below · R1 ≠ R3 overlay PULSE.

- **R0** `iomesh ttfh --unit` — offline smoke (no broker)
- **R1** `iomesh ttfh --live` — fail-open consume probe (**EMPTY** unless decoded messages; never invent **PULSE**; not overlay PULSE)
- [`scripts/ttfh-demo.sh`](https://github.com/iome-sh/iomesh-tui/blob/main/scripts/ttfh-demo.sh) in the TUI repo — unit then optional live
- **R2** `/memory ingest` — three RCA-shaped turns (local overlay stays **private**)
- **R2** `/memory digest --require-sources mesh,private` — **cite-both or explicit miss** (miss is **named**: `no_mesh_pulse` when `missing=mesh`, etc.)
- **R2** Short-term: `/memory patterns` — ops **Beta** · empty ≠ invent · never APPLY
- **R2** Long-term: `/memory facts-as-of --as-of <RFC3339>` — palace · **not Memory GA**
- **R3** `/dashboard` consume — parked · required for E-G1 · `--live` decoded-N is not this
- **R4** After PULSE: `iomesh memory pull` — dual_write **OFF** · pull ≠ Connected

Cite-both needs a mesh-class receipt **and** a private-class receipt in the
digest window. Catalog list is not consume. Grant-only is not cite-both. An
explicit miss is success for this flag; inventing mesh is not.

`--unit` wins over `--live` (stay offline). `--live` is a light consume probe
only: **EMPTY** until decoded broker messages; unreachable/no endpoint stays
EMPTY and fail-open. Never invent PULSE. dual_write **OFF**. **not** Memory GA.
This optional host path is **not** E-G1.

### Local-only cite-both miss (`no_mesh_pulse`)

With no mesh receipts (local overlay only), cite-both must miss mesh.
That transcript is **`no_mesh_pulse`** (`missing=mesh`). Do **not** stamp
mesh on the local overlay to force cite-both. Catalog / grant /
`source=external` never satisfy cite-both. The kernel does **not**
classify miss classes.

```text
/memory ingest
# turn 1: PagerDuty page: webhook ingress 5xx
#         → provenance.source_hint=private  tag=source_hint:private
# turn 2: HMAC-verified delivery HTTP 200 is not a consume receipt
#         → provenance.source_hint=private  tag=source_hint:private
# turn 3: CreateConsumer 500 when consumers.mode is NULL
#         → provenance.source_hint=private  tag=source_hint:private

/memory digest --require-sources mesh,private
require-sources: miss · required=mesh,private · cited=private · missing=mesh · miss_class=no_mesh_pulse · receipt window newest-first · n=3 · mesh not in this receipt set · local palace on disk
```

That miss is **success** for the flag (no mesh-class receipt in a local-only
palace). `miss_class=` is a **host** digest token (TUI this wave); the
kernel does not emit it. This page is documentation of the commands, not a
live mesh session.

### Named miss classes (V2-A)

Digest can already cite-both or miss. V2-A **names** the miss. These are
host digest tokens (copy, **not** classifiers). The kernel does **not**
classify miss classes. Not a new kernel SoR.

**Classes:** `no_mesh_pulse` · `no_private_overlay` · `conflict` ·
`insufficient_signal` · `linked_pr_miss` · `public_vs_internal` ·
`no_memo` · `crm_only_restatement`.

Mapping already true on the host digest:

- `missing=mesh` → `no_mesh_pulse` (typical local overlay / kit)
- `missing=private` → `no_private_overlay`
- empty / rejected patterns → `insufficient_signal` (host hyphenated
  `insufficient-signal · nothing reliable today`)
- `conflict` / `linked_pr_miss` / `public_vs_internal` / `no_memo` /
  `crm_only_restatement` are **vocabulary**, not kernel-computed

Honesty:

- `linked_pr_miss` is a private eval column on the SRE recipe · **not**
  an MTTR claim
- `public_vs_internal` only if a public status page exists
- `crm_only_restatement` is overlay restating CRM without a pulse ·
  overlay does **not** GET Salesforce/CRM
- `no_memo` is a missing living memo (RevOps **V2-C** sitting; not this
  kernel page)
- do **not** stamp mesh on overlay

**Not** Memory GA (public MIT ≠ GA). **Not** overlay PULSE. leftover_is_bind
stays **OPEN**. **Not** V2-B (IngestTurn / SessionIDs / provenance).
**Not** V2-C RevOps. **not E-G1.** dual_write **OFF**.

Cost-max stays the same on the host path: hash embedder, no Qdrant, no cloud
palace. Optional Ollama is a TUI pin, not a kernel requirement.

## E-G1 is not this page

**E-G1** is a **real laptop** run: **PULSE + 3 RCA + cite-both-or-miss**.

This page does **not** satisfy E-G1. A unit test does **not** satisfy E-G1.
`go run ./examples/ttfh_rca` does **not** satisfy E-G1. Optional TUI/MCP
pins, `iomesh ttfh --unit` / `--live`, and slash-command names do **not**
satisfy E-G1.

Do not treat a docs PR, a README table row, or CI green as E-G1 closed.

## Notes

- dual_write **OFF** (host policy, not a kernel product flag)
- **not** Memory GA
- inspectable filesystem palace remains the source of truth
- one process per palace root (multi-process writers unsupported)
- LongMemEval is a different eval — see [`LONGMEMEVAL.md`](LONGMEMEVAL.md);
  a methodology card does not move TTFH / cite-both
