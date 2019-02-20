#!/usr/bin/env bash
scripts=$(dirname "$0")

tfplugin status -ready-for-modules | xargs $scripts/switch-to-modules.sh
