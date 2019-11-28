package cmd

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
}

type Contributors struct {
	Contributors []*Contributor
}

func (c *Contributor) contribution(repo Repository) *Contribution {
	found, ok := c.Contributions[repo]
	if ok {
		return found
	}

	cb := &Contribution{}
	c.Contributions[repo] = cb

	return cb
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
}
