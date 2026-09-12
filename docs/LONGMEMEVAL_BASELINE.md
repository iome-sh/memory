# LongMemEval locked mixed baseline (internal)

**Not a README number** · **not official V1**.

Same 12 question IDs. Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. Retrieve `session_id` = official `conv_id`. Isolated palace per embed mode.

Official V1 remains: BGE-small-en-v1.5 ONNX + mixed n=500 + this judge, reproduced twice.

## Wave F — T1 clothing-errand clusters (kernel `aa64dbc` / #115 on #114, 2026-09-12T20:35Z–20:57Z)

Kernel `#115` (`AssembleCountEvidence` clusters dry-clean / return / pick-up) on `#114` (harness-only batch ONNX retrieve scoring). Health `embed_mode` verified. **Not a README number.**

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **2/2** (clothes pass, projects pass) |
| MiniLM ONNX | **12/12** | 1.000 | `onnx-minilm-l6-v2` | **2/2** |
| BAAI BGE ONNX | **12/12** | 1.000 | `onnx-bge-small-en-v1.5` | **2/2** |

Clothes (`0a995998`, gold 3) pass all three (return + pickup exchanged boots + dry-cleaning). Projects (`6d550036`, gold 2) pass all three. Hash miss is temporal `gpt4_2487a7cb` (event order inverted). Other types 2/2. n=12 clothes residual from Wave E is closed. Not official V1.

## n=60 scout — `origin/main` `1f5bb20` (kernel `f99c140` + Wave E docs, 2026-09-12T20:18Z–20:29Z)

Internal locked mixed **n=60** (`testdata/longmemeval_baseline_ids_n60.json`, 10 of each of 6 types). First 12 IDs = n=12 lock. Reader `gpt-4o-mini`. Judge `gpt-4o-2024-08-06`. Isolated palace. Health `embed_mode` verified. **BGE skipped** (too slow; parent runs BGE after kernel PRs). **Not a README number** · **not official V1**.

| Embed | Judge-true | Rate | health `embed_mode` | knowledge-update | multi-session | ss-assistant | ss-preference | ss-user | temporal |
|-------|------------|------|---------------------|------------------|---------------|--------------|---------------|---------|----------|
| hash | **48/60** | 0.800 | `hash` | **9/10** | **7/10** | **10/10** | **6/10** | **10/10** | **6/10** |
| MiniLM ONNX | **47/60** | 0.783 | `onnx-minilm-l6-v2` | **9/10** | **6/10** | **10/10** | **6/10** | **10/10** | **6/10** |
| BAAI BGE ONNX | SKIP | — | not run | — | — | — | — | — | — |

n=12 prefix **did not fully hold** vs Wave E (`f99c140`): Wave E hash **11/12** / MiniLM **12/12**. This scout hash **10/12** / MiniLM **10/12**. Extra miss `gpt4_2487a7cb` (event order inverted) on both; clothes `0a995998` gold 3 also miss on MiniLM this run (counted 2). Projects `6d550036` still pass. Reader/judge variance on the lock; do not treat Wave E MiniLM 12/12 as reproduced.

Miss IDs shared: `08f4fc43` `0a995998` `0edc2aef` `2a1811e2` `2c63a862` `35a27287` `852ce960` `aae3761f` `afdc33df` `caf03d32` `gpt4_2487a7cb` `gpt4_59c863d7`. MiniLM-only: `3a704032` (plants). Hash-only: none.

What n=60 breaks: incomplete multi-session counts (clothes 3, kits 5, driving hours, MiniLM plants 3); temporal date-diff / order; preference-conditioned answers 4/10 both; one stale knowledge-update (`852ce960` $350k vs $400k). ss-user and ss-assistant 10/10.

Artifacts (gitignored palaces): worktree `data/baseline-n60-scout/` (health, hypotheses JSONL, eval-results-gpt-4o, server logs). Summary `/tmp/lme-n60-scout-1f5bb20-summary.md`.

## Wave E — T1 count-assembly (kernel `f99c140` / #111, 2026-09-12T19:43Z–19:57Z)

Kernel `#111` (`AssembleCountEvidence` + palace-wide count-fact union). Health `embed_mode` verified.

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **11/12** | 0.917 | `hash` | **1/2** (clothes miss, projects pass) |
| MiniLM ONNX | **12/12** | 1.000 | `onnx-minilm-l6-v2` | **2/2** |
| BAAI BGE ONNX | **11/12** | 0.917 | `onnx-bge-small-en-v1.5` | **1/2** (clothes miss, projects pass) |

Clothes (`0a995998`, gold 3) pass on MiniLM (return + pickup exchanged boots + dry-cleaning). Hash counts pickup + dry-clean = 2; BGE counts pickup + return = 2. Projects (`6d550036`, gold 2) pass all three (Marketing Research led-team + Data Mining solo). Other types 2/2. T1 done-when (MiniLM **and** BGE multi-session no longer 0/2) **met**. Not a README number.

## Wave D — T1 overlap-rank + led extract (kernel `2695e02` / #110, 2026-09-12T19:25Z–19:36Z)

Kernel `#110` only. Does **not** include `f99c140` (#111 count-assembly).

| Embed | Judge-true | Rate | health `embed_mode` | multi-session |
|-------|------------|------|---------------------|---------------|
| hash | **9/12** | 0.750 | `hash` | **0/2** (clothes miss, projects miss) |
| MiniLM ONNX | **11/12** | 0.917 | `onnx-minilm-l6-v2` | **1/2** (clothes pass, projects miss) |
| BAAI BGE ONNX | **10/12** | 0.833 | `onnx-bge-small-en-v1.5` | **0/2** (clothes miss, projects miss) |

Clothes (`0a995998`, gold 3) pass on MiniLM only. Projects (`6d550036`, gold 2) fail all three. Hash also misses temporal `gpt4_2487a7cb`. T1 done-when (MiniLM **and** BGE multi-session no longer 0/2) **not met**. Not a README number.

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
- Wave F is kernel `aa64dbc` (#115 on #114). MiniLM/BGE 12/12 MS 2/2; hash 11/12 MS 2/2 (temporal `gpt4_2487a7cb`). Clothes residual closed. n=60 scout remains on kernel `f99c140`. Not a README number.
