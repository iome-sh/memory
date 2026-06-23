#!/usr/bin/env bash
# Run official LongMemEval evaluate_qa.py against generated hypotheses.
# Usage: scripts/longmemeval_judge.sh gpt-4o-mini hypotheses.jsonl data/longmemeval_oracle.json
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <judge-model> <hypotheses.jsonl> <oracle.json>" >&2
  echo "example: $0 gpt-4o-mini hypotheses.jsonl data/longmemeval_oracle.json" >&2
  exit 2
fi

JUDGE_MODEL="$1"
HYPOTHESES="$2"
ORACLE="$3"

if [[ ! -f "${HYPOTHESES}" ]]; then
  echo "longmemeval-judge: hypotheses not found: ${HYPOTHESES}" >&2
  exit 1
fi
if [[ ! -f "${ORACLE}" ]]; then
  echo "longmemeval-judge: oracle dataset not found: ${ORACLE}" >&2
  exit 1
fi

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "longmemeval-judge: SKIP — OPENAI_API_KEY unset (judge requires OpenAI)" >&2
  exit 0
fi

bash "${ROOT}/scripts/clone_longmemeval_eval.sh"

EVAL_DIR="${ROOT}/third_party/LongMemEval/src/evaluation"
EVAL_SCRIPT="${EVAL_DIR}/evaluate_qa.py"
if [[ ! -f "${EVAL_SCRIPT}" ]]; then
  echo "longmemeval-judge: evaluate_qa.py not found at ${EVAL_SCRIPT}" >&2
  exit 1
fi

echo "longmemeval-judge: model=${JUDGE_MODEL} hypotheses=${HYPOTHESES} oracle=${ORACLE}"
(
  cd "${EVAL_DIR}"
  python3 evaluate_qa.py "${JUDGE_MODEL}" "${HYPOTHESES}" "${ORACLE}"
)

METRICS_SCRIPT="${EVAL_DIR}/print_qa_metrics.py"
if [[ -f "${METRICS_SCRIPT}" ]]; then
  LOG_FILE="${HYPOTHESES}.log"
  if [[ -f "${LOG_FILE}" ]]; then
    echo "longmemeval-judge: printing metrics from ${LOG_FILE}"
    (
      cd "${EVAL_DIR}"
      python3 print_qa_metrics.py "${JUDGE_MODEL}" "${LOG_FILE}" "${ORACLE}"
    )
  fi
fi