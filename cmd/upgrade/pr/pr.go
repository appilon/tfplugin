package pr

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/appilon/tfplugin/util"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"github.com/pkg/browser"
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
	var branch string
	var message string
	var remote string
	var open bool
	var openBrowser bool
	var base string
	var user string
	var title string
	var provider string
	flags.StringVar(&branch, "branch", "tfplugin-upgrade-"+time.Now().Format("2006-01-02-15-04-05"), "name of branch to create")
	flags.StringVar(&message, "message", "sdk upgrade", "commit message")
	flags.StringVar(&remote, "remote", "origin", "remote to push to")
	flags.StringVar(&provider, "provider", "", "provider to pull request")
	flags.StringVar(&base, "base", "master", "base branch to open pull request against")
	flags.StringVar(&user, "user", "", "github user/org for cross account pull requests")
	flags.StringVar(&title, "title", "sdk upgrade", "title of the pull request")
	flags.BoolVar(&open, "open", false, "open pull request automatically, requires GITHUB_PERSONAL_TOKEN")
	flags.BoolVar(&openBrowser, "browser", false, "open created pull request in browser")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	if err = util.Run(providerPath, "git", "checkout", "-b", branch); err != nil {
		log.Printf("Error creating git branch %q: %s", branch, err)
		return 1
	}

	if err = util.Run(providerPath, "git", "add", "--all"); err != nil {
		log.Printf("Error adding files: %s", err)
		return 1
	}

	if err = util.Run(providerPath, "git", "commit", "-m", message); err != nil {
		log.Printf("Error committing: %s", err)
		return 1
	}

	if err = util.Run(providerPath, "git", "push", remote, branch); err != nil {
		log.Printf("Error pushing to %s: %s", remote, err)
		return 1
	}

	if open {
		if err = openPullRequest(providerPath, base, branch, user, title, openBrowser); err != nil {
			log.Printf("Error opening pull request: %s", err)
			return 1
		}
	}

	return 0
}

func openPullRequest(providerPath, base, head, user, title string, openBrowser bool) error {
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

	owner, repo, err := getGitHubDetails(providerPath)
	if err != nil {
		return err
	}

	if user != "" {
		head = user + ":" + head
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
	})
	if err != nil {
		return err
	}
	fmt.Printf("\nPull request created, view at: %s\n", pr.GetHTMLURL())

	if openBrowser {
		err = browser.OpenURL(pr.GetHTMLURL())
	}
	return err
}

func getGitHubDetails(providerPath string) (string, string, error) {
	// format is .../owner/repo
	parts := strings.Split(providerPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("%s should follow '.../owner/repo' format", providerPath)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
