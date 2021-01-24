package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
)

// Commenter is API to post a comment to GitHub
type Commenter interface {
	Create(ctx context.Context, cmt comment.Comment) error
	List(ctx context.Context, pr comment.PullRequest) ([]comment.IssueComment, error)
	HideComment(ctx context.Context, nodeID string) error
	GetAuthenticatedUser(ctx context.Context) (string, error)
}

type CommentController struct {
	Commenter Commenter
	Expr      Expr
	Getenv    func(string) string
	Platform  Platform
}

func (ctrl *CommentController) Post(ctx context.Context, cmt comment.Comment, hiddenParam map[string]interface{}) error {
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}

const (
	embeddedCommentPrefix    = "<!-- github-comment: "
	embeddedCommentSuffix    = " -->"
	lenEmbeddedCommentPrefix = len(embeddedCommentPrefix)
	lenEmbeddedCommentSuffix = len(embeddedCommentSuffix)
)

func extractMetaFromComment(body string, metadata map[string]interface{}) bool {
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, embeddedCommentPrefix) {
			continue
		}
		if !strings.HasSuffix(line, embeddedCommentSuffix) {
			continue
		}
		if err := json.Unmarshal([]byte(line[lenEmbeddedCommentPrefix:len(line)-lenEmbeddedCommentSuffix]), &metadata); err != nil {
			continue
		}
		return true
	}
	return false
}

func (ctrl *CommentController) complementMetaData(metadata map[string]interface{}) {
	if metadata == nil {
		return
	}
	if ctrl.Platform == nil {
		return
	}
	switch ctrl.Platform.CI() {
	case "circleci":
		metadata["JobName"] = ctrl.Getenv("CIRCLE_JOB")
		metadata["JobID"] = ctrl.Getenv("CIRCLE_WORKFLOW_JOB_ID")
	case "drone":
		metadata["WorkflowName"] = ctrl.Getenv("DRONE_STATE_NAME")
		metadata["JobName"] = ctrl.Getenv("DRONE_STEP_NAME")
	case "github-actions":
		metadata["WorkflowName"] = ctrl.Getenv("GITHUB_WORKFLOW")
		metadata["JobName"] = ctrl.Getenv("GITHUB_JOB")
	case "codebuild":
		metadata["JobID"] = ctrl.Getenv("CODEBUILD_BUILD_ID")
	}
}
