#!/usr/bin/env bash
set -euo pipefail

mapfile -t repo_files < <(git ls-files -co --exclude-standard)

if [ "${#repo_files[@]}" -eq 0 ]; then
  echo "no repository files found to check"
  exit 0
fi

pre-commit run --files "${repo_files[@]}"

echo
echo "default-stage hooks passed"
echo "manual-stage hook available: pre-commit run govulncheck --hook-stage manual --all-files"
