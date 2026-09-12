#!/usr/bin/env bash
# Locked mixed n=12 LongMemEval baseline: hash vs MiniLM vs BAAI BGE-small-en-v1.5.
#
# Same question IDs, same reader (gpt-4o-mini), same judge (gpt-4o-2024-08-06),
# session_id=conv_id, isolated palace per embed mode.
#
# Not official V1 (n=12, not 500; MiniLM column is not the BGE pin).
# Not a README number. Not Memory GA. Hash-overlap unpublished.
# dual_write OFF. Optional; not part of make ci.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

log() { printf 'longmemeval-baseline: %s\n' "$*" >&2; }

DATASET="${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json}"
IDS="${LONGMEMEVAL_IDS_FILE:-testdata/longmemeval_baseline_ids.json}"
OUTDIR="${LONGMEMEVAL_BASELINE_DIR:-data/baseline}"
ADDR="${LONGMEMEVAL_ADDR:-:8769}"
SERVER_URL="${LONGMEMEVAL_SERVER:-http://127.0.0.1:8769}"
WORKERS="${LONGMEMEVAL_QA_WORKERS:-2}"
JUDGE="${LONGMEMEVAL_JUDGE_MODEL:-gpt-4o-2024-08-06}"
READER="${OPENAI_MODEL:-gpt-4o-mini}"
DOC_OUT="${LONGMEMEVAL_BASELINE_DOC:-docs/LONGMEMEVAL_BASELINE.md}"
MINILM="${ROOT}/testdata/models/KnightsAnalytics_all-MiniLM-L6-v2"
BAAI="${ROOT}/testdata/models/BAAI_bge-small-en-v1.5"

if [[ ! -f "${DATASET}" ]]; then
  log "SKIP: missing ${DATASET} (gitignored oracle). Not a CI failure."
  exit 0
fi
if [[ ! -f "${IDS}" ]]; then
  log "missing ids file ${IDS}"
  exit 1
fi
if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  log "SKIP: OPENAI_API_KEY unset"
  exit 0
fi

if [[ -x "${ROOT}/.venv/bin/python" ]]; then
  export PATH="${ROOT}/.venv/bin:${PATH}"
elif [[ -n "${LONGMEMEVAL_PYTHON:-}" ]]; then
  PYDIR="$(cd "$(dirname "${LONGMEMEVAL_PYTHON}")" && pwd)"
  export PATH="${PYDIR}:${PATH}"
fi
if ! python3 -c "import openai,requests,tqdm,backoff,numpy" 2>/dev/null; then
  log "python deps missing; pip install -r requirements-bench.txt"
  exit 1
fi

mkdir -p "${OUTDIR}"

if [[ ! -f "${BAAI}/model.onnx" ]]; then
  log "fetching BAAI BGE-small-en-v1.5 into hugot layout (KnightsAnalytics repo is 404)"
  go run ./scripts/download_onnx_model.go >/tmp/lme-bge-path.txt || true
fi
if [[ ! -f "${BAAI}/model.onnx" ]]; then
  log "BGE layout missing after download helper; BGE column will SKIP"
fi

PORT="${ADDR#:}"
kill_server() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "${SERVER_PID}" 2>/dev/null; then
    kill "${SERVER_PID}" 2>/dev/null || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
  SERVER_PID=""
  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids="$(lsof -tiTCP:"${PORT}" -sTCP:LISTEN 2>/dev/null || true)"
    if [[ -n "${pids}" ]]; then
      # shellcheck disable=SC2086
      kill ${pids} 2>/dev/null || true
      sleep 1
      pids="$(lsof -tiTCP:"${PORT}" -sTCP:LISTEN 2>/dev/null || true)"
      if [[ -n "${pids}" ]]; then
        # shellcheck disable=SC2086
        kill -9 ${pids} 2>/dev/null || true
      fi
    fi
  fi
}
trap kill_server EXIT

wait_health() {
  local want="$1"
  local i body mode
  for i in $(seq 1 60); do
    body="$(curl -fsS "${SERVER_URL}/health" 2>/dev/null || true)"
    if [[ -n "${body}" ]]; then
      mode="$(python3 -c 'import json,sys; print(json.loads(sys.argv[1]).get("embed_mode",""))' "${body}" 2>/dev/null || true)"
      if [[ "${mode}" == "${want}" ]]; then
        printf '%s\n' "${body}"
        return 0
      fi
    fi
    sleep 1
  done
  log "server health embed_mode want=${want} last=${body:-none} at ${SERVER_URL}"
  return 1
}

