#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

: "${ETHEREUM_CREATE2_VERIFY_NETWORK:?ETHEREUM_CREATE2_VERIFY_NETWORK is required}"
: "${ETHEREUM_CREATE2_VERIFY_RPC_URL:?ETHEREUM_CREATE2_VERIFY_RPC_URL is required}"
: "${ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY:?ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY is required}"
: "${ETHEREUM_CREATE2_VERIFY_COLLECTOR_ADDRESS:?ETHEREUM_CREATE2_VERIFY_COLLECTOR_ADDRESS is required}"

FUND_AMOUNT_WEI="${ETHEREUM_CREATE2_VERIFY_FUND_AMOUNT_WEI:-1}"

pushd "${REPO_ROOT}" >/dev/null

go run ./cmd/ethereum-create2-tool build

args=(
  verify-chain
  --network "${ETHEREUM_CREATE2_VERIFY_NETWORK}"
  --rpc-url "${ETHEREUM_CREATE2_VERIFY_RPC_URL}"
  --operator-private-key "${ETHEREUM_CREATE2_VERIFY_OPERATOR_PRIVATE_KEY}"
  --collector "${ETHEREUM_CREATE2_VERIFY_COLLECTOR_ADDRESS}"
  --fund-amount-wei "${FUND_AMOUNT_WEI}"
)

if [ -n "${ETHEREUM_CREATE2_VERIFY_FACTORY_ADDRESS:-}" ]; then
  args+=(--factory "${ETHEREUM_CREATE2_VERIFY_FACTORY_ADDRESS}")
fi

if [ -n "${ETHEREUM_CREATE2_VERIFY_SALT:-}" ]; then
  args+=(--salt "${ETHEREUM_CREATE2_VERIFY_SALT}")
fi

go run ./cmd/ethereum-create2-tool "${args[@]}"

popd >/dev/null
