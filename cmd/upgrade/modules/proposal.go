package modules

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/appilon/tfplugin/svc"
	"github.com/appilon/tfplugin/util"
	"github.com/google/go-github/github"
)

const IssueTitle = "[PROPOSAL] Switch to Go Modules"
const issueBody = `As part of the preparation for Terraform v0.12, we would like to migrate all providers to use [Go Modules](https://github.com/golang/go/wiki/Modules). We plan to continue checking dependencies into vendor/ to remain compatible with existing tooling/CI for a period of time, however go modules will be used for management. Go Modules is the official solution for the go programming language, we understand some providers might not want this change yet, however we encourage providers to begin looking towards the switch as this is how we will be managing all Go projects in the future. Would maintainers please react with :+1: for support, or :-1: if you wish to have this provider omitted from the first wave of pull requests. If your provider is in support, we would ask that you avoid merging any pull requests that mutate the dependencies while the Go Modules PR is open (in fact a total codefreeze would be even more helpful), otherwise we will need to close that PR and re-run %go mod init%. Once merged, dependencies can be added or updated as follows:

%%%
$ GO111MODULE=on go get github.com/some/module@master
$ GO111MODULE=on go mod tidy
$ GO111MODULE=on go mod vendor
%%%

GO111MODULE=on might be unnecessary depending on your environment, this example will fetch a module @ master and record it in your project's go.mod and go.sum files. It's a good idea to tidy up afterward and then copy the dependencies into vendor/. To remove dependencies from your project, simply remove all usage from your codebase and run:

%%%
$ GO111MODULE=on go mod tidy
$ GO111MODULE=on go mod vendor
%%%

Thank you sincerely for all your time, contributions, and cooperation!`

func proposeGoModules(providerPath string) int {
	if _, err := os.Stat(filepath.Join(providerPath, "go.mod")); !os.IsNotExist(err) {
		log.Printf("%s/go.mod exists or some other error occured... skipping", providerPath)
		return 0
	}

	owner, repo, err := util.GetGitHubDetails(providerPath)
	if err != nil {
		log.Printf("Error determining repo details: %s", err)
		return 1
	}

	if issueNo, err := issueExists(owner, repo, IssueTitle); err != nil {
		log.Printf("Error searching for GH issue w/ title %q: %s", IssueTitle, err)
		return 1
	} else if issueNo > 0 {
		log.Printf("Issue #%d already exists... skipping", issueNo)
		return 0
	}

	issueNo, err := openIssue(owner, repo, IssueTitle, issueBody)

	if err != nil {
		log.Printf("Error opening GH issue: %s", err)
		return 1
	}

	log.Printf("Opened GH issue#%d", issueNo)

	return 0
}

func pullRequestExists(owner, repo, title string) (int, error) {
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "open",
	}

	for {
		prs, resp, err := svc.Github().PullRequests.List(context.TODO(), owner, repo, opt)
		if err != nil {
			return 0, err
		}

		for _, pr := range prs {
			if strings.Contains(pr.GetTitle(), title) {
				return pr.GetNumber(), nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return 0, nil
}

func issueExists(owner, repo, title string) (int, error) {
	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "open",
	}

	for {
		issues, resp, err := svc.Github().Issues.ListByRepo(context.TODO(), owner, repo, opt)
		if err != nil {
			return 0, err
		}

		for _, issue := range issues {
			if strings.Contains(issue.GetTitle(), title) {
				return issue.GetNumber(), nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return 0, nil
}

func openIssue(owner, repo, title, body string) (int, error) {
	issue, _, err := svc.Github().Issues.Create(context.TODO(), owner, repo, &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(strings.Replace(issueBody, "%", "`", -1)),
	})

	if err != nil {
		return 0, err
	}

	return issue.GetNumber(), nil
}
