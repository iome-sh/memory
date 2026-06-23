#!/usr/bin/env bash
# Download official LongMemEval cleaned datasets into data/.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${ROOT}/data"
mkdir -p "${DATA_DIR}"

BASE_URL="https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/resolve/main"

download() {
  local name="$1"
  local dest="${DATA_DIR}/${name}"
  if [[ -f "${dest}" ]]; then
    echo "already present: ${dest}"
    return 0
  fi
  echo "downloading ${name} ..."
  if command -v wget >/dev/null 2>&1; then
    wget -q --show-progress -O "${dest}" "${BASE_URL}/${name}"
  elif command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "${dest}" "${BASE_URL}/${name}"
  else
    echo "error: need wget or curl" >&2
    exit 1
  fi
}

download "longmemeval_oracle.json"
download "longmemeval_s_cleaned.json"

echo "datasets ready under ${DATA_DIR}/"