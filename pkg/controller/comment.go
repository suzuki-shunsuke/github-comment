package controller

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/github-comment-metadata/metadata"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
)

// GitHub is API to post a comment to GitHub
type GitHub interface {
	CreateComment(ctx context.Context, cmt *github.Comment) error
	ListComments(ctx context.Context, pr *github.PullRequest) ([]*github.IssueComment, error)
	HideComment(ctx context.Context, nodeID string) error
	GetAuthenticatedUser(ctx context.Context) (string, error)
	PRNumberWithSHA(ctx context.Context, owner, repo, sha string) (int, error)
}

type CommentController struct {
	GitHub   GitHub
	Expr     Expr
	Getenv   func(string) string
	Platform Platform
}

func (c *CommentController) Post(ctx context.Context, cmt *github.Comment) error {
	if err := c.GitHub.CreateComment(ctx, cmt); err != nil {
		return fmt.Errorf("send a comment: %w", err)
	}
	return nil
}

func extractMetaFromComment(body string, data *map[string]any) bool {
	f, _ := metadata.Extract(body, data)
	return f
}

func (c *CommentController) complementMetaData(data map[string]any) {
	if data == nil {
		return
	}
	if c.Platform == nil {
		return
	}
	_ = metadata.SetCIEnv(c.Platform.CI(), c.Getenv, data)
}

func (c *CommentController) getEmbeddedComment(data map[string]any) (string, error) {
	c.complementMetaData(data)
	return metadata.Convert(data) //nolint:wrapcheck
}
