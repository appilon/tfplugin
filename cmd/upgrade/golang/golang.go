package golang

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/appilon/tfplugin/util"
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
	var fromStr string
	var toStr string
	var provider string
	var commit bool
	var message string
	var fmt bool
	var fix bool
	flags.StringVar(&fromStr, "from", "", "current version of go")
	flags.StringVar(&toStr, "to", strings.TrimPrefix(runtime.Version(), "go"), "version of go upgrading to")
	flags.StringVar(&provider, "provider", "", "provider to upgrade")
	flags.BoolVar(&commit, "commit", false, "changes will be committed")
	flags.StringVar(&message, "message", "", "specify commit message")
	flags.BoolVar(&fmt, "fmt", false, "Run go fmt on provider")
	flags.BoolVar(&fix, "fix", false, "Run go fix on provider")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	var from *version.Version
	if fromStr != "" {
		from, err = version.NewVersion(fromStr)
		if err != nil {
			log.Printf("Error parsing -from: %s", err)
			return 1
		}
	} else {
		from, err = detectGoVersionFromTravis(providerPath)
		if err != nil {
			log.Printf("Could not detect go version from .travis.yml: %s", err)
			return 1
		}
	}

	to, err := version.NewVersion(toStr)
	if err != nil {
		log.Printf("Error parsing -to: %s", err)
		return 1
	}

	if err := updateTravis(providerPath, to.String()); err != nil {
		log.Printf("Error updating .travis.yml: %s", err)
		return 1
	}

	if err := updateReadme(providerPath, from, to); err != nil {
		log.Printf("Error updating README.md: %s", err)
		return 1
	}

	if fmt || fix {
		packageName, err := util.GetPackageName(providerPath)
		if err != nil {
			log.Printf("Error determining package name: %s", err)
			return 1
		}

		if fix {
			if err := util.Run(os.Environ(), providerPath, "go", "tool", "fix", "./"+packageName); err != nil {
				log.Printf("Error running go tool fix: %s", err)
				return 1
			}
		}

		if fmt {
			if err := util.Run(os.Environ(), providerPath, "gofmt", "-s", "-w", "./"+packageName); err != nil {
				log.Printf("Error running gofmt: %s", err)
				return 1
			}
		}
	}

	if commit {
		if err = util.Run(os.Environ(), providerPath, "git", "add", "--all"); err != nil {
			log.Printf("Error adding files: %s", err)
			return 1
		}

		if message == "" {
			message = "provider: Require Go " + toStr + " in TravisCI and README\n"
			if fix {
				message += "provider: Run go fix\n"
			}
			if fmt {
				message += "provider: Run go fmt\n"
			}
		}

		if err = util.Run(os.Environ(), providerPath, "git", "commit", "-m", message); err != nil {
			log.Printf("Error committing: %s", err)
			return 1
		}
	}

	return 0
}

func updateReadme(providerPath string, from, to *version.Version) error {
	fromSegments := from.Segments()
	toSegments := to.Segments()
	search := fmt.Sprintf("%d.%d", fromSegments[0], fromSegments[1])
	replace := fmt.Sprintf("%d.%d", toSegments[0], toSegments[1])

	filename := filepath.Join(providerPath, "README.md")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	out := strings.Replace(string(content), search, replace, -1)
	return ioutil.WriteFile(filename, []byte(out), 0644)
}

func detectGoVersionFromTravis(providerPath string) (*version.Version, error) {
	filename := filepath.Join(providerPath, ".travis.yml")
	content, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		// support alternative extension
		filename = filepath.Join(providerPath, ".travis.yaml")
		content, err = ioutil.ReadFile(filename)
	} else if err != nil {
		return nil, err
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
		return nil, errors.New("no 'language: go' in .travis.yml")
	}

	v := strings.TrimLeft(lines[goLine+1], ` -"`)
	v = strings.TrimRight(v, ` "x.`)

	return version.NewVersion(v)
}

func updateTravis(providerPath string, v string) error {
	filename := filepath.Join(providerPath, ".travis.yml")
	content, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		// support alternative extension
		filename = filepath.Join(providerPath, ".travis.yaml")
		content, err = ioutil.ReadFile(filename)
	} else if err != nil {
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
