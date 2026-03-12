#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKER_DIR="$ROOT_DIR/deployments/cloudflare/payrune-api"
DEPLOY_ARGS=("$@")
SECRET_ARGS=()
SYNC_VARS=(
	"POSTGRES_CONNECTION_STRING"
	"BITCOIN_MAINNET_LEGACY_XPUB"
	"BITCOIN_MAINNET_SEGWIT_XPUB"
	"BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"
	"BITCOIN_MAINNET_TAPROOT_XPUB"
	"BITCOIN_TESTNET4_LEGACY_XPUB"
	"BITCOIN_TESTNET4_SEGWIT_XPUB"
	"BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"
	"BITCOIN_TESTNET4_TAPROOT_XPUB"
)

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

extract_secret_args() {
	local i arg

	for ((i = 0; i < ${#DEPLOY_ARGS[@]}; i++)); do
		arg="${DEPLOY_ARGS[$i]}"

		case "$arg" in
		--env)
			if ((i + 1 >= ${#DEPLOY_ARGS[@]})); then
				fail "--env requires a value"
			fi
			SECRET_ARGS+=("$arg" "${DEPLOY_ARGS[$((i + 1))]}")
			((i++))
			;;
		--env=*)
			SECRET_ARGS+=("$arg")
			;;
		esac
	done
}

sync_secret_from_env() {
	local name="$1"
	local value="${!name:-}"

	if [[ -z "$value" ]]; then
		return
	fi

	info "Syncing Wrangler secret: $name"
	printf '%s' "$value" | npm exec -- wrangler secret put "$name" "${SECRET_ARGS[@]}"
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

extract_secret_args

step "Preparing deploy inputs"
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
for name in "${SYNC_VARS[@]}"; do
	sync_secret_from_env "$name"
done

step "Deploying Worker"
npm exec -- wrangler deploy "${DEPLOY_ARGS[@]}"
success "Worker deploy finished."
