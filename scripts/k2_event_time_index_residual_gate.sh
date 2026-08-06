#!/usr/bin/env bash
# k2_event_time_index_residual_gate.sh — s1303 offline residual
# free eng residual pin s1303 · free eng concurrent s1301+ after free-floor s1299 · lag s1300
# SSOT for K2 ListMemoryWithOptions / timeline path residual honesty:
#   filters before Limit shipped (underfill class) · O(n) FS scan residual-honest
#   full event-time index residual · not shipped · residual ≠ invent index green
#   host memory_timeline uses list path · TUI s1296 / concurrent s1301 mention only
# Honesty: kernel ≠ product Memory GA · not Memory GA · dual_write OFF · no invent GA ·
#   residual PASS ≠ live dogfood · RESULT PASS / RESULT OK honesty chain
# Soft skip: SKIP_K2_EVENT_TIME_INDEX=1
#
# Usage:
#   ./scripts/k2_event_time_index_residual_gate.sh
#   make residual-gate
#   make k2-event-time-index-residual-gate
#   SKIP_K2_EVENT_TIME_INDEX=1 ./scripts/k2_event_time_index_residual_gate.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FAIL=0
WARN=0

log() { printf 'k2-event-time-index-residual: %s\n' "$*" >&2; }
pass() { log "PASS: $*"; }
warn() { log "WARN: $*"; WARN=$((WARN + 1)); }
fail() { log "FAIL: $*"; FAIL=$((FAIL + 1)); }

if [[ "${SKIP_K2_EVENT_TIME_INDEX:-}" == "1" ]]; then
  log "SKIP_K2_EVENT_TIME_INDEX=1 soft skip"
  log "RESULT OK (soft skip)"
  log "non-claim: soft skip ≠ invent Memory GA · residual ≠ invent index green · residual PASS ≠ live dogfood"
  exit 0
fi

log "offline residual SSOT for K2 ListMemoryWithOptions / event-time index residual (no dogfood / Memory GA invent required)"
log "non-claim: full event-time index residual · residual ≠ invent index green · kernel ≠ product Memory GA · not Memory GA"

DOC="docs/operations/k2-event-time-index-residual.md"
MAKEFILE="Makefile"
GATE_SCRIPT="scripts/k2_event_time_index_residual_gate.sh"

need_grep() {
  local file="$1" needle="$2" label="$3"
  if [[ ! -f "$file" ]]; then
    fail "skip needle ${label} (missing ${file})"
    return
  fi
  if grep -qF "$needle" "$file"; then
    pass "needle ${label}"
  else
    fail "missing needle ${label} in ${file}"
  fi
}

need_path() {
  local path="$1" label="$2"
  if [[ -e "$path" ]]; then
    pass "path ${label}"
  else
    fail "missing path ${label} (${path})"
  fi
}

need_tree_grep() {
  local needle="$1" label="$2"
  if grep -R -n --include='*.go' -F "$needle" . >/dev/null 2>&1; then
    pass "tree ${label}"
  else
    fail "tree missing ${label} (${needle})"
  fi
}

# Residual doc presence + honesty needles
need_path "$DOC" "k2 event-time index residual SSOT"
need_path "$GATE_SCRIPT" "k2 event-time index residual gate script"
need_grep "$DOC" 's1303' "doc free eng residual pin s1303"
need_grep "$DOC" 's1301+' "doc free eng concurrent s1301+"
need_grep "$DOC" 's1299' "doc free-floor s1299"
need_grep "$DOC" 's1300' "doc lag residual s1300"
need_grep "$DOC" 's1296' "doc concurrent TUI s1296 mention"
need_grep "$DOC" 's1297' "doc peer s1297 inventory residual"
need_grep "$DOC" 's611' "doc kernel peer s611"
need_grep "$DOC" 'O(n)' "doc O(n) FS scan residual class"
need_grep "$DOC" 'event-time index residual' "doc event-time index residual"
need_grep "$DOC" 'ListMemoryWithOptions' "doc ListMemoryWithOptions"
need_grep "$DOC" 'ListMemoryOptions' "doc ListMemoryOptions"
need_grep "$DOC" 'filters before limit' "doc filters before limit (shipped)"
need_grep "$DOC" 'Memory GA' "doc Memory GA honesty"
need_grep "$DOC" 'not Memory GA' "doc not Memory GA"
need_grep "$DOC" 'kernel ≠ product Memory GA' "doc kernel ≠ product Memory GA"
need_grep "$DOC" 'residual ≠ invent index green' "doc residual ≠ invent index green"
need_grep "$DOC" 'dual_write OFF' "doc dual_write OFF"
need_grep "$DOC" 'no invent GA' "doc no invent GA"
need_grep "$DOC" 'PASS ≠ live dogfood' "doc PASS ≠ live dogfood"
need_grep "$DOC" 'memory_timeline' "doc host peer memory_timeline mention"
need_grep "$DOC" 'SKIP_K2_EVENT_TIME_INDEX' "doc soft skip env"
need_grep "$DOC" 'residual-gate' "doc Makefile residual-gate target"
need_grep "$DOC" 'k2-event-time-index-residual-gate' "doc Makefile k2 residual-gate target"
need_grep "$DOC" 'RESULT PASS' "doc RESULT PASS honesty chain"

# Kernel tree symbols (Go)
need_tree_grep 'ListMemoryWithOptions' "ListMemoryWithOptions"
need_tree_grep 'ListMemoryOptions' "ListMemoryOptions"

# Makefile target when present
if [[ -f "$MAKEFILE" ]]; then
  if grep -qE 'residual-gate' "$MAKEFILE" 2>/dev/null; then
    pass "needle Makefile residual-gate target"
  else
    fail "Makefile missing residual-gate target"
  fi
  if grep -qE 'k2-event-time-index-residual-gate|k2_event_time_index_residual_gate' "$MAKEFILE" 2>/dev/null; then
    pass "needle Makefile k2-event-time-index residual gate"
  else
    fail "Makefile missing k2-event-time-index residual gate target"
  fi
else
  warn "no Makefile (document bash path only)"
fi

log "WARN count=${WARN} FAIL count=${FAIL}"
if [[ "$FAIL" -gt 0 ]]; then
  log "RESULT FAIL"
  exit 1
fi
log "RESULT PASS"
log "RESULT OK honesty chain: ListMemoryWithOptions filters before limit shipped · O(n) FS scan · event-time index residual · residual ≠ invent index green · kernel ≠ product Memory GA · dual_write OFF · no invent GA · residual PASS ≠ live dogfood"
exit 0
