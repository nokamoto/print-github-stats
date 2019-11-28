package cmd

import "time"

type Repository string

type Contribution struct {
	Approve         int
	Deletions       int
	Additions       int
	MergedDeletions int
	MergedAdditions int
	Reviews         int
	Comments        int
}

type Contributor struct {
	Name          string
	Contributions map[Repository]*Contribution
	Summary       *Contribution
}

type Contributors struct {
	Org          string
	Start        time.Time
	End          time.Time
	Contributors []*Contributor
}

func (c *Contributor) contribution(repo Repository) *Contribution {
	found, ok := c.Contributions[repo]
	if ok {
		return found
	}

	cb := &Contribution{}
	c.Contributions[repo] = cb
	c.Summary = new(Contribution)

	return cb
}

func (c *Contributor) summarize() {
	for _, cb := range c.Contributions {
		c.Summary.Approve += cb.Approve
		c.Summary.Deletions += cb.Deletions
		c.Summary.Additions += cb.Additions
		c.Summary.MergedDeletions += cb.MergedDeletions
		c.Summary.MergedAdditions += cb.MergedAdditions
		c.Summary.Comments += cb.Comments
		c.Summary.Reviews += cb.Reviews
	}
}

func (cs *Contributors) contributor(name string) *Contributor {
	for _, c := range cs.Contributors {
		if c.Name == name {
			return c
		}
	}

	c := &Contributor{Name: name, Contributions: make(map[Repository]*Contribution)}
	cs.Contributors = append(cs.Contributors, c)

	return c
}

func (cs *Contributors) stats(pull PullRequestState) {
	author := pull.Pull.GetUser().GetLogin()
	authorContribution := cs.contributor(author).contribution(pull.Repository)

	if pull.Pull.GetMerged() {
		authorContribution.MergedAdditions += pull.Pull.GetAdditions()
		authorContribution.MergedDeletions += pull.Pull.GetDeletions()
	}

	authorContribution.Additions += pull.Pull.GetAdditions()
	authorContribution.Deletions += pull.Pull.GetDeletions()

	approvedUniq := make(map[string]struct{})
	for _, review := range pull.Reviews {
		who := review.GetUser().GetLogin()

		if review.GetState() == "APPROVED" {
			approvedUniq[who] = struct{}{}
		}

		cs.contributor(who).contribution(pull.Repository).Reviews += 1
	}

	for who := range approvedUniq {
		cs.contributor(who).contribution(pull.Repository).Approve += 1
	}

	for _, comment := range pull.Comments {
		who := comment.GetUser().GetLogin()
		cs.contributor(who).contribution(pull.Repository).Comments += 1
	}

	for _, c := range cs.Contributors {
		c.summarize()
	}
}
