#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash scripts/ethereum_create2_factory_deploy.sh [--dry-run|--broadcast]

Required env:
  ETHEREUM_SWEEP_RPC_URL
  ETHEREUM_SWEEP_FROM_ADDRESS

Optional env:
  ETHEREUM_SWEEP_DERIVATION_PATH

Notes:
  - The script deploys the checked-in `Create2ReceiverFactoryV1` artifact.
  - It resolves the target network from chain ID and updates checked-in metadata on successful
    broadcast.
  - It validates the connected Ledger sender before deciding whether to broadcast.
  - Default mode is `--dry-run`.
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

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
artifact_path="${repo_root}/internal/infrastructure/ethereumcreate2assets/artifacts/Create2ReceiverFactoryV1.json"
metadata_dir="${repo_root}/internal/infrastructure/ethereumcreate2assets/metadata"

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

require_command jq
require_command cast

require_env ETHEREUM_SWEEP_RPC_URL
require_env ETHEREUM_SWEEP_FROM_ADDRESS

from_address="$(normalize_hex "${ETHEREUM_SWEEP_FROM_ADDRESS}")"
derivation_path="$(trim "${ETHEREUM_SWEEP_DERIVATION_PATH:-}")"

[[ "$from_address" =~ ^0x[0-9a-f]{40}$ ]] || die "ETHEREUM_SWEEP_FROM_ADDRESS must be a 20-byte hex address"
[[ -f "$artifact_path" ]] || die "factory artifact is missing: $artifact_path"

contract_name="$(jq -r '.contractName // empty' "$artifact_path")"
creation_code_hex="$(jq -r '.creationCodeHex // empty' "$artifact_path" | tr '[:upper:]' '[:lower:]')"

[[ "$contract_name" == "Create2ReceiverFactory" ]] || die "unexpected contractName in artifact: ${contract_name:-<empty>}"
[[ "$creation_code_hex" =~ ^0x[0-9a-f]+$ ]] || die "artifact creationCodeHex is invalid"

chain_id="$(cast chain-id --rpc-url "$ETHEREUM_SWEEP_RPC_URL" | tail -n 1 | tr -d '\r' | tr -d '[:space:]')"
[[ "$chain_id" =~ ^[0-9]+$ ]] || die "unexpected chain id output: ${chain_id:-<empty>}"
network="$(resolve_network_from_chain_id "$chain_id")"
metadata_path="${metadata_dir}/${network}.json"
[[ -f "$metadata_path" ]] || die "metadata file is missing for network ${network}: $metadata_path"

metadata_network="$(jq -r '.network // empty' "$metadata_path" | tr '[:upper:]' '[:lower:]')"
current_factory_address="$(jq -r '.factoryAddress // empty' "$metadata_path" | tr '[:upper:]' '[:lower:]')"

[[ "$metadata_network" == "$network" ]] || die "metadata network mismatch in ${metadata_path}"
[[ "$current_factory_address" =~ ^0x[0-9a-f]{40}$ ]] || die "metadata factoryAddress is invalid in ${metadata_path}"

ledger_cmd=(cast wallet address --ledger)
if [[ -n "$derivation_path" ]]; then
  ledger_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi

ledger_sender="$("${ledger_cmd[@]}" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]')"
[[ "$ledger_sender" =~ ^0x[0-9a-f]{40}$ ]] || die "Ledger sender is invalid: ${ledger_sender:-<empty>}"
[[ "$ledger_sender" == "$from_address" ]] || die "Ledger sender ${ledger_sender} does not match ETHEREUM_SWEEP_FROM_ADDRESS ${from_address}"

send_cmd=(cast send --json --rpc-url "$ETHEREUM_SWEEP_RPC_URL" --from "$from_address" --ledger)
if [[ -n "$derivation_path" ]]; then
  send_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi
send_cmd+=(--create "$creation_code_hex")

mode="dry-run"
if [[ "$broadcast" -eq 1 ]]; then
  mode="broadcast"
fi

printf 'mode: %s\n' "$mode"
printf 'network: %s\n' "$network"
printf 'chain_id: %s\n' "$chain_id"
printf 'metadata_path: %s\n' "$metadata_path"
printf 'current_factory_address: %s\n' "$current_factory_address"
printf 'artifact_path: %s\n' "$artifact_path"
printf 'contract_name: %s\n' "$contract_name"
printf 'rpc_url: %s\n' "$ETHEREUM_SWEEP_RPC_URL"
printf 'ledger_sender: %s\n' "$ledger_sender"
printf 'creation_code_hex_length: %s\n' "${#creation_code_hex}"
printf 'command:'
for arg in "${send_cmd[@]}"; do
  printf ' %q' "$arg"
done
printf '\n'

if [[ "$broadcast" -eq 1 ]]; then
  broadcast_output="$("${send_cmd[@]}")"
  printf '%s\n' "$broadcast_output"

  deployed_factory_address="$(printf '%s\n' "$broadcast_output" | jq -r 'select(type == "object") | .contractAddress // empty' | tail -n 1 | tr '[:upper:]' '[:lower:]')"
  [[ "$deployed_factory_address" =~ ^0x[0-9a-f]{40}$ ]] || die "could not read deployed factory address from cast output"

  tmp_metadata="$(mktemp)"
  jq \
    --arg address "$deployed_factory_address" \
    --arg mode "deployed" \
    '.factoryAddress = $address | .mode = $mode' \
    "$metadata_path" >"$tmp_metadata"
  mv "$tmp_metadata" "$metadata_path"

  printf 'updated_factory_address: %s\n' "$deployed_factory_address"
  printf 'updated_metadata_path: %s\n' "$metadata_path"
else
  printf 'dry-run: not broadcasting\n'
fi
