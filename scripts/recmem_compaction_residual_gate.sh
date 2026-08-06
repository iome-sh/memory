#!/usr/bin/env bash
# recmem_compaction_residual_gate.sh — s1313 offline residual
# free eng residual pin s1313 · free eng concurrent s1311+ after free-floor s1309 · lag s1310
# SSOT for RecMem / compaction kernel path residual honesty:
#   AutoRecMemCompaction / PerformCompaction / CompactionConfig exist as kernel primitives
#   AutoRecMemCompaction shipped partial · host memory_trigger_compact advisory (publish trigger)
#   TUI requires HITL (peer s1311) · Compaction PASS ≠ invent Memory GA token-reduction
#   Phase 3 semantic refine residual if still planned · dual_write OFF host concern
# Honesty: kernel ≠ product Memory GA · not Memory GA · dual_write OFF · no invent GA ·
#   residual PASS ≠ live dogfood · RESULT PASS / RESULT OK honesty chain
# Soft skip: SKIP_RECMEM_COMPACTION=1
#
# Usage:
#   ./scripts/recmem_compaction_residual_gate.sh
#   make residual-gate
#   make recmem-compaction-residual-gate
#   SKIP_RECMEM_COMPACTION=1 ./scripts/recmem_compaction_residual_gate.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FAIL=0
WARN=0

log() { printf 'recmem-compaction-residual: %s\n' "$*" >&2; }
pass() { log "PASS: $*"; }
warn() { log "WARN: $*"; WARN=$((WARN + 1)); }
fail() { log "FAIL: $*"; FAIL=$((FAIL + 1)); }

if [[ "${SKIP_RECMEM_COMPACTION:-}" == "1" ]]; then
  log "SKIP_RECMEM_COMPACTION=1 soft skip"
  log "RESULT OK (soft skip)"
  log "non-claim: soft skip ≠ invent Memory GA · Compaction PASS ≠ invent Memory GA token-reduction · residual PASS ≠ live dogfood"
  exit 0
fi

log "offline residual SSOT for RecMem / compaction kernel path (no dogfood / Memory GA invent required)"
log "non-claim: AutoRecMemCompaction shipped partial · trigger advisory host · HITL on TUI · Compaction PASS ≠ invent Memory GA token-reduction · kernel ≠ product Memory GA · not Memory GA"

DOC="docs/operations/recmem-compaction-residual.md"
PLAN="docs/recmem-integration-plan.md"
MAKEFILE="Makefile"
GATE_SCRIPT="scripts/recmem_compaction_residual_gate.sh"

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
need_path "$DOC" "recmem compaction residual SSOT"
need_path "$GATE_SCRIPT" "recmem compaction residual gate script"
need_path "$PLAN" "recmem integration plan (historical)"
need_grep "$DOC" 's1313' "doc free eng residual pin s1313"
need_grep "$DOC" 's1311+' "doc free eng concurrent s1311+"
need_grep "$DOC" 's1309' "doc free-floor s1309"
need_grep "$DOC" 's1310' "doc lag residual s1310"
need_grep "$DOC" 's1311' "doc concurrent TUI s1311 HITL mention"
need_grep "$DOC" 's1297' "doc peer s1297 inventory residual"
need_grep "$DOC" 's1303' "doc peer s1303 K2 residual"
need_grep "$DOC" 'AutoRecMemCompaction' "doc AutoRecMemCompaction"
need_grep "$DOC" 'PerformCompaction' "doc PerformCompaction"
need_grep "$DOC" 'CompactionConfig' "doc CompactionConfig"
need_grep "$DOC" 'shipped partial' "doc AutoRecMemCompaction shipped partial"
need_grep "$DOC" 'memory_trigger_compact' "doc host memory_trigger_compact"
need_grep "$DOC" 'advisory' "doc trigger advisory"
need_grep "$DOC" 'HITL' "doc TUI HITL"
need_grep "$DOC" 'Phase 3' "doc Phase 3 semantic refine residual"
need_grep "$DOC" 'semantic refine' "doc semantic refine residual class"
need_grep "$DOC" 'token-reduction' "doc token-reduction non-claim"
need_grep "$DOC" 'Memory GA' "doc Memory GA honesty"
need_grep "$DOC" 'not Memory GA' "doc not Memory GA"
need_grep "$DOC" 'kernel ≠ product Memory GA' "doc kernel ≠ product Memory GA"
need_grep "$DOC" 'dual_write OFF' "doc dual_write OFF"
need_grep "$DOC" 'no invent GA' "doc no invent GA"
need_grep "$DOC" 'PASS ≠ live dogfood' "doc PASS ≠ live dogfood"
need_grep "$DOC" 'recmem-integration-plan.md' "doc link historical plan"
need_grep "$DOC" 'SKIP_RECMEM_COMPACTION' "doc soft skip env"
need_grep "$DOC" 'residual-gate' "doc Makefile residual-gate target"
need_grep "$DOC" 'recmem-compaction-residual-gate' "doc Makefile recmem residual-gate target"
need_grep "$DOC" 'RESULT PASS' "doc RESULT PASS honesty chain"

# Kernel tree symbols (Go) — at least one of AutoRecMemCompaction / PerformCompaction / CompactionConfig
TREE_HIT=0
if grep -R -n --include='*.go' -F 'AutoRecMemCompaction' . >/dev/null 2>&1; then
  pass "tree AutoRecMemCompaction"
  TREE_HIT=1
else
  warn "tree missing AutoRecMemCompaction"
fi
if grep -R -n --include='*.go' -F 'PerformCompaction' . >/dev/null 2>&1; then
  pass "tree PerformCompaction"
  TREE_HIT=1
else
  warn "tree missing PerformCompaction"
fi
if grep -R -n --include='*.go' -F 'CompactionConfig' . >/dev/null 2>&1; then
  pass "tree CompactionConfig"
  TREE_HIT=1
else
  warn "tree missing CompactionConfig"
fi
if [[ "$TREE_HIT" -eq 0 ]]; then
  fail "tree missing AutoRecMemCompaction or PerformCompaction or CompactionConfig"
else
  pass "tree compaction primitive present (AutoRecMemCompaction or PerformCompaction or CompactionConfig)"
fi

# Makefile target when present
if [[ -f "$MAKEFILE" ]]; then
  if grep -qE 'residual-gate' "$MAKEFILE" 2>/dev/null; then
    pass "needle Makefile residual-gate target"
  else
    fail "Makefile missing residual-gate target"
  fi
  if grep -qE 'recmem-compaction-residual-gate|recmem_compaction_residual_gate' "$MAKEFILE" 2>/dev/null; then
    pass "needle Makefile recmem-compaction residual gate"
  else
    fail "Makefile missing recmem-compaction residual gate target"
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
log "RESULT OK honesty chain: AutoRecMemCompaction shipped partial · PerformCompaction · CompactionConfig · memory_trigger_compact advisory · TUI HITL s1311 · Compaction PASS ≠ invent Memory GA token-reduction · Phase 3 semantic refine residual · kernel ≠ product Memory GA · dual_write OFF · no invent GA · residual PASS ≠ live dogfood"
exit 0
