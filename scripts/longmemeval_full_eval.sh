#!/usr/bin/env bash
# Full LongMemEval pipeline: dataset download, offline recall bench, optional QA+judge.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

log() { echo "longmemeval-full-eval: $*"; }

DATASET="${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json}"
HYPOTHESES="${LONGMEMEVAL_HYPOTHESES:-hypotheses.jsonl}"
JUDGE_MODEL="${LONGMEMEVAL_JUDGE_MODEL:-gpt-4o-mini}"
QA_LIMIT="${LONGMEMEVAL_QA_LIMIT:-0}"
QA_WORKERS="${LONGMEMEVAL_QA_WORKERS:-4}"

log "phase 1 — download dataset if missing"
bash scripts/download_longmemeval_dataset.sh

log "phase 2 — offline ONNX recall bench (full oracle split; ~500q, ~3-5 min — progress on stderr)"
LONGMEMEVAL_DATASET="${DATASET}" LONGMEMEVAL_LIMIT="${LONGMEMEVAL_BENCH_LIMIT:-0}" \
  bash scripts/longmemeval_recall_bench.sh

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  log "phase 3 — SKIP QA+judge (OPENAI_API_KEY unset)"
  log "to run full 500-q eval with OpenAI judge:"
  log "  export OPENAI_API_KEY=sk-..."
  log "  export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_all-MiniLM-L6-v2"
  log "  go run cmd/longmemeval-server/main.go &"
  log "  make longmemeval-qa-generate LONGMEMEVAL_QA_LIMIT=500"
  log "  make longmemeval-judge"
  exit 0
fi

log "phase 3 — QA generation + official judge (requires running server)"
if ! curl -fsS "http://localhost:8765/health" >/dev/null 2>&1; then
  log "server not running at http://localhost:8765 — start it first:"
  log "  export MEMORY_ONNX_MODEL_PATH=testdata/models/KnightsAnalytics_all-MiniLM-L6-v2"
  log "  go run cmd/longmemeval-server/main.go"
  log "then: make longmemeval-qa-generate && make longmemeval-judge"
  exit 0
fi

LIMIT_ARG=()
if [[ "${QA_LIMIT}" != "0" ]]; then
  LIMIT_ARG=(--limit "${QA_LIMIT}")
fi

python3 scripts/longmemeval_qa_generate.py \
  --dataset "${DATASET}" \
  --output "${HYPOTHESES}" \
  --workers "${QA_WORKERS}" \
  "${LIMIT_ARG[@]}"

bash scripts/longmemeval_judge.sh "${JUDGE_MODEL}" "${HYPOTHESES}" "${DATASET}"
log "done"