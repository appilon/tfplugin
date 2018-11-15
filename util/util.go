package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FindProvider(providerPath string) (string, error) {
	if providerPath == "" {
		return os.Getwd()
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errors.New("GOPATH is empty")
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

func GetPackageName(providerPath string) (string, error) {
	lastDash := strings.LastIndexByte(providerPath, '-')
	if lastDash == -1 || len(providerPath) == lastDash+1 {
		return "", fmt.Errorf("%s does not follow plugin naming convention terraform-{type}-{name}", providerPath)
	}
	return providerPath[lastDash+1:], nil
}

func Run(env []string, dir, name string, arg ...string) error {
	os.Stderr.WriteString(fmt.Sprintf("==> %s %s\n", name, strings.Join(arg, " ")))
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = env
	return cmd.Run()
}
