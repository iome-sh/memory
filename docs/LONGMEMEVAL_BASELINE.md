# LongMemEval locked mixed baseline (internal)

Kernel-only · **not Memory GA** · dual_write **OFF** · **not a README number** · **not official V1**.

Comparable columns on the **same 12 question IDs** (stratified mixed, 2 of each type).
Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. `session_id` = official `conv_id`.
Isolated palace per embed mode (`LONGMEMEVAL_PALACE_ROOT`).

Official V1 remains: BGE-small-en-v1.5 ONNX + mixed sample of the **full** oracle (n=500) + this judge, reproduced twice.
This n=12 table is the **improvement baseline**, not that card.

| Field | Value |
|-------|--------|
| Date (UTC) | 2026-09-12T17:33:19Z |
| Kernel SHA | `cb93b086c310a7a45b5d25f854fe5e6fcedda242` |
| Kernel describe | v1.5.11-7-gcb93b08 |
| Slice | locked mixed n=12 (`testdata/longmemeval_baseline_ids.json`) |
| Histogram | knowledge-update 2, multi-session 2, single-session-assistant 2, single-session-preference 2, single-session-user 2, temporal-reasoning 2 |
| Reader | `gpt-4o-mini` (not the official V1 pin; official pin is the judge) |
| Judge | `gpt-4o-2024-08-06` |
| Qdrant | off |
| BGE source | `BAAI/bge-small-en-v1.5` `onnx/model.onnx` reshaped to hugot layout. `KnightsAnalytics/bge-small-en-v1.5` does not exist (HF 404). |
| Health `embed_mode` | hash → `hash` · MiniLM → `onnx-minilm-l6-v2` · BGE → `onnx-bge-small-en-v1.5` (isolated `LONGMEMEVAL_PALACE_ROOT` per mode) |

## Scores

| Embed | Judge-true | Rate | By type |
|-------|------------|------|---------|
| hash | **9/12** | 0.750 | knowledge-update 2/2, multi-session 0/2, single-session-assistant 2/2, single-session-preference 2/2, single-session-user 2/2, temporal-reasoning 1/2 |
| minilm | **10/12** | 0.833 | knowledge-update 2/2, multi-session 0/2, single-session-assistant 2/2, single-session-preference 2/2, single-session-user 2/2, temporal-reasoning 2/2 |
| bge | **10/12** | 0.833 | knowledge-update 2/2, multi-session 0/2, single-session-assistant 2/2, single-session-preference 2/2, single-session-user 2/2, temporal-reasoning 2/2 |

Hash is keyword-first retrieve (default embedder). MiniLM is in-tree ONNX, **not** the official V1 embed pin.
BGE is BAAI ONNX in hugot layout (root `model.onnx`).

**Improvement target on this ID list:** `multi-session` is **0/2** on every embedder. Hash also misses one `temporal-reasoning` (1/2) that MiniLM and BGE get. Next scale: same IDs protocol at n=60, then n=500 official-shaped mixed.

Reproduce (needs gitignored `data/longmemeval_oracle.json` + `OPENAI_API_KEY`; ONNX weights stay gitignored under `testdata/models/`):

```bash
make longmemeval-baseline
# next scale (not official V1): LONGMEMEVAL_IDS_FILE=testdata/longmemeval_baseline_ids_n60.json make longmemeval-baseline
```

## Auth requirements (checked 2026-09-12)

| Resource | Auth | Live check |
|----------|------|------------|
| `BAAI/bge-small-en-v1.5` ONNX | **None.** Public. `HF_TOKEN` optional (rate limits). | API **200** with and without token. `config.json` resolve **200** unauth. |
| `KnightsAnalytics/bge-small-en-v1.5` | N/A — **repo missing** | Unauth **401** (HF generic wall). **Valid token → 404 Repository not found.** Not a login miss. |
| `KnightsAnalytics/all-MiniLM-L6-v2` | None (public). In-tree copy needs no network. | API **200** with and without token. |
| Oracle `data/longmemeval_oracle.json` | None. Local gitignored file (~15 MB). | No network. |
| Generate (`gpt-4o-mini`) | **`OPENAI_API_KEY` required** | Unset → generate exits 1. |
| Judge (`gpt-4o-2024-08-06`) | **`OPENAI_API_KEY` required** | Unset → judge SKIP exit 0 (not a CI failure). |
| Invalid/expired `HF_TOKEN` | Must not block public BAAI | Helper retries **once without** `Authorization` on HTTP 401/403. 404 is not retried. |

`leftover_is_bind` / `IOMESH_*_WEBHOOK_SECRET` / YAML APPLY are **unrelated** to this kernel bench. dual_write OFF.

## Honesty

- leftover_is_bind OPEN · dual_write OFF · not Memory GA · hash-overlap unpublished
- n=12 is not overall V1 · prefix-n is not mixed · do not put these rates on the README
- E-G1 is a laptop TTFH clock; this table does not move it

