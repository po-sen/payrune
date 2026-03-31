#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash scripts/ethereum_create2_sweep.sh [--dry-run|--broadcast]

Required env:
  DATABASE_URL
  ETHEREUM_SWEEP_RPC_URL
  ETHEREUM_SWEEP_FROM_ADDRESS

Selector env:
  Set exactly one of:
    ETHEREUM_SWEEP_PAYMENT_ADDRESS_ID
    ETHEREUM_SWEEP_ADDRESS

Optional env:
  ETHEREUM_SWEEP_DERIVATION_PATH

Notes:
  - The script loads one issued allocation row from `address_policy_allocations`.
  - It reads `sweep_material_json` as the only operator-facing recovery payload.
  - It rejects rows that are not Ethereum CREATE2 sweep material.
  - It checks the connected Ledger sender before deciding whether to broadcast.
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

require_command psql
require_command jq
require_command cast

require_env DATABASE_URL
require_env ETHEREUM_SWEEP_RPC_URL
require_env ETHEREUM_SWEEP_FROM_ADDRESS

payment_address_id="$(trim "${ETHEREUM_SWEEP_PAYMENT_ADDRESS_ID:-}")"
address_selector="$(trim "${ETHEREUM_SWEEP_ADDRESS:-}")"
derivation_path="$(trim "${ETHEREUM_SWEEP_DERIVATION_PATH:-}")"
from_address="$(normalize_hex "${ETHEREUM_SWEEP_FROM_ADDRESS}")"

if [[ -n "$payment_address_id" && -n "$address_selector" ]]; then
  die "set exactly one selector: ETHEREUM_SWEEP_PAYMENT_ADDRESS_ID or ETHEREUM_SWEEP_ADDRESS"
fi
if [[ -z "$payment_address_id" && -z "$address_selector" ]]; then
  die "one selector is required: ETHEREUM_SWEEP_PAYMENT_ADDRESS_ID or ETHEREUM_SWEEP_ADDRESS"
fi

if [[ ! "$from_address" =~ ^0x[0-9a-f]{40}$ ]]; then
  die "ETHEREUM_SWEEP_FROM_ADDRESS must be a 20-byte hex address"
fi

selector_name=""
selector_value=""
query=""
if [[ -n "$payment_address_id" ]]; then
  [[ "$payment_address_id" =~ ^[0-9]+$ ]] || die "ETHEREUM_SWEEP_PAYMENT_ADDRESS_ID must be an integer"
  selector_name="payment_address_id"
  selector_value="$payment_address_id"
  query="
    SELECT id,
           chain,
           network,
           scheme,
           address,
           COALESCE(sweep_material_json::text, '')
      FROM address_policy_allocations
     WHERE allocation_status = 'issued'
       AND id = ${payment_address_id}
     LIMIT 2
  "
else
  address_selector="$(normalize_hex "$address_selector")"
  [[ "$address_selector" =~ ^0x[0-9a-f]{40}$ ]] || die "ETHEREUM_SWEEP_ADDRESS must be a 20-byte hex address"
  selector_name="address"
  selector_value="$address_selector"
  query="
    SELECT id,
           chain,
           network,
           scheme,
           address,
           COALESCE(sweep_material_json::text, '')
      FROM address_policy_allocations
     WHERE allocation_status = 'issued'
       AND lower(address) = '${address_selector}'
     LIMIT 2
  "
fi

mapfile -t rows < <(psql "$DATABASE_URL" -X -A -t -F $'\t' -v ON_ERROR_STOP=1 -c "$query")

