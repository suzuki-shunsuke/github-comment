package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
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
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	skipHideComment := false
	nodeIDs, err := ctrl.listHiddenComments(ctx, cmt, hiddenParam)
	if err != nil {
		skipHideComment = true
		logE.WithError(err).Error("list hidden comments")
	}
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	if !skipHideComment {
		logE.WithFields(logrus.Fields{
			"count":    len(nodeIDs),
			"node_ids": nodeIDs,
		}).Debug("comments which would be hidden")
		ctrl.hideComments(ctx, nodeIDs)
	}
	return nil
}

func (ctrl *CommentController) listHiddenComments(ctx context.Context, cmt comment.Comment, hiddenParam map[string]interface{}) ([]string, error) {
	return listHiddenComments(
		ctx, ctrl.Commenter, ctrl.Expr, ctrl.Getenv, cmt, hiddenParam)
}

func (ctrl *CommentController) hideComments(ctx context.Context, nodeIDs []string) {
	hideComments(ctx, ctrl.Commenter, nodeIDs)
}

func hideComments(ctx context.Context, commenter Commenter, nodeIDs []string) {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	for _, nodeID := range nodeIDs {
		if err := commenter.HideComment(ctx, nodeID); err != nil {
			logE.WithError(err).WithFields(logrus.Fields{
				"node_id": nodeID,
			}).Error("hide an old comment")
			continue
		}
		logE.WithFields(logrus.Fields{
			"node_id": nodeID,
		}).Debug("hide an old comment")
	}
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

func listHiddenComments( //nolint:funlen
	ctx context.Context,
	commenter Commenter, exp Expr,
	getEnv func(string) string,
	cmt comment.Comment,
	paramExpr map[string]interface{},
) ([]string, error) {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	if cmt.HideOldComment == "" {
		logE.Debug("hide_old_comment isn't set")
		return nil, nil
	}
	login, err := commenter.GetAuthenticatedUser(ctx)
	if err != nil {
		logE.WithError(err).Warn("get an authenticated user")
	}

	comments, err := commenter.List(ctx, comment.PullRequest{
		Org:      cmt.Org,
		Repo:     cmt.Repo,
		PRNumber: cmt.PRNumber,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	logE.WithFields(logrus.Fields{
		"count":     len(comments),
		"org":       cmt.Org,
		"repo":      cmt.Repo,
		"pr_number": cmt.PRNumber,
	}).Debug("get comments")

	nodeIDs := []string{}
	prg, err := exp.Compile(cmt.HideOldComment)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, comment := range comments {
		nodeID := comment.ID
		// TODO remove these filters
		if isExcludedComment(comment, login) {
			logE.WithFields(logrus.Fields{
				"node_id": nodeID,
				"login":   login,
			}).Debug("exclude a comment")
			continue
		}

		metadata := map[string]interface{}{}
		hasMeta := extractMetaFromComment(comment.Body, metadata)
		param := map[string]interface{}{
			"Comment": map[string]interface{}{
				"Body": comment.Body,
				// "CreatedAt": comment.CreatedAt,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]interface{}{
				"Org":      cmt.Org,
				"Repo":     cmt.Repo,
				"PRNumber": cmt.PRNumber,
				"SHA1":     cmt.SHA1,
			},
			"Vars": cmt.Vars,
			"PostedComment": map[string]interface{}{
				"Body":        cmt.Body,
				"TemplateKey": cmt.TemplateKey,
			},
			"Env": getEnv,
		}
		for k, v := range paramExpr {
			param[k] = v
		}

		logE.WithFields(logrus.Fields{
			"node_id":          nodeID,
			"hide_old_comment": cmt.HideOldComment,
			"param":            param,
		}).Debug("judge whether an existing comment is hidden")
		f, err := prg.Run(param)
		if err != nil {
			logE.WithError(err).WithFields(logrus.Fields{
				"node_id": nodeID,
			}).Error("judge whether an existing comment is hidden")
			continue
		}
		if !f {
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, nil
}

func isExcludedComment(cmt comment.IssueComment, login string) bool {
	if !cmt.ViewerCanMinimize {
		return true
	}
	if cmt.IsMinimized {
		return true
	}
	// GitHub Actions's GITHUB_TOKEN secret doesn't have a permission to get an authenticated user.
	// So if `login` is empty, we give up filtering comments by login.
	if login != "" && cmt.Author.Login != login {
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
		metadata["job_name"] = ctrl.Getenv("CIRCLE_JOB")
		metadata["job_id"] = ctrl.Getenv("CIRCLE_WORKFLOW_JOB_ID")
	case "drone":
		metadata["workflow_name"] = ctrl.Getenv("DRONE_STATE_NAME")
		metadata["job_name"] = ctrl.Getenv("DRONE_STEP_NAME")
	case "github-actions":
		metadata["workflow_name"] = ctrl.Getenv("GITHUB_WORKFLOW")
		metadata["job_name"] = ctrl.Getenv("GITHUB_JOB")
	case "codebuild":
		metadata["job_id"] = ctrl.Getenv("CODEBUILD_BUILD_ID")
	}
}
