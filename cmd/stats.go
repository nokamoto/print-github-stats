package cmd

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"time"
)

const (
	debugKey  = "debug"
	urlKey    = "url"
	orgKey    = "org"
	startKey  = "start-date"
	endKey    = "end-date"
	outputKey = "output"
	sleepKey  = "sleep"
)

const (
	jsonValue = "json"
	csvValue  = "csv"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Prints contributions to a github repository",
	Long:  `Prints contributions to a github repository.`,
	Run:   stats,
}

func init() {
	flags := statsCmd.Flags()

	flags.String(urlKey, "https://<enterprise>/api/v3/", `github api url
https://developer.github.com/v3
https://developer.github.com/enterprise/2.19/v3/enterprise-admin
`)
	viper.BindPFlag(urlKey, flags.Lookup(urlKey))

	flags.String(orgKey, "<org>", `github organization`)
	viper.BindPFlag(orgKey, flags.Lookup(orgKey))

	flags.Bool(debugKey, false, "debug mode")
	viper.BindPFlag(debugKey, flags.Lookup(debugKey))

	now := time.Now()
	flags.String(endKey, fmt.Sprintf("%d-%2d-%2d", now.Year(), now.Month(), now.Day()), `retrieves pull requests created before the date`)
	viper.BindPFlag(endKey, flags.Lookup(endKey))

	now = now.AddDate(0, -1, 0)
	flags.String(startKey, fmt.Sprintf("%d-%2d-%2d", now.Year(), now.Month(), now.Day()), `retrieves pull requests created after the date`)
	viper.BindPFlag(startKey, flags.Lookup(startKey))

	flags.StringP(outputKey, "o", "json", `Output format: json or csv`)
	viper.BindPFlag(outputKey, flags.Lookup(outputKey))

	flags.String(sleepKey, "1s", `Adds the delay between HTTP requests
https://golang.org/pkg/time/#ParseDuration`)
	viper.BindPFlag(sleepKey, flags.Lookup(sleepKey))
}

func listRepos(ctx context.Context, client *github.Client, org string) []*github.Repository {
	opts := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 30}}

	var repos []*github.Repository
	for {
		result, res, err := client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			fatal("failed: %v", err)
		}

		debug("%v", *res)

		repos = append(repos, result...)

		if res.NextPage == 0 {
			break
		}

		opts.Page = res.NextPage
	}

	return repos
}

type PullRequestState struct {
	Repository Repository
	Pull       *github.PullRequest
	Reviews    []*github.PullRequestReview
	Comments   []*github.PullRequestComment
}

func sleep() {
	d, err := time.ParseDuration(viper.GetString(sleepKey))
	if err != nil {
		fatal("invalid duration: %v", err)
	}

	time.Sleep(d)
}

func listReviews(ctx context.Context, client *github.Client, org string, repo string, num int) []*github.PullRequestReview {
	opts := &github.ListOptions{
		PerPage: 30,
	}
	var reviews []*github.PullRequestReview
	for {
		sleep()

		result, res, err := client.PullRequests.ListReviews(ctx, org, repo, num, opts)
		if err != nil {
			fatal("failed: %v", err)
		}

		debug("%v", res)

		debug("%d review(s)", len(result))

		reviews = append(reviews, result...)

		if res.NextPage == 0 {
			break
		}

		opts.Page = res.NextPage
	}
	return reviews
}

func listComments(ctx context.Context, client *github.Client, org string, repo string, num int) []*github.PullRequestComment {
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 30,
		},
	}
	var comments []*github.PullRequestComment
	for {
		sleep()

		result, res, err := client.PullRequests.ListComments(ctx, org, repo, num, opts)
		if err != nil {
			fatal("failed: %v", err)
		}

		debug("%v", res)

		debug("%d comment(s)", len(result))

		comments = append(comments, result...)

		if res.NextPage == 0 {
			break
		}

		opts.Page = res.NextPage
	}
	return comments
}

func describePull(ctx context.Context, client *github.Client, org string, repo string, num int) PullRequestState {
	sleep()

	result, res, err := client.PullRequests.Get(ctx, org, repo, num)
	if err != nil {
		fatal("failed: %v", err)
	}

	debug("%v", *res)

	reviews := listReviews(ctx, client, org, repo, num)

	comments := listComments(ctx, client, org, repo, num)

	return PullRequestState{
		Repository: Repository(repo),
		Pull:       result,
		Reviews:    reviews,
		Comments:   comments,
	}
}

func listPulls(ctx context.Context, client *github.Client, org string, repo string, st, et time.Time) []PullRequestState {
	var pulls []PullRequestState

	opts := &github.PullRequestListOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: 30},
	}

	for {
		sleep()

		result, res, err := client.PullRequests.List(ctx, org, repo, opts)
		if err != nil {
			fatal("failed: %v", err)
		}

		debug("%v", *res)

		out := false
		for _, pull := range result {
			if pull.GetCreatedAt().Before(st) {
				debug("break: #%d created at %v", pull.GetNumber(), pull.GetCreatedAt())
				out = true
				break
			}
			if pull.GetCreatedAt().Before(et) {
				debug("#%d created at %v", pull.GetNumber(), pull.GetCreatedAt())
				pulls = append(pulls, describePull(ctx, client, org, repo, pull.GetNumber()))
			} else {
				debug("skip: #%d created at %v", pull.GetNumber(), pull.GetCreatedAt())
			}
		}
		if out {
			break
		}

		if res.NextPage == 0 {
			break
		}

		opts.Page = res.NextPage
	}

	return pulls
}

func stats(_ *cobra.Command, _ []string) {
	rawUrl := viper.GetString(urlKey)
	org := viper.GetString(orgKey)
	start := viper.GetString(startKey)
	end := viper.GetString(endKey)
	output := viper.GetString(outputKey)

	layout := "2006-01-02"
	st, err := time.Parse(layout, start)
	if err != nil {
		fatal("failed: %v", err)
	}

	et, err := time.Parse(layout, end)
	if err != nil {
		fatal("failed: %v", err)
	}

	if st.After(et) {
		fatal("start %v >= end %v", st, et)
	}

	if output != jsonValue && output != csvValue {
		fatal("invalid output: %s", output)
	}

	ctx := context.Background()

	debug("%s", rawUrl)
	client := github.NewClient(nil)

	baseUrl, err := url.Parse(rawUrl)
	if err != nil {
		fatal("failed: %v", err)
	}

	client.BaseURL = baseUrl

	debug("GET /orgs/%s/repos", org)

	repos := listRepos(ctx, client, org)

	contributors := &Contributors{
		Org:   org,
		Start: start,
		End:   end,
	}

	for _, repo := range repos {
		debug("GET /repos/%s/%s/pulls (from %s to %s)", org, *repo.Name, start, end)

		pulls := listPulls(ctx, client, org, *repo.Name, st, et)

		for _, pull := range pulls {
			contributors.stats(pull)
		}
	}

	switch output {
	case jsonValue:
		contributors.printJSON()
	case csvValue:
		contributors.printCSV()
	}
}
