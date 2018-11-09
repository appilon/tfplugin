package pr

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/appilon/tfplugin/cmd"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"golang.org/x/oauth2"
)

const CommandName = "upgrade pr"

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
	var branch string
	var message string
	var remote string
	var open bool
	var providerPath string
	flags.StringVar(&branch, "branch", "tfplugin-upgrade-"+time.Now().Format("2006-01-02-15-04-05"), "name of branch to create")
	flags.StringVar(&message, "message", "sdk upgrade", "commit message")
	flags.StringVar(&remote, "remote", "origin", "remote to push to")
	flags.StringVar(&providerPath, "provider", "", "provider to PR")
	flags.BoolVar(&open, "open", false, "Optionally open PR automatically, requires GITHUB_PERSONAL_TOKEN")
	flags.Parse(args)

	providerPath, err = cmd.FindProvider(providerPath)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	if err = run(providerPath, "git", "checkout", "-b", branch); err != nil {
		log.Printf("Error creating git branch %q: %s", branch, err)
		return 1
	}

	if err = run(providerPath, "git", "add", "--all"); err != nil {
		log.Printf("Error adding files: %s", err)
		return 1
	}

	if err = run(providerPath, "git", "commit", "-m", message); err != nil {
		log.Printf("Error committing: %s", err)
		return 1
	}

	if err = run(providerPath, "git", "push", remote, branch); err != nil {
		log.Printf("Error pushing to %s: %s", remote, err)
		return 1
	}

	if open {
		if err = openPullRequest(providerPath, remote, branch); err != nil {
			log.Printf("Error opening pull request: %s", err)
			return 1
		}
	}

	return 0
}

func run(dir, name string, arg ...string) error {
	fmt.Printf("==> %s %s\n", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func openPullRequest(providerPath, remote, branch string) error {
	token := os.Getenv("GITHUB_PERSONAL_TOKEN")
	if token == "" {
		return errors.New("No GITHUB_PERSONAL_TOKEN set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return nil
}
