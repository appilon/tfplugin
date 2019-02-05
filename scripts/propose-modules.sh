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
        tfplugin upgrade modules -propose
        popd
        sleep 5 # GH API rate limits
    fi
done

popd
