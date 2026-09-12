#!/usr/bin/env bash
# Official LongMemEval V1 mixed-run methodology card harness.
#
# Optional; not part of make ci. Missing oracle is SKIP (exit 0), not a CI failure.
#
# Official V1 = upstream evaluate_qa.py + judge gpt-4o-2024-08-06 + mixed sample
# + ONNX + session_id on retrieve + data/longmemeval_oracle.json.
#
# testdata/longmemeval_oracle_subset.json is 3 single-session-user items —
# that is not mixed official V1.
#
# Makefile default LONGMEMEVAL_JUDGE_MODEL=gpt-4o-mini is a cheap local path —
# not official V1. This card always records the official pin.
#
# Usage:
#   make longmemeval-v1-card
#   scripts/longmemeval_v1_card.sh
#   LONGMEMEVAL_V1_RUN=1 LONGMEMEVAL_QA_LIMIT=12 scripts/longmemeval_v1_card.sh
#
# Env:
#   LONGMEMEVAL_DATASET      default data/longmemeval_oracle.json
#   LONGMEMEVAL_QA_SAMPLE    default mixed (prefix is not official V1)
#   LONGMEMEVAL_QA_LIMIT     mixed n (0 or unset = full file for histogram)
#   LONGMEMEVAL_V1_CARD_OUT  optional write path (stdout always; default no file)
#   LONGMEMEVAL_V1_RUN       1 = generate+judge mixed sample (needs key + ONNX + server)
#   MEMORY_ONNX_MODEL_PATH   required for a scored official run (hash is not V1)
#
# Honesty: kernel-only · not Memory GA · dual_write OFF · no published score.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

log() { printf 'longmemeval-v1-card: %s\n' "$*" >&2; }

OFFICIAL_JUDGE="gpt-4o-2024-08-06"
CHEAP_JUDGE="gpt-4o-mini"
DATASET="${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json}"
SUBSET_REL="testdata/longmemeval_oracle_subset.json"
SAMPLE="${LONGMEMEVAL_QA_SAMPLE:-mixed}"
LIMIT="${LONGMEMEVAL_QA_LIMIT:-0}"
RUN="${LONGMEMEVAL_V1_RUN:-0}"
OUT="${LONGMEMEVAL_V1_CARD_OUT:-}"
HYPOTHESES="${LONGMEMEVAL_HYPOTHESES:-hypotheses.jsonl}"

SHA="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
TAG="$(git describe --tags --exact-match 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo untagged)"
DATE_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
DATE_DAY="$(date -u +%Y-%m-%d)"

ONNX_PATH="${MEMORY_ONNX_MODEL_PATH:-}"
if [[ -z "${ONNX_PATH}" && -d "${ROOT}/testdata/models/KnightsAnalytics_bge-small-en-v1.5" ]]; then
  ONNX_PATH="${ROOT}/testdata/models/KnightsAnalytics_bge-small-en-v1.5"
fi
if [[ -n "${ONNX_PATH}" ]]; then
  EMBED_MODE="ONNX (${ONNX_PATH})"
else
  EMBED_MODE="ONNX required for official V1 (hash is not official; MEMORY_ONNX_MODEL_PATH unset)"
fi

realpath_py() {
  python3 -c 'import os,sys; print(os.path.realpath(sys.argv[1]))' "$1"
}

subset_note() {
  cat <<'EOF'
in-repo testdata/longmemeval_oracle_subset.json is 3 single-session-user items — that is not mixed official V1.
Prefix-n on the official file is temporal-first, not overall V1. Hash overlap is not official V1.
EOF
}

emit_and_maybe_write() {
  local body="$1"
  printf '%s\n' "${body}"
  if [[ -n "${OUT}" ]]; then
    mkdir -p "$(dirname "${OUT}")"
    printf '%s\n' "${body}" >"${OUT}"
    log "wrote ${OUT}"
  fi
}

skip_card() {
  local reason="$1"
  log "SKIP — ${reason}"
  subset_note | while IFS= read -r line; do log "${line}"; done
  local body
  body="$(cat <<EOF
# LongMemEval V1 methodology card — SKIPPED

status: SKIP
reason: ${reason}
date_utc: ${DATE_UTC}
kernel_commit_sha: ${SHA}
kernel_tag: ${TAG}
dataset: ${DATASET}
in_repo_subset: ${SUBSET_REL}
subset_note: 3 single-session-user items — not mixed official V1
sample: mixed (required for official V1; not prefix-n)
embed_mode: ${EMBED_MODE}
judge_model_pin: ${OFFICIAL_JUDGE}
makefile_default_judge: ${CHEAP_JUDGE} (cheap local path — not official V1)
judge_script: third_party/LongMemEval/src/evaluation/evaluate_qa.py
session_scope: session_id = official conv_id on /retrieve
dual_write: OFF
product_claim: not Memory GA
label: INTERNAL unpublished · not Memory GA · not hash-overlap

Honesty: missing/non-official oracle is not a CI failure. Do not treat the in-repo
subset as official V1. Do not publish a number on README. Do not vendor LongMemEval-M
(~2.7 GB) or V2 (~7 GB).
EOF
)"
  emit_and_maybe_write "${body}"
  exit 0
}

