#!/usr/bin/env bash
# advanced_agent_inventory_residual_gate.sh — s1297 offline residual
# free eng residual pin s1297 · free eng concurrent s1296+ after free-floor s1294 · lag s1295
# SSOT for kernel advanced agent inventory (not product Memory GA):
#   MultiHopRetrieve + PreferShorterHops (s1067/s1278) · multi-hop lite ≠ full graph RAG
#   ListFactsAsOf / EntryValidAt · K4 bi-temporal lite ≠ dual-clock Graphiti
#   SupersedeEntityFacts · A3 lite ≠ NLP contradiction
#   ListMemoryWithOptions · K2 timeline filters before limit · full FS event-time index residual
# Host/TUI peers mention only: memory_timeline · memory_related · memory_facts_as_of ·
#   memory_supersede_entity · TUI s1296 timeline/compact-status
# Honesty: kernel ≠ product Memory GA · dual_write OFF · no invent GA ·
#   residual PASS ≠ live dogfood · RESULT PASS / RESULT OK honesty chain
# Soft skip: SKIP_ADVANCED_AGENT_INVENTORY=1
#
# Usage:
#   ./scripts/advanced_agent_inventory_residual_gate.sh
#   make residual-gate
#   SKIP_ADVANCED_AGENT_INVENTORY=1 ./scripts/advanced_agent_inventory_residual_gate.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FAIL=0
WARN=0

log() { printf 'advanced-agent-inventory-residual: %s\n' "$*" >&2; }
pass() { log "PASS: $*"; }
warn() { log "WARN: $*"; WARN=$((WARN + 1)); }
fail() { log "FAIL: $*"; FAIL=$((FAIL + 1)); }

if [[ "${SKIP_ADVANCED_AGENT_INVENTORY:-}" == "1" ]]; then
  log "SKIP_ADVANCED_AGENT_INVENTORY=1 soft skip"
  log "RESULT OK (soft skip)"
  log "non-claim: soft skip ≠ invent Memory GA · residual PASS ≠ live dogfood"
  exit 0
fi

log "offline residual SSOT for kernel advanced agent inventory (no dogfood / Memory GA invent required)"
log "non-claim: multi-hop lite ≠ full graph RAG · K4 bi-temporal lite ≠ dual-clock Graphiti · A3 lite ≠ NLP contradiction · kernel ≠ product Memory GA"

DOC="docs/operations/advanced-agent-inventory-residual.md"
MAKEFILE="Makefile"
GATE_SCRIPT="scripts/advanced_agent_inventory_residual_gate.sh"

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

# Residual doc presence + honesty / inventory needles
need_path "$DOC" "advanced agent inventory residual SSOT"
need_path "$GATE_SCRIPT" "advanced agent inventory residual gate script"
need_grep "$DOC" 's1297' "doc free eng residual pin s1297"
need_grep "$DOC" 's1296+' "doc free eng concurrent s1296+"
need_grep "$DOC" 's1294' "doc free-floor s1294"
need_grep "$DOC" 's1295' "doc lag residual s1295"
need_grep "$DOC" 's1296' "doc concurrent TUI s1296 mention"
need_grep "$DOC" 'MultiHopRetrieve' "doc MultiHopRetrieve inventory"
need_grep "$DOC" 'PreferShorterHops' "doc PreferShorterHops inventory"
need_grep "$DOC" 'ListFactsAsOf' "doc ListFactsAsOf inventory"
need_grep "$DOC" 'EntryValidAt' "doc EntryValidAt inventory"
need_grep "$DOC" 'SupersedeEntityFacts' "doc SupersedeEntityFacts inventory"
need_grep "$DOC" 'ListMemoryWithOptions' "doc ListMemoryWithOptions inventory"
need_grep "$DOC" 'multi-hop lite' "doc multi-hop lite honesty"
need_grep "$DOC" 'full graph RAG' "doc full graph RAG honesty"
need_grep "$DOC" 'bi-temporal lite' "doc bi-temporal lite honesty"
need_grep "$DOC" 'dual-clock Graphiti' "doc dual-clock Graphiti honesty"
need_grep "$DOC" 'A3 lite' "doc A3 lite honesty"
need_grep "$DOC" 'NLP contradiction' "doc NLP contradiction honesty"
need_grep "$DOC" 'Memory GA' "doc Memory GA honesty"
need_grep "$DOC" 'kernel ≠ product Memory GA' "doc kernel ≠ product Memory GA"
need_grep "$DOC" 'dual_write OFF' "doc dual_write OFF"
need_grep "$DOC" 'no invent GA' "doc no invent GA"
need_grep "$DOC" 'PASS ≠ live dogfood' "doc PASS ≠ live dogfood"
need_grep "$DOC" 'memory_timeline' "doc host peer memory_timeline mention"
need_grep "$DOC" 'memory_related' "doc host peer memory_related mention"
need_grep "$DOC" 'memory_facts_as_of' "doc host peer memory_facts_as_of mention"
need_grep "$DOC" 'memory_supersede_entity' "doc host peer memory_supersede_entity mention"
need_grep "$DOC" 'SKIP_ADVANCED_AGENT_INVENTORY' "doc soft skip env"
need_grep "$DOC" 'residual-gate' "doc Makefile residual-gate target"
need_grep "$DOC" 'RESULT PASS' "doc RESULT PASS honesty chain"

# Kernel tree symbols (Go)
need_tree_grep 'PreferShorterHops' "PreferShorterHops"
need_tree_grep 'MultiHopRetrieve' "MultiHopRetrieve"
need_tree_grep 'ListFactsAsOf' "ListFactsAsOf"
need_tree_grep 'SupersedeEntityFacts' "SupersedeEntityFacts"
need_tree_grep 'ListMemoryWithOptions' "ListMemoryWithOptions"

# Makefile target when present
if [[ -f "$MAKEFILE" ]]; then
  if grep -qE 'residual-gate' "$MAKEFILE" 2>/dev/null; then
    pass "needle Makefile residual-gate target"
  else
    fail "Makefile missing residual-gate target"
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
log "RESULT OK honesty chain: multi-hop lite ≠ full graph RAG · K4 bi-temporal lite ≠ dual-clock Graphiti · A3 lite ≠ NLP contradiction · kernel ≠ product Memory GA · dual_write OFF · no invent GA · residual PASS ≠ live dogfood"
exit 0
