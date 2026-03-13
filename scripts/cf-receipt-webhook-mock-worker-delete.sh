#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKER_DIR="$ROOT_DIR/deployments/cloudflare/receipt-webhook-mock"

cd "$WORKER_DIR"
npm exec -- wrangler delete