if [[ ! -f "${DATASET}" ]]; then
  skip_card "official oracle missing at ${DATASET} (gitignored; not vendored)"
fi

DATASET_REAL="$(realpath_py "${DATASET}")"
SUBSET_REAL="$(realpath_py "${SUBSET_REL}")"
if [[ "${DATASET_REAL}" == "${SUBSET_REAL}" ]]; then
  skip_card "refusing ${SUBSET_REL} as official V1 (3 single-session-user items ≠ mixed V1)"
fi

# Histogram + official-shape check (stdlib json; uses scripts/longmemeval_sample.py).
SLICE_JSON="$(python3 - "${DATASET}" "${SAMPLE}" "${LIMIT}" "${SUBSET_REL}" <<'PY'
import json, os, sys

dataset, sample, limit_s, subset_rel = sys.argv[1:5]
sys.path.insert(0, "scripts")
from longmemeval_sample import apply_limit, type_histogram  # noqa: E402

with open(dataset, encoding="utf-8") as f:
    data = json.load(f)
if not isinstance(data, list):
    for key in ("examples", "data", "items", "questions"):
        if isinstance(data, dict) and key in data:
            data = data[key]
            break
if not isinstance(data, list):
    print(json.dumps({"ok": False, "reason": "oracle JSON is not a list of examples"}))
    raise SystemExit(0)

limit = int(limit_s or "0")
full_hist = type_histogram(data)
slice_ = apply_limit(data, limit, sample, warn=lambda m: print(m, file=sys.stderr))
hist = type_histogram(slice_)
types = sorted(hist)
reason = ""
ok = True
if os.path.realpath(dataset) == os.path.realpath(subset_rel):
    ok, reason = False, "in-repo subset is not official V1"
elif len(data) < 50:
    ok, reason = False, f"n={len(data)} looks like a subset, not longmemeval_oracle.json (500)"
elif len(full_hist) <= 1:
    ok, reason = False, f"single question_type {full_hist} — not mixed official V1"
elif "temporal-reasoning" not in full_hist or "single-session-user" not in full_hist:
    ok, reason = False, f"type set {types} is not the official V1 mix"
print(json.dumps({
    "ok": ok,
    "reason": reason,
    "n_file": len(data),
    "n_slice": len(slice_),
    "sample": sample,
    "limit": limit,
    "histogram_file": full_hist,
    "histogram_slice": hist,
}))
PY
)"

python3 -c 'import json,sys; json.loads(sys.argv[1])' "${SLICE_JSON}" >/dev/null

OK="$(python3 -c 'import json,sys; print("true" if json.loads(sys.argv[1]).get("ok") else "false")' "${SLICE_JSON}")"
if [[ "${OK}" != "true" ]]; then
  REASON="$(python3 -c 'import json,sys; print(json.loads(sys.argv[1]).get("reason") or "not official V1")' "${SLICE_JSON}")"
  skip_card "${REASON}"
fi

N_FILE="$(python3 -c 'import json,sys; print(json.loads(sys.argv[1])["n_file"])' "${SLICE_JSON}")"
N_SLICE="$(python3 -c 'import json,sys; print(json.loads(sys.argv[1])["n_slice"])' "${SLICE_JSON}")"
HIST_FILE="$(python3 -c 'import json,sys; print(json.dumps(json.loads(sys.argv[1])["histogram_file"], sort_keys=True))' "${SLICE_JSON}")"
HIST_SLICE="$(python3 -c 'import json,sys; print(json.dumps(json.loads(sys.argv[1])["histogram_slice"], sort_keys=True))' "${SLICE_JSON}")"

if [[ "${SAMPLE}" != "mixed" ]]; then
  log "warning: sample=${SAMPLE} is not mixed; official V1 requires --sample mixed (prefix-n is temporal-first)"
fi

SCORED="no"
SCORE_NOTE="methodology card only (no official-judge sample this invocation)"

