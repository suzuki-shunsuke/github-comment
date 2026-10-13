package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type HideController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If there is the standard input, it is treated as the comment template
	HasStdin func() bool
	Stderr   io.Writer
	GitHub   GitHub
	Platform Platform
	Config   *config.Config
	Expr     Expr
}

func (c *HideController) Hide(ctx context.Context, logger *slog.Logger, opts *option.HideOptions) error {
	param, err := c.getParamListHiddenComments(ctx, logger, opts)
	if err != nil {
		return err
	}
	nodeIDs, err := listCommentsByCondition(ctx, logger, c.GitHub, c.Expr, param, isExcludedComment)
	if err != nil {
		return err
	}
	logger.Debug("comments which would be hidden",
		"count", len(nodeIDs),
		"node_ids", nodeIDs,
	)
	c.hideComments(ctx, logger, nodeIDs)
	return nil
}

func (c *HideController) getParamListHiddenComments(ctx context.Context, logger *slog.Logger, opts *option.HideOptions) (*ParamListComments, error) { //nolint:cyclop,funlen,dupl
	param := &ParamListComments{}

	cfg := c.Config

	if cfg.Base != nil {
		if opts.Org == "" {
			opts.Org = cfg.Base.Org
		}
		if opts.Repo == "" {
			opts.Repo = cfg.Base.Repo
		}
	}

	if c.Platform != nil {
		if err := c.Platform.ComplementHide(opts); err != nil {
			return nil, fmt.Errorf("failed to complement opts with platform built in environment variables: %w", err)
		}
	}

	if opts.PRNumber == 0 && opts.SHA1 != "" {
		prNum, err := c.GitHub.PRNumberWithSHA(ctx, opts.Org, opts.Repo, opts.SHA1)
		if err != nil {
			slogerr.WithError(logger, err).Warn("list associated prs",
				"org", opts.Org,
				"repo", opts.Repo,
				"sha", opts.SHA1,
			)
		}
		if prNum > 0 {
			opts.PRNumber = prNum
		}
	}

	if err := option.ValidateHide(opts); err != nil {
		return param, fmt.Errorf("opts is invalid: %w", err)
	}

	hideCondition := opts.Condition
	if hideCondition == "" {
		a, ok := c.Config.Hide[opts.HideKey]
		if !ok {
			return param, errors.New("invalid hide-key: " + opts.HideKey)
		}
		hideCondition = a
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]any, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	return &ParamListComments{
		PRNumber:  opts.PRNumber,
		Org:       opts.Org,
		Repo:      opts.Repo,
		SHA1:      opts.SHA1,
		Condition: hideCondition,
		Vars:      cfg.Vars,
		ExprParams: map[string]any{
			"HideKey": opts.HideKey,
		},
	}, nil
}

func (c *HideController) hideComments(ctx context.Context, logger *slog.Logger, nodeIDs []string) {
	commentHidden := false
	for _, nodeID := range nodeIDs {
		if err := c.GitHub.HideComment(ctx, nodeID); err != nil {
			slogerr.WithError(logger, err).Error("hide an old comment",
				"node_id", nodeID,
			)
			continue
		}
		commentHidden = true
		logger.Info("hide an old comment",
			"node_id", nodeID,
		)
	}
	if !commentHidden {
		logger.Info("no comment is hidden")
	}
}

func isExcludedComment(cmt *github.IssueComment, login string) bool {
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
