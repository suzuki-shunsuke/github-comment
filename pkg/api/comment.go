package api

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/github-comment-metadata/metadata"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
)

// Commenter is API to post a comment to GitHub
type Commenter interface {
	Create(ctx context.Context, cmt *comment.Comment) error
	List(ctx context.Context, pr *comment.PullRequest) ([]*comment.IssueComment, error)
	HideComment(ctx context.Context, nodeID string) error
	GetAuthenticatedUser(ctx context.Context) (string, error)
}

type CommentController struct {
	Commenter Commenter
	Expr      Expr
	Getenv    func(string) string
	Platform  Platform
}

func (ctrl *CommentController) Post(ctx context.Context, cmt *comment.Comment, hiddenParam map[string]interface{}) error {
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
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
