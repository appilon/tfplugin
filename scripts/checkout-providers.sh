#!/usr/bin/env bash
set -Eeuxo pipefail

mkdir -p $GOPATH/src/github.com/terraform-providers
pushd $GOPATH/src/github.com/terraform-providers

curl 'https://api.github.com/orgs/terraform-providers/repos?per_page=200' > repos.json

for git_uri in $(cat repos.json | jq -r '.[] | .ssh_url'); do
    repo_dir=$(basename $git_uri .git)
    if [[ ! -d $repo_dir ]]
    then
        git clone --depth 1 $git_uri
    fi
done

popd
