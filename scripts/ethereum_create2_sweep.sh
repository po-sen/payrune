#!/usr/bin/env bash
set -euo pipefail

max_batch_size=25

usage() {
  cat <<'EOF'
Usage:
  bash scripts/ethereum_create2_sweep.sh [--dry-run|--broadcast]

Required env:
  DATABASE_URL
  ETHEREUM_SWEEP_RPC_URL
  ETHEREUM_SWEEP_FROM_ADDRESS

Selector env:
  Set exactly one selector family:
    ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS   comma-separated issued allocation ids
    ETHEREUM_SWEEP_ADDRESSES             comma-separated issued receiver addresses

Optional env:
  ETHEREUM_SWEEP_DERIVATION_PATH

Notes:
  - The script loads only explicitly selected issued allocation rows.
  - It reads `sweep_material_json` as the operator-facing recovery payload.
  - It rejects rows that are not Ethereum CREATE2 sweep material.
  - It routes recovery through the factory recorded in the selected rows' `sweep_material_json`.
  - It still loads checked-in metadata for the selected network so operators can compare the row
    factory against the network's current issuance factory.
  - It can recover predicted CREATE2 addresses that still have ETH but no deployed receiver code.
  - It rejects selected receivers whose current on-chain balance is zero.
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

display_asset_reference() {
  local value
  value="$(trim "${1:-}")"
  if [[ -n "$value" ]]; then
    printf '%s' "$value"
  else
    printf '%s' "<native>"
  fi
}

parse_uint_output() {
  local value
  value="$(trim "${1:-}")"
  value="${value%%[*}"
  value="${value%% *}"
  printf '%s' "$(trim "$value")"
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

load_metadata_for_network() {
  local network="$1"
  metadata_path="${metadata_dir}/${network}.json"
  [[ -f "$metadata_path" ]] || die "metadata file is missing for network ${network}: $metadata_path"

  metadata_network="$(jq -r '.network // empty' "$metadata_path" | tr '[:upper:]' '[:lower:]')"
  current_factory_address="$(jq -r '.factoryAddress // empty' "$metadata_path" | tr '[:upper:]' '[:lower:]')"

  [[ "$metadata_network" == "$network" ]] || die "metadata network mismatch in ${metadata_path}"
  [[ "$current_factory_address" =~ ^0x[0-9a-f]{40}$ ]] || die "metadata factoryAddress is invalid in ${metadata_path}"
}

array_contains() {
  local needle="$1"
  shift

  local item
  for item in "$@"; do
    if [[ "$item" == "$needle" ]]; then
      return 0
    fi
  done
  return 1
}

join_by() {
  local separator="$1"
  shift

  local output=""
  local item
  for item in "$@"; do
    if [[ -z "$output" ]]; then
      output="$item"
    else
      output="${output}${separator}${item}"
    fi
  done

  printf '%s' "$output"
}

parse_payment_address_ids() {
  local raw_csv="$1"
  local -a raw_items=()
  local item=""
  local normalized=""

  selector_payment_address_ids=()
  IFS=',' read -r -a raw_items <<<"$raw_csv"
  for item in "${raw_items[@]}"; do
    normalized="$(trim "$item")"
    [[ -n "$normalized" ]] || die "ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS contains an empty item"
    [[ "$normalized" =~ ^[0-9]+$ ]] || die "ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS contains a non-integer item: $normalized"
    if array_contains "$normalized" "${selector_payment_address_ids[@]}"; then
      die "ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS contains a duplicate id: $normalized"
    fi
    selector_payment_address_ids+=("$normalized")
  done

  [[ "${#selector_payment_address_ids[@]}" -gt 0 ]] || die "ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS is empty"
  [[ "${#selector_payment_address_ids[@]}" -le "$max_batch_size" ]] || die "batch exceeds max size ${max_batch_size}"
}

parse_addresses() {
  local raw_csv="$1"
  local -a raw_items=()
  local item=""
  local normalized=""

  selector_addresses=()
  IFS=',' read -r -a raw_items <<<"$raw_csv"
  for item in "${raw_items[@]}"; do
    normalized="$(normalize_hex "$item")"
    [[ -n "$normalized" ]] || die "ETHEREUM_SWEEP_ADDRESSES contains an empty item"
    [[ "$normalized" =~ ^0x[0-9a-f]{40}$ ]] || die "ETHEREUM_SWEEP_ADDRESSES contains an invalid address: $normalized"
    if array_contains "$normalized" "${selector_addresses[@]}"; then
      die "ETHEREUM_SWEEP_ADDRESSES contains a duplicate address: $normalized"
    fi
    selector_addresses+=("$normalized")
  done

  [[ "${#selector_addresses[@]}" -gt 0 ]] || die "ETHEREUM_SWEEP_ADDRESSES is empty"
  [[ "${#selector_addresses[@]}" -le "$max_batch_size" ]] || die "batch exceeds max size ${max_batch_size}"
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
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

require_command psql
require_command jq
require_command cast

require_env DATABASE_URL
require_env ETHEREUM_SWEEP_RPC_URL
require_env ETHEREUM_SWEEP_FROM_ADDRESS

payment_address_ids_raw="$(trim "${ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS:-}")"
addresses_raw="$(trim "${ETHEREUM_SWEEP_ADDRESSES:-}")"
derivation_path="$(trim "${ETHEREUM_SWEEP_DERIVATION_PATH:-}")"
from_address="$(normalize_hex "${ETHEREUM_SWEEP_FROM_ADDRESS}")"

if [[ -n "$payment_address_ids_raw" && -n "$addresses_raw" ]]; then
  die "set exactly one selector family: ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS or ETHEREUM_SWEEP_ADDRESSES"
fi
if [[ -z "$payment_address_ids_raw" && -z "$addresses_raw" ]]; then
  die "one selector family is required: ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS or ETHEREUM_SWEEP_ADDRESSES"
fi
if [[ ! "$from_address" =~ ^0x[0-9a-f]{40}$ ]]; then
  die "ETHEREUM_SWEEP_FROM_ADDRESS must be a 20-byte hex address"
fi

selector_name=""
selector_count=0
query=""
if [[ -n "$payment_address_ids_raw" ]]; then
  parse_payment_address_ids "$payment_address_ids_raw"
  selector_name="payment_address_ids"
  selector_count="${#selector_payment_address_ids[@]}"
  query="
    SELECT id,
           chain,
           network,
           scheme,
           address,
           COALESCE(sweep_material_json::text, '')
      FROM address_policy_allocations
     WHERE allocation_status = 'issued'
       AND id IN ($(join_by , "${selector_payment_address_ids[@]}"))
     ORDER BY id ASC
  "
else
  parse_addresses "$addresses_raw"
  selector_name="addresses"
  selector_count="${#selector_addresses[@]}"

  sql_address_literals=()
  for address in "${selector_addresses[@]}"; do
    sql_address_literals+=("'${address}'")
  done

  query="
    SELECT id,
           chain,
           network,
           scheme,
           address,
           COALESCE(sweep_material_json::text, '')
      FROM address_policy_allocations
     WHERE allocation_status = 'issued'
       AND lower(address) IN ($(join_by , "${sql_address_literals[@]}"))
     ORDER BY id ASC
  "
fi

mapfile -t rows < <(psql "$DATABASE_URL" -X -A -t -F $'\t' -v ON_ERROR_STOP=1 -c "$query")

if [[ ${#rows[@]} -eq 0 ]]; then
  die "no issued allocation rows matched the provided ${selector_name}"
fi
if [[ ${#rows[@]} -ne "$selector_count" ]]; then
  die "selector count ${selector_count} does not match issued row count ${#rows[@]}"
fi
if [[ ${#rows[@]} -gt "$max_batch_size" ]]; then
  die "batch exceeds max size ${max_batch_size}"
fi

selected_ids=()
selected_receivers=()
selected_receiver_balances=()
selected_receiver_code_states=()
selected_create2_salts=()
selected_init_codes=()
selected_undeployed_ids=()
selected_undeployed_receivers=()
selected_undeployed_create2_salts=()
selected_undeployed_init_codes=()
selected_network=""
selected_asset_reference=""
selected_asset_reference_initialized=0
selected_factory_address=""
metadata_path=""
metadata_network=""
current_factory_address=""
row_id_seen=()
receiver_seen=()
zero_balance_ids=()
zero_balance_receivers=()

for row in "${rows[@]}"; do
  IFS=$'\t' read -r row_id row_chain row_network row_scheme row_address sweep_material_json <<<"$row"

  row_id="$(trim "$row_id")"
  row_chain="$(trim "$row_chain")"
  row_network="$(trim "$row_network")"
  row_scheme="$(trim "$row_scheme")"
  row_address="$(normalize_hex "$row_address")"
  sweep_material_json="$(trim "$sweep_material_json")"

  [[ -n "$row_id" ]] || die "selected row is missing id"
  [[ "$row_id" =~ ^[0-9]+$ ]] || die "selected row has invalid id: ${row_id:-<empty>}"
  [[ "$row_chain" == "ethereum" ]] || die "selected row ${row_id} is not ethereum: ${row_chain:-<empty>}"
  [[ "$row_scheme" == "create2" ]] || die "selected row ${row_id} is not create2: ${row_scheme:-<empty>}"
  [[ "$row_address" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row ${row_id} address is invalid"
  [[ -n "$sweep_material_json" ]] || die "selected row ${row_id} has empty sweep_material_json"

  if array_contains "$row_id" "${row_id_seen[@]}"; then
    die "selected rows contain duplicate payment_address_id: $row_id"
  fi
  row_id_seen+=("$row_id")

  if array_contains "$row_address" "${receiver_seen[@]}"; then
    die "selected rows contain duplicate receiver address: $row_address"
  fi
  receiver_seen+=("$row_address")

  printf '%s' "$sweep_material_json" | jq -e . >/dev/null || die "selected row ${row_id} has invalid sweep_material_json"

  material_type="$(printf '%s' "$sweep_material_json" | jq -r '.material_type // empty')"
  material_version="$(printf '%s' "$sweep_material_json" | jq -r '.material_version // empty')"
  material_chain="$(printf '%s' "$sweep_material_json" | jq -r '.chain // empty')"
  material_network="$(printf '%s' "$sweep_material_json" | jq -r '.network // empty')"
  material_asset_reference="$(printf '%s' "$sweep_material_json" | jq -r '.asset_reference // empty' | tr '[:upper:]' '[:lower:]')"
  material_address="$(printf '%s' "$sweep_material_json" | jq -r '.address // empty' | tr '[:upper:]' '[:lower:]')"
  predicted_address="$(printf '%s' "$sweep_material_json" | jq -r '.predicted_address // empty' | tr '[:upper:]' '[:lower:]')"
  factory_address="$(printf '%s' "$sweep_material_json" | jq -r '.factory_address // empty' | tr '[:upper:]' '[:lower:]')"
  collector_address="$(printf '%s' "$sweep_material_json" | jq -r '.collector_address // empty' | tr '[:upper:]' '[:lower:]')"
  create2_salt="$(printf '%s' "$sweep_material_json" | jq -r '.create2_salt // empty' | tr '[:upper:]' '[:lower:]')"
  init_code_hex="$(printf '%s' "$sweep_material_json" | jq -r '.init_code_hex // empty' | tr '[:upper:]' '[:lower:]')"
  init_code_hash="$(printf '%s' "$sweep_material_json" | jq -r '.init_code_hash // empty' | tr '[:upper:]' '[:lower:]')"

  [[ "$material_type" == "ethereum_create2" ]] || die "selected row ${row_id} has unexpected material_type: ${material_type:-<empty>}"
  [[ "$material_version" == "1" ]] || die "selected row ${row_id} has unexpected material_version: ${material_version:-<empty>}"
  [[ "$material_chain" == "ethereum" ]] || die "selected row ${row_id} has unexpected material chain: ${material_chain:-<empty>}"
  [[ "$material_network" == "$row_network" ]] || die "selected row ${row_id} material network does not match row network"
  normalized_material_asset_reference=""
  if [[ -n "$material_asset_reference" ]]; then
    [[ "$material_asset_reference" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row ${row_id} has invalid asset_reference"
    normalized_material_asset_reference="$material_asset_reference"
  fi
  [[ "$material_address" == "$row_address" ]] || die "selected row ${row_id} material address does not match row address"
  [[ "$predicted_address" == "$row_address" ]] || die "selected row ${row_id} predicted address does not match row address"
  [[ "$factory_address" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row ${row_id} has invalid factory_address"
  [[ "$collector_address" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row ${row_id} has invalid collector_address"
  [[ "$create2_salt" =~ ^0x[0-9a-f]{64}$ ]] || die "selected row ${row_id} has invalid create2_salt"
  [[ "$init_code_hex" =~ ^0x[0-9a-f]+$ ]] || die "selected row ${row_id} has invalid init_code_hex"
  [[ "$init_code_hash" =~ ^0x[0-9a-f]{64}$ ]] || die "selected row ${row_id} has invalid init_code_hash"

  if ! computed_init_code_hash="$(cast keccak "$init_code_hex" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')"; then
    die "failed to recompute init_code_hash for selected row ${row_id}"
  fi
  [[ "$computed_init_code_hash" =~ ^0x[0-9a-f]{64}$ ]] || die "selected row ${row_id} returned invalid recomputed init_code_hash"
  [[ "$computed_init_code_hash" == "$init_code_hash" ]] || die "selected row ${row_id} init_code_hash does not match init_code_hex"

  if ! computed_receiver_address="$(cast create2 --deployer "$factory_address" --salt "$create2_salt" --init-code-hash "$init_code_hash" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')"; then
    die "failed to recompute CREATE2 receiver address for selected row ${row_id}"
  fi
  [[ "$computed_receiver_address" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row ${row_id} returned invalid computed CREATE2 receiver address"
  [[ "$computed_receiver_address" == "$row_address" ]] || die "selected row ${row_id} computed CREATE2 receiver address does not match row address"

  if [[ -z "$selected_network" ]]; then
    selected_network="$row_network"
    load_metadata_for_network "$selected_network"
  elif [[ "$row_network" != "$selected_network" ]]; then
    die "selected rows span multiple networks: ${selected_network} and ${row_network}"
  fi
  if [[ "$selected_asset_reference_initialized" -eq 0 ]]; then
    selected_asset_reference="$normalized_material_asset_reference"
    selected_asset_reference_initialized=1
  elif [[ "$normalized_material_asset_reference" != "$selected_asset_reference" ]]; then
    die "selected rows mix asset references: $(display_asset_reference "$selected_asset_reference") and $(display_asset_reference "$normalized_material_asset_reference"); split native and token rows into separate sweeps"
  fi
  if [[ -z "$selected_factory_address" ]]; then
    selected_factory_address="$factory_address"
  elif [[ "$factory_address" != "$selected_factory_address" ]]; then
    die "selected rows span multiple factory addresses: ${selected_factory_address} and ${factory_address}"
  fi

  balance_value=""
  balance_label=""
  if [[ -n "$normalized_material_asset_reference" ]]; then
    if ! balance_value="$(cast call "$normalized_material_asset_reference" 'balanceOf(address)(uint256)' "$row_address" --rpc-url "$ETHEREUM_SWEEP_RPC_URL" | tail -n 1 | tr -d '\r')"; then
      die "failed to query token balance for selected row ${row_id}"
    fi
    balance_value="$(parse_uint_output "$balance_value" | tr -d '[:space:]')"
    balance_label="token_minor"
  else
    if ! balance_value="$(cast balance --rpc-url "$ETHEREUM_SWEEP_RPC_URL" "$row_address" | tail -n 1 | tr -d '\r' | tr -d '[:space:]')"; then
      die "failed to query on-chain balance for selected row ${row_id}"
    fi
    balance_label="wei"
  fi
  [[ "$balance_value" =~ ^[0-9]+$ ]] || die "selected row ${row_id} returned invalid ${balance_label} balance: ${balance_value:-<empty>}"
  if [[ "$balance_value" == "0" ]]; then
    zero_balance_ids+=("$row_id")
    zero_balance_receivers+=("$row_address")
  fi

  if ! receiver_code_hex="$(cast code "$row_address" --rpc-url "$ETHEREUM_SWEEP_RPC_URL" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')"; then
    die "failed to query receiver code for selected row ${row_id}"
  fi
  [[ "$receiver_code_hex" =~ ^0x[0-9a-f]*$ ]] || die "selected row ${row_id} returned invalid receiver code"
  if [[ "$receiver_code_hex" == "0x" ]]; then
    selected_receiver_code_states+=("${row_id}:undeployed")
    selected_undeployed_ids+=("$row_id")
    selected_undeployed_receivers+=("$row_address")
    selected_undeployed_create2_salts+=("$create2_salt")
    selected_undeployed_init_codes+=("$init_code_hex")
  else
    if ! deployed_collector="$(cast call "$row_address" 'collector()(address)' --rpc-url "$ETHEREUM_SWEEP_RPC_URL" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')"; then
      die "selected row ${row_id} receiver contract does not expose collector()"
    fi
    [[ "$deployed_collector" =~ ^0x[0-9a-f]{40}$ ]] || die "selected row ${row_id} receiver contract returned invalid collector()"
    [[ "$deployed_collector" == "$collector_address" ]] || die "selected row ${row_id} receiver collector() does not match recorded collector_address"
    selected_receiver_code_states+=("${row_id}:deployed")
  fi

  selected_ids+=("$row_id")
  selected_receivers+=("$row_address")
  selected_receiver_balances+=("${row_id}:${balance_value}")
  selected_create2_salts+=("$create2_salt")
  selected_init_codes+=("$init_code_hex")
done

if [[ "${#zero_balance_ids[@]}" -gt 0 ]]; then
  die "selected rows include zero-balance receivers; remove them before sweep. payment_address_ids=[$(join_by , "${zero_balance_ids[@]}")] receiver_addresses=[$(join_by , "${zero_balance_receivers[@]}")]"
fi

if ! factory_code_hex="$(cast code "$selected_factory_address" --rpc-url "$ETHEREUM_SWEEP_RPC_URL" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')"; then
  die "failed to query factory code for selected factory ${selected_factory_address}"
fi
[[ "$factory_code_hex" =~ ^0x[0-9a-f]*$ ]] || die "selected factory ${selected_factory_address} returned invalid code"
[[ "$factory_code_hex" != "0x" ]] || die "selected factory ${selected_factory_address} is not deployed on-chain"

ledger_cmd=(cast wallet address --ledger)
if [[ -n "$derivation_path" ]]; then
  ledger_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi

ledger_sender="$("${ledger_cmd[@]}" | tail -n 1 | tr -d '\r' | tr '[:upper:]' '[:lower:]')"
[[ "$ledger_sender" =~ ^0x[0-9a-f]{40}$ ]] || die "Ledger sender is invalid: ${ledger_sender:-<empty>}"
[[ "$ledger_sender" == "$from_address" ]] || die "Ledger sender ${ledger_sender} does not match ETHEREUM_SWEEP_FROM_ADDRESS ${from_address}"

send_cmd=(cast send --rpc-url "$ETHEREUM_SWEEP_RPC_URL" --from "$from_address" --ledger)
if [[ -n "$derivation_path" ]]; then
  send_cmd+=(--mnemonic-derivation-path "$derivation_path")
fi

mode="dry-run"
if [[ "$broadcast" -eq 1 ]]; then
  mode="broadcast"
fi

printf 'mode: %s\n' "$mode"
printf 'selector: %s\n' "$selector_name"
printf 'network: %s\n' "$selected_network"
if [[ -n "$selected_asset_reference" ]]; then
  printf 'asset_reference: %s\n' "$selected_asset_reference"
else
  printf 'asset_reference: %s\n' "<native>"
fi
printf 'metadata_path: %s\n' "$metadata_path"
printf 'metadata_factory_address: %s\n' "$current_factory_address"
printf 'selected_factory_address: %s\n' "$selected_factory_address"
if [[ "$selected_factory_address" == "$current_factory_address" ]]; then
  printf 'factory_selection_source: %s\n' "current_metadata"
else
  printf 'factory_selection_source: %s\n' "row_material"
fi
printf 'payment_address_ids: [%s]\n' "$(join_by , "${selected_ids[@]}")"
printf 'receiver_count: %s\n' "${#selected_receivers[@]}"
printf 'receiver_addresses: [%s]\n' "$(join_by , "${selected_receivers[@]}")"
if [[ -n "$selected_asset_reference" ]]; then
  printf 'receiver_token_balances_minor: [%s]\n' "$(join_by , "${selected_receiver_balances[@]}")"
else
  printf 'receiver_balances_wei: [%s]\n' "$(join_by , "${selected_receiver_balances[@]}")"
fi
printf 'receiver_code_states: [%s]\n' "$(join_by , "${selected_receiver_code_states[@]}")"
printf 'ledger_sender: %s\n' "$ledger_sender"

if [[ -n "$selected_asset_reference" ]]; then
  recovery_path="create2_batch_token_sweep"
  call_signature="sweepERC20(bytes32[],bytes[],address)"
  salt_arg="[$(join_by , "${selected_create2_salts[@]}")]"
  init_code_arg="[$(join_by , "${selected_init_codes[@]}")]"
  send_cmd+=("$selected_factory_address" "$call_signature" "$salt_arg" "$init_code_arg" "$selected_asset_reference")

  printf 'recovery_path: %s\n' "$recovery_path"
  printf 'call_signature: %s\n' "$call_signature"
  if [[ "${#selected_undeployed_ids[@]}" -gt 0 ]]; then
    printf 'deploy_payment_address_ids: [%s]\n' "$(join_by , "${selected_undeployed_ids[@]}")"
    printf 'deploy_receiver_addresses: [%s]\n' "$(join_by , "${selected_undeployed_receivers[@]}")"
  fi
  printf 'signature_count_estimate: %s\n' "1"
  printf 'command:'
  for arg in "${send_cmd[@]}"; do
    printf ' %q' "$arg"
  done
  printf '\n'

  if [[ "$broadcast" -eq 1 ]]; then
    "${send_cmd[@]}"
  fi
else
  recovery_path="create2_batch_sweep"
  call_signature="sweep(bytes32[],bytes[])"
  salt_arg="[$(join_by , "${selected_create2_salts[@]}")]"
  init_code_arg="[$(join_by , "${selected_init_codes[@]}")]"
  send_cmd+=("$selected_factory_address" "$call_signature" "$salt_arg" "$init_code_arg")

  printf 'recovery_path: %s\n' "$recovery_path"
  printf 'call_signature: %s\n' "$call_signature"
  printf 'signature_count_estimate: %s\n' "1"
  printf 'command:'
  for arg in "${send_cmd[@]}"; do
    printf ' %q' "$arg"
  done
  printf '\n'

  if [[ "$broadcast" -eq 1 ]]; then
    "${send_cmd[@]}"
  fi
fi

if [[ "$broadcast" -eq 0 ]]; then
  printf 'dry-run: not broadcasting\n'
fi
