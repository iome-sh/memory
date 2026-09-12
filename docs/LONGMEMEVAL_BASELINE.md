# LongMemEval locked mixed baseline (internal)

**Not a README number** · **not official V1**.

Same 12 question IDs. Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. Retrieve `session_id` = official `conv_id`. Isolated palace per embed mode.

Official V1 remains: BGE-small-en-v1.5 ONNX + mixed n=500 + this judge, reproduced twice.

## Wave D — T1 overlap-rank + led extract (kernel `2695e02` / #110, 2026-09-12T19:36:11Z)

Kernel `#110` only. Does **not** include `f99c140` (#111 count-assembly). Health `embed_mode` verified.

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **9/12** | 0.750 | `hash` | **0/2** (clothes miss, projects miss) |
| MiniLM ONNX | **11/12** | 0.917 | `onnx-minilm-l6-v2` | **1/2** (clothes pass, projects miss) |
| BAAI BGE ONNX | **10/12** | 0.833 | `onnx-bge-small-en-v1.5` | **0/2** (clothes miss, projects miss) |

Clothes (`0a995998`, gold 3): MiniLM counts boots pick-up + return + dry-cleaning; hash and BGE only boots pick-up + return (lost dry-cleaning vs Wave C). Projects (`6d550036`, gold 2) still miss all three. Hash temporal **1/2** (`gpt4_2487a7cb`). Versus Wave C (`a25a883`), overlap-rank (#110) helped MiniLM and dropped hash/BGE clothes. T1 done-when (MiniLM **and** BGE multi-session no longer 0/2) **not met**. Not a README number.

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
- Wave D is kernel `2695e02` (#110 only). Next: remesure n=12 on `f99c140` (#111 count-assembly, landed). Do not invent a Wave E row until that run exists. T1 done-when not met (MiniLM 1/2, BGE still 0/2). Not a README number.
