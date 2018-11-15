package sdk

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/appilon/tfplugin/cmd/upgrade/modules"
	"github.com/appilon/tfplugin/util"
	"github.com/mitchellh/cli"
)

const CommandName = "upgrade sdk"
const TerraformRepo = "github.com/hashicorp/terraform"

type command struct{}

func (c *command) Help() string {
	return ""
}

func (c *command) Synopsis() string {
	return ""
}

func CommandFactory() (cli.Command, error) {
	return &command{}, nil
}

func (c *command) Run(args []string) int {
	flags := flag.NewFlagSet(CommandName, flag.ExitOnError)
	var to string
	var provider string
	var depTool string
	var commit bool
	var message string
	flags.StringVar(&to, "to", "latest", "version of the terraform sdk to upgrade to")
	flags.StringVar(&provider, "provider", "", "provider to upgrade")
	flags.StringVar(&depTool, "dep-tool", "modules", "dependency tool for the provider")
	flags.BoolVar(&commit, "commit", false, "changes will be committed")
	flags.StringVar(&message, "message", "", "specify commit message")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	if err = updateSDK(providerPath, to, depTool); err != nil {
		log.Printf("Error updating sdk to %s: %s", to, err)
		return 1
	}

	if commit {
		if err = util.Run(os.Environ(), providerPath, "git", "add", "--all"); err != nil {
			log.Printf("Error adding files: %s", err)
			return 1
		}

		if message == "" {
			var command string

			switch depTool {
			case "govendor":
				// TODO
			case "dep":
				// TODO
			case "modules":
				command = "go get " + TerraformRepo + "@" + to
			}

			message = fmt.Sprintf("deps: %s@%s\nUpdated via: %s\n", TerraformRepo, to, command)
		}

		if err = util.Run(os.Environ(), providerPath, "git", "commit", "-m", message); err != nil {
			log.Printf("Error committing: %s", err)
			return 1
		}
	}

	return 0
}

func updateSDK(providerPath, version, depTool string) error {
	switch depTool {
	case "govendor":
		// TODO
	case "dep":
		// TODO
	case "modules":
		if err := util.Run(modules.Env(), providerPath, "go", "get", TerraformRepo+"@"+version); err != nil {
			return fmt.Errorf("Error fetching %s@%s: %s", TerraformRepo, version, err)
		}

		if err := util.Run(modules.Env(), providerPath, "go", "mod", "vendor"); err != nil {
			return fmt.Errorf("Error running go mod vendor in %s: %s", providerPath, err)
		}
	}
	return nil
}
