#!/usr/bin/env bash
# Build release artifacts under dist/ for a tagged version.
# Usage: hack/release-assets.sh <version> <image-repo>
# Example: hack/release-assets.sh 0.1.0 ghcr.io/platformrelay/ibm-mq-mcp
set -euo pipefail

VERSION="${1:?version required (e.g. 0.1.0)}"
IMAGE="${2:?image repository required (e.g. ghcr.io/platformrelay/ibm-mq-mcp)}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="${ROOT}/dist"
WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

mkdir -p "${DIST}"

BINARY=ibm-mq-mcp
PLATFORMS=(
  linux/amd64
  linux/arm64
  darwin/amd64
  darwin/arm64
)

for platform in "${PLATFORMS[@]}"; do
  GOOS="${platform%/*}"
  GOARCH="${platform#*/}"
  out_dir="${WORK}/${BINARY}_${VERSION}_${GOOS}_${GOARCH}"
  mkdir -p "${out_dir}"
  (
    cd "${ROOT}"
    CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
      go build -trimpath -ldflags="-w -s" -o "${out_dir}/${BINARY}" ./cmd/ibm-mq-mcp
  )
  tar -C "${WORK}" -czf "${DIST}/${BINARY}_${VERSION}_${GOOS}_${GOARCH}.tar.gz" \
    "${BINARY}_${VERSION}_${GOOS}_${GOARCH}/${BINARY}"
  echo "built ${DIST}/${BINARY}_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
done

(
  cd "${DIST}"
  if [[ ! -f "sbom.spdx.json" ]]; then
    echo "sbom.spdx.json missing from ${DIST}/ (required for checksums)" >&2
    exit 1
  fi
  files=("${BINARY}"_*_"${VERSION}"_*.tar.gz "sbom.spdx.json")
  sha256sum "${files[@]}" >checksums.txt
)

echo "release assets written to ${DIST}/ (image ref: ${IMAGE}:${VERSION})"