if [[ "${RUN}" == "1" || "${RUN}" == "true" ]]; then
  log "LONGMEMEVAL_V1_RUN=1 — attempting mixed official-judge sample"
  if [[ -z "${OPENAI_API_KEY:-}" ]]; then
    log "scored run SKIP — OPENAI_API_KEY unset"
    SCORE_NOTE="scored run skipped: OPENAI_API_KEY unset"
  elif [[ -z "${ONNX_PATH}" ]]; then
    log "scored run SKIP — MEMORY_ONNX_MODEL_PATH unset (hash is not official V1)"
    SCORE_NOTE="scored run skipped: ONNX required (hash is not official V1)"
  elif ! curl -fsS "http://localhost:8765/health" >/dev/null 2>&1; then
    log "scored run SKIP — longmemeval-server not running at http://localhost:8765"
    log "start: MEMORY_ONNX_MODEL_PATH=${ONNX_PATH} go run ./cmd/longmemeval-server"
    SCORE_NOTE="scored run skipped: server not running at :8765"
  else
    HEALTH="$(curl -fsS "http://localhost:8765/health" || true)"
    SERVER_EMBED="$(python3 -c 'import json,sys; print(json.loads(sys.argv[1] or "{}").get("embed_mode") or "unknown")' "${HEALTH}")"
    log "server health embed_mode=${SERVER_EMBED}"
    if [[ "${SERVER_EMBED}" == "hash" ]]; then
      log "scored run SKIP — server embed_mode=hash (official V1 is ONNX)"
      SCORE_NOTE="scored run skipped: server embed_mode=hash"
    else
      RUN_LIMIT="${LIMIT}"
      if [[ "${RUN_LIMIT}" == "0" ]]; then
        RUN_LIMIT=12
        log "LONGMEMEVAL_QA_LIMIT unset/0; using mixed n=12 for scored sample (not full 500)"
      fi
      export MEMORY_ONNX_MODEL_PATH="${ONNX_PATH}"
      export LONGMEMEVAL_QA_SAMPLE="mixed"
      python3 scripts/longmemeval_qa_generate.py \
        --dataset "${DATASET}" \
        --output "${HYPOTHESES}" \
        --sample mixed \
        --limit "${RUN_LIMIT}" \
        --workers "${LONGMEMEVAL_QA_WORKERS:-2}"
      bash scripts/longmemeval_judge.sh "${OFFICIAL_JUDGE}" "${HYPOTHESES}" "${DATASET}"
      SCORED="yes"
      N_SLICE="${RUN_LIMIT}"
      HIST_SLICE="$(python3 - "${DATASET}" "${RUN_LIMIT}" <<'PY'
import json, os, sys
sys.path.insert(0, "scripts")
from longmemeval_sample import apply_limit, type_histogram
with open(sys.argv[1], encoding="utf-8") as f:
    data = json.load(f)
slice_ = apply_limit(data, int(sys.argv[2]), "mixed")
print(json.dumps(type_histogram(slice_), sort_keys=True))
PY
)"
      SCORE_NOTE="mixed official-judge sample completed (INTERNAL unpublished · not Memory GA · not hash-overlap · not a README number)"
      log "${SCORE_NOTE}"
    fi
  fi
fi

BODY="$(cat <<EOF
# LongMemEval V1 methodology card

status: OK
label: INTERNAL unpublished · not Memory GA · not hash-overlap
date_utc: ${DATE_UTC}
date: ${DATE_DAY}
kernel_commit_sha: ${SHA}
kernel_tag: ${TAG}
dataset: ${DATASET}
dataset_variant: LongMemEval oracle JSON (longmemeval_oracle.json)
n_file: ${N_FILE}
sample: ${SAMPLE} (stratified by question_type — not prefix-n)
n_slice: ${N_SLICE}
type_histogram_file: ${HIST_FILE}
type_histogram_slice: ${HIST_SLICE}
session_scope: session_id = official conv_id on /retrieve
embed_mode: ${EMBED_MODE}
reader: scripts/longmemeval_qa_generate.py + kernel SearchMemoryWithOptions
judge_model_pin: ${OFFICIAL_JUDGE}
makefile_default_judge: ${CHEAP_JUDGE} (cheap local path — not official V1)
judge_script: third_party/LongMemEval/src/evaluation/evaluate_qa.py
qdrant: off unless LONGMEMEVAL_QDRANT_URL is set (not required)
dual_write: OFF
product_claim: not Memory GA
scored_official_judge_sample: ${SCORED}
score_note: ${SCORE_NOTE}

Honesty: this card is not a README number, not Memory GA, and not a leaderboard submit.
in-repo ${SUBSET_REL} is 3 single-session-user items — not mixed official V1.
Do not vendor LongMemEval-M (~2.7 GB) or V2 (~7 GB).
EOF
)"

emit_and_maybe_write "${BODY}"
log "done (scored=${SCORED})"
exit 0
