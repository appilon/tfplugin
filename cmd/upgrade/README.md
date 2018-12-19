## 0.12 TF SDK Upgrade
This is a guide to upgrade a provider to the Terraform 0.12 "SDK". This guide will assume you want to upgrade to the latest version of go and switch to using go modules for dependency management. If this isn't the case, upgrading the SDK is still likely possible however the steps might vary. The guide will first describe the manual steps, and then it will demonstrate how `tfplugin upgrade` will mostly automatically handle things.

### Switch to a new branch
It probably goes without saying, but it would be best to work off a fresh feature branch (or multiple branches).

### Upgrade to go1.11
Go 1.11 brings [go modules](https://github.com/golang/go/wiki/Modules), the "official" dependency management system for go. It will deprecate the usage of relying on the `GOPATH` and a project's `vendor` folder. Even if you do not wish to switch to go modules yet, there are module related implications to just upgrading to go1.11.

#### Update Provider README
Go into the readme and replace any references to the required version of go to 1.11

#### Update .travis.yml
Although our acceptance tests run in TeamCity, pre-checks for pull requests run in Travis. You will need to update the version of go required to 1.11, you also need to force go modules to be off. This is accomplished by adding an environment variable

```yaml
env:
  - GO111MODULE=off
```
or
```yaml
env:
  global:
    - GO111MODULE=off
```

This will ensure the `go get` behavior and build system remains on the "legacy" `GOPATH` and `vendor/` system.

#### Run go fix (optional)
It's always good to ensure the provider codebase uses the latest APIs and syntax after a version upgrade.

```
$ go tool fix ./<provider_package>
```

#### Run gofmt
Our CI checks that the provider passes `gofmt`, so make sure to run `gofmt`.

```
$ gofmt -s -w ./<provider_package>
```

#### TFPLUGIN
`tfplugin upgrade` can automatically detect the version of go specified in the travis config, and from there update your README.md, .travis.yml to `tfplugin`'s `runtime.Version()`. it can also run `go tool fix` and `gofmt` for you, as well as construct a decent commit message.

```
$ tfplugin upgrade go -fix -fmt -commit
```

If you would like to be explicit about the version of go in the README.md, and explicit with the version you are upgrading to use the flags `-from` and `-to`

```
$ tfplugin upgrade go -from 1.8 -to 1.11
```

### Switch to modules
You can switch a provider to start using modules, however our build systems will still be using `vendor/` for the time being.

#### Purge govendor usage (or any other dependency manager)
Most providers were managed with `govendor` and even if they weren't there is likely some remnant of it in the project makefile and/or travis config. PLEASE NOTE: as of writing `tfplugin` does not perform this yet, however it is in progress.

##### In the travis config
* Remove any lines resembling `go get github.com/kardianos/govendor`
* Remove any lines resembling `make vendor-status`

##### In the Makefile or GNUmakefile
* Remove any lines resembling `go get github.com/kardianos/govendor`
* Remove the `vendor-status` target and any references to it throughout the file
* Remove any other `govendor` usage/figure out the `go mod` equivalent

There is likely a `go mod` equivalent to `govendor status` we could substitute in instead if we wanted to keep this check. If I learn it from someone or look into it I will update the guide.

#### Run go mod
`go mod` will do most of the work importing from whatever previous tool was in place. The main annoyance is if your machine has the provider on the GOPATH you will need to force `GO111MODULE=on` for these commands to work.

#### init/import

```
$ GO111MODULE=on go mod init
```

#### purge vendor/
It's a good idea to nuke `vendor/`

```
rm -rf vendor/
```

#### tidy (optional)
Tidy will add any missing imports and prune your project and update the `go.mod` and `go.sum` files.

```
$ GO111MODULE=on go mod tidy
```

#### vendor
We still use vendoring in the end, this will copy the exact dependencies into `vendor/`.

```
$ GO111MODULE=on go mod vendor
```

#### TFPLUGIN
`tfplugin upgrade` can run these commands and commit for you.

```
$ tfplugin upgrade modules -commit
```

### Upgrade to Terraform 0.12
As of writing it's advised to fetch the latest code `@master` to get all the latest bug fixes. In the future we will suggest upgrading to the latest official release.

```
$ GO111MODULE=on go get github.com/hashicorp/terraform@master
```

I have noticed for some providers this leads to a compilation error in a transitive dependency? It might be worth specifying `-u` which will upgrade transitive dependencies to their latest `MINOR` release, this might solve the issue however in my specific experiences it just made another transitive dep fail to compile.

```
$ GO111MODULE=on go get -u github.com/hashicorp/terraform@master
```

Remember to copy to `vendor/`

```
$ GO111MODULE=on go mod vendor
```

#### TFPLUGIN
`tfplugin upgrade` can run these commands and create a commit for you. By default it will specify the `latest` target (which will get the latest release, I'm not fully versed in modules yet, not sure if that includes major releases). Regardless we suggest getting `master`, this can be specified with the `-to` flag.

```
$ tfplugin upgrade sdk -to master -commit
```

### Run acceptance tests
A provider compiled with the TF 0.12 SDK should still work with TF v0.11 and HCL1 configurations (Please let us know if it doesn't!!!). However a provider's acceptance test configurations will need to be upgraded to HCL2 syntax. This is because the acceptance tests run in-process against the vendored test harness, which in turn calls into the vendored Terraform Core, which is v0.12 and no longer parses HCL1.

If you have the ability to run the acceptance tests locally via `make testacc` do so and see what happens. You can expect a variety of configuration failures and its time to work through them. If you are a HashiCorp employee you can login to TeamCity OSS and run the provider acceptance tests (if you have no local setup), just make sure to specify your branch and _set the version of go to 1.11.4_.

### Run Travis
To run travis (to my knowledge) you will need to create a pull request. `tfplugin` can create one if you specify an access token with `GITHUB_PERSONAL_TOKEN`, just note the pull request will be coming from the associated GitHub account.

```
$ tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)"
```

Specifying `-open` will open the newly created pull request webpage in your default browser. Specifying `-title` will let you set the title. The remote can be specified with `-remote` and for cross-account PRs specify `-user`.

```
$ tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)" -title="new code" -remote=appilon -user=appilon
```

### HCL upgrades
TODO: I will try and learn and detail the kinds of configuration upgrades that come up frequently. 