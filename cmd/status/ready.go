package status

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/appilon/tfplugin/cmd/upgrade/modules"
	"github.com/appilon/tfplugin/svc"
	"github.com/appilon/tfplugin/util"
	"github.com/google/go-github/github"
)

func listReadyProviders() int {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 200},
	}
	q := fmt.Sprintf(`org:terraform-providers "%s" in:title is:issue is:open`, modules.IssueTitle)
	for {
		res, resp, err := svc.Github().Search.Issues(context.TODO(), q, opt)
		for i := range res.Issues {
			printIfUpvoted(&res.Issues[i])
		}
		if err != nil {
			log.Printf("Error searching for issues: %s", err)
			return 1
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return 0
}

func printIfUpvoted(issue *github.Issue) {
	var upvotes int
	var downvotes int

	owner, repo, err := util.GetGitHubDetails(issue.GetRepositoryURL())
	if err != nil {
		log.Printf("Error parsing owner/repo: %s", err)
		return
	}
	id := issue.GetNumber()

	defer func() {
		if upvotes > downvotes {
			fmt.Printf("github.com/%s/%s", owner, repo)
			fmt.Println()
		}
	}()

	// issues returned from search aren't fully populated
	issue, _, err = svc.Github().Issues.Get(context.TODO(), owner, repo, id)
	if err != nil {
		log.Printf("Error fully fetching %s/%s/%d: %s", owner, repo, id, err)
		return
	}

	upvotes += issue.Reactions.GetPlusOne()
	downvotes += issue.Reactions.GetMinusOne()

	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Direction:   "asc",
	}

	for {
		comments, resp, err := svc.Github().Issues.ListComments(context.TODO(), owner, repo, issue.GetNumber(), opt)
		for _, comment := range comments {
			// strip quoted text to avoid the original issue message from counting as an upvote or downvote
			msg := removeQuotedText(comment.GetBody())
			// checking for emojis is weird.... to my knowledge skin tone modifiers have an extra character
			// so this should match all variations of thumbs up?
			if strings.Contains(msg, "👍") {
				upvotes++
			}
			if strings.Contains(msg, "👎") {
				downvotes++
			}
		}
		if err != nil {
			log.Printf("Error fetching comments for %s/%s/%d: %s", owner, repo, issue.GetNumber(), err)
			return
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

func removeQuotedText(body string) string {
	var cleaned []string
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), ">") {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}
