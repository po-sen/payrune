#!/usr/bin/env bash
set -euo pipefail

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT

cp go.mod "${tmp_dir}/go.mod.before"
cp go.sum "${tmp_dir}/go.sum.before"

go mod tidy

diff -u "${tmp_dir}/go.mod.before" go.mod
diff -u "${tmp_dir}/go.sum.before" go.sum
