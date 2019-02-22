#!/usr/bin/env bash
scripts=$(dirname "$0")

tfplugin status -no-response-for-modules | xargs -L 1 $scripts/switch-to-modules.sh
