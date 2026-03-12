#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKER_DIR="$ROOT_DIR/deployments/cloudflare/payrune-poller"

if [[ $# -ne 1 ]]; then
	echo "usage: $0 <mainnet|testnet4>" >&2
	exit 1
fi

TARGET_NETWORK="$1"
case "$TARGET_NETWORK" in
mainnet|testnet4) ;;
*)
	echo "unsupported poller target: $TARGET_NETWORK" >&2
	exit 1
	;;
esac

cd "$WORKER_DIR"
npm exec -- wrangler delete --env "$TARGET_NETWORK"
