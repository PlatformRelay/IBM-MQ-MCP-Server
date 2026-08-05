#!/usr/bin/env bash
# Fail if gofmt would change any tracked Go file.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

# Portable: avoid mapfile (Bash 4+) for macOS system bash.
files="$(git ls-files '*.go')"
if [[ -z "${files}" ]]; then
  echo "no Go files" >&2
  exit 0
fi
# shellcheck disable=SC2086
out="$(gofmt -l ${files})"
if [[ -n "${out}" ]]; then
  echo "gofmt needed on:" >&2
  echo "${out}" >&2
  exit 1
fi
echo "gofmt clean"
