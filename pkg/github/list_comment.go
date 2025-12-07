package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

type PullRequest struct {
	PRNumber int
	Org      string
	Repo     string
}

func (c *Client) listIssueComment(ctx context.Context, pr *PullRequest) ([]*IssueComment, error) { //nolint:dupl
	// https://github.com/shurcooL/githubv4#pagination
	var q struct {
		Repository struct {
			Issue struct {
				Comments struct {
					Nodes    []*IssueComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]any{
		"repositoryOwner": githubv4.String(pr.Org),
		"repositoryName":  githubv4.String(pr.Repo),
		"issueNumber":     githubv4.Int(pr.PRNumber),
		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var allComments []*IssueComment
	for {
		if err := c.ghV4.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
		}
		allComments = append(allComments, q.Repository.Issue.Comments.Nodes...)
		if !q.Repository.Issue.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.Issue.Comments.PageInfo.EndCursor)
	}
	return allComments, nil
}

func (c *Client) listPRComment(ctx context.Context, pr *PullRequest) ([]*IssueComment, error) { //nolint:dupl
	// https://github.com/shurcooL/githubv4#pagination
	var q struct {
		Repository struct {
			PullRequest struct {
				Comments struct {
					Nodes    []*IssueComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
			} `graphql:"pullRequest(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]any{
		"repositoryOwner": githubv4.String(pr.Org),
		"repositoryName":  githubv4.String(pr.Repo),
		"issueNumber":     githubv4.Int(pr.PRNumber),
		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var allComments []*IssueComment
	for {
		if err := c.ghV4.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
		}
		allComments = append(allComments, q.Repository.PullRequest.Comments.Nodes...)
		if !q.Repository.PullRequest.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.PullRequest.Comments.PageInfo.EndCursor)
	}
	return allComments, nil
}

func (c *Client) ListComments(ctx context.Context, pr *PullRequest) ([]*IssueComment, error) {
	cmts, prErr := c.listPRComment(ctx, pr)
	if prErr == nil {
		return cmts, nil
	}
	cmts, err := c.listIssueComment(ctx, pr)
	if err == nil {
		return cmts, nil
	}
	return nil, fmt.Errorf("get pull request or issue comments: %w, %w", prErr, err)
}
