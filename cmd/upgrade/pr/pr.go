package pr

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/appilon/tfplugin/svc"
	"github.com/appilon/tfplugin/util"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"github.com/pkg/browser"
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
	var closes int
	var provider string
	flags.StringVar(&branch, "branch", "", "name of branch to pull request")
	flags.StringVar(&remote, "remote", "origin", "remote to push to")
	flags.StringVar(&provider, "provider", "", "provider to pull request")
	flags.StringVar(&base, "base", "master", "base branch to open pull request against")
	flags.StringVar(&user, "user", "", "github user/org for cross account pull requests")
	flags.StringVar(&title, "title", "[AUTOMATED] tfplugin pull request", "title of the pull request")
	flags.IntVar(&closes, "closes", 0, "PR closes issue #")
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

	var body string
	if closes > 0 {
		body = fmt.Sprintf(`Closes #%d`, closes)
	}
	if pr, err := openPullRequest(providerPath, base, branch, user, title, body); err != nil {
		log.Printf("Error opening pull request: %s", err)
		return 1
	} else if open {
		if err := browser.OpenURL(pr.GetHTMLURL()); err != nil {
			log.Printf("Error opening %s in browser: %s", pr.GetHTMLURL(), err)
		}
	}

	return 0
}

func openPullRequest(providerPath, base, head, user, title string, body string) (*github.PullRequest, error) {
	owner, repo, err := util.GetGitHubDetails(providerPath)
	if err != nil {
		return nil, err
	}

	if user != "" {
		head = user + ":" + head
	}

	pr, _, err := svc.Github().PullRequests.Create(context.TODO(), owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	})
	if err != nil {
		return nil, err
	}
	os.Stderr.WriteString(fmt.Sprintf("\nPull request created, view at: %s\n", pr.GetHTMLURL()))

	return pr, nil
}
