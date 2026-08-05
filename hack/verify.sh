#!/usr/bin/env bash
# Local verify: modules tidy + format check (fast preflight).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

go mod tidy
if ! git diff --quiet -- go.mod go.sum; then
  echo "go.mod/go.sum not tidy; run go mod tidy and commit" >&2
  git diff -- go.mod go.sum >&2 || true
  exit 1
fi

"${ROOT}/hack/format-check.sh"
echo "verify: ok"
