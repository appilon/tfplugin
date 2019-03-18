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

var disableGoModulesFor = []string{"go get", "go generate", "gometalinter --install"}

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
	var propose bool
	flags.StringVar(&provider, "provider", "", "provider to switch to go modules")
	flags.BoolVar(&propose, "propose", false, "open issue proposing switch to go modules")
	flags.BoolVar(&commit, "commit", false, "changes will be committed")
	flags.StringVar(&message, "message", "deps: use go modules for dep mgmt\nrun go mod tidy\nremove govendor from makefile and travis config\nset appropriate env vars for go modules\n", "specify commit message")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	if propose {
		return proposeGoModules(providerPath)
	}

	// check if repo already uses modules
	if _, err := os.Stat(filepath.Join(providerPath, "go.mod")); err == nil {
		log.Printf("Provider is already on go modules")
		return 1
	}

	// check if modules PR is currently open
	if owner, repo, err := util.GetGitHubDetails(providerPath); err != nil {
		log.Printf("Error getting owner/repo info: %s", err)
		return 1
	} else if prNo, err := PullRequestExists(owner, repo, "[MODULES]"); err != nil {
		log.Printf("Error looking up pull requests for %s/%s: %s", owner, repo, err)
		return 1
	} else if prNo > 0 {
		log.Printf("Provider already has modules pull request: https://github.com/%s/%s/pulls/%d", owner, repo, prNo)
		return 1
	}

	// switch to modules

	if err := util.Run(Env(), providerPath, "go", "mod", "init"); err != nil {
		log.Printf("Error running go mod init in %s: %s", providerPath, err)
		return 1
	}

	if err := os.RemoveAll(filepath.Join(providerPath, "Gopkg.lock")); err != nil {
		log.Printf("Error deleting Gopkg.lock: %s", err)
		return 1
	}

	if err := os.RemoveAll(filepath.Join(providerPath, "Gopkg.toml")); err != nil {
		log.Printf("Error deleting Gopkg.toml: %s", err)
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

	if err := removeGovendorDepFromTravis(providerPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error removing govendor from travis config in %s: %s", providerPath, err)
		return 1
	} else if err != nil {
		log.Printf("No travis file.. skipping step")
	}

	if err := removeGovendorDepFromMakefile(providerPath); err != nil {
		log.Printf("Error removing govendor from makefile in %s: %s", providerPath, err)
		return 1
	}

	if err := setModulesEnvVarsInTravis(providerPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error setting module related env vars in travis file %s: %s", providerPath, err)
		return 1
	} else if err != nil {
		log.Printf("No travis file.. skipping step")
	}

	if err := turnOffModulesForCertainCommandsInMakefile(providerPath); err != nil {
		log.Printf("Error disabling modules for certain commands in makefile %s: %s", providerPath, err)
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

func setModulesEnvVarsInTravis(providerPath string) error {
	filename, content, err := util.ReadOneOf(providerPath, ".travis.yml", ".travis.yaml")
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	// need to ensure newline before EOF
	if lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}

	if util.SearchLines(lines, "GOFLAGS=-mod=vendor", 0) == -1 {
		envLine := util.SearchLines(lines, "env:", 0)
		if envLine == -1 {
			last := len(lines) - 1
			lines = util.InsertLineBefore(lines, last, "env:")
			envLine = util.SearchLines(lines, "env:", last)
		}

		globalLine := util.SearchLines(lines, "global:", envLine)
		if globalLine == -1 {
			lines = util.InsertLineBefore(lines, envLine+1, "  - GOFLAGS=-mod=vendor GO111MODULE=on")
		} else {
			lines = util.InsertLineBefore(lines, globalLine+1, "    - GOFLAGS=-mod=vendor GO111MODULE=on")
		}
	}

	return ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}

func turnOffModulesForCertainCommandsInMakefile(providerPath string) error {
	filename, content, err := util.ReadOneOf(providerPath, "Makefile", "GNUmakefile")
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	// disable modules for commands known to not work in module mode
	for i, line := range lines {
		for _, command := range disableGoModulesFor {
			index := strings.Index(line, command)
			if index > -1 && !strings.Contains(line[:index], "GO111MODULE=off") {
				lines[i] = line[:index] + "GO111MODULE=off " + line[index:]
			}
		}
	}

	return ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}

func removeLineContaining(lines []string, search string) []string {
	if l := util.SearchLines(lines, search, 0); l > -1 {
		lines = util.DeleteLines(lines, l)
	}
	return lines
}

func removeGovendorDepFromTravis(providerPath string) error {
	filename, content, err := util.ReadOneOf(providerPath, ".travis.yml", ".travis.yaml")
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	lines = removeLineContaining(lines, "github.com/kardianos/govendor")
	lines = removeLineContaining(lines, "github.com/golang/dep/cmd/dep")

	if vendorStatusLine := util.SearchLines(lines, "vendor-status", 0); vendorStatusLine > -1 {
		lines = util.DeleteLines(lines, vendorStatusLine)
	}

	return ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}

func removeGovendorDepFromMakefile(providerPath string) error {
	filename, content, err := util.ReadOneOf(providerPath, "Makefile", "GNUmakefile")
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	lines = removeLineContaining(lines, "github.com/kardianos/govendor")
	lines = removeLineContaining(lines, "github.com/golang/dep/cmd/dep")

	if vendorStatusTargetLine := util.SearchLines(lines, "vendor-status:", 0); vendorStatusTargetLine > -1 {
		lines = util.DeleteLines(lines, vendorStatusTargetLine, vendorStatusTargetLine+1)
	}

	out := strings.Join(lines, "\n")
	out = strings.Replace(out, " vendor-status ", " ", -1)
	out = strings.Replace(out, " vendor-status", "", -1)

	return ioutil.WriteFile(filename, []byte(out), 0644)
}

func Env() []string {
	return append(os.Environ(), "GO111MODULE=on")
}
