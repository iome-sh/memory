#!/usr/bin/env bash
# Print shell exports for CGO ORT builds. Usage: eval "$(./scripts/ort_cgo_env.sh)"
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LIB_DIR="${MEMORY_ORT_DEPS_DIR:-${ROOT}/testdata/ort-deps/lib}"

if [[ ! -f "${LIB_DIR}/libtokenizers.a" ]]; then
  echo "echo 'error: missing ${LIB_DIR}/libtokenizers.a — run: ./scripts/download_ort_deps.sh' >&2" >&2
  echo "return 1 2>/dev/null || exit 1"
  exit 1
fi

case "$(uname -s)" in
  Darwin)
    if [[ ! -f "${LIB_DIR}/libonnxruntime.dylib" ]] && ! ls "${LIB_DIR}"/libonnxruntime*.dylib >/dev/null 2>&1; then
      echo "echo 'error: missing libonnxruntime.dylib in ${LIB_DIR}' >&2" >&2
      echo "return 1 2>/dev/null || exit 1"
      exit 1
    fi
    ;;
  Linux)
    if [[ ! -f "${LIB_DIR}/libonnxruntime.so" ]] && ! ls "${LIB_DIR}"/libonnxruntime*.so* >/dev/null 2>&1; then
      echo "echo 'error: missing libonnxruntime.so in ${LIB_DIR}' >&2" >&2
      echo "return 1 2>/dev/null || exit 1"
      exit 1
    fi
    ;;
esac

printf 'export CGO_ENABLED=1\n'
printf 'export CGO_LDFLAGS="-L%s -ltokenizers"\n' "${LIB_DIR}"
printf 'export MEMORY_ORT_LIBRARY_DIR=%s\n' "${LIB_DIR}"
printf 'export DYLD_LIBRARY_PATH=%s:${DYLD_LIBRARY_PATH:-}\n' "${LIB_DIR}"
printf 'export LD_LIBRARY_PATH=%s:${LD_LIBRARY_PATH:-}\n' "${LIB_DIR}"