#!/usr/bin/env bash
# Shallow clone LongMemEval evaluation scripts into third_party/LongMemEval.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="${ROOT}/third_party/LongMemEval"
REPO="https://github.com/xiaowu0162/LongMemEval.git"

if [[ -d "${DEST}/.git" ]]; then
  echo "longmemeval-clone: already present at ${DEST}"
  exit 0
fi

mkdir -p "${ROOT}/third_party"
echo "longmemeval-clone: shallow cloning ${REPO} -> ${DEST}"
git clone --depth 1 "${REPO}" "${DEST}"
echo "longmemeval-clone: done"