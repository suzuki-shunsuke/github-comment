package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"
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
	HideOldComment string
	TemplateKey    string
	Vars           map[string]interface{}
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

func (client *Client) sendIssueComment(ctx context.Context, cmt *Comment, body string) error {
	if cmt.CommentID != 0 {
		if _, _, err := client.issue.EditComment(ctx, cmt.Org, cmt.Repo, cmt.CommentID, &github.IssueComment{
			Body: github.String(body),
		}); err != nil {
			return fmt.Errorf("edit a issue or pull request comment by GitHub API: %w", err)
		}
		return nil
	}
	if _, _, err := client.issue.CreateComment(ctx, cmt.Org, cmt.Repo, cmt.PRNumber, &github.IssueComment{
		Body: github.String(body),
	}); err != nil {
		return fmt.Errorf("create a comment to issue or pull request by GitHub API: %w", err)
	}
	return nil
}

func (client *Client) sendCommitComment(ctx context.Context, cmt *Comment, body string) error {
	if cmt.CommentID != 0 {
		if _, _, err := client.repo.UpdateComment(ctx, cmt.Org, cmt.Repo, cmt.CommentID, &github.RepositoryComment{
			Body: github.String(body),
		}); err != nil {
			return fmt.Errorf("update a commit comment by GitHub API: %w", err)
		}
		return nil
	}
	if _, _, err := client.repo.CreateComment(ctx, cmt.Org, cmt.Repo, cmt.SHA1, &github.RepositoryComment{
		Body: github.String(body),
	}); err != nil {
		return fmt.Errorf("create a commit comment by GitHub API: %w", err)
	}
	return nil
}

func (client *Client) createComment(ctx context.Context, cmt *Comment, tooLong bool) error {
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
		return client.sendIssueComment(ctx, cmt, body)
	}
	return client.sendCommitComment(ctx, cmt, body)
}

func (client *Client) CreateComment(ctx context.Context, cmt *Comment) error {
	return client.createComment(ctx, cmt, len(cmt.Body) > 65536) //nolint:gomnd
}
