#!/usr/bin/env bash
# Scrub staged (or tree) files for forbidden secret-like patterns.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

MODE="${1:-staged}" # staged | tree
PATTERNS_FILE="${ROOT}/hack/scrub-patterns.txt"

if [[ ! -f "${PATTERNS_FILE}" ]]; then
  echo "missing ${PATTERNS_FILE}" >&2
  exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

if [[ "${MODE}" == "tree" ]]; then
  git ls-files >"${tmp}.raw"
else
  git diff --cached --name-only --diff-filter=ACMR >"${tmp}.raw"
fi

# Exclude scrub tooling itself so pattern literals cannot self-match (Kollect pattern).
grep -Ev '^hack/scrub(-patterns\.txt|\.sh)$' "${tmp}.raw" >"${tmp}" || true
rm -f "${tmp}.raw"

if [[ ! -s "${tmp}" ]]; then
  echo "scrub: nothing to scan"
  exit 0
fi

failed=0
while IFS= read -r pattern || [[ -n "${pattern}" ]]; do
  [[ -z "${pattern}" || "${pattern}" =~ ^# ]] && continue
  # Portable across GNU/BSD xargs: feed paths on stdin.
  if tr '\n' '\0' <"${tmp}" | xargs -0 grep -nE "${pattern}" -- 2>/dev/null; then
    echo "scrub: matched forbidden pattern: ${pattern}" >&2
    failed=1
  fi
done < "${PATTERNS_FILE}"

if [[ "${failed}" -ne 0 ]]; then
  exit 1
fi
echo "scrub: clean (${MODE})"
