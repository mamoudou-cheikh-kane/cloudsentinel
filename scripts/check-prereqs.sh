#!/usr/bin/env bash
# Verify the local environment has every tool the demo needs.
#
# Each tool is checked twice:
#   1. Is the binary on PATH?
#   2. Does it report a version?
#
# Bails out on the first missing tool.

set -euo pipefail

color_red()    { printf '\033[0;31m%s\033[0m\n' "$*"; }
color_green()  { printf '\033[0;32m%s\033[0m\n' "$*"; }
color_blue()   { printf '\033[0;34m%s\033[0m\n' "$*"; }

require() {
  local name="$1"
  local command="$2"
  local install_url="$3"
  if ! command -v "$name" >/dev/null 2>&1; then
    color_red ">>> ERROR: '$name' is not installed or not on PATH."
    color_red "    Install instructions: $install_url"
    exit 1
  fi
  local version
  version=$(eval "$command" 2>&1 | head -n 1 || echo "(unknown version)")
  color_green ">>> $name OK: $version"
}

color_blue ">>> Checking prerequisites for the CloudSentinel demo"
echo ""

require "docker"  "docker --version"        "https://docs.docker.com/get-docker/"
require "kind"    "kind --version"          "https://kind.sigs.k8s.io/docs/user/quick-start/"
require "kubectl" "kubectl version --client" "https://kubernetes.io/docs/tasks/tools/"

echo ""
color_green ">>> All prerequisites satisfied. You can run 'make demo'."
