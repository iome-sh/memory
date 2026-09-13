# LongMemEval methodology card

Kernel-only · **not Memory GA** · dual_write **OFF**.

This page is the methodology card for any LongMemEval work on
`github.com/iome-sh/memory`. It exists so a number cannot ship without a
harness description. **No official score is published here** / README / site /
deck. A hash-overlap figure is an internal gate only — it is not official V1
and must not appear on the README, the site, or a deck.

Companion walking skeleton (consume-clock / TTFH / cite-both) is
[`examples/ttfh_rca`](../examples/ttfh_rca). A LongMemEval number does not move
those clocks. **Not Memory GA.**

Locked mixed **n=12** / **n=60** hash vs MiniLM vs BAAI BGE comparison (same
IDs, same judge): [`docs/LONGMEMEVAL_BASELINE.md`](LONGMEMEVAL_BASELINE.md).
Those tables are an **improvement baseline**, not official V1 and not a README
number. The first official V1 BGE mixed n=500 scored run is in **Internal run
log** below (**INTERNAL unpublished**). Reproduce twice before any public
figure. `KnightsAnalytics/bge-small-en-v1.5` is not a published Hugging Face
repo (auth 404). Comparable BGE ONNX is `BAAI/bge-small-en-v1.5` reshaped into
hugot layout (`make longmemeval-baseline`).

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

Reproduce twice before any public number. A published V1 figure, if it ever
ships, is labelled **“kernel retrieve + reader.”**

Do not vendor the 7 GB V2 tree. Do not publish judge-free hash-overlap as
accuracy. Do not claim we beat Mem0, Graphiti, or Letta on a vendor harness.

## Card fields (fill when a scored run exists)

Record these on every official-judge run. First official V1 BGE mixed-500 is
filled below. **INTERNAL unpublished.** Not a README number.

| Field | Value |
|-------|--------|
| Date (UTC) | generate 2026-09-13T03:55:05Z → 2026-09-13T07:22:29Z; judge done 2026-09-13T07:27:03Z |
| Kernel commit SHA | `9bee5428a6ff6d221ae57dca15da4a279ebca0f6` (`v1.5.12-20-g9bee542`) |
| Kernel tag | v1.5.12-20-g9bee542. `#139` (`e075fdd`) is docs/script only and does not change this score. |
| Dataset variant | gitignored `data/longmemeval_oracle.json` n=500 |
| Sample | `--sample mixed` **no `--limit`** (full file) |
| n (questions) | 500 complete (0 missing) |
| Type histogram | `temporal-reasoning` 133, `multi-session` 133, `knowledge-update` 78, `single-session-preference` 30, `single-session-assistant` 56, `single-session-user` 70 |
| Session scope | `session_id` = official `conv_id` on `/retrieve` |
| Embed mode | health `onnx-bge-small-en-v1.5`; all 500 hyp rows `embed_mode=onnx-bge-small-en-v1.5`; `BAAI/bge-small-en-v1.5` hugot layout (not KnightsAnalytics 404). PersistEmbeddings **OFF**. **Not hash.** |
| Reader | `scripts/longmemeval_qa_generate.py` + kernel `SearchMemoryWithOptions`; `OPENAI_MODEL=gpt-4o-mini` |
| Judge model pin | `gpt-4o-2024-08-06` via `evaluate_qa.py` (zoo key `gpt-4o`) |
| Judge script | upstream `third_party/LongMemEval/src/evaluation/evaluate_qa.py` |
| Qdrant | off |
| dual_write | OFF |
| Palace | isolated; port `:8781`; log `/tmp/lme-v1-official-run.log` |
| Status | **FIRST official V1 BGE mixed-500 · INTERNAL unpublished · not README · not Memory GA · reproduce twice before public** |

This is the **first** official V1 BGE mixed-500 scored run. **INTERNAL
unpublished.** Not a README number. Not Memory GA. Reproduce twice before any
public figure. Hash overlap unpublished. n=12 / n=60 cards stay improvement
baseline ≠ this V1 run. TTFH / cite-both is a different clock.

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

