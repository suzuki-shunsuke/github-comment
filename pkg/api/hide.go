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
	Config    *config.Config
	Expr      Expr
}

func (ctrl *HideController) Hide(ctx context.Context, opts *option.HideOptions) error {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	param, err := ctrl.getParamListHiddenComments(opts)
	if err != nil {
		return err
	}
	nodeIDs, err := listHiddenComments(
		ctx, ctrl.Commenter, ctrl.Expr, param, nil)
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

func (ctrl *HideController) getParamListHiddenComments(opts *option.HideOptions) (*ParamListHiddenComments, error) {
	param := &ParamListHiddenComments{}
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementHide(opts); err != nil {
			return nil, fmt.Errorf("failed to complement opts with platform built in environment variables: %w", err)
		}
	}

	cfg := ctrl.Config

	if opts.Org == "" {
		opts.Org = cfg.Base.Org
	}
	if opts.Repo == "" {
		opts.Repo = cfg.Base.Repo
	}

	if err := option.ValidateHide(opts); err != nil {
		return param, fmt.Errorf("opts is invalid: %w", err)
	}

	hideCondition := opts.Condition
	if hideCondition == "" {
		a, ok := ctrl.Config.Hide[opts.HideKey]
		if !ok {
			return param, errors.New("invalid hide-key: " + opts.HideKey)
		}
		hideCondition = a
	}

	return &ParamListHiddenComments{
		PRNumber:  opts.PRNumber,
		Org:       opts.Org,
		Repo:      opts.Repo,
		SHA1:      opts.SHA1,
		Condition: hideCondition,
		HideKey:   opts.HideKey,
	}, nil
}

func hideComments(ctx context.Context, commenter Commenter, nodeIDs []string) {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	commentHidden := false
	for _, nodeID := range nodeIDs {
		if err := commenter.HideComment(ctx, nodeID); err != nil {
			logE.WithError(err).WithFields(logrus.Fields{
				"node_id": nodeID,
			}).Error("hide an old comment")
			continue
		}
		commentHidden = true
		logE.WithFields(logrus.Fields{
			"node_id": nodeID,
		}).Info("hide an old comment")
	}
	if !commentHidden {
		logE.Info("no comment is hidden")
	}
}

type ParamListHiddenComments struct {
	Condition string
	HideKey   string
	Org       string
	Repo      string
	SHA1      string
	PRNumber  int
	Vars      map[string]interface{}
}

func listHiddenComments( //nolint:funlen
	ctx context.Context,
	commenter Commenter, exp Expr,
	param *ParamListHiddenComments,
	paramExpr map[string]interface{},
) ([]string, error) {
	logE := logrus.WithFields(logrus.Fields{
		"program": "github-comment",
	})
	if param.Condition == "" {
		logE.Debug("the condition to hide comments isn't set")
		return nil, nil
	}
	login, err := commenter.GetAuthenticatedUser(ctx)
	if err != nil {
		logE.WithError(err).Warn("get an authenticated user")
	}

	comments, err := commenter.List(ctx, &comment.PullRequest{
		Org:      param.Org,
		Repo:     param.Repo,
		PRNumber: param.PRNumber,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	logE.WithFields(logrus.Fields{
		"count":     len(comments),
		"org":       param.Org,
		"repo":      param.Repo,
		"pr_number": param.PRNumber,
	}).Debug("get comments")

	nodeIDs := []string{}
	prg, err := exp.Compile(param.Condition)
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
		hasMeta := extractMetaFromComment(comment.Body, &metadata)
		paramMap := map[string]interface{}{
			"Comment": map[string]interface{}{
				"Body": comment.Body,
				// "CreatedAt": comment.CreatedAt,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]interface{}{
				"Org":      param.Org,
				"Repo":     param.Repo,
				"PRNumber": param.PRNumber,
				"SHA1":     param.SHA1,
			},
			"HideKey": param.HideKey,
			"Vars":    param.Vars,
		}
		for k, v := range paramExpr {
			paramMap[k] = v
		}

		logE.WithFields(logrus.Fields{
			"node_id":   nodeID,
			"condition": param.Condition,
			"param":     paramMap,
		}).Debug("judge whether an existing comment is hidden")
		f, err := prg.Run(paramMap)
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

func isExcludedComment(cmt *comment.IssueComment, login string) bool {
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
