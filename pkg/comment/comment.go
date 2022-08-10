package comment

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/github-comment-metadata/metadata"
	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
	"github.com/suzuki-shunsuke/github-comment/pkg/github"
)

type Controller struct {
	GitHub   domain.GitHub
	Expr     domain.Expr
	Getenv   func(string) string
	Platform domain.Platform
}

func (ctrl *Controller) Post(ctx context.Context, cmt *github.Comment, hiddenParam map[string]interface{}) error {
	if err := ctrl.GitHub.CreateComment(ctx, cmt); err != nil {
		return fmt.Errorf("send a comment: %w", err)
	}
	return nil
}

func ExtractMetaFromComment(body string, data *map[string]interface{}) bool {
	f, _ := metadata.Extract(body, data)
	return f
}

func (ctrl *Controller) complementMetaData(data map[string]interface{}) {
	if data == nil {
		return
	}
	if ctrl.Platform == nil {
		return
	}
	_ = metadata.SetCIEnv(ctrl.Platform.CI(), ctrl.Getenv, data)
}

func (ctrl *Controller) GetEmbeddedComment(data map[string]interface{}) (string, error) {
	ctrl.complementMetaData(data)
	return metadata.Convert(data) //nolint:wrapcheck
}