run_mode() {
  local mode="$1"
  local palace="${OUTDIR}/palace-${mode}"
  local hyp="${OUTDIR}/hypotheses-${mode}.jsonl"
  rm -rf "${palace}"
  mkdir -p "${palace}"
  kill_server
  unset MEMORY_ONNX_MODEL_PATH || true
  case "${mode}" in
    hash) ;;
    minilm)
      export MEMORY_ONNX_MODEL_PATH="${MINILM}"
      ;;
    bge)
      export MEMORY_ONNX_MODEL_PATH="${BAAI}"
      ;;
    *)
      log "unknown mode ${mode}"
      return 1
      ;;
  esac
  local want_mode="hash"
  case "${mode}" in
    minilm) want_mode="onnx-minilm-l6-v2" ;;
    bge) want_mode="onnx-bge-small-en-v1.5" ;;
  esac
  log "mode=${mode} want=${want_mode} palace=${palace} onnx=${MEMORY_ONNX_MODEL_PATH:-hash}"
  mkdir -p bin
  if [[ ! -x bin/longmemeval-server ]]; then
    go build -o bin/longmemeval-server ./cmd/longmemeval-server
  fi
  LONGMEMEVAL_PALACE_ROOT="${palace}" LONGMEMEVAL_ADDR="${ADDR}" \
    ./bin/longmemeval-server >"${OUTDIR}/server-${mode}.log" 2>&1 &
  SERVER_PID=$!
  if ! wait_health "${want_mode}"; then
    log "server log:"; tail -40 "${OUTDIR}/server-${mode}.log" >&2 || true
    return 1
  fi
  curl -fsS "${SERVER_URL}/health" | tee "${OUTDIR}/health-${mode}.json" >&2
  OPENAI_MODEL="${READER}" LONGMEMEVAL_SERVER="${SERVER_URL}" \
    python3 scripts/longmemeval_qa_generate.py \
      --dataset "${DATASET}" \
      --output "${hyp}" \
      --ids-file "${IDS}" \
      --workers "${WORKERS}" \
      --server "${SERVER_URL}"
  bash scripts/longmemeval_judge.sh "${JUDGE}" "${hyp}" "${DATASET}"
}

MODES="${LONGMEMEVAL_BASELINE_MODES:-hash minilm bge}"
for mode in ${MODES}; do
  case "${mode}" in
    hash|minilm) run_mode "${mode}" ;;
    bge)
      if [[ -f "${BAAI}/model.onnx" ]]; then
        LONGMEMEVAL_QA_WORKERS="${LONGMEMEVAL_BGE_WORKERS:-1}"
        WORKERS="${LONGMEMEVAL_BGE_WORKERS:-1}"
        run_mode bge
      else
        log "skip bge column (no model.onnx)"
      fi
      ;;
    *) log "skip unknown mode ${mode}" ;;
  esac
done
kill_server

python3 - "${OUTDIR}" "${DOC_OUT}" "${IDS}" "${JUDGE}" "${READER}" << 'PY'
import json, sys, collections, datetime, subprocess, os
from pathlib import Path
outdir, doc_out, ids_path, judge, reader = sys.argv[1:6]
ids = json.loads(Path(ids_path).read_text())
qid_types = {}
# types from first available hyp file
rows_by_mode = {}
for mode in ("hash", "minilm", "bge"):
    # judge writes hyp.eval-results-gpt-4o when zoo key is gpt-4o
    cands = list(Path(outdir).glob(f"hypotheses-{mode}.jsonl.eval-results-*"))
    hyp = Path(outdir) / f"hypotheses-{mode}.jsonl"
    path = cands[0] if cands else hyp
    if not path.exists():
        rows_by_mode[mode] = None
        continue
    rows = [json.loads(l) for l in path.read_text().splitlines() if l.strip()]
    rows_by_mode[mode] = rows
    for r in rows:
        qid_types[r.get("question_id")] = r.get("question_type") or qid_types.get(r.get("question_id"))

