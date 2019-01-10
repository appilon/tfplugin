package modules

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	var removeGovendor bool
	flags.StringVar(&provider, "provider", "", "provider to switch to go modules")
	flags.BoolVar(&commit, "commit", false, "changes will be committed")
	flags.StringVar(&message, "message", "deps: use go modules for dep mgmt\n", "specify commit message")
	flags.BoolVar(&removeGovendor, "remove-govendor", false, "remove govendor from makefile and travis config")
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

	if removeGovendor {
		if err := removeGovendorFromTravis(providerPath); err != nil {
			log.Printf("Error removing govendor from travis config in %s: %s", providerPath, err)
			return 1
		}

		if err := removeGovendorFromMakefile(providerPath); err != nil {
			log.Printf("Error removing govendor from makefile in %s: %s", providerPath, err)
			return 1
		}
	}

	if commit {
		if removeGovendor {
			message += "remove govendor from makefile and travis config\n"
		}

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

func removeGovendorFromTravis(providerPath string) error {
	filename, content, err := util.ReadOneOf(providerPath, ".travis.yml", ".travis.yaml")
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	if goGetLine := util.SearchLines(lines, "github.com/kardianos/govendor", 0); goGetLine > -1 {
		lines = util.DeleteLines(lines, goGetLine)
	}

	if vendorStatusLine := util.SearchLines(lines, "vendor-status", 0); vendorStatusLine > -1 {
		lines = util.DeleteLines(lines, vendorStatusLine)
	}

	return ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}

func removeGovendorFromMakefile(providerPath string) error {
	filename, content, err := util.ReadOneOf(providerPath, "Makefile", "GNUmakefile")
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	if goGetLine := util.SearchLines(lines, "github.com/kardianos/govendor", 0); goGetLine > -1 {
		lines = util.DeleteLines(lines, goGetLine)
	}

	if vendorStatusTargetLine := util.SearchLines(lines, "vendor-status:", 0); vendorStatusTargetLine > -1 {
		lines = util.DeleteLines(lines, vendorStatusTargetLine, vendorStatusTargetLine+1)
	}

	out := strings.Replace(strings.Join(lines, "\n"), "vendor-status", "", -1)

	return ioutil.WriteFile(filename, []byte(out), 0644)
}

func Env() []string {
	return append(os.Environ(), "GO111MODULE=on")
}
