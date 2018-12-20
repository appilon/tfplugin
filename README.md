# TFPLUGIN
Swiss army knife tool for Terraform Plugin development

## Installation
```
$ go get -u github.com/appilon/tfplugin
```

## Schema Extraction
```
$ tfplugin schema github.com/mitchellh/terraform-provider-netlify > provider.json
```

### Requirements
* The provider exists on your GOPATH (supports multiple gopaths like `GOPATH=~/go:~/git/go`)
* The provider follows Terraform plugin naming convention of `terraform-{type}-{name}`
* The provider exports a provider function ex: `netlify.Provider` type: `func() terraform.ResourceProvider`. The package name is extracted from the Terraform plugin naming convention.
* The `terraform.ResourceProvider` interface is satisfied via `*schema.Provider` (provider is cast to that type)
* Due to vendoring, for the extraction trick to work `hashicorp/terraform` must be on the GOPATH
```
$ go get -u github.com/hashicorp/terraform
```

## Documentation Generation
```
$ cat provider.json | tfplugin docs -resource=netlify_hook
```

The doc generation is very much still a work in progress.

## Provider auto upgrade
providers can be converted to go modules, have go version bumped in Travis and README, as well as the version of the vendored Terraform SDK bumped. See [scripts/upgrade-providers.sh](scripts/upgrade-providers.sh) as an example. For a more detailed walkthrough specific to the important 0.12 upgrade see [this](cmd/upgrade)
