#!/usr/bin/env bash
# Merge git-cliff output with install instructions for GitHub Releases.
# Required env: CHANGELOG_SECTION, VERSION, TAG, IMAGE_REPO, IMAGE_DIGEST,
# GITHUB_REPOSITORY
set -euo pipefail

CHANGELOG_SECTION="${CHANGELOG_SECTION:?CHANGELOG_SECTION required}"
VERSION="${VERSION:?VERSION required}"
TAG="${TAG:?TAG required}"
IMAGE_REPO="${IMAGE_REPO:?IMAGE_REPO required}"
IMAGE_DIGEST="${IMAGE_DIGEST:?IMAGE_DIGEST required}"
GITHUB_REPOSITORY="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY required}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-release-body.md}"

export VERSION TAG IMAGE_REPO IMAGE_DIGEST GITHUB_REPOSITORY
{
  cat "${CHANGELOG_SECTION}"
  printf '\n---\n\n'
  envsubst <"${ROOT}/.github/release-notes-install.md"
} >"${OUT}"

echo "wrote ${OUT}"
