#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKER_DIR="$ROOT_DIR/deployments/cloudflare/payrune-api"

if [[ -t 1 ]]; then
	COLOR_BLUE=$'\033[1;34m'
	COLOR_GREEN=$'\033[0;32m'
	COLOR_RESET=$'\033[0m'
else
	COLOR_BLUE=''
	COLOR_GREEN=''
	COLOR_RESET=''
fi

step() {
	printf '%s==> %s%s\n' "$COLOR_BLUE" "$1" "$COLOR_RESET"
}

success() {
	printf '%sOK: %s%s\n' "$COLOR_GREEN" "$1" "$COLOR_RESET"
}

cd "$WORKER_DIR"
step "Deleting Worker"
npm exec -- wrangler delete "$@"
success "Worker delete finished."
