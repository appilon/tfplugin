## 0.12 TF SDK Upgrade
This is a guide to upgrade a provider to the Terraform 0.12 "SDK". This guide will assume you want to upgrade to the latest version of go and switch to using go modules for dependency management. If this isn't the case, upgrading the SDK is still possible however the steps might vary.

## Steps
* [Create a new branch](#create-a-new-branch)
* [Upgrade to go1.11](#upgrade-to-go1.11)
* [Switch to go modules](#switch-to-go-modules)
* [Upgrade to Terraform 0.12](#upgrade-terraform)
* [Merge changes](#merge)
* [Update Changelog](#update-changelog)

## Create a new branch
It probably goes without saying, but it would be best to work off a fresh feature branch (or branches).

## Upgrade to go1.11
Go 1.11 brings [go modules](https://github.com/golang/go/wiki/Modules), the "official" dependency management system for go. It will deprecate the usage of relying on the `GOPATH` and a project's `vendor` folder. Even if you do not wish to switch to go modules yet, there are module related implications to just upgrading to go1.11.

### Update Provider README
Go into the readme and replace any references to the required version of go to 1.11

### Update .travis.yml
Although our acceptance tests run in TeamCity, pre-checks for pull requests run in Travis. You will need to update the version of go required to 1.11.

### Run go fix (optional)
It's always good to ensure the provider codebase uses the latest APIs and syntax after a version upgrade.

```
$ go tool fix ./<provider_package>
```

### Run gofmt
Our CI checks that the provider passes `gofmt`, so make sure to run `gofmt`.

```
$ gofmt -s -w ./<provider_package>
```

## Switch to go modules
You can switch a provider to start using modules, however our build systems will still be using `vendor/` for the time being.

Our dependencies are currently all checked into `vendor/`. We need to instruct go to use that folder for compilation. This is accomplished by setting the environment variables `GO111MODULE=on` and `GOFLAGS=-mod=vendor` before running commands. 

You could achieve a similar solution with `GO111MODULE=off`, however there is a subtle difference that in module mode with `-mod=vendor`, all transitive dependencies are compiled from the top level `vendor/` folder and only that folder. The legacy mode `GO111MODULE=off` would find transitive dependencies for a dependency if there were any within a `vendor/` folder in THAT dependency.

For tooling installs/generate, you will need to prefix `go get` `go generate` and `gometalinter --install` with `GO111MODULE=off`, this is to force the legacy mode, this is required to properly install those CLI tools or run `go generate`.

### Prepare clean Go environment

A few of us have experienced weird go module behaviours (e.g. corrupted cache) in the past,
which lead us to believe it's best to have clean Go environment for any larger-scale
changes involving go modules.

Here's an example of how to prepare such environment via Docker:

```bash
dgo() {
    local DESIRED_GO_VERSION="latest"
    local NAME=$(basename "$PWD")
    echo "Launching golang:${DESIRED_GO_VERSION} for ${NAME} ..."
    docker run --rm -ti -e GO111MODULE=on -v "$PWD":/usr/src/${NAME} -w /usr/src/${NAME} golang:${DESIRED_GO_VERSION} bash
}
```

Then you can just run `dgo` and carry on per instructions below.

### Run go mod
`go mod` will do most of the work importing from whatever previous tool was in place. If the provider is in the $GOPATH you will need the environment variable `GO111MODULE=on` for these commands to work.

```
$ GO111MODULE=on go mod init
rm -rf vendor/
$ GO111MODULE=on go mod tidy
$ GO111MODULE=on go mod vendor
```

### Purge govendor usage (or any other dependency manager)
Most providers were managed with `govendor` and even if they weren't there is likely some remnant of it in the project makefile and/or travis config.

#### Update travis config
Remove the following lines (if present) from `.travis.yml`:
* `go get github.com/kardianos/govendor`
* `make vendor-status`

#### In the Makefile or GNUmakefile
Remove the following lines (if present) from `Makefile` or `GNUMakefile`:
* `go get github.com/kardianos/govendor`
* Remove the `vendor-status` target and any references to it throughout the file

## Upgrade Terraform
As of writing it's advised to fetch the branch `pluginsdk-v0.12-early2` to get all the latest bug fixes. In the future we will suggest upgrading to the latest official release.

```
$ GO111MODULE=on go get github.com/hashicorp/terraform@pluginsdk-v0.12-early2
$ GO111MODULE=on go mod tidy
$ GO111MODULE=on go mod vendor
```

### Run tests
A provider compiled with the TF 0.12 SDK should still work with TF v0.11 and HCL1 configurations (Please let us know if it doesn't!!!). However a provider's acceptance test configurations will need to be upgraded to HCL2 syntax. This is because the acceptance tests run in-process against the vendored test harness, which in turn calls into the vendored Terraform Core, which is v0.12 and no longer parses HCL1.

Ensure that the vendored go modules are used: 
```
GO111MODULE=on GOFLAGS=-mod=vendor make test
```

#### Run acceptance tests
If you have the ability to run the acceptance tests locally via `make testacc` do so and see what happens. You can expect a variety of configuration failures and its time to work through them. If you are a HashiCorp employee you can login to TeamCity OSS and run the provider acceptance tests (if you have no local setup), just make sure to specify your branch and set the version of go appropriately.

```
GO111MODULE=on GOFLAGS=-mod=vendor make testacc
```

### Merge
Open a PR
Verify that travis passes
Merge PR 

### Update Changelog
Update CHANGELOG.md


Example addition to CHANGELOG.md:

```
2.0.0 (Unreleased)

IMPROVEMENTS:

The provider is now compatible with Terraform v0.12, while retaining compatibility with prior versions.
```
