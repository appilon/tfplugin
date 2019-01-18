## TFPLUGIN
This doc was originally meant to describe using `tfplugin upgrade` to automate parts of a provider switch to go1.11/Modules and TF SDK 0.12. It then became the guide for manually performing the steps and what to do when you encounter a particular error/issue. That information will likely be moved outside of this repo, for now however they have been split into the following files:

* [COMMON ISSUES](COMMON_ISSUES.md)
* [MANUAL UPGRADE](MANUAL_UPGRADE.md)

### Upgrading Go
```
$ tfplugin upgrade go -fix -fmt -commit
````

This will detect the current version of Go from the travis config, then replace it to `runtime.Version()` in the travis config and README.md. The version to update from and to update to can be specific with `-from` and `-to`. This will also run `go tool fix` and `gofmt`. It will construct a commit message based on what ran, you can overwrite this message with `-message`, this applies to both `tfplugin upgrade modules -commit` and `tfplugin upgrade sdk -commit` as well.

### Switching to modules
```
$ tfplugin upgrade modules -remove-govendor -commit
```

This will run `go mod init` `go mod tidy` and `go mod vendor` for you. It will also rip out any usage of `govendor` from the makefile and .travis.yml.

### Upgrade to Terraform 0.12 SDK
```
$ tfplugin upgrade sdk -to pluginsdk-v0.12-early2 -commit
```

This will vendor install the sdk, tidy, and vendor the sdk.

### Open Pull request
```
$ tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)"
```
You can open a PR to a provider if you configure a `GITHUB_PERSONAL_TOKEN`. Specifying `-open` will open the newly created pull request webpage in your default browser. Specifying `-title` will let you set the title. The remote can be specified with `-remote` and for cross-account PRs specify `-user`.

```
$ tfplugin upgrade pr -branch="$(git rev-parse --abbrev-ref HEAD)" -title="new code" -remote=appilon -user=appilon
```
