# LongMemEval locked mixed baseline (internal)

**Not a README number** · **not official V1**.

Comparable columns on the **same 12 question IDs** (stratified mixed, 2 of each type).
Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. Retrieve `session_id` = official `conv_id`.
Isolated palace per embed mode (`LONGMEMEVAL_PALACE_ROOT`).

Official V1 remains: BGE-small-en-v1.5 ONNX + mixed sample of the **full** oracle (n=500) + this judge, reproduced twice.

## Scores — T1 retrieve (kernel `ef6a3e9`, 2026-09-12T18:52Z)

After T1 (`SessionIDs` / `conv:` grouping / session-diverse Limit + per-turn haystack `session_id` ingest). Health `embed_mode` verified per column.

| Embed | Judge-true | Rate | multi-session | temporal-reasoning |
|-------|------------|------|---------------|-------------------|
| hash | **9/12** | 0.750 | **0/2** | 1/2 |
| MiniLM ONNX | **10/12** | 0.833 | **0/2** | 2/2 |
| BAAI BGE ONNX | **10/12** | 0.833 | **0/2** | 2/2 |

Other types 2/2 on MiniLM and BGE. Hash also 2/2 except temporal 1/2.

**Retrieve check:** both multi-session items now pull **all inner haystack sessions** (3 and 4 session ids in the k=40 set). Pre-T1 retrieve was a single flattened `SessionID=conv_id`. Judge still 0/2: the reader does not assemble the count (`3` clothes, `2` projects) from chatter-heavy snippets. Next kernel slice: promote `turn_fact` / `fact_augmented` children on count queries.

## Scores — pre-T1 (kernel `cb93b08`, 2026-09-12T17:33Z)

| Embed | Judge-true | Rate | multi-session |
|-------|------------|------|---------------|
| hash | **9/12** | 0.750 | 0/2 |
| MiniLM ONNX | **10/12** | 0.833 | 0/2 |
| BAAI BGE ONNX | **10/12** | 0.833 | 0/2 |

Headline rates did not move. T1 changed retrieve coverage, not the judge score.

## Protocol

| Field | Value |
|-------|--------|
| Slice | locked mixed n=12 (`testdata/longmemeval_baseline_ids.json`) |
| Histogram | knowledge-update 2, multi-session 2, single-session-assistant 2, single-session-preference 2, single-session-user 2, temporal-reasoning 2 |
| Reader | `gpt-4o-mini` |
| Judge | `gpt-4o-2024-08-06` |
| Qdrant | off |
| BGE source | `BAAI/bge-small-en-v1.5` `onnx/model.onnx` reshaped to hugot layout |

```bash
make longmemeval-baseline
# next scale: LONGMEMEVAL_IDS_FILE=testdata/longmemeval_baseline_ids_n60.json make longmemeval-baseline
```

## Auth requirements (checked 2026-09-12)

| Resource | Auth | Live check |
|----------|------|------------|
| `BAAI/bge-small-en-v1.5` ONNX | **None.** Public. `HF_TOKEN` optional (rate limits). | API **200** with and without token. |
| `KnightsAnalytics/bge-small-en-v1.5` | N/A — **repo missing** | Valid token → **404**. Not a login miss. |
| MiniLM in-tree | None | No network if cached. |
| Oracle JSON | None | Local gitignored file. |
| Generate / judge | **`OPENAI_API_KEY` required** | Unset → generate fails; judge SKIP. |

## Notes

- hash-overlap unpublished
- n=12 is not overall V1 · do not put these rates on the README
- this table does not move the TTFH / cite-both walking skeleton
