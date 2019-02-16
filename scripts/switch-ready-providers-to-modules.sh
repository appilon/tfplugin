#!/usr/bin/env bash
set -Eeuo pipefail

scripts=$(dirname "$0")
source "$scripts/skip-providers.sh"

tfplugin status -ready-for-modules | xargs $scripts/switch-to-modules.sh