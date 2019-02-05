package svc

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var gh *github.Client

func Github() *github.Client {
	if gh == nil {
		token := os.Getenv("GITHUB_PERSONAL_TOKEN")
		if token == "" {
			panic("No GITHUB_PERSONAL_TOKEN set")
		}

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		gh = github.NewClient(tc)
	}

	return gh
}
