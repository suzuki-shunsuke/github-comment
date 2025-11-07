package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v77/github"
)

func (c *Client) PRNumberWithSHA(ctx context.Context, owner, repo, sha string) (int, error) {
	prs, _, err := c.pr.ListPullRequestsWithCommit(ctx, owner, repo, sha, &github.ListOptions{
		PerPage: 1,
	})
	if err != nil {
		return 0, fmt.Errorf("list associated pull requests: %w", err)
	}
	if len(prs) == 0 {
		return 0, errors.New("associated pull request isn't found")
	}
	return prs[0].GetNumber(), nil
}
