package golang

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/appilon/tfplugin/cmd"
	version "github.com/hashicorp/go-version"
	"github.com/mitchellh/cli"
)

const CommandName = "upgrade go"

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
	var err error

	var fromStr string
	var toStr string
	var providerPath string
	flags.StringVar(&fromStr, "from", "", "current version of go")
	flags.StringVar(&toStr, "to", strings.TrimPrefix(runtime.Version(), "go"), "version of go upgrading to")
	flags.StringVar(&providerPath, "provider", "", "provider to upgrade")
	flags.Parse(args)

	if fromStr == "" {
		log.Print("-from must be set")
		return 1
	}

	from, err := version.NewVersion(fromStr)
	if err != nil {
		log.Printf("Error parsing -from: %s", err)
		return 1
	}

	to, err := version.NewVersion(toStr)
	if err != nil {
		log.Printf("Error parsing -to: %s", err)
		return 1
	}

	if providerPath == "" {
		providerPath, err = os.Getwd()
		if err != nil {
			log.Printf("Error getting working directory: %s", err)
			return 1
		}
	} else {
		provider := providerPath
		providerPath, err = cmd.FindProviderInGoPath(provider)
		if err != nil {
			log.Printf("Error finding %s in GOPATH: %s", provider, err)
			return 1
		}
	}

	if err = updateTravis(providerPath, to.String()); err != nil {
		log.Printf("Error updating .travis.yml: %s", err)
		return 1
	}

	if err = updateReadme(providerPath, from, to); err != nil {
		log.Printf("Error updating README.md: %s", err)
		return 1
	}

	return 0
}

func updateReadme(providerPath string, from, to *version.Version) error {
	fromSegments := from.Segments()
	toSegments := to.Segments()
	search := fmt.Sprintf("%d.%d", fromSegments[0], fromSegments[1])
	replace := fmt.Sprintf("%d.%d", toSegments[0], toSegments[1])

	filename := path.Join(providerPath, "README.md")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	out := strings.Replace(string(content), search, replace, -1)
	return ioutil.WriteFile(filename, []byte(out), 0644)
}

func updateTravis(providerPath string, v string) error {
	filename := path.Join(providerPath, ".travis.yml")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var goLine int
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "language: go") {
			goLine = i + 1
			break
		}
	}
	if goLine == 0 {
		return errors.New("no 'language: go' in .travis.yml")
	}
	lines[goLine+1] = fmt.Sprintf("- %s", v)

	out := strings.Join(lines, "\n")
	return ioutil.WriteFile(filename, []byte(out), 0644)
}
