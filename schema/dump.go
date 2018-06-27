package schema

import (
	"encoding/json"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

const dumper = `
package main

import (
	"log"
	"github.com/appilon/tfdev/schema"
	p "%s/%s"
)

func main() {
	if err := schema.Dump(p.Provider); err != nil {
		log.Fatal(err)
	}
}
`

// This is called by the generated file, working dir set to provider dir
func Dump(providerFunc func() terraform.ResourceProvider) error {
	provider := providerFunc().(*schema.Provider)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	return enc.Encode(NewProviderDump(provider))
}
