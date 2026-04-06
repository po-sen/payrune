#!/usr/bin/env bash
set -euo pipefail

default_mainnet_usdt_asset_reference="0xdac17f958d2ee523a2206206994597c13d831ec7"
default_sepolia_usdt_asset_reference="0xd077a400968890eacc75cdc901f0356c943e4fdb"

usage() {
  cat <<'EOF'
Usage:
  bash scripts/ethereum_usdt_pay_with_ledger.sh [--dry-run|--broadcast]

Required env:
  ETHEREUM_PAYMENT_RPC_URL
  ETHEREUM_PAYMENT_FROM_ADDRESS
  ETHEREUM_PAYMENT_TO_ADDRESS
  ETHEREUM_PAYMENT_AMOUNT_MINOR

Optional env:
  ETHEREUM_PAYMENT_DERIVATION_PATH
  ETHEREUM_PAYMENT_ASSET_REFERENCE
  ETHEREUM_MAINNET_USDT_ASSET_REFERENCE
  ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE

Notes:
  - The helper sends `transfer(address,uint256)` for USDT using a Ledger signer.
  - `ETHEREUM_PAYMENT_AMOUNT_MINOR` must be an integer in minor units. USDT uses 6 decimals.
  - Default mode is `--dry-run`.
  - `--broadcast` validates the connected Ledger sender before sending the transaction.
EOF
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

trim() {
  local value="${1:-}"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

normalize_hex() {
  local value
  value="$(trim "${1:-}")"
  printf '%s' "${value,,}"
}

require_command() {
  local name="$1"
  command -v "$name" >/dev/null 2>&1 || die "$name is required"
}

require_env() {
  local name="$1"
  local value
  value="$(trim "${!name:-}")"
  [[ -n "$value" ]] || die "$name is required"
}

resolve_network_from_chain_id() {
  local chain_id="$1"
  case "$chain_id" in
    1)
      printf 'mainnet'
      ;;
    11155111)
      printf 'sepolia'
      ;;
    *)
      die "unsupported Ethereum chain id: ${chain_id:-<empty>}"
      ;;
  esac
}

resolve_usdt_asset_reference_for_network() {
  local network="$1"
  local explicit_override=""
  local env_override=""

  explicit_override="$(normalize_hex "${ETHEREUM_PAYMENT_ASSET_REFERENCE:-}")"
  if [[ -n "$explicit_override" ]]; then
    resolved_asset_reference="$explicit_override"
    resolved_asset_reference_source="ETHEREUM_PAYMENT_ASSET_REFERENCE"
    return 0
  fi

  case "$network" in
    mainnet)
      env_override="$(normalize_hex "${ETHEREUM_MAINNET_USDT_ASSET_REFERENCE:-}")"
      if [[ -n "$env_override" ]]; then
        resolved_asset_reference="$env_override"
        resolved_asset_reference_source="ETHEREUM_MAINNET_USDT_ASSET_REFERENCE"
      else
        resolved_asset_reference="$default_mainnet_usdt_asset_reference"
        resolved_asset_reference_source="default_mainnet"
      fi
      ;;
    sepolia)
      env_override="$(normalize_hex "${ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE:-}")"
      if [[ -n "$env_override" ]]; then
        resolved_asset_reference="$env_override"
        resolved_asset_reference_source="ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE"
      else
        resolved_asset_reference="$default_sepolia_usdt_asset_reference"
        resolved_asset_reference_source="default_sepolia"
      fi
      ;;
    *)
      die "unsupported Ethereum network: ${network:-<empty>}"
      ;;
  esac
}

broadcast=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      broadcast=0
      shift
      ;;
    --broadcast)
      broadcast=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      die "unknown argument: $1"
      ;;
  esac
done

require_command cast

require_env ETHEREUM_PAYMENT_RPC_URL
require_env ETHEREUM_PAYMENT_FROM_ADDRESS
require_env ETHEREUM_PAYMENT_TO_ADDRESS
require_env ETHEREUM_PAYMENT_AMOUNT_MINOR

rpc_url="$(trim "${ETHEREUM_PAYMENT_RPC_URL}")"
from_address="$(normalize_hex "${ETHEREUM_PAYMENT_FROM_ADDRESS}")"
to_address="$(normalize_hex "${ETHEREUM_PAYMENT_TO_ADDRESS}")"
amount_minor="$(trim "${ETHEREUM_PAYMENT_AMOUNT_MINOR}")"
derivation_path="$(trim "${ETHEREUM_PAYMENT_DERIVATION_PATH:-}")"

[[ "$from_address" =~ ^0x[0-9a-f]{40}$ ]] || die "ETHEREUM_PAYMENT_FROM_ADDRESS must be a 20-byte hex address"
[[ "$to_address" =~ ^0x[0-9a-f]{40}$ ]] || die "ETHEREUM_PAYMENT_TO_ADDRESS must be a 20-byte hex address"
[[ "$amount_minor" =~ ^[0-9]+$ ]] || die "ETHEREUM_PAYMENT_AMOUNT_MINOR must be a base-10 integer"
[[ "$amount_minor" != "0" ]] || die "ETHEREUM_PAYMENT_AMOUNT_MINOR must be greater than zero"

chain_id="$(cast chain-id --rpc-url "$rpc_url" | tail -n 1 | tr -d '\r' | tr -d '[:space:]')"
[[ "$chain_id" =~ ^[0-9]+$ ]] || die "unexpected chain id output: ${chain_id:-<empty>}"
network="$(resolve_network_from_chain_id "$chain_id")"

resolved_asset_reference=""
resolved_asset_reference_source=""
resolve_usdt_asset_reference_for_network "$network"
[[ "$resolved_asset_reference" =~ ^0x[0-9a-f]{40}$ ]] || die "resolved asset reference is invalid: ${resolved_asset_reference:-<empty>}"

send_cmd=(cast send --rpc-url "$rpc_url" --from "$from_address" --ledger)
if [[ -n "$derivation_path" ]]; then
  send_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi
send_cmd+=("$resolved_asset_reference" "transfer(address,uint256)" "$to_address" "$amount_minor")

mode="dry-run"
if [[ "$broadcast" -eq 1 ]]; then
  mode="broadcast"
fi

printf 'mode: %s\n' "$mode"
printf 'network: %s\n' "$network"
printf 'chain_id: %s\n' "$chain_id"
printf 'asset_reference: %s\n' "$resolved_asset_reference"
printf 'asset_reference_source: %s\n' "$resolved_asset_reference_source"
printf 'from_address: %s\n' "$from_address"
printf 'to_address: %s\n' "$to_address"
printf 'amount_minor: %s\n' "$amount_minor"
printf 'command:'
for arg in "${send_cmd[@]}"; do
  printf ' %q' "$arg"
done
printf '\n'

if [[ "$broadcast" -eq 0 ]]; then
  printf 'dry-run: not broadcasting\n'
  exit 0
fi

ledger_cmd=(cast wallet address --ledger)
if [[ -n "$derivation_path" ]]; then
  ledger_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi

ledger_sender="$("${ledger_cmd[@]}" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]')"
[[ "$ledger_sender" =~ ^0x[0-9a-f]{40}$ ]] || die "Ledger sender is invalid: ${ledger_sender:-<empty>}"
[[ "$ledger_sender" == "$from_address" ]] || die "Ledger sender ${ledger_sender} does not match ETHEREUM_PAYMENT_FROM_ADDRESS ${from_address}"

printf 'ledger_sender: %s\n' "$ledger_sender"

broadcast_output="$("${send_cmd[@]}")"
printf '%s\n' "$broadcast_output"
