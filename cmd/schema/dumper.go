package schema

const dumper = `
package main

import (
	"log"
	"os"
	"github.com/appilon/tfplugin/schema"
	p "%s/%s"
)

func main() {
	if err := schema.NewProviderDump(p.Provider).ToJSON(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
`