## Internal run log

Label: **unpublished · not hash-overlap**. This section is not a README number.
The first official V1 BGE mixed-500 card is recorded below (**INTERNAL
unpublished**). MiniLM n=12 is **not** official V1. n=12 / n=60 stay
improvement baseline ≠ this V1 run. **Not Memory GA.** TTFH / cite-both is a
different clock. Reproduce twice before any public figure.

### 2026-09-13 — official V1 BGE mixed n=500 (INTERNAL unpublished)

First official V1 scored run. Isolated palace; port `:8781`. Log
`/tmp/lme-v1-official-run.log`. Hypotheses gitignored:
`data/v1-official/hypotheses-bge.jsonl` (+ `.eval-results-gpt-4o`) — **do not
commit**. Hash overlap unpublished. n=12 / n=60 cards stay improvement
baseline ≠ this V1 run. **Not Memory GA.** TTFH / cite-both is a different
clock. **Not a README number.** Reproduce twice before any public figure.

| Field | Value |
|-------|--------|
| Date (UTC) | generate 2026-09-13T03:55:05Z → 2026-09-13T07:22:29Z; judge done 2026-09-13T07:27:03Z |
| Kernel commit SHA | `9bee5428a6ff6d221ae57dca15da4a279ebca0f6` (`v1.5.12-20-g9bee542`) |
| Kernel tag | v1.5.12-20-g9bee542. `#139` (`e075fdd`) is docs/script only and does not change this score. |
| Dataset variant | gitignored `data/longmemeval_oracle.json` n=500 |
| Sample | `--sample mixed` **no `--limit`** (full file) |
| n (questions) | 500 complete (0 missing) |
| Type histogram | `temporal-reasoning` 133, `multi-session` 133, `knowledge-update` 78, `single-session-preference` 30, `single-session-assistant` 56, `single-session-user` 70 |
| Session scope | `session_id` = official `conv_id` |
| Embed mode | health `onnx-bge-small-en-v1.5`; all 500 hyp rows `embed_mode=onnx-bge-small-en-v1.5`; `BAAI/bge-small-en-v1.5` hugot layout (not KnightsAnalytics 404) |
| PersistEmbeddings | OFF |
| Qdrant | off |
| dual_write | OFF |
| Reader | `scripts/longmemeval_qa_generate.py` + `SearchMemoryWithOptions`; `OPENAI_MODEL=gpt-4o-mini` |
| Judge model pin | `gpt-4o-2024-08-06` via `evaluate_qa.py` (zoo key `gpt-4o`) |
| Judge script | `scripts/longmemeval_judge.sh` → `third_party/LongMemEval/src/evaluation/evaluate_qa.py` |
| Palace | isolated; port `:8781` |
| Log | `/tmp/lme-v1-official-run.log` |
| Status | **FIRST official V1 BGE mixed-500 · INTERNAL unpublished · not README · not Memory GA · reproduce twice before public** |

| Metric | Value |
|--------|--------|
| Overall Accuracy | **0.776** = **388/500** |
| Task-averaged Accuracy | **0.802** |
| Abstention Accuracy | **0.6333 (30)** |

By type (judge-true):

| Type | Judge-true | Rate |
|------|------------|------|
| single-session-user | **68/70** | 0.9714 |
| single-session-assistant | **54/56** | 0.9643 |
| knowledge-update | **63/78** | 0.8077 |
| multi-session | **99/133** | 0.7444 |
| single-session-preference | **21/30** | 0.700 |
| temporal-reasoning | **83/133** | 0.6241 |

n=500 complete (0 missing). Overall **388/500**. **Not a README number.** **Not
Memory GA.**

### 2026-09-12 — official BGE-small-en-v1.5 ONNX retry (SKIP)

Retried `go run ./scripts/download_onnx_model.go` (hugot `DownloadModel` of `KnightsAnalytics/bge-small-en-v1.5`). Hugging Face returned **401** (`Invalid username or password`) for hugot and unauthenticated `curl`. Authenticated `curl` / `hf download` returned **404** (`Repository not found` / `Model not found`). Hugging Face model page 404; KnightsAnalytics org listing has MiniLM but no `bge-small-en-v1.5`. No BGE weights were written under `testdata/models/`. In-tree MiniLM was not deleted. Official V1 embed pin remains unmet. **No scored BGE card. No README number. Hash overlap unpublished. MiniLM ≠ BGE pin.**

