#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash scripts/ethereum_create2_build_artifacts.sh [--solc-image REF] [--solc-platform PLATFORM]

Options:
  --solc-image REF      docker image ref for solc
  --solc-platform REF   docker platform for solc

Notes:
  - Rebuilds the checked-in CREATE2 Solidity artifacts under
    `internal/infrastructure/ethereumcreate2assets/artifacts/`.
  - This is a repo maintenance helper. Operators do not need it for deploy or sweep.
EOF
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_command() {
  local name="$1"
  command -v "$name" >/dev/null 2>&1 || die "$name is required"
}

join_by() {
  local separator="$1"
  shift

  local output=""
  local item=""
  for item in "$@"; do
    if [[ -z "$output" ]]; then
      output="$item"
    else
      output="${output}${separator}${item}"
    fi
  done

  printf '%s' "$output"
}

default_solc_image_ref="ghcr.io/argotorg/solc@sha256:6263a14716bf74f01cc80e86e0fcd28a5bae4d4aca46cc8aa6f4c2d6608ab143"
default_solc_platform="linux/amd64"

solc_image_ref="$default_solc_image_ref"
solc_platform="$default_solc_platform"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --solc-image)
      [[ $# -ge 2 ]] || die "--solc-image requires a value"
      solc_image_ref="$2"
      shift 2
      ;;
    --solc-platform)
      [[ $# -ge 2 ]] || die "--solc-platform requires a value"
      solc_platform="$2"
      shift 2
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

require_command docker
require_command jq

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
assets_dir="${repo_root}/internal/infrastructure/ethereumcreate2assets"
contracts_dir="${assets_dir}/contracts"
artifacts_dir="${assets_dir}/artifacts"

[[ -d "$contracts_dir" ]] || die "contracts directory is missing: $contracts_dir"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

solc_input_path="${tmp_dir}/solc-input.json"
solc_output_path="${tmp_dir}/solc-output.json"
solc_stderr_path="${tmp_dir}/solc-stderr.txt"

sources_json='{}'
mapfile -t contract_files < <(find "$contracts_dir" -maxdepth 1 -type f -name '*.sol' | sort)
[[ "${#contract_files[@]}" -gt 0 ]] || die "no Solidity sources found in $contracts_dir"

for contract_file in "${contract_files[@]}"; do
  source_name="$(basename "$contract_file")"
  sources_json="$(
    jq \
      --arg source_name "$source_name" \
      --rawfile content "$contract_file" \
      '. + {($source_name): {content: $content}}' <<<"$sources_json"
  )"
done

jq -n \
  --argjson sources "$sources_json" \
  '{
    language: "Solidity",
    sources: $sources,
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      },
      outputSelection: {
        "*": {
          "*": [
            "abi",
            "evm.bytecode.object",
            "evm.deployedBytecode.object"
          ]
        }
      }
    }
  }' >"$solc_input_path"

docker_args=(run --rm -i)
if [[ -n "${solc_platform}" ]]; then
  docker_args+=(--platform "$solc_platform")
fi

if ! docker "${docker_args[@]}" "$solc_image_ref" --standard-json <"$solc_input_path" >"$solc_output_path" 2>"$solc_stderr_path"; then
  stderr_output="$(tr -d '\r' <"$solc_stderr_path")"
  if [[ -n "$stderr_output" ]]; then
    die "$(printf 'compile solidity contracts failed\n%s' "$stderr_output")"
  fi
  die "compile solidity contracts failed"
fi

warning_messages="$(
  jq -r '
    [.errors[]? | select(
      (.formattedMessage // "" | type == "string") and
      ((.formattedMessage // "") | length > 0) and
      ((.severity // "") | ascii_downcase) != "error"
    ) | .formattedMessage] | join("\n")
  ' "$solc_output_path"
)"
if [[ -n "$warning_messages" ]]; then
  printf '%s\n' "$warning_messages" >&2
fi

fatal_messages="$(
  jq -r '
    [.errors[]? | select(
      (.formattedMessage // "" | type == "string") and
      ((.formattedMessage // "") | length > 0) and
      ((.severity // "") | ascii_downcase) == "error"
    ) | .formattedMessage] | join("\n")
  ' "$solc_output_path"
)"
if [[ -n "$fatal_messages" ]]; then
  die "$(printf 'solidity compile errors:\n%s' "$fatal_messages")"
fi

version_args=(run --rm)
if [[ -n "${solc_platform}" ]]; then
  version_args+=(--platform "$solc_platform")
fi
version_args+=("$solc_image_ref" --version)

if ! compiler_version_output="$(docker "${version_args[@]}" 2>&1)"; then
  die "$(printf 'read solc compiler version failed\n%s' "$compiler_version_output")"
fi

compiler_version="$(
  printf '%s\n' "$compiler_version_output" |
    awk -F'Version:' '/Version:/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' |
    tail -n 1
)"
[[ -n "$compiler_version" ]] || die "read solc compiler version failed: missing version line"

mkdir -p "$artifacts_dir"

artifact_definitions=(
  "Create2ReceiverFactory.sol|Create2ReceiverFactory|Create2ReceiverFactoryV1.json"
  "FixedCollectorReceiver.sol|FixedCollectorReceiver|FixedCollectorReceiverV1.json"
)
built_contract_names=()

for definition in "${artifact_definitions[@]}"; do
  IFS='|' read -r source_name contract_name file_name <<<"$definition"
  artifact_path="${artifacts_dir}/${file_name}"
  built_contract_names+=("$contract_name")

  if ! jq \
    --arg source_name "$source_name" \
    --arg contract_name "$contract_name" \
    --arg compiler_version "$compiler_version" \
    '
      .contracts[$source_name][$contract_name] as $compiled
      | if $compiled == null then
          error("missing compiled contract " + $contract_name + " from " + $source_name)
        else
          .
        end
      | ($compiled.evm.bytecode.object // "") as $creation_code
      | ($compiled.evm.deployedBytecode.object // "") as $runtime_code
      | if ($creation_code | type) != "string" or ($runtime_code | type) != "string" or ($creation_code | length) == 0 or ($runtime_code | length) == 0 then
          error("compiled contract " + $contract_name + " is missing bytecode")
        else
          {
            sourceName: $source_name,
            contractName: $contract_name,
            compilerVersion: $compiler_version,
            abi: $compiled.abi,
            creationCodeHex: ("0x" + $creation_code),
            runtimeCodeHex: ("0x" + $runtime_code)
          }
        end
    ' "$solc_output_path" >"$artifact_path"; then
    die "failed to write artifact: ${file_name}"
  fi
done

printf 'rebuilt_artifacts_dir: %s\n' "$artifacts_dir"
printf 'compiler_version: %s\n' "$compiler_version"
printf 'contracts: %s\n' "$(join_by ', ' "${built_contract_names[@]}")"
