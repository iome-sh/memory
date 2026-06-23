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
  CACHED="${ROOT}/testdata/models/KnightsAnalytics_all-MiniLM-L6-v2"
  if [[ -d "${CACHED}" ]]; then
    export MEMORY_ONNX_MODEL_PATH="${CACHED}"
  fi
fi

JSON_REPORT="${LONGMEMEVAL_JSON_REPORT:-}"
QUIET="${LONGMEMEVAL_QUIET:-}"

echo "longmemeval recall bench: dataset=${DATASET} topk=${TOPK} min-recall=${MIN_RECALL}"
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