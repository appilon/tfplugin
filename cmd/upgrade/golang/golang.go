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

func updateTravis(providerPath string, to *version.Version) error {
	filename := filepath.Join(providerPath, ".travis.yml")
	content, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		// support alternative extension
		filename = filepath.Join(providerPath, ".travis.yaml")
		content, err = ioutil.ReadFile(filename)
	} else if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	if contains(lines, "language: go", 0) == -1 {
		return errors.New("no 'language: go' in travis config")
	}

	goLine := contains(lines, "go:", 0)
	if goLine == -1 {
		return errors.New("no 'go:' in travis config")
	}

	lines = set(lines, goLine+1, fmt.Sprintf("  - %s", to))

	if to.Compare(Go111) >= 0 && contains(lines, "GO111MODULE=off", 0) == -1 {
		envLine := contains(lines, "env:", 0)
		if envLine == -1 {
			last := len(lines) - 1
			lines = insertBefore(lines, last, "env:")
			envLine = contains(lines, "env:", last)
		}

		globalLine := contains(lines, "global:", envLine)
		if globalLine == -1 {
			lines = insertBefore(lines, envLine+1, "  - GO111MODULE=off")
		} else {
			lines = insertBefore(lines, globalLine+1, "    - GO111MODULE=off")
		}
	}

	return ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}

func contains(lines []string, search string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], search) {
			return i
		}
	}
	return -1
}

func set(lines []string, index int, line string) []string {
	if index < len(lines) {
		lines[index] = line
	} else {
		lines = append(lines, line)
	}
	return lines
}

// taken from https://github.com/golang/go/wiki/SliceTricks
func insertBefore(lines []string, index int, line string) []string {
	return append(lines[:index], append([]string{line}, lines[index:]...)...)
}
