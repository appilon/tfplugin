#!/usr/bin/env bash
set -Eeuxo pipefail

pushd $GOPATH/src/github.com/terraform-providers

for git_uri in $(cat repos.json | jq -r '.[] | .ssh_url'); do
    repo_dir=$(basename $git_uri .git)
    if [[ -d $repo_dir ]]
    then
        pushd $repo_dir
        git checkout -b "tfplugin-$(date +%F)"
        tfplugin upgrade go -fix -fmt -commit
        tfplugin upgrade modules -commit
        tfplugin upgrade sdk -to v0.12.0-alpha2 -commit
        tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)"
        popd
    fi
done

popd
