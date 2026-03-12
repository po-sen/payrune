#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

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
	COLOR_RED=$'\033[0;31m'
	COLOR_RESET=$'\033[0m'
else
	COLOR_BLUE=''
	COLOR_CYAN=''
	COLOR_GREEN=''
	COLOR_RED=''
	COLOR_RESET=''
fi

step() {
	printf '%s==> %s%s\n' "$COLOR_BLUE" "$1" "$COLOR_RESET"
}

info() {
	printf '%s%s%s\n' "$COLOR_CYAN" "$1" "$COLOR_RESET"
}

success() {
	printf '%sOK: %s%s\n' "$COLOR_GREEN" "$1" "$COLOR_RESET"
}

fail() {
	printf '%sERROR: %s%s\n' "$COLOR_RED" "$1" "$COLOR_RESET" >&2
	exit 1
}

require_env() {
	local name="$1"

	if [[ -n "${!name:-}" ]]; then
		info "Using $name from shell env or .env.cloudflare."
		return
	fi

	fail "$name is required. Set it in shell env or .env.cloudflare."
}

step "Preparing PostgreSQL migration"
info "Auto-loading .env.cloudflare when present."
require_env "POSTGRES_CONNECTION_STRING"

step "Running PostgreSQL migrations"
(
	cd "$ROOT_DIR"
	DATABASE_URL="$POSTGRES_CONNECTION_STRING" GOCACHE=/tmp/go-build go run ./cmd/migrate up
)
success "PostgreSQL migrations completed."
