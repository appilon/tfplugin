#!/usr/bin/env bash

scripts=$(dirname "$0")
source "$scripts/skip-providers.sh"

pushd $GOPATH/src/github.com/terraform-providers

echo "provider,go_version,uses_modules,sdk_version" > ./providers.csv

for git_uri in $(cat repos.json | jq -r '.[] | select( .archived == false ) | .ssh_url'); do
    repo_dir=$(basename $git_uri .git)
    skip=$(skip_provider "$repo_dir")
    if [[ -d $repo_dir ]] && [[ -z "$skip" ]]
    then
        pushd $repo_dir
        git checkout -f master
        git pull
        tfplugin status >> ../providers.csv
        popd
    fi
done

popd
