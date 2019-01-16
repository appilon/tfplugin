#!/usr/bin/env bash
set -Eeuo pipefail

scripts=$(dirname "$0")
source "$scripts/skip-providers.sh"

pushd $GOPATH/src/github.com/terraform-providers

for git_uri in $(cat repos.json | jq -r '.[] | select( .archived == false ) | .ssh_url'); do
    repo_dir=$(basename $git_uri .git)
    skip=$(skip_provider "$repo_dir")
    if [[ -d $repo_dir ]] && [[ -z "$skip" ]]
    then
        pushd $repo_dir
        git checkout -f master
        git pull
        git checkout -b "v0.12-upgrade-$(date +%F)"
        tfplugin upgrade go -fix -fmt -commit
        tfplugin upgrade modules -commit
        tfplugin upgrade sdk -to pluginsdk-v0.12-early2 -commit
        tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)"
        popd
    fi
done

popd
