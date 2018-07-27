package schema

const dumper = `
package main

import (
	"log"
	"os"
	"github.com/appilon/tfdev/schema"
	p "%s/%s"
)

func main() {
	if err := schema.NewProviderDump(p.Provider).ToJSON(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
`
