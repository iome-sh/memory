# LongMemEval locked mixed baseline (internal)

**Not a README number** · **not official V1**.

Same 12 question IDs. Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. Retrieve `session_id` = official `conv_id`. Isolated palace per embed mode.

Official V1 remains: BGE-small-en-v1.5 ONNX + mixed n=500 + this judge, reproduced twice.

## Wave C — T1 + count-query fact promotion (kernel `a25a883`, 2026-09-12T19:12Z)

| Embed | Judge-true | Rate | multi-session |
|-------|------------|------|---------------|
| hash | **11/12** | 0.917 | **1/2** (clothes pass, projects miss) |
| MiniLM ONNX | **10/12** | 0.833 | **0/2** |
| BAAI BGE ONNX | **11/12** | 0.917 | **1/2** (clothes pass, projects miss) |

Clothes (`0a995998`, gold 3) now counts boots pick-up + return + dry-cleaning on hash and BGE. Projects (`6d550036`, gold 2) still fails: reader sees no “led/leading” count. Other types 2/2. Hash temporal 2/2 this run.

## Wave B — T1 retrieve only (kernel `ef6a3e9`, 2026-09-12T18:52Z)

| Embed | Judge-true | multi-session |
|-------|------------|---------------|
| hash | 9/12 | 0/2 |
| MiniLM ONNX | 10/12 | 0/2 |
| BAAI BGE ONNX | 10/12 | 0/2 |

Retrieve covered all inner haystack sessions; judge still 0/2.

## Wave A — pre-T1 (kernel `cb93b08`, 2026-09-12T17:33Z)

| Embed | Judge-true | multi-session |
|-------|------------|---------------|
| hash | 9/12 | 0/2 |
| MiniLM ONNX | 10/12 | 0/2 |
| BAAI BGE ONNX | 10/12 | 0/2 |

## Protocol

| Field | Value |
|-------|--------|
| Slice | locked mixed n=12 (`testdata/longmemeval_baseline_ids.json`) |
| Histogram | 2 of each of 6 types |
| BGE source | `BAAI/bge-small-en-v1.5` hugot layout |

```bash
make longmemeval-baseline
```

## Notes

- hash-overlap unpublished · n=12 is not overall V1 · not a README number
- TTFH / cite-both walking skeleton is a different clock
