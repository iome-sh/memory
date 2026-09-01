#!/usr/bin/env bash
# public_flip_readiness_gate.sh — s1467 offline M4 public-flip readiness residual
# free eng residual pin s1467 · free eng concurrent s1467+ after free-floor s1465 · lag s1466
# peers s1468 (mcp) · s1469 (TUI) · s1470 (aion residual) mention only · free-floor peer s1471 · free eng s1473+
#
# SSOT for Option A M4 *readiness* (not the flip):
#   docs/PUBLIC_FLIP_READINESS.md + OPEN_SOURCE_AUDIT + LICENSE/SECURITY/CI present
#   needles: public · residual PASS ≠ public flip · kernel first · not Memory GA · s1467
# Honesty: public · residual PASS ≠ public flip · not Memory GA · dual_write OFF ·
#   aion stays private · M4 readiness ≠ M4 complete / invent public · does NOT flip visibility
# Soft skip: SKIP_PUBLIC_FLIP_READINESS=1
#
# Usage:
#   ./scripts/public_flip_readiness_gate.sh
#   make public-flip-readiness-gate
#   SKIP_PUBLIC_FLIP_READINESS=1 ./scripts/public_flip_readiness_gate.sh
#
# Offline greps only — no network required beyond repo files · does NOT change GitHub visibility.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FAIL=0
WARN=0

log() { printf 'public-flip-readiness: %s\n' "$*" >&2; }
pass() { log "PASS: $*"; }
warn() { log "WARN: $*"; WARN=$((WARN + 1)); }
fail() { log "FAIL: $*"; FAIL=$((FAIL + 1)); }

if [[ "${SKIP_PUBLIC_FLIP_READINESS:-}" == "1" ]]; then
  log "SKIP_PUBLIC_FLIP_READINESS=1 soft skip"
  log "RESULT OK (soft skip)"
  log "non-claim: soft skip ≠ invent public flip · residual PASS ≠ public flip · public · not Memory GA"
  exit 0
fi

log "offline M4 public-flip readiness residual (no visibility flip / no network beyond repo files)"
log "non-claim: residual PASS ≠ public flip · public · kernel first · not Memory GA · dual_write OFF · aion stays private · M4 readiness ≠ invent public"

DOC="docs/PUBLIC_FLIP_READINESS.md"
AUDIT="docs/OPEN_SOURCE_AUDIT.md"
LICENSE="LICENSE"
SECURITY="SECURITY.md"
CI_WF=".github/workflows/ci.yml"
MAKEFILE="Makefile"
GATE_SCRIPT="scripts/public_flip_readiness_gate.sh"

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

forbid_grep() {
  local file="$1" needle="$2" label="$3"
  if [[ ! -f "$file" ]]; then
    fail "skip forbid ${label} (missing ${file})"
    return
  fi
  if grep -qF "$needle" "$file"; then
    fail "forbidden needle ${label} in ${file}"
  else
    pass "forbid ${label}"
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

# Required artifacts present
need_path "$DOC" "PUBLIC_FLIP_READINESS.md"
need_path "$AUDIT" "OPEN_SOURCE_AUDIT.md"
need_path "$LICENSE" "LICENSE"
need_path "$SECURITY" "SECURITY.md"
need_path "$CI_WF" "CI workflow"
need_path "$GATE_SCRIPT" "public flip readiness gate script"

# Core readiness needles in PUBLIC_FLIP_READINESS.md
need_grep "$DOC" 's1467' "doc free eng residual pin s1467"
need_grep "$DOC" 's1467+' "doc free eng concurrent s1467+"
need_grep "$DOC" 's1465' "doc free-floor s1465"
need_grep "$DOC" 's1466' "doc lag s1466"
need_grep "$DOC" 's1468' "doc peer s1468 mcp mention"
need_grep "$DOC" 's1469' "doc peer s1469 TUI mention"
need_grep "$DOC" 's1470' "doc peer s1470 aion residual mention"
need_grep "$DOC" 's1471' "doc free-floor peer s1471"
need_grep "$DOC" 's1473+' "doc free eng s1473+"
need_grep "$DOC" 'public' "doc public"
need_grep "$DOC" 'residual PASS ≠ public flip' "doc residual PASS ≠ public flip"
need_grep "$DOC" 'kernel first' "doc kernel first"
need_grep "$DOC" 'not Memory GA' "doc not Memory GA"
need_grep "$DOC" 'dual_write OFF' "doc dual_write OFF"
need_grep "$DOC" 'aion stays private' "doc aion stays private"
need_grep "$DOC" 'iomesh-memory-mcp' "doc MCP host naming"
need_grep "$DOC" 'OPEN_SOURCE_AUDIT.md' "doc link OPEN_SOURCE_AUDIT"
need_grep "$DOC" 'Palace sunset' "doc Palace sunset"
need_grep "$DOC" 'open boxes stay open' "doc open boxes stay open"
need_grep "$DOC" 'mesh optional' "doc mesh optional"
forbid_grep "$DOC" '$88' "doc no \$88 rate"
forbid_grep "$DOC" '$119' "doc no \$119 rate"
forbid_grep "$AUDIT" '$88' "audit no \$88 rate"
forbid_grep "$AUDIT" '$119' "audit no \$119 rate"
need_grep "$DOC" 'public-flip-readiness-gate' "doc Makefile public-flip-readiness-gate target"
need_grep "$DOC" 'SKIP_PUBLIC_FLIP_READINESS' "doc soft skip env"
need_grep "$DOC" 'M4 readiness ≠ M4 complete' "doc M4 readiness ≠ M4 complete"
need_grep "$DOC" 'does **not** flip' "doc does not flip visibility honesty"

# OPEN_SOURCE_AUDIT continuum stamp + link + honesty
need_grep "$AUDIT" 's1467' "audit continuum s1467"
need_grep "$AUDIT" 'public' "audit public"
need_grep "$AUDIT" 'PUBLIC_FLIP_READINESS.md' "audit links PUBLIC_FLIP_READINESS"
need_grep "$AUDIT" 'residual PASS ≠ public flip' "audit residual PASS ≠ public flip"
need_grep "$AUDIT" 'not Memory GA' "audit not Memory GA"

# Makefile target
if [[ -f "$MAKEFILE" ]]; then
  if grep -qE 'public-flip-readiness-gate|public_flip_readiness_gate' "$MAKEFILE" 2>/dev/null; then
    pass "needle Makefile public-flip-readiness-gate target"
  else
    fail "Makefile missing public-flip-readiness-gate target"
  fi
else
  warn "no Makefile (document bash path only)"
fi

# Non-claim: gate must not invoke visibility flip commands (static self-check).
# Only non-comment executable lines are considered (lines not starting with optional whitespace + #).
if grep -E '^[[:space:]]*(gh[[:space:]]+repo[[:space:]]+edit|gh[[:space:]]+api)' "$GATE_SCRIPT" \
  | grep -vE '^[[:space:]]*#' \
  | grep -qiE 'visibility' 2>/dev/null; then
  fail "gate script must not invoke visibility flip commands"
else
  pass "gate script does not invoke visibility flip commands"
fi

log "WARN count=${WARN} FAIL count=${FAIL}"
if [[ "$FAIL" -gt 0 ]]; then
  log "RESULT FAIL"
  exit 1
fi
log "RESULT PASS"
log "RESULT OK honesty chain: public · residual PASS ≠ public flip · kernel first · not Memory GA · dual_write OFF · aion stays private · M4 readiness ≠ invent public · open boxes stay open · Palace sunset · mesh optional · s1467"
exit 0
