#!/usr/bin/env bash
# Offline LongMemEval recall benchmark (no OpenAI).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

DATASET="${LONGMEMEVAL_DATASET:-testdata/longmemeval_oracle_subset.json}"
LIMIT="${LONGMEMEVAL_LIMIT:-0}"
TOPK="${LONGMEMEVAL_TOPK:-5}"
MIN_RECALL="${LONGMEMEVAL_MIN_RECALL:-0.6}"

if [[ -z "${MEMORY_ONNX_MODEL_PATH:-}" ]]; then
  CACHED="${ROOT}/testdata/models/KnightsAnalytics_bge-small-en-v1.5"
  if [[ -d "${CACHED}" ]]; then
    export MEMORY_ONNX_MODEL_PATH="${CACHED}"
  fi
fi

JSON_REPORT="${LONGMEMEVAL_JSON_REPORT:-}"
QUIET="${LONGMEMEVAL_QUIET:-}"

echo "longmemeval recall bench: dataset=${DATASET} topk=${TOPK} min-recall=${MIN_RECALL}" >&2
echo "longmemeval recall bench: printed recall = top-k gold-answer string overlap (judge-free). Not official V1 gpt-4o QA. Not V2 LAFS." >&2
if [[ "${LIMIT}" == "0" && "${DATASET}" == *oracle* ]]; then
  echo "longmemeval recall bench: full oracle run (~500 questions, ~3-5 min ONNX); progress on stderr" >&2
fi
ARGS=(
  -dataset "${DATASET}"
  -limit "${LIMIT}"
  -topk "${TOPK}"
  -min-recall "${MIN_RECALL}"
)
if [[ -n "${JSON_REPORT}" ]]; then
  ARGS+=(-json-report "${JSON_REPORT}")
fi
if [[ "${QUIET}" == "1" || "${QUIET}" == "true" ]]; then
  ARGS+=(-quiet)
fi
go run ./cmd/longmemeval-bench "${ARGS[@]}"