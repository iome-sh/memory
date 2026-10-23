#!/usr/bin/env bash
# Scans tracked files and pull request text for strings listed in the
# LEAK_SCAN_PATTERNS secret. Only path:line locations are printed; the
# patterns and the matched text never reach the log.
#
# Secret format: one extended regular expression (grep -E) per line, matched
# case-insensitively. Blank lines and lines starting with # are ignored.
set -uo pipefail

if [ -z "${LEAK_SCAN_PATTERNS:-}" ]; then
  echo "::error::LEAK_SCAN_PATTERNS is empty; refusing to report a pass without scanning."
  exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
pat="$work/patterns"

printf '%s\n' "$LEAK_SCAN_PATTERNS" | sed -e 's/\r$//' | grep -vE '^[[:space:]]*(#|$)' > "$pat" || true
unset LEAK_SCAN_PATTERNS

# Mask every pattern line in case anything downstream prints it.
while IFS= read -r p; do
  echo "::add-mask::${p}"
done < "$pat"

count="$(wc -l < "$pat" | tr -d ' ')"
if [ "$count" -eq 0 ]; then
  echo "::error::LEAK_SCAN_PATTERNS has no usable patterns; refusing to report a pass without scanning."
  exit 1
fi

# Reject an invalid pattern list instead of treating a grep error as "no match".
grep -qiE -f "$pat" /dev/null 2>/dev/null
if [ $? -gt 1 ]; then
  echo "::error::LEAK_SCAN_PATTERNS contains an invalid extended regular expression."
  exit 1
fi

hits="$work/hits"
: > "$hits"

# Tracked files. Raw grep output (which contains matched text) stays on disk;
# only path:line is kept.
git grep -n -I -i -E --no-color -f "$pat" > "$work/raw" 2>/dev/null
rc=$?
if [ "$rc" -gt 1 ]; then
  echo "::error::Tree scan failed (git grep exit ${rc})."
  exit 1
fi
cut -d: -f1,2 "$work/raw" >> "$hits"
rm -f "$work/raw"

# Pull request title and body, when this run is for a pull request.
scan_text() {
  local label="$1" value="$2" f="$work/text"
  [ -n "$value" ] || return 0
  printf '%s\n' "$value" | sed -e 's/\r$//' > "$f"
  grep -n -i -E -f "$pat" "$f" 2>/dev/null | cut -d: -f1 | sed -e "s|^|${label}:|" >> "$hits"
  rm -f "$f"
}
scan_text "pull-request-title" "${PR_TITLE:-}"
scan_text "pull-request-body" "${PR_BODY:-}"

n="$(wc -l < "$hits" | tr -d ' ')"
if [ "$n" -gt 0 ]; then
  echo "Found ${n} disallowed string location(s):"
  while IFS= read -r loc; do
    file="${loc%:*}"
    line="${loc##*:}"
    echo "  ${loc}"
    case "$file" in
      pull-request-*) echo "::error title=leak-scan::Disallowed string in ${file} at line ${line}" ;;
      *) echo "::error file=${file},line=${line},title=leak-scan::Disallowed string at ${loc}" ;;
    esac
  done < "$hits"
  echo "### leak-scan: ${n} location(s) found" >> "$GITHUB_STEP_SUMMARY"
  exit 1
fi

echo "Scanned $(git ls-files | wc -l | tr -d ' ') tracked files and the pull request text against ${count} pattern(s): no matches."
echo "### leak-scan: scanned against ${count} pattern(s), no matches" >> "$GITHUB_STEP_SUMMARY"