def score(rows):
    if not rows:
        return None
    by = collections.defaultdict(lambda: [0, 0])
    ok = 0
    for r in rows:
        lab = r.get("autoeval_label") or {}
        hit = bool(lab.get("label")) if isinstance(lab, dict) else False
        t = r.get("question_type") or "?"
        by[t][1] += 1
        by[t][0] += int(hit)
        ok += int(hit)
    return ok, len(rows), dict(by)

sha = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
tag = subprocess.check_output(["git", "describe", "--tags", "--always"], text=True).strip()
now = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

def fmt_mode(mode, sc):
    if not sc:
        return "| %s | SKIP | — | — |" % mode
    ok, n, by = sc
    types = ", ".join("%s %d/%d" % (t, by[t][0], by[t][1]) for t in sorted(by))
    return "| %s | **%d/%d** | %.3f | %s |" % (mode, ok, n, ok / n if n else 0, types)

lines = []
lines.append("# LongMemEval locked mixed baseline (internal)")
lines.append("")
lines.append("Kernel-only · **not Memory GA** · dual_write **OFF** · **not a README number** · **not official V1**.")
lines.append("")
lines.append("Comparable columns on the **same 12 question IDs** (stratified mixed, 2 of each type).")
lines.append("Reader `gpt-4o-mini`. Judge **`gpt-4o-2024-08-06`**. `session_id` = official `conv_id`.")
lines.append("Isolated palace per embed mode (`LONGMEMEVAL_PALACE_ROOT`).")
lines.append("")
lines.append("Official V1 remains: BGE-small-en-v1.5 ONNX + mixed sample of the **full** oracle (n=500) + this judge, reproduced twice.")
lines.append("This n=12 table is the **improvement baseline**, not that card.")
lines.append("")
lines.append("| Field | Value |")
lines.append("|-------|--------|")
lines.append("| Date (UTC) | %s |" % now)
lines.append("| Kernel SHA | `%s` |" % sha)
lines.append("| Kernel describe | %s |" % tag)
lines.append("| Slice | locked mixed n=12 (`testdata/longmemeval_baseline_ids.json`) |")
lines.append("| Histogram | knowledge-update 2, multi-session 2, single-session-assistant 2, single-session-preference 2, single-session-user 2, temporal-reasoning 2 |")
lines.append("| Reader | `%s` (not the official V1 pin; official pin is the judge) |" % reader)
lines.append("| Judge | `%s` |" % judge)
lines.append("| Qdrant | off |")
lines.append("| BGE source | `BAAI/bge-small-en-v1.5` `onnx/model.onnx` reshaped to hugot layout. `KnightsAnalytics/bge-small-en-v1.5` does not exist (HF 404). |")
lines.append("")
lines.append("## Scores")
lines.append("")
lines.append("| Embed | Judge-true | Rate | By type |")
lines.append("|-------|------------|------|---------|")
for mode in ("hash", "minilm", "bge"):
    rows = rows_by_mode.get(mode)
    sc = score(rows)
    n_ids = len(ids.get("question_ids") or [])
    extra = ""
    if rows is not None and sc and sc[1] != n_ids:
        extra = " **incomplete (%d/%d IDs; do not treat rate as comparable)**" % (sc[1], n_ids)
        lines.append(fmt_mode(mode, sc)[:-1] + extra + " |")
    else:
        lines.append(fmt_mode(mode, sc))
lines.append("")
lines.append("Hash is keyword-first retrieve (default embedder). MiniLM is in-tree ONNX, **not** the official V1 embed pin.")
lines.append("BGE is BAAI ONNX in hugot layout (root `model.onnx`). Improve the **BGE** column on this ID list; then scale n=60 / n=500.")
lines.append("")
lines.append("## Honesty")
lines.append("")
lines.append("- leftover_is_bind OPEN · dual_write OFF · not Memory GA · hash-overlap unpublished")
lines.append("- n=12 is not overall V1 · prefix-n is not mixed · do not put these rates on the README")
lines.append("- E-G1 is a laptop TTFH clock; this table does not move it")
lines.append("")
text = "\n".join(lines) + "\n"
Path(doc_out).write_text(text)
print(text)
PY

log "wrote ${DOC_OUT}"
