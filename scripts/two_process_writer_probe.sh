#!/usr/bin/env bash
# two_process_writer_probe.sh — last-write-wins evidence on one palace root.
#
# Two concurrent compiled writers share BaseDir: distinct entry IDs plus one
# colliding ID. Prints last-write-wins and whether relations/entity-graph.json
# and indexes/event-time.json remain valid JSON.
#
# Always exits 0 (probe, not a CI gate). Multi-process writers remain
# unsupported. This does not invent tenancy. Probe ≠ flock · flock is not
# shipped · probe ≠ Memory GA · dual_write OFF.
#
# Usage:
#   bash scripts/two_process_writer_probe.sh
#   MEMORY_TWO_PROCESS_PROBE=1 go test -count=1 -run '^TestTwoProcessWriterProbe$'
#
# Env:
#   MEMORY_TWO_PROCESS_PROBE_N  unique writes per writer (default 24)
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

log() { printf 'two-process-writer-probe: %s\n' "$*" >&2; }

log "multi-process writers remain unsupported; this does not invent tenancy"
log "probe ≠ flock · flock is not shipped · probe ≠ Memory GA · dual_write OFF"

N="${MEMORY_TWO_PROCESS_PROBE_N:-24}"
WORKDIR="$(mktemp -d "${TMPDIR:-/tmp}/memory-two-process-probe.XXXXXX")" || {
  log "mktemp failed (probe, not a CI fail)"
  exit 0
}
cleanup() { rm -rf "${WORKDIR}"; }
trap cleanup EXIT

BIN="${WORKDIR}/two-process-writer-probe"
PALACE="${WORKDIR}/palace"
mkdir -p "${PALACE}" || {
  log "mkdir palace failed (probe, not a CI fail)"
  exit 0
}

if ! go build -o "${BIN}" ./cmd/two-process-writer-probe; then
  log "go build failed (probe, not a CI fail)"
  exit 0
fi

status_a=0
status_b=0
"${BIN}" -base "${PALACE}" -writer A -n "${N}" &
pid_a=$!
"${BIN}" -base "${PALACE}" -writer B -n "${N}" &
pid_b=$!
wait "${pid_a}" || status_a=$?
wait "${pid_b}" || status_b=$?

log "writer A pid=${pid_a} exit=${status_a}"
log "writer B pid=${pid_b} exit=${status_b}"

"${BIN}" -inspect -base "${PALACE}" -n "${N}" || log "inspect exited $? (probe continues)"

log "done; always exit 0. Single-writer contract unchanged."
exit 0
