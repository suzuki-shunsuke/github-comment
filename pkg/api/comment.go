package api

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

func (ctrl *CommentController) Post(ctx context.Context, cmt *github.Comment) error {
	if err := ctrl.GitHub.CreateComment(ctx, cmt); err != nil {
		return fmt.Errorf("send a comment: %w", err)
	}
	return nil
}

func extractMetaFromComment(body string, data *map[string]interface{}) bool {
	f, _ := metadata.Extract(body, data)
	return f
}

func (ctrl *CommentController) complementMetaData(data map[string]interface{}) {
	if data == nil {
		return
	}
	if ctrl.Platform == nil {
		return
	}
	_ = metadata.SetCIEnv(ctrl.Platform.CI(), ctrl.Getenv, data)
}

func (ctrl *CommentController) getEmbeddedComment(data map[string]interface{}) (string, error) {
	ctrl.complementMetaData(data)
	return metadata.Convert(data) //nolint:wrapcheck
}
