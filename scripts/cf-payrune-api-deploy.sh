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
	"BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	"BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	"BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	"BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
)

if [[ -t 1 ]]; then
	COLOR_BLUE=$'\033[1;34m'
	COLOR_CYAN=$'\033[0;36m'
	COLOR_GREEN=$'\033[0;32m'
	COLOR_YELLOW=$'\033[0;33m'
	COLOR_RED=$'\033[0;31m'
	COLOR_WHITE=$'\033[1;37m'
	COLOR_RESET=$'\033[0m'
else
	COLOR_BLUE=''
	COLOR_CYAN=''
	COLOR_GREEN=''
	COLOR_YELLOW=''
	COLOR_RED=''
	COLOR_WHITE=''
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

prompt_text() {
	printf '%s? %s%s' "$COLOR_WHITE" "$1" "$COLOR_RESET"
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

prompt_optional_value() {
	local name="$1"
	local prompt="$2"
	local value

	if [[ -n "${!name:-}" ]]; then
		info "Using $name from shell env."
		return
	fi

	if [[ ! -t 0 ]]; then
		warn "$name is not set and no interactive terminal is available. Secret sync will be skipped for it."
		return
	fi

	prompt_text "$prompt (leave blank to skip): "
	read -r value
	if [[ -n "$value" ]]; then
		printf -v "$name" '%s' "$value"
		export "$name"
		success "Captured $name for this deploy."
		return
	fi

	warn "Skipping $name secret sync."
}

prompt_yes_no() {
	local prompt="$1"
	local answer

	if [[ ! -t 0 ]]; then
		return 1
	fi

	prompt_text "$prompt [y/N]: "
	read -r answer
	case "${answer,,}" in
	y|yes) return 0 ;;
	*) return 1 ;;
	esac
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

extract_secret_args

step "Preparing deploy inputs"
info "PostgreSQL connection string controls two things in this deploy:"
info "1. run migrations before deploy"
info "2. sync the Worker secret POSTGRES_CONNECTION_STRING"
prompt_optional_value "POSTGRES_CONNECTION_STRING" "PostgreSQL connection string for migration + Worker secret"

if [[ -n "${POSTGRES_CONNECTION_STRING:-}" ]]; then
	success "Migration plan: PostgreSQL migrations WILL run before Worker deploy."
else
	warn "Migration plan: PostgreSQL migrations will be SKIPPED. Existing database schema and Worker secret will be left as-is."
fi

if prompt_yes_no "Update optional xpub/config secrets during this deploy?"; then
	for name in \
		BITCOIN_MAINNET_LEGACY_XPUB \
		BITCOIN_MAINNET_SEGWIT_XPUB \
		BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB \
		BITCOIN_MAINNET_TAPROOT_XPUB \
		BITCOIN_TESTNET4_LEGACY_XPUB \
		BITCOIN_TESTNET4_SEGWIT_XPUB \
		BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB \
		BITCOIN_TESTNET4_TAPROOT_XPUB \
		BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS \
		BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS \
		BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER \
		BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER; do
		prompt_optional_value "$name" "$name"
	done
fi

cd "$WORKER_DIR"

if [[ -n "${POSTGRES_CONNECTION_STRING:-}" ]]; then
	step "Running PostgreSQL migrations"
	(
		cd "$ROOT_DIR"
		DATABASE_URL="$POSTGRES_CONNECTION_STRING" GOCACHE=/tmp/go-build go run ./cmd/migrate up
	)
	success "PostgreSQL migrations completed."
else
	warn "Skipping PostgreSQL migration."
	warn "Skipping Worker secret sync for POSTGRES_CONNECTION_STRING."
fi

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
