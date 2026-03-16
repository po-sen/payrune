#!/usr/bin/env bash
set -euo pipefail

npm --prefix deployments/ethereum ci
npm --prefix deployments/ethereum test
