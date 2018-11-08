package schema

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/appilon/tfplugin/cmd"
	"github.com/mitchellh/cli"
)

const CommandName = "schema"

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

func getPackageName(providerPath string) (string, error) {
	lastDash := strings.LastIndexByte(providerPath, '-')
	if lastDash == -1 || len(providerPath) == lastDash+1 {
		return "", fmt.Errorf("%s does not follow provider naming convention terraform-provider-name", providerPath)
	}
	return providerPath[lastDash+1:], nil
}

func moveVendoredTerraform(path string, undo bool) (bool, error) {
	oldPath := filepath.Join(path, "vendor/github.com/hashicorp/terraform")
	newPath := oldPath
	if undo {
		oldPath += "x"
	} else {
		newPath += "x"
	}

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return false, nil
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return false, err
	}
	return true, nil
}

func (c *command) Run(args []string) int {
	if len(args) != 1 {
		return cli.RunResultHelp
	}

	providerPath := args[0]
	fullPath, err := cmd.FindProviderInGoPath(providerPath)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	packageName, err := getPackageName(providerPath)
	if err != nil {
		log.Printf("Error determining package exporting provider: %s", err)
		return 1
	}

	file, err := os.Create(filepath.Join(fullPath, "dumper.go"))
	if err != nil {
		log.Printf("Cannot create dumper.go: %s", err)
		return 1
	}

	if _, err = fmt.Fprintf(file, dumper, providerPath, packageName); err != nil {
		log.Printf("Could not write to dumper.go: %s", err)
		return 1
	}

	cmd := exec.Command("go", "run", "dumper.go")
	cmd.Dir = fullPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// sigh, golang vendoring woes
	moved, err := moveVendoredTerraform(fullPath, false)
	if err != nil {
		log.Printf("Error moving vendored terraform: %s", err)
		return 1
	}

	// going forward errors don't exit right away, attempt cleanup
	var status int

	if err = cmd.Run(); err != nil {
		log.Printf("go run dumper.go exited with error: %s", err)
		status = 1
	}

	if err = file.Close(); err != nil {
		log.Printf("Error closing dumper.go: %s", err)
		status = 1
	}

	if err = os.Remove(file.Name()); err != nil {
		log.Printf("Error deleting %s: %s", file.Name(), err)
		status = 1
	}

	if moved {
		if _, err = moveVendoredTerraform(fullPath, true); err != nil {
			log.Printf("Error moving back vendored terraform: %s", err)
			status = 1
		}
	}

	return status
}
