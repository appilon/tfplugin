package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func FindProviderInGoPath(providerPath string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		errors.New("GOPATH is empty")
	}
	gopaths := filepath.SplitList(gopath)

	for _, p := range gopaths {
		fullPath := filepath.Join(p, "src", providerPath)
		info, err := os.Stat(fullPath)

		if err == nil {
			if !info.IsDir() {
				return "", fmt.Errorf("%s is not a directory", fullPath)
			} else {
				return fullPath, nil
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("Could not find %s in GOPATH: %s", providerPath, gopath)
}
