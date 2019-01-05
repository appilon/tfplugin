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

var Go111 *version.Version

type command struct{}

func init() {
	Go111 = version.Must(version.NewVersion("1.11.0"))
}

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

	if err := updateTravis(providerPath, to); err != nil {
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
			message = "provider: Require Go " + to.String() + " in TravisCI and README\n"
			if to.Compare(Go111) >= 0 {
				message += "NOTE: ensured GO111MODULE=off in travis config to enforce legacy go get behavior\n"
			}
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

func majorMinor(v *version.Version) string {
	segments := v.Segments()
	return fmt.Sprintf("%d.%d", segments[0], segments[1])
}

func updateReadme(providerPath string, from, to *version.Version) error {
	search := majorMinor(from)
	replace := majorMinor(to)

	filename := filepath.Join(providerPath, "README.md")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	out := strings.Replace(string(content), search, replace, -1)
	return ioutil.WriteFile(filename, []byte(out), 0644)
}

func detectGoVersionFromTravis(providerPath string) (*version.Version, error) {
	_, content, err := util.ReadOneOf(providerPath, ".travis.yml", ".travis.yaml")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	if util.SearchLines(lines, "language: go", 0) == -1 {
		return nil, errors.New("no 'language: go' in travis config")
	}

	goLine := util.SearchLines(lines, "go:", 0)
	if goLine == -1 {
		return nil, errors.New("no 'go:' in travis config")
	}

	v := strings.TrimLeft(lines[goLine+1], ` -"`)
	v = strings.TrimRight(v, ` "x.`)

	return version.NewVersion(v)
}

func updateTravis(providerPath string, to *version.Version) error {
	filename, content, err := util.ReadOneOf(providerPath, ".travis.yml", ".travis.yaml")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	if util.SearchLines(lines, "language: go", 0) == -1 {
		return errors.New("no 'language: go' in travis config")
	}

	goLine := util.SearchLines(lines, "go:", 0)
	if goLine == -1 {
		return errors.New("no 'go:' in travis config")
	}

	lines = util.SetLine(lines, goLine+1, fmt.Sprintf(`  - "%s.x"`, majorMinor(to)))

	if to.Compare(Go111) >= 0 && util.SearchLines(lines, "GO111MODULE=off", 0) == -1 {
		envLine := util.SearchLines(lines, "env:", 0)
		if envLine == -1 {
			last := len(lines) - 1
			lines = util.InsertLineBefore(lines, last, "env:")
			envLine = util.SearchLines(lines, "env:", last)
		}

		globalLine := util.SearchLines(lines, "global:", envLine)
		if globalLine == -1 {
			lines = util.InsertLineBefore(lines, envLine+1, "  - GO111MODULE=off")
		} else {
			lines = util.InsertLineBefore(lines, globalLine+1, "    - GO111MODULE=off")
		}
	}

	return ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}
