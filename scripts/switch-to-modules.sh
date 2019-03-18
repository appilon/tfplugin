#!/usr/bin/env bash

set -Eeuo pipefail

scripts=$(dirname "$0")
source "$scripts/skip-providers.sh"

repo_dir=$(basename $1)
skip=$(skip_provider "$repo_dir")
if [[ ! -z "$skip" ]]
then
    echo "skipping...$skip"
    exit 0
fi

pushd $GOPATH/src/$1

git checkout -f master
git pull
git branch -D "go-modules-$(date +%F)" || true
git checkout -b "go-modules-$(date +%F)"
# this is mainly to encode the .go-version, will create commit message
# stating it did a few things that are likely noop
# as most PRs have already been merged switching to go 1.11
tfplugin upgrade go -to="1.11.5" -fix -fmt -encode -commit
tfplugin upgrade modules -commit
# upgrade to v0.11 sdk which has a clean transitive dependency story that uses go modules
tfplugin upgrade sdk -to=sdk-v0.11-with-go-modules -commit
tfplugin upgrade pr -closes="$(tfplugin status -proposal)" -branch="$(git rev-parse --abbrev-ref HEAD)" -title="[MODULES] Re-open Switch to Go Modules"

popd