if [[ ${#rows[@]} -eq 0 ]]; then
  die "no issued allocation row matched ${selector_name}=${selector_value}"
fi
if [[ ${#rows[@]} -gt 1 ]]; then
  die "selector ${selector_name}=${selector_value} matched multiple issued allocation rows"
fi

IFS=$'\t' read -r row_id row_chain row_network row_scheme row_address sweep_material_json <<<"${rows[0]}"

row_id="$(trim "$row_id")"
row_chain="$(trim "$row_chain")"
row_network="$(trim "$row_network")"
row_scheme="$(trim "$row_scheme")"
row_address="$(normalize_hex "$row_address")"
sweep_material_json="$(trim "$sweep_material_json")"

[[ -n "$row_id" ]] || die "selected row is missing id"
[[ "$row_chain" == "ethereum" ]] || die "selected row is not ethereum: ${row_chain:-<empty>}"
[[ "$row_scheme" == "create2" ]] || die "selected row is not create2: ${row_scheme:-<empty>}"
[[ "$row_address" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row address is invalid"
[[ -n "$sweep_material_json" ]] || die "selected row has empty sweep_material_json"

printf '%s' "$sweep_material_json" | jq -e . >/dev/null || die "selected row has invalid sweep_material_json"

material_type="$(printf '%s' "$sweep_material_json" | jq -r '.material_type // empty')"
material_version="$(printf '%s' "$sweep_material_json" | jq -r '.material_version // empty')"
material_chain="$(printf '%s' "$sweep_material_json" | jq -r '.chain // empty')"
material_network="$(printf '%s' "$sweep_material_json" | jq -r '.network // empty')"
material_address="$(printf '%s' "$sweep_material_json" | jq -r '.address // empty' | tr '[:upper:]' '[:lower:]')"
predicted_address="$(printf '%s' "$sweep_material_json" | jq -r '.predicted_address // empty' | tr '[:upper:]' '[:lower:]')"
factory_address="$(printf '%s' "$sweep_material_json" | jq -r '.factory_address // empty' | tr '[:upper:]' '[:lower:]')"
collector_address="$(printf '%s' "$sweep_material_json" | jq -r '.collector_address // empty' | tr '[:upper:]' '[:lower:]')"
create2_salt="$(printf '%s' "$sweep_material_json" | jq -r '.create2_salt // empty' | tr '[:upper:]' '[:lower:]')"
init_code_hex="$(printf '%s' "$sweep_material_json" | jq -r '.init_code_hex // empty' | tr '[:upper:]' '[:lower:]')"
init_code_hash="$(printf '%s' "$sweep_material_json" | jq -r '.init_code_hash // empty' | tr '[:upper:]' '[:lower:]')"

[[ "$material_type" == "ethereum_create2" ]] || die "unexpected material_type: ${material_type:-<empty>}"
[[ "$material_version" == "1" ]] || die "unexpected material_version: ${material_version:-<empty>}"
[[ "$material_chain" == "ethereum" ]] || die "unexpected material chain: ${material_chain:-<empty>}"
[[ "$material_network" == "$row_network" ]] || die "material network does not match row network"
[[ "$material_address" == "$row_address" ]] || die "material address does not match row address"
[[ "$predicted_address" == "$row_address" ]] || die "predicted address does not match row address"
[[ "$factory_address" =~ ^0x[0-9a-f]{40}$ ]] || die "factory_address is invalid"
[[ "$collector_address" =~ ^0x[0-9a-f]{40}$ ]] || die "collector_address is invalid"
[[ "$create2_salt" =~ ^0x[0-9a-f]{64}$ ]] || die "create2_salt is invalid"
[[ "$init_code_hex" =~ ^0x[0-9a-f]+$ ]] || die "init_code_hex is invalid"
[[ "$init_code_hash" =~ ^0x[0-9a-f]{64}$ ]] || die "init_code_hash is invalid"

ledger_cmd=(cast wallet address --ledger)
if [[ -n "$derivation_path" ]]; then
  ledger_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi

ledger_sender="$("${ledger_cmd[@]}" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]')"
[[ "$ledger_sender" =~ ^0x[0-9a-f]{40}$ ]] || die "Ledger sender is invalid: ${ledger_sender:-<empty>}"
[[ "$ledger_sender" == "$from_address" ]] || die "Ledger sender ${ledger_sender} does not match ETHEREUM_SWEEP_FROM_ADDRESS ${from_address}"

send_cmd=(cast send "$predicted_address" "sweep()" --rpc-url "$ETHEREUM_SWEEP_RPC_URL" --from "$from_address" --ledger)
if [[ -n "$derivation_path" ]]; then
  send_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi

mode="dry-run"
if [[ "$broadcast" -eq 1 ]]; then
  mode="broadcast"
fi

printf 'mode: %s\n' "$mode"
printf 'selector: %s=%s\n' "$selector_name" "$selector_value"
printf 'payment_address_id: %s\n' "$row_id"
printf 'network: %s\n' "$row_network"
printf 'receiver_address: %s\n' "$row_address"
printf 'ledger_sender: %s\n' "$ledger_sender"
printf 'sweep_material_json: %s\n' "$(printf '%s' "$sweep_material_json" | jq -c .)"
printf 'command:'
for arg in "${send_cmd[@]}"; do
  printf ' %q' "$arg"
done
printf '\n'

if [[ "$broadcast" -eq 1 ]]; then
  "${send_cmd[@]}"
else
  printf 'dry-run: not broadcasting\n'
fi
