#!/usr/bin/env bash
set -Eeuo pipefail

scripts=$(dirname "$0")
source "$scripts/skip-providers.sh"

mkdir -p $GOPATH/src/github.com/terraform-providers
pushd $GOPATH/src/github.com/terraform-providers

curl 'https://api.github.com/orgs/terraform-providers/repos?per_page=200' > repos.json

for git_uri in $(cat repos.json | jq -r '.[] | select( .archived == false ) | .ssh_url'); do
    repo_dir=$(basename $git_uri .git)
    skip=$(skip_provider "$repo_dir")
    if [[ ! -d $repo_dir ]] && [[ -z "$skip" ]]
    then
        git clone --depth 1 --no-single-branch $git_uri
    fi
done

popd
