# LongMemEval methodology card

This page describes how to run LongMemEval against `github.com/iome-sh/memory`
and what each kind of run means. It exists so that any figure is always
reported together with its harness description. **No score is published
here.** A hash-overlap figure is a local smoke check only; it is not
official V1 accuracy.

The companion walking skeleton is [`examples/ttfh_rca`](../examples/ttfh_rca).
A LongMemEval number does not change that example's timings.

`KnightsAnalytics/bge-small-en-v1.5` is not a published Hugging Face repo
(404). The comparable BGE ONNX model is `BAAI/bge-small-en-v1.5`, reshaped
into hugot layout (`make longmemeval-baseline`).

## What official means

| Label | What it is | What it is not |
|-------|------------|----------------|
| **In-repo overlap smoke** | `make longmemeval-smoke` / `longmemeval-recall-gate` / `longmemeval-bench`. Printed `aggregate recall` is **judge-free top-k gold-answer string overlap**. Default embedder is **hash**. | Official V1 QA accuracy. A leaderboard number. |
| **Official V1** | Upstream `evaluate_qa.py` with judge **`gpt-4o-2024-08-06`** against `longmemeval_oracle.json` (then S). Reader is the kernel retrieve path plus the generate script. `session_id` on retrieve. Embed mode **ONNX**, not hash. Mixed-type **full oracle (n=500)** (`--sample mixed`; do not pass `--limit`). | Prefix-n on the official file (temporal-first). Mixed n=12 (a sample, not V1). Hash overlap. A leaderboard number. |
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

A published V1 figure, if one ever ships, is labelled **“kernel retrieve +
reader”** and comes with the card fields below.

Do not vendor the 7 GB V2 tree. Do not publish judge-free hash-overlap as
accuracy. Do not claim we beat Mem0, Graphiti, or Letta on a vendor harness.

## Card fields (fill when a scored run is published)

Record these on every official-judge run:

| Field | Value |
|-------|--------|
| Date (UTC) | |
| Kernel commit SHA | |
| Kernel tag | |
| Dataset variant | e.g. `data/longmemeval_oracle.json` (gitignored) |
| Sample | `--sample mixed`, no `--limit` for the full file |
| n (questions) | |
| Type histogram | |
| Session scope | `session_id` = official `conv_id` on `/retrieve` |
| Embed mode | ONNX model name (not hash) |
| Reader | `scripts/longmemeval_qa_generate.py` + kernel `SearchMemoryWithOptions` |
| Judge model pin | `gpt-4o-2024-08-06` via `evaluate_qa.py` |
| Judge script | upstream `third_party/LongMemEval/src/evaluation/evaluate_qa.py` |
| Qdrant | on / off |
| Result | overall and per-type accuracy |

## How to run (operators)

```bash
# Offline overlap smoke (no OpenAI, hash default) — not official V1
make longmemeval-smoke
make longmemeval-recall-gate

# Methodology card (optional; not part of make ci).
# Missing data/longmemeval_oracle.json → SKIP exit 0 (not a CI failure).
# testdata/longmemeval_oracle_subset.json is 3 single-session-user items — not mixed official V1.
# Official judge pin is gpt-4o-2024-08-06. Makefile default gpt-4o-mini is a cheap local path — not official V1.
make longmemeval-v1-card

# Official V1 files (does not vendor LongMemEval-M ~2.7 GB or V2 ~7 GB).
# data/ is gitignored. Oracle JSON is ~15 MB; s_cleaned is ~277 MB.
make download-dataset   # → data/longmemeval_oracle.json

# Generate hypotheses against a running local harness
export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_bge-small-en-v1.5
# If BGE is unavailable (Hugging Face may 401 / 404; 2026-09-12 retry still SKIP), in-tree MiniLM is the local ONNX path:
#   testdata/models/KnightsAnalytics_all-MiniLM-L6-v2
# MiniLM is not the official V1 BGE-small-en-v1.5 embed pin.
go run ./cmd/longmemeval-server
make longmemeval-qa-generate LONGMEMEVAL_QA_SAMPLE=mixed
LONGMEMEVAL_JUDGE_MODEL=gpt-4o-2024-08-06 make longmemeval-judge
# scripts/longmemeval_judge.sh maps gpt-4o-2024-08-06 → upstream evaluate_qa.py zoo key gpt-4o.

# Official V1 scored run: mixed full oracle n=500 (do not set LONGMEMEVAL_QA_LIMIT).
LONGMEMEVAL_V1_RUN=1 make longmemeval-v1-card
# Mixed sample n=12 is not official V1:
LONGMEMEVAL_V1_RUN=1 LONGMEMEVAL_QA_LIMIT=12 make longmemeval-v1-card
```

`--limit N` on `scripts/longmemeval_qa_generate.py` is **dataset prefix order**.
Official V1 starts with `temporal-reasoning`, so a small n is not a mixed V1
score. Use `--sample mixed` (or `LONGMEMEVAL_QA_SAMPLE=mixed`) and print the
type histogram. `LONGMEMEVAL_V1_RUN=1` without `LONGMEMEVAL_QA_LIMIT` is the
full mixed oracle (**n=500**). `LONGMEMEVAL_QA_LIMIT=12` is a mixed sample,
**not** official V1.

`/retrieve` accepts `session_id`. Shared-palace QA without it is other-session
dominated.

Haystack dates accept official cleaned `2006/01/02 (Mon) 15:04` as well as
RFC3339.

## Notes

- The inspectable filesystem palace remains the source of truth.
- Hash embeddings must never be persisted as `QueryVec` / stored vectors.
- `make longmemeval-v1-card` is optional and is **not** part of `make ci`.
- Hypotheses, eval results and the oracle JSON stay in gitignored `data/`;
  do not commit them.