| Field | Value |
|-------|--------|
| Date (UTC) | 2026-09-12T05:21:26Z |
| Kernel commit SHA | `e70343eab80e71c5157d6e34e3eaab39e98f73cf` (tag **v1.5.11**) |
| Kernel tag | v1.5.11 |
| Status | **SKIP** — BGE-small-en-v1.5 ONNX still unavailable (hugot **401**; HF **404**) |
| Embed mode | **not** official V1 BGE pin (download failed; MiniLM internal card below is a different embedder) |
| Judge model pin | `gpt-4o-2024-08-06` (not run — no BGE weights) |

### 2026-09-12 — mixed MiniLM ONNX generate+judge (INTERNAL unpublished)

Methodology proof, not a leaderboard. Oracle JSON stayed in gitignored `data/` (not committed). Hypotheses JSONL and eval-results stayed gitignored. LongMemEval-M (~2.7 GB) and V2 (~7 GB) were not vendored. Hash overlap unpublished. **Not BGE official pin. Not a README number.**

| Field | Value |
|-------|--------|
| Date (UTC) | 2026-09-12T05:05:50Z |
| Kernel commit SHA | `471dce551a2b81626f41f23463ad95ed98c755f8` (tag **v1.5.11**) |
| Kernel tag | v1.5.11 |
| Dataset variant | `longmemeval_oracle.json` **not committed** (gitignored `data/`, ~15 MB, 500 questions) |
| Sample | `mixed` (`LONGMEMEVAL_QA_SAMPLE=mixed`, stratified by `question_type`) — **not** prefix-n |
| n (questions) | 12 |
| Type histogram | `knowledge-update` 2, `multi-session` 2, `single-session-assistant` 2, `single-session-preference` 2, `single-session-user` 2, `temporal-reasoning` 2 |
| Session scope | `session_id` = official `conv_id` on `/retrieve` |
| Embed mode | **ONNX MiniLM-L6-v2** (384-d, in-tree `testdata/models/KnightsAnalytics_all-MiniLM-L6-v2`, hugot GoMLX). **Not hash. Not BGE-small-en-v1.5. Not official V1 embed pin.** BGE still unavailable (Hugging Face hugot download **401**; authenticated lookup **404** as of 2026-09-12 retry). |
| Reader | `scripts/longmemeval_qa_generate.py` + kernel `SearchMemoryWithOptions` (`cmd/longmemeval-server` `/retrieve`) |
| Judge model pin | `gpt-4o-2024-08-06` (upstream `evaluate_qa.py` zoo key `gpt-4o`; Makefile default `gpt-4o-mini` is **not** official V1) |
| Judge script | `scripts/longmemeval_judge.sh` → `third_party/LongMemEval/src/evaluation/evaluate_qa.py` |
| Qdrant | off |
| Status | scored official-judge mixed sample completed — **unpublished · not a README number** |

In-repo `testdata/longmemeval_oracle_subset.json` remains **3 `single-session-user` items** — that is **not** this mixed slice and **not** mixed official V1. Oracle JSON stays gitignored (`data/`). No accuracy number is published here or on the README.

## Notes

- Inspectable filesystem palace remains the source of truth.
- Hash embeddings must never be persisted as `QueryVec` / stored vectors.
- This card is not a fundraising exhibit.
- Host walking skeleton (TUI `/memory digest --require-sources mesh,private`)
  is cite-both of mesh pull + private palace — a different clock from this eval.
- `make longmemeval-v1-card` is optional and is **not** part of `make ci`.
- First official V1 BGE mixed n=500 is **INTERNAL unpublished** (388/500).
  Not a README number. Not Memory GA. Reproduce twice before public.
  n=12 / n=60 remain improvement baseline ≠ this V1 run.
