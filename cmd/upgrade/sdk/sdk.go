package sdk

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/appilon/tfplugin/util"
	"github.com/mitchellh/cli"
)

const CommandName = "upgrade sdk"

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
	flags.StringVar(&to, "to", "latest", "version of the terraform sdk to upgrade to")
	flags.StringVar(&provider, "provider", "", "provider to upgrade")
	flags.StringVar(&depTool, "dep-tool", "modules", "dependency tool for the provider, will default to modules and attempt autoconversion")
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

	return 0
}

func updateSDK(providerPath, version, depTool string) error {
	switch depTool {
	case "govendor":
		// TODO
	case "dep":
		// TODO
	case "modules":
		fallthrough
	default:
		var firstTimeModules bool
		if _, err := os.Stat(path.Join(providerPath, "go.mod")); os.IsNotExist(err) {
			firstTimeModules = true
			if err := util.Run(modulesEnv(), providerPath, "go", "mod", "init"); err != nil {
				return fmt.Errorf("Error running go mod init in %s: %s", providerPath, err)
			}

			if err := os.RemoveAll(path.Join(providerPath, "vendor")); err != nil {
				return fmt.Errorf("Error purging vendor/ from %s: %s", providerPath, err)
			}
		}

		if err := util.Run(modulesEnv(), providerPath, "go", "get", "github.com/hashicorp/terraform@"+version); err != nil {
			return fmt.Errorf("Error fetching github.com/hashicorp/terraform@%s: %s", version, err)
		}

		if firstTimeModules {
			if err := util.Run(modulesEnv(), providerPath, "go", "mod", "tidy"); err != nil {
				return fmt.Errorf("Error running go mod tidy in %s: %s", providerPath, err)
			}
		}

		if err := util.Run(modulesEnv(), providerPath, "go", "mod", "vendor"); err != nil {
			return fmt.Errorf("Error running go mod vendor in %s: %s", providerPath, err)
		}
	}
	return nil
}

func modulesEnv() []string {
	var env []string
	for _, pair := range os.Environ() {
		if !strings.HasPrefix(pair, "GOPATH") {
			env = append(env, pair)
		}
	}
	return append(env, "GO11MODULE=on")
}
