## 0.12 TF SDK Upgrade
This is a guide to upgrade a provider to the Terraform 0.12 "SDK". This guide will assume you want to upgrade to the latest version of go and switch to using go modules for dependency management. If this isn't the case, upgrading the SDK is still likely possible however the steps might vary. The guide will first describe the manual steps, and then it will demonstrate how `tfplugin upgrade` will mostly automatically handle things.

### Switch to a new branch
It probably goes without saying, but it would be best to work off a fresh feature branch (or multiple branches).

### Upgrade to go1.11
Go 1.11 brings [go modules](https://github.com/golang/go/wiki/Modules), the "official" dependency management system for go. It will deprecate the usage of relying on the `GOPATH` and a project's `vendor` folder. Even if you do not wish to switch to go modules yet, there are module related implications to just upgrading to go1.11.

#### Update Provider README
Go into the readme and replace any references to the required version of go to 1.11

#### Update .travis.yml
Although our acceptance tests run in TeamCity, pre-checks for pull requests run in Travis. You will need to update the version of go required to 1.11.

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

If you would like to be explicit about the version of go in the README.md, and explicit with the version you are upgrading to use the flags `-from` and `-to`.

```
$ tfplugin upgrade go -from 1.8 -to 1.11
```

### Switch to modules
You can switch a provider to start using modules, however our build systems will still be using `vendor/` for the time being.

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

#### tidy (optional but do it)
Tidy will add any missing imports and prune your project and update the `go.mod` and `go.sum` files.

```
$ GO111MODULE=on go mod tidy
```

#### vendor
We still use vendoring in the end, this will copy the exact dependencies into `vendor/`.

```
$ GO111MODULE=on go mod vendor
```

#### Fix go get (and go generate) behavior now that we are module aware
.travis.yml and the makefile will need to have all `go get`, `go generate`, and specifically `gometaliner --install` commands prefixed with `GO111MODULE=off`. This is because TravisCI will default to module mode, fetching CLI tools will fail in this mode. Alternatively you could just configure the whole environment for module mode off. As of writing `tfplugin` does not perform this (it can easily just haven't gotten to it).

#### Purge govendor usage (or any other dependency manager)
Most providers were managed with `govendor` and even if they weren't there is likely some remnant of it in the project makefile and/or travis config.

##### In the travis config
* Remove any lines resembling `go get github.com/kardianos/govendor`
* Remove any lines resembling `make vendor-status`

##### In the Makefile or GNUmakefile
* Remove any lines resembling `go get github.com/kardianos/govendor`
* Remove the `vendor-status` target and any references to it throughout the file

We can likely emulate a similar behavior of `govendor status` using `go mod` and `git`.

#### TFPLUGIN
`tfplugin upgrade` can run these commands, remove `govendor` and commit for you.

```
$ tfplugin upgrade modules -remove-govendor -commit
```

### Upgrade to Terraform 0.12
As of writing it's advised to fetch the branch `pluginsdk-v0.12-early2` to get all the latest bug fixes. In the future we will suggest upgrading to the latest official release.

```
$ GO111MODULE=on go get github.com/hashicorp/terraform@master
```

I have noticed for some providers this leads to a compilation error in a transitive dependency? It might be worth specifying `-u` which will upgrade transitive dependencies to their latest `MINOR` release, this might solve the issue however in my specific experiences it just made another transitive dep fail to compile.

```
$ GO111MODULE=on go get -u github.com/hashicorp/terraform@pluginsdk-v0.12-early2
```

Remember to copy to `vendor/`

```
$ GO111MODULE=on go mod vendor
```

#### TFPLUGIN
`tfplugin upgrade` can run these commands and create a commit for you. By default it will specify the `latest` target (which will get the latest release, I'm not fully versed in modules yet, not sure if that includes major releases). Regardless we suggest getting `pluginsdk-v0.12-early2`, this can be specified with the `-to` flag.

```
$ tfplugin upgrade sdk -to pluginsdk-v0.12-early2 -commit
```

### Run acceptance tests
A provider compiled with the TF 0.12 SDK should still work with TF v0.11 and HCL1 configurations (Please let us know if it doesn't!!!). However a provider's acceptance test configurations will need to be upgraded to HCL2 syntax. This is because the acceptance tests run in-process against the vendored test harness, which in turn calls into the vendored Terraform Core, which is v0.12 and no longer parses HCL1.

If you have the ability to run the acceptance tests locally via `make testacc` do so and see what happens. You can expect a variety of configuration failures and its time to work through them. If you are a HashiCorp employee you can login to TeamCity OSS and run the provider acceptance tests (if you have no local setup), just make sure to specify your branch and set the version of go appropriately.

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
These notes are very rough but they describe common issues and the fixes for them.

1)
```
"id": {
                Type:     schema.TypeString,
                Computed: true,
            },
```
This is an automatic attribute, do not declare it or InternalValidate() will fail

2)
The most common issue will be nested types such as Map and List now requiring an equals assignment, and nested types that are resources now needing to be a true nested blocks (no equals). Martin's words explain it best:

>From a provider developer's perspective, that should be described as: any collection that has its `Elem` set to a `schema.Schema` is an "attribute" which must be set with the equals sign, while any where `Elem` is set to a `schema.Resource` is a "nested block" which must _not_ have the equals sign

3)
```
resource "grafana_alert_notification" "test" {
    type = "email"
    name = "terraform-acc-test"
    settings {
			"addresses" = "foo@bar.test"
			"uploadImage" = "false"
			"autoResolve" = "true"
		}
}
```
In this situation we first encountered a parsing error claiming attributes cannot be quoted. This is because Terraform hasn't processed the schema yet and doesn't realize the real error is #2 ^. If you do remove the quotes (which are allowed for TypeMap), you will get the same error message as #2.

4)
Many of our test cases will check that a list is empty. In 0.12 if the config had no attribute at all it will not be in the state, therefore we _had_ failures as a result of the key/path not even existing. This has since been shimmed, however it might be good to switch these checks to use `TestCheckNoResourceAttr`
```
{
				Config: testAccOrganizationConfig_usersUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationCheckExists("grafana_organization.test", &org),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "name", "terraform-acc-test",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "admins.#", "0",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "editors.#", "1",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "editors.0", "john.doe@example.com",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "viewers.#", "0",
					),
				),
			},
```