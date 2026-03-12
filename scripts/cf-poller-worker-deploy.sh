#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKER_DIR="$ROOT_DIR/deployments/cloudflare/payrune-poller"

load_root_cloudflare_env() {
	local env_file="$1"
	local line key value

	if [[ ! -f "$env_file" ]]; then
		return
	fi

	while IFS= read -r line || [[ -n "$line" ]]; do
		line="${line#"${line%%[![:space:]]*}"}"
		line="${line%"${line##*[![:space:]]}"}"
		if [[ -z "$line" || "${line:0:1}" == "#" || "$line" != *=* ]]; then
			continue
		fi

		key="${line%%=*}"
		value="${line#*=}"
		key="${key#"${key%%[![:space:]]*}"}"
		key="${key%"${key##*[![:space:]]}"}"
		value="${value#"${value%%[![:space:]]*}"}"
		value="${value%"${value##*[![:space:]]}"}"

		if [[ -z "$key" || -n "${!key+x}" ]]; then
			continue
		fi
		if [[ "${value:0:1}" == '"' && "${value: -1}" == '"' ]]; then
			value="${value:1:${#value}-2}"
		elif [[ "${value:0:1}" == "'" && "${value: -1}" == "'" ]]; then
			value="${value:1:${#value}-2}"
		fi

		printf -v "$key" '%s' "$value"
		export "$key"
	done <"$env_file"
}

load_root_cloudflare_env "$ROOT_DIR/.env.cloudflare"

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

if [[ -t 1 ]]; then
	COLOR_BLUE=$'\033[1;34m'
	COLOR_CYAN=$'\033[0;36m'
	COLOR_GREEN=$'\033[0;32m'
	COLOR_YELLOW=$'\033[0;33m'
	COLOR_RED=$'\033[0;31m'
	COLOR_RESET=$'\033[0m'
else
	COLOR_BLUE=''
	COLOR_CYAN=''
	COLOR_GREEN=''
	COLOR_YELLOW=''
	COLOR_RED=''
	COLOR_RESET=''
fi

step() {
	printf '%s==> %s%s\n' "$COLOR_BLUE" "$1" "$COLOR_RESET"
}

info() {
	printf '%s%s%s\n' "$COLOR_CYAN" "$1" "$COLOR_RESET"
}

warn() {
	printf '%sWARN: %s%s\n' "$COLOR_YELLOW" "$1" "$COLOR_RESET"
}

success() {
	printf '%sOK: %s%s\n' "$COLOR_GREEN" "$1" "$COLOR_RESET"
}

fail() {
	printf '%sERROR: %s%s\n' "$COLOR_RED" "$1" "$COLOR_RESET" >&2
	exit 1
}

sync_secret_from_env() {
	local name="$1"
	local value="${!name:-}"

	if [[ -z "$value" ]]; then
		return
	fi

	info "Syncing Wrangler secret: $name"
	printf '%s' "$value" | npm exec -- wrangler secret put "$name" --env "$TARGET_NETWORK"
	success "Wrangler secret synced: $name"
}

require_env() {
	local name="$1"

	if [[ -n "${!name:-}" ]]; then
		info "Using $name from shell env or .env.cloudflare."
		return
	fi

	fail "$name is required. Set it in shell env or .env.cloudflare."
}

poller_secret_names=(
	"POSTGRES_CONNECTION_STRING"
)

if [[ "$TARGET_NETWORK" == "mainnet" ]]; then
	network_secret_names=(
		"BITCOIN_MAINNET_ESPLORA_USER"
		"BITCOIN_MAINNET_ESPLORA_PASSWORD"
	)
else
	network_secret_names=(
		"BITCOIN_TESTNET4_ESPLORA_USER"
		"BITCOIN_TESTNET4_ESPLORA_PASSWORD"
	)
fi

step "Preparing ${TARGET_NETWORK} poller deploy inputs"
info "Auto-loading .env.cloudflare when present."
info "POSTGRES_CONNECTION_STRING is required here to sync the Worker secret."
info "Run make cf-migrate separately when you need database migrations."
require_env "POSTGRES_CONNECTION_STRING"
success "Worker secret plan: POSTGRES_CONNECTION_STRING WILL be synced."

cd "$WORKER_DIR"

step "Installing Worker deployment dependencies"
npm install
success "Worker deployment dependencies installed."

step "Checking Worker source"
npm run check
success "Worker source check passed."

step "Running Worker tests"
npm test
success "Worker tests passed."

step "Syncing selected Worker secrets"
for name in "${poller_secret_names[@]}" "${network_secret_names[@]}"; do
	sync_secret_from_env "$name"
done

step "Deploying ${TARGET_NETWORK} poller Worker"
npm exec -- wrangler deploy --env "$TARGET_NETWORK"
success "${TARGET_NETWORK} poller deploy finished."
