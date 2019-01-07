package pr

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
	var remote string
	var open bool
	var base string
	var user string
	var title string
	var provider string
	flags.StringVar(&branch, "branch", "", "name of branch to pull request")
	flags.StringVar(&remote, "remote", "origin", "remote to push to")
	flags.StringVar(&provider, "provider", "", "provider to pull request")
	flags.StringVar(&base, "base", "master", "base branch to open pull request against")
	flags.StringVar(&user, "user", "", "github user/org for cross account pull requests")
	flags.StringVar(&title, "title", "[AUTOMATED] sdk upgrade", "title of the pull request")
	flags.BoolVar(&open, "open", false, "open created pull request in browser")
	flags.Parse(args)

	providerPath, err := util.FindProvider(provider)
	if err != nil {
		log.Printf("Error finding provider: %s", err)
		return 1
	}

	if err := util.Run(os.Environ(), providerPath, "git", "push", remote, branch); err != nil {
		log.Printf("Error pushing to %s/%s: %s", remote, branch, err)
		return 1
	}

	if pr, err := openPullRequest(providerPath, base, branch, user, title); err != nil {
		log.Printf("Error opening pull request: %s", err)
		return 1
	} else if open {
		if err := browser.OpenURL(pr.GetHTMLURL()); err != nil {
			log.Printf("Error opening %s in browser: %s", pr.GetHTMLURL(), err)
		}
	}

	return 0
}

func openPullRequest(providerPath, base, head, user, title string) (*github.PullRequest, error) {
	token := os.Getenv("GITHUB_PERSONAL_TOKEN")
	if token == "" {
		return nil, errors.New("No GITHUB_PERSONAL_TOKEN set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner, repo, err := getGitHubDetails(providerPath)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	os.Stderr.WriteString(fmt.Sprintf("\nPull request created, view at: %s\n", pr.GetHTMLURL()))

	return pr, nil
}

func getGitHubDetails(providerPath string) (string, string, error) {
	// format is .../owner/repo
	parts := strings.Split(providerPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("%s should follow '.../owner/repo' format", providerPath)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
