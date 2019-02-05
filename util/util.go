package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

func GetGitHubDetails(providerPath string) (string, string, error) {
	// format is .../owner/repo
	parts := strings.Split(providerPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("%s should follow '.../owner/repo' format", providerPath)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
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

func ReadOneOf(dir string, filenames ...string) (fullpath string, content []byte, err error) {
	for _, filename := range filenames {
		fullpath = filepath.Join(dir, filename)
		content, err = ioutil.ReadFile(fullpath)
		if err == nil {
			break
		}
	}
	return
}

func SearchLines(lines []string, search string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], search) {
			return i
		}
	}
	return -1
}

func SetLine(lines []string, index int, line string) []string {
	if index < len(lines) {
		lines[index] = line
	} else {
		lines = append(lines, line)
	}
	return lines
}

// taken from https://github.com/golang/go/wiki/SliceTricks
func InsertLineBefore(lines []string, index int, line string) []string {
	return append(lines[:index], append([]string{line}, lines[index:]...)...)
}

func DeleteLines(lines []string, toDelete ...int) []string {
	sort.Ints(toDelete)
	for i, index := range toDelete {
		index -= i
		lines = append(lines[:index], lines[index+1:]...)
	}
	return lines
}
