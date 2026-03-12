#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/deployments/cloudflare/payrune-api/src/generated"
WASM_PATH="$OUTPUT_DIR/payrune_api_worker.wasm"
WASM_EXEC_PATH="$OUTPUT_DIR/wasm_exec.js"
LEGACY_WASM_MODULE_PATH="$OUTPUT_DIR/payrune_api_worker_wasm.mjs"

mkdir -p "$OUTPUT_DIR"
rm -f "$LEGACY_WASM_MODULE_PATH"

(cd "$ROOT_DIR" && GOCACHE=/tmp/go-build GOOS=js GOARCH=wasm go build -trimpath -o "$WASM_PATH" ./cmd/api-worker)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$WASM_EXEC_PATH"
