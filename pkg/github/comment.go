package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v76/github"
	"github.com/sirupsen/logrus"
)

type Comment struct {
	PRNumber       int
	CommentID      int64
	Org            string
	Repo           string
	Body           string
	BodyForTooLong string
	SHA1           string
	TemplateKey    string
	Vars           map[string]any
}

// `graphql:"IssueComment(isMinimized: false, viewerCanMinimize: true)"`
type IssueComment struct {
	ID         string
	DatabaseID int64
	Body       string
	Author     struct {
		Login string
	}
	CreatedAt string
	// TODO remove
	IsMinimized       bool
	ViewerCanMinimize bool
}

func (c *Client) sendIssueComment(ctx context.Context, cmt *Comment, body string) error {
	if cmt.CommentID != 0 {
		if _, _, err := c.issue.EditComment(ctx, cmt.Org, cmt.Repo, cmt.CommentID, &github.IssueComment{
			Body: github.Ptr(body),
		}); err != nil {
			return fmt.Errorf("edit a issue or pull request comment by GitHub API: %w", err)
		}
		return nil
	}
	if _, _, err := c.issue.CreateComment(ctx, cmt.Org, cmt.Repo, cmt.PRNumber, &github.IssueComment{
		Body: github.Ptr(body),
	}); err != nil {
		return fmt.Errorf("create a comment to issue or pull request by GitHub API: %w", err)
	}
	return nil
}

func (c *Client) sendCommitComment(ctx context.Context, cmt *Comment, body string) error {
	if cmt.CommentID != 0 {
		if _, _, err := c.repo.UpdateComment(ctx, cmt.Org, cmt.Repo, cmt.CommentID, &github.RepositoryComment{
			Body: github.Ptr(body),
		}); err != nil {
			return fmt.Errorf("update a commit comment by GitHub API: %w", err)
		}
		return nil
	}
	if _, _, err := c.repo.CreateComment(ctx, cmt.Org, cmt.Repo, cmt.SHA1, &github.RepositoryComment{
		Body: github.Ptr(body),
	}); err != nil {
		return fmt.Errorf("create a commit comment by GitHub API: %w", err)
	}
	return nil
}

func (c *Client) createComment(ctx context.Context, cmt *Comment, tooLong bool) error {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	body := cmt.Body
	if tooLong {
		logE.WithFields(logrus.Fields{
			"body_length": len(body),
		}).Warn("body is too long so it is replaced with `BodyForTooLong`")
		body = cmt.BodyForTooLong
	}
	if cmt.PRNumber != 0 {
		return c.sendIssueComment(ctx, cmt, body)
	}
	return c.sendCommitComment(ctx, cmt, body)
}

func (c *Client) CreateComment(ctx context.Context, cmt *Comment) error {
	return c.createComment(ctx, cmt, len(cmt.Body) > 65536) //nolint:mnd
}
