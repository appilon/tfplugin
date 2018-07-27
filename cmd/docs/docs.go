package docs

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/appilon/tfdev/schema"
	"github.com/mitchellh/cli"
)

type command struct {
	name string
}

func (c *command) Help() string {
	return ""
}

func (c *command) Synopsis() string {
	return ""
}

func CommandFactory() (cli.Command, error) {
	return &command{
		name: "docs",
	}, nil
}

func (c *command) Run(args []string) int {
	flags := flag.NewFlagSet(c.name, flag.ExitOnError)
	var datasource string
	var resource string
	flags.StringVar(&datasource, "datasource", "", "data source name")
	flags.StringVar(&resource, "resource", "", "resource name")
	flags.Parse(args)

	provider := &schema.ProviderDump{}
	err := json.NewDecoder(os.Stdin).Decode(provider)
	if err != nil {
		log.Printf("Error decoding provider json: %s", err)
		return 1
	}

	var resourceMap map[string]*schema.ResourceDump
	var mapType string
	var name string
	if datasource != "" {
		resourceMap = provider.DataSourcesMap
		mapType = "data source"
		name = datasource
	} else if resource != "" {
		resourceMap = provider.ResourcesMap
		mapType = "resource"
		name = resource
	} else {
		return cli.RunResultHelp
	}

	r, exists := resourceMap[name]
	if !exists {
		log.Printf("Error '%s' is not a %s", name, mapType)
		return 1
	}

	if err := PrintDoc(r); err != nil {
		log.Printf("Error printing doc for %s: %s", name, err)
		return 1
	}

	return 0
}
