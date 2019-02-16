#!/usr/bin/env bash
set -Eeuo pipefail

pushd $GOPATH/src/$1

git checkout -f master
git pull
git checkout -b "go-modules-$(date +%F)"
# this is mainly to encode the .go-version, will create commit message
# stating it did a few things that are likely noop
# as most PRs have already been merged switching to go 1.11
tfplugin upgrade go -to="1.11.5" -fix -fmt -encode -commit
tfplugin upgrade modules -commit
tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)" -title="[MODULES] Switch to Go Modules"

popd
