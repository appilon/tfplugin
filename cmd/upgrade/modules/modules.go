package modules

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/appilon/tfplugin/util"
	"github.com/mitchellh/cli"
)

const CommandName = "upgrade modules"

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
	var commit bool
	var message string
	flags.StringVar(&provider, "provider", "", "provider to switch to go modules")
	flags.BoolVar(&commit, "commit", false, "changes will be committed")
	flags.StringVar(&message, "message", "vendor with go modules", "specify commit message")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	if err := util.Run(Env(), providerPath, "go", "mod", "init"); err != nil {
		log.Printf("Error running go mod init in %s: %s", providerPath, err)
		return 1
	}

	if err := os.RemoveAll(filepath.Join(providerPath, "vendor")); err != nil {
		log.Printf("Error purging vendor/ from %s: %s", providerPath, err)
		return 1
	}

	if err := util.Run(Env(), providerPath, "go", "mod", "tidy"); err != nil {
		log.Printf("Error running go mod tidy in %s: %s", providerPath, err)
		return 1
	}

	if err := util.Run(Env(), providerPath, "go", "mod", "vendor"); err != nil {
		log.Printf("Error running go mod vendor in %s: %s", providerPath, err)
		return 1
	}

	if commit {
		if err = util.Run(os.Environ(), providerPath, "git", "add", "--all"); err != nil {
			log.Printf("Error adding files: %s", err)
			return 1
		}

		if err = util.Run(os.Environ(), providerPath, "git", "commit", "-m", message); err != nil {
			log.Printf("Error committing: %s", err)
			return 1
		}
	}

	return 0
}

func Env() []string {
	return append(os.Environ(), "GO111MODULE=on")
}
