#!/usr/bin/env bash
# longmemeval_server_boot_smoke.sh — boot a built longmemeval-server and
# require GET /health to return 200 within the timeout.
#
# Hash embeddings only (MEMORY_ONNX_MODEL_PATH unset), no Qdrant, temp palace
# root. Exits non-zero if the binary exits early or /health is not 200 in time.
#
# Usage:
#   go build -o /tmp/longmemeval-server ./cmd/longmemeval-server
#   bash scripts/longmemeval_server_boot_smoke.sh /tmp/longmemeval-server
#
# Env:
#   BOOT_SMOKE_ADDR     listen addr (default 127.0.0.1:18765)
#   BOOT_SMOKE_TIMEOUT  seconds to wait for /health (default 30)
set -euo pipefail

BIN="${1:?usage: $0 <path-to-longmemeval-server>}"
ADDR="${BOOT_SMOKE_ADDR:-127.0.0.1:18765}"
TIMEOUT="${BOOT_SMOKE_TIMEOUT:-30}"
PALACE="$(mktemp -d)"
LOG="${PALACE}/server.log"

env -u MEMORY_ONNX_MODEL_PATH -u LONGMEMEVAL_QDRANT_URL \
  LONGMEMEVAL_ADDR="$ADDR" LONGMEMEVAL_PALACE_ROOT="$PALACE" \
  "$BIN" >"$LOG" 2>&1 &
PID=$!

cleanup() {
  kill "$PID" 2>/dev/null || true
  wait "$PID" 2>/dev/null || true
  rm -rf "$PALACE"
}
trap cleanup EXIT

fail() {
  echo "::error::$1"
  echo "--- server log ---"
  cat "$LOG" || true
  exit 1
}

deadline=$((SECONDS + TIMEOUT))
code=000
while [ "$SECONDS" -lt "$deadline" ]; do
  if ! kill -0 "$PID" 2>/dev/null; then
    fail "longmemeval-server exited before /health returned 200"
  fi
  code="$(curl -s -o "${PALACE}/health.json" -w '%{http_code}' "http://${ADDR}/health" || true)"
  if [ "$code" = "200" ]; then
    echo "GET /health -> 200: $(cat "${PALACE}/health.json")"
    exit 0
  fi
  sleep 1
done
fail "GET /health did not return 200 within ${TIMEOUT}s (last status ${code})"
