#!/usr/bin/env bash
# Coverage floor for packages that currently have statements under test.
# Default scope: packages with production observability and MCP wiring under test.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

COVERAGE_MIN="${COVERAGE_MIN:-50}"
COVERPROFILE="${COVERPROFILE:-coverage.out}"
PACKAGES="${COVER_PACKAGES:-./internal/mcpserver/... ./internal/observability/... ./internal/adapter/opshttp/... ./internal/adapter/remotemcp/... ./cmd/ibm-mq-mcp/...}"

CGO_ENABLED="${CGO_ENABLED:-0}" go test ${RACE_FLAGS:-} ${PACKAGES} \
  -count=1 \
  -coverprofile="${COVERPROFILE}" \
  -covermode=atomic

total="$(go tool cover -func="${COVERPROFILE}" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
if [[ -z "${total}" ]]; then
  echo "failed to parse coverage total from ${COVERPROFILE}" >&2
  exit 1
fi
awk -v total="${total}" -v min="${COVERAGE_MIN}" 'BEGIN {
  if (total+0 < min+0) {
    printf "coverage %.1f%% is below floor %s%%\n", total, min > "/dev/stderr"
    exit 1
  }
  printf "coverage %.1f%% (floor %s%%)\n", total, min
}'
