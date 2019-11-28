package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/google/go-github/v28/github"
	"net/url"
	"time"
)

const (
	startKey = "start-date"
	endKey = "end-date"
)

var statsCmd =  &cobra.Command{
	Use:   "stats",
	Short: "Prints contributions to a github repository",
	Long: `Prints contributions to a github repository.`,
	Run: stats,
}

func init() {
	flags := statsCmd.Flags()

	now := time.Now()
	flags.String(endKey, fmt.Sprintf("%d-%2d-%2d", now.Year(), now.Month(), now.Day()), `retrieves pull requests created before the date`)
	viper.BindPFlag(endKey, flags.Lookup(endKey))

	now = now.AddDate(0, -1, 0)
	flags.String(startKey, fmt.Sprintf("%d-%2d-%2d", now.Year(), now.Month(), now.Day()), `retrieves pull requests created after the date`)
	viper.BindPFlag(startKey, flags.Lookup(startKey))
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
	Pull *github.PullRequest
	Reviews []*github.PullRequestReview
}

func describePull(ctx context.Context, client *github.Client, org string, repo string, num int) PullRequestState {
	result, res, err := client.PullRequests.Get(ctx, org, repo, num)
	if err != nil {
		fatal("failed: %v", err)
	}

	debug("%v", *res)

	debug("%d %d %d", result.GetDeletions(), result.GetAdditions(), result.GetMerged())

	opts := &github.ListOptions{
		PerPage: 30,
	}
	var reviews []*github.PullRequestReview
	for {
		result, res, err := client.PullRequests.ListReviews(ctx, org, repo, num, opts)
		if err != nil {
			fatal("failed: %v", err)
		}

		debug("%v", res)

		reviews = append(reviews, result...)

		if res.NextPage == 0 {
			break
		}

		opts.Page = res.NextPage
	}

	return PullRequestState{
		Repository: Repository(repo),
		Pull: result,
		Reviews: reviews,
	}
}

func listPulls(ctx context.Context, client *github.Client, org string, repo string, st, et time.Time) []PullRequestState {
	var pulls []PullRequestState

	opts := &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{PerPage: 30},
	}

	for {
		result, res, err := client.PullRequests.List(ctx, org, repo, opts)
		if err != nil {
			fatal("failed: %v", err)
		}

		debug("%v", *res)

		out := false
		for _, pull := range result {
			if pull.GetCreatedAt().Before(st) {
				out = true
				break
			}
			if pull.GetCreatedAt().After(et) {
				// ignore
				break
			}

			pulls = append(pulls, describePull(ctx, client, org, repo, pull.GetNumber()))
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

	contributors := &Contributors{}

	for _, repo := range repos {
		debug("GET /repos/%s/%s/pulls (from %s to %s)", org, *repo.Name, start, end)

		pulls := listPulls(ctx, client, org, *repo.Name, st, et)

		for _, pull := range pulls {
			contributors.stats(pull)
		}
	}

	data, err := json.Marshal(contributors)
	if err != nil {
		fatal("failed: %v", err)
	}

	fmt.Println(string(data))
}
