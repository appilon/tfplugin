# TFDEV
Swiss army knife tool for Terraform and plugin development

## Installation
```
$ go get -u github.com/appilon/tfdev
```

## Schema Extraction
```
$ tfdev schema github.com/mitchellh/terraform-provider-netlify > provider.json
```

### Requirements
* The provider exists on your GOPATH (supports multiple gopaths like `GOPATH=~/go:~/git/go`)
* The provider follows Terraform plugin naming convention of `terraform-provider-name`
* The provider exports a provider function ex: `netlify.Provider` type: `func() terraform.ResourceProvider`. The package name is extracted from the Terraform plugin naming convention.
* The `terraform.ResourceProvider` interface is satisfied via `*schema.Provider` (provider is cast to that type)

## Documentation Generation
```
$ cat provider.json | tfdev docs -resource=netlify_hook
```

The doc generation is very much still a work in progress.