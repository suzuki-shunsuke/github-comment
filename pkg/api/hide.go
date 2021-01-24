package api

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type HideController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If thre is the standard input, it is treated as the comment template
	HasStdin  func() bool
	Stderr    io.Writer
	Commenter Commenter
	Platform  Platform
	Config    config.Config
	Expr      Expr
}

func (ctrl *HideController) Hide(ctx context.Context, opts option.HideOptions) error {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	cmt, err := ctrl.getCommentParams(opts)
	if err != nil {
		return err
	}
	nodeIDs, err := listHiddenComments(
		ctx, ctrl.Commenter, ctrl.Expr, ctrl.Getenv, cmt, nil)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"count":    len(nodeIDs),
		"node_ids": nodeIDs,
	}).Debug("comments which would be hidden")
	hideComments(ctx, ctrl.Commenter, nodeIDs)
	return nil
}

func (ctrl *HideController) getCommentParams(opts option.HideOptions) (comment.Comment, error) {
	cmt := comment.Comment{}
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementHide(&opts); err != nil {
			return cmt, fmt.Errorf("failed to complement opts with platform built in environment variables: %w", err)
		}
	}

	cfg := ctrl.Config

	if opts.Org == "" {
		opts.Org = cfg.Base.Org
	}
	if opts.Repo == "" {
		opts.Repo = cfg.Base.Repo
	}

	hideCondition, ok := ctrl.Config.Hide[opts.HideKey]
	if !ok {
		return cmt, errors.New("invalid hide-key: " + opts.HideKey)
	}

	return comment.Comment{
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		SHA1:           opts.SHA1,
		HideOldComment: hideCondition,
	}, nil
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
