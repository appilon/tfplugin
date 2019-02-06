package status

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/appilon/tfplugin/cmd/upgrade/golang"
	"github.com/appilon/tfplugin/util"
	"github.com/mitchellh/cli"
	"github.com/radeksimko/go-mod-diff/go-internal/modfile"
)

const CommandName = "status"

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
	var provider string
	flags.StringVar(&provider, "provider", "", "provider to analyze")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	providerName, err := util.GetPackageName(providerPath)
	if err != nil {
		log.Printf("Error determining provider name: %s", err)
		return 1
	}

	goVersion := "?"
	usesModules := false
	sdkVersion := "?"

	defer func() {
		fmt.Printf("%s,%s,%t,%s", providerName, goVersion, usesModules, sdkVersion)
		fmt.Println()
	}()

	if v, err := golang.DetectGoVersionFromTravis(providerPath); err == nil {
		goVersion = v.Original()
	} else {
		log.Printf("Error determining go version: %s", err)
	}

	if _, err := os.Stat(filepath.Join(providerPath, "go.mod")); err == nil {
		usesModules = true
	} else if !os.IsNotExist(err) {
		log.Printf("Error checking for go.mod: %s", err)
	}

	// for now only pull sdk version if using go modules
	if usesModules {
		gomodPath := filepath.Join(providerPath, "go.mod")
		data, err := ioutil.ReadFile(gomodPath)
		if err != nil {
			log.Printf("Error reading go.mod file: %s", err)
			return 1
		}
		f, err := modfile.Parse(gomodPath, data, nil)
		if err != nil {
			log.Printf("Error parsing go.mod file: %s", err)
			return 1
		}

		for _, r := range f.Require {
			if r.Mod.Path == "github.com/hashicorp/terraform" {
				sdkVersion = r.Mod.Version
				break
			}
		}
	}

	return 0
}
