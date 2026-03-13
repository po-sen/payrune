#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/deployments/cloudflare/payrune-webhook-dispatcher/src/generated"
WASM_PATH="$OUTPUT_DIR/payrune_webhook_dispatcher_worker.wasm"
WASM_EXEC_PATH="$OUTPUT_DIR/wasm_exec.js"

mkdir -p "$OUTPUT_DIR"

(cd "$ROOT_DIR" && GOCACHE=/tmp/go-build GOOS=js GOARCH=wasm go build -trimpath -o "$WASM_PATH" ./cmd/webhook-dispatcher-worker)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$WASM_EXEC_PATH"
