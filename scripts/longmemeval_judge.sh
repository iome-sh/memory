#!/usr/bin/env bash
# Run official LongMemEval evaluate_qa.py against generated hypotheses.
# Usage: scripts/longmemeval_judge.sh gpt-4o-mini hypotheses.jsonl data/longmemeval_oracle.json
#
# Official V1 pin is gpt-4o-2024-08-06. Upstream evaluate_qa.py model_zoo keys are
# gpt-4o (→ gpt-4o-2024-08-06) and gpt-4o-mini (→ gpt-4o-mini-2024-07-18).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <judge-model> <hypotheses.jsonl> <oracle.json>" >&2
  echo "example: $0 gpt-4o-mini hypotheses.jsonl data/longmemeval_oracle.json" >&2
  echo "official V1 pin: $0 gpt-4o-2024-08-06 hypotheses.jsonl data/longmemeval_oracle.json" >&2
  exit 2
fi

JUDGE_MODEL="$1"
HYPOTHESES="$2"
ORACLE="$3"

realpath_py() {
  python3 -c 'import os,sys; print(os.path.realpath(sys.argv[1]))' "$1"
}

if [[ ! -f "${HYPOTHESES}" ]]; then
  echo "longmemeval-judge: hypotheses not found: ${HYPOTHESES}" >&2
  exit 1
fi
if [[ ! -f "${ORACLE}" ]]; then
  echo "longmemeval-judge: oracle dataset not found: ${ORACLE}" >&2
  exit 1
fi

HYPOTHESES="$(realpath_py "${HYPOTHESES}")"
ORACLE="$(realpath_py "${ORACLE}")"

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "longmemeval-judge: SKIP — OPENAI_API_KEY unset (judge requires OpenAI)" >&2
  exit 0
fi

# Map official pin / aliases onto upstream model_zoo keys.
ZOO_KEY="${JUDGE_MODEL}"
case "${JUDGE_MODEL}" in
  gpt-4o-2024-08-06|gpt-4o) ZOO_KEY="gpt-4o" ;;
  gpt-4o-mini|gpt-4o-mini-2024-07-18) ZOO_KEY="gpt-4o-mini" ;;
esac

bash "${ROOT}/scripts/clone_longmemeval_eval.sh"

EVAL_DIR="${ROOT}/third_party/LongMemEval/src/evaluation"
EVAL_SCRIPT="${EVAL_DIR}/evaluate_qa.py"
if [[ ! -f "${EVAL_SCRIPT}" ]]; then
  echo "longmemeval-judge: evaluate_qa.py not found at ${EVAL_SCRIPT}" >&2
  exit 1
fi

echo "longmemeval-judge: model=${JUDGE_MODEL} zoo_key=${ZOO_KEY} hypotheses=${HYPOTHESES} oracle=${ORACLE}"
(
  cd "${EVAL_DIR}"
  python3 evaluate_qa.py "${ZOO_KEY}" "${HYPOTHESES}" "${ORACLE}"
)

METRICS_SCRIPT="${EVAL_DIR}/print_qa_metrics.py"
RESULT_FILE="${HYPOTHESES}.eval-results-${ZOO_KEY}"
if [[ -f "${METRICS_SCRIPT}" && -f "${RESULT_FILE}" ]]; then
  echo "longmemeval-judge: printing metrics from ${RESULT_FILE}"
  (
    cd "${EVAL_DIR}"
    python3 print_qa_metrics.py "${RESULT_FILE}" "${ORACLE}"
  )
fi