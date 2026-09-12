# LongMemEval methodology card

Kernel-only · **not Memory GA** · dual_write **OFF**.

This page is the methodology card for any LongMemEval work on
`github.com/iome-sh/memory`. It exists so a number cannot ship without a
harness description. **No official score is published here.** A hash-overlap
figure is an internal gate only — it is not official V1 and must not appear on
the README, the site, or a deck.

Companion walking skeleton (consume-clock / TTFH / cite-both) is
[`examples/ttfh_rca`](../examples/ttfh_rca). A LongMemEval number does not move
those clocks.

## What official means

| Label | What it is | What it is not |
|-------|------------|----------------|
| **In-repo overlap smoke** | `make longmemeval-smoke` / `longmemeval-recall-gate` / `longmemeval-bench`. Printed `aggregate recall` is **judge-free top-k gold-answer string overlap**. Default embedder is **hash**. | Official V1 QA accuracy. A leaderboard number. |
| **Official V1** | Upstream `evaluate_qa.py` with judge **`gpt-4o-2024-08-06`** against `longmemeval_oracle.json` (then S). Reader is the kernel retrieve path plus the generate script. `session_id` on retrieve. Embed mode **ONNX**, not hash. Mixed-type sample (`--sample mixed`). | Prefix-n on the official file (that slice is temporal-first). Hash overlap. Memory GA. |
| **Official V2** | Later, separate harness (LAFS). This kernel can load V2 file layout (`make longmemeval-v2-bench`) without vendoring the ~7 GB snapshot. | A substitute for V1. A published Gain figure from this repo. |

Makefile default `LONGMEMEVAL_JUDGE_MODEL` is **`gpt-4o-mini`** — a cheap local
path, **not** official V1. Set the official pin explicitly:

```bash
export LONGMEMEVAL_JUDGE_MODEL=gpt-4o-2024-08-06
export LONGMEMEVAL_QA_SAMPLE=mixed
export MEMORY_ONNX_MODEL_PATH=/path/to/bge-small-en-v1.5   # not hash
# session_id is passed by scripts/longmemeval_qa_generate.py (conv_id / question_id)
make longmemeval-qa-generate
make longmemeval-judge
```

Reproduce twice before any public number. A published V1 figure, if it ever
ships, is labelled **“kernel retrieve + reader, not Memory GA.”**

Do not vendor the 7 GB V2 tree. Do not publish judge-free hash-overlap as
accuracy. Do not claim we beat Mem0, Graphiti, or Letta on a vendor harness.

## Card fields (fill when a scored run exists)

Record these on every official-judge run. Leave values blank until the run is
real.

| Field | Value |
|-------|--------|
| Date (UTC) | — |
| Kernel commit SHA | — |
| Kernel tag | v1.5.10 (or the tag under test) |
| Dataset variant | LongMemEval-S / oracle JSON (`longmemeval_oracle.json`) |
| Sample | `mixed` (stratified by `question_type`) — **not** prefix-n |
| n (questions) | — |
| Type histogram | — (print from `--sample mixed`) |
| Session scope | `session_id` = official `conv_id` on `/retrieve` |
| Embed mode | ONNX (BGE-small-en-v1.5, 384-d). **Not hash.** |
| Reader | `scripts/longmemeval_qa_generate.py` + kernel `SearchMemoryWithOptions` |
| Judge model pin | `gpt-4o-2024-08-06` |
| Judge script | upstream `third_party/LongMemEval/src/evaluation/evaluate_qa.py` |
| Qdrant | off unless `LONGMEMEVAL_QDRANT_URL` is set (not required) |
| dual_write | OFF |
| Product claim | **not Memory GA** |

First scored official V1 mixed run is **internal**. Public number optional and
labelled. Hash overlap stays unpublished.

## How to run (operators)

```bash
# Offline overlap smoke (no OpenAI, hash default) — not official V1
make longmemeval-smoke
make longmemeval-recall-gate

# Official V1 files (does not vendor 7 GB)
make download-dataset   # → data/longmemeval_oracle.json

# Generate hypotheses against a running local harness
export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5
go run ./cmd/longmemeval-server
make longmemeval-qa-generate LONGMEMEVAL_QA_SAMPLE=mixed
LONGMEMEVAL_JUDGE_MODEL=gpt-4o-2024-08-06 make longmemeval-judge
```

`--limit N` on `scripts/longmemeval_qa_generate.py` is **dataset prefix order**.
Official V1 starts with `temporal-reasoning`, so a small n is not a mixed V1
score. Use `--sample mixed` (or `LONGMEMEVAL_QA_SAMPLE=mixed`) and print the
type histogram.

`/retrieve` accepts `session_id`. Shared-palace QA without it is other-session
dominated.

Haystack dates accept official cleaned `2006/01/02 (Mon) 15:04` as well as
RFC3339.

## Honesty

- Inspectable filesystem palace remains the source of truth.
- Hash embeddings must never be persisted as `QueryVec` / stored vectors.
- This card is not a Memory GA announcement and not a seed-deck exhibit.
- Host walking skeleton (TUI `/memory digest --require-sources mesh,private`)
  is cite-both of mesh pull + private palace — a different clock from this eval.
