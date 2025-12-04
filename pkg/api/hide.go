package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"

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
	Logger   *slog.Logger
}

func (c *HideController) Hide(ctx context.Context, opts *option.HideOptions) error {
	param, err := c.getParamListHiddenComments(ctx, opts)
	if err != nil {
		return err
	}
	nodeIDs, err := c.listHiddenComments(ctx, param, nil)
	if err != nil {
		return err
	}
	c.Logger.Debug("comments which would be hidden",
		"count", len(nodeIDs),
		"node_ids", nodeIDs,
	)
	c.hideComments(ctx, nodeIDs)
	return nil
}

func (c *HideController) getParamListHiddenComments(ctx context.Context, opts *option.HideOptions) (*ParamListHiddenComments, error) { //nolint:cyclop,funlen
	param := &ParamListHiddenComments{}

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
			slogerr.WithError(c.Logger, err).Warn("list associated prs",
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

	return &ParamListHiddenComments{
		PRNumber:  opts.PRNumber,
		Org:       opts.Org,
		Repo:      opts.Repo,
		SHA1:      opts.SHA1,
		Condition: hideCondition,
		HideKey:   opts.HideKey,
		Vars:      cfg.Vars,
	}, nil
}

func (c *HideController) hideComments(ctx context.Context, nodeIDs []string) {
	commentHidden := false
	for _, nodeID := range nodeIDs {
		if err := c.GitHub.HideComment(ctx, nodeID); err != nil {
			slogerr.WithError(c.Logger, err).Error("hide an old comment",
				"node_id", nodeID,
			)
			continue
		}
		commentHidden = true
		c.Logger.Info("hide an old comment",
			"node_id", nodeID,
		)
	}
	if !commentHidden {
		c.Logger.Info("no comment is hidden")
	}
}

type ParamListHiddenComments struct {
	Condition string
	HideKey   string
	Org       string
	Repo      string
	SHA1      string
	PRNumber  int
	Vars      map[string]any
}

func (c *HideController) listHiddenComments( //nolint:funlen
	ctx context.Context,
	param *ParamListHiddenComments,
	paramExpr map[string]any,
) ([]string, error) {
	if param.Condition == "" {
		c.Logger.Debug("the condition to hide comments isn't set")
		return nil, nil
	}
	login, err := c.GitHub.GetAuthenticatedUser(ctx)
	if err != nil {
		slogerr.WithError(c.Logger, err).Warn("get an authenticated user")
	}

	comments, err := c.GitHub.ListComments(ctx, &github.PullRequest{
		Org:      param.Org,
		Repo:     param.Repo,
		PRNumber: param.PRNumber,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	c.Logger.Debug("get comments",
		"count", len(comments),
		"org", param.Org,
		"repo", param.Repo,
		"pr_number", param.PRNumber,
	)

	nodeIDs := []string{}
	prg, err := c.Expr.Compile(param.Condition)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, comment := range comments {
		nodeID := comment.ID
		// TODO remove these filters
		if isExcludedComment(comment, login) {
			c.Logger.Debug("exclude a comment",
				"node_id", nodeID,
				"login", login,
			)
			continue
		}

		metadata := map[string]any{}
		hasMeta := extractMetaFromComment(comment.Body, &metadata)
		paramMap := map[string]any{
			"Comment": map[string]any{
				"Body": comment.Body,
				// "CreatedAt": comment.CreatedAt,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]any{
				"Org":      param.Org,
				"Repo":     param.Repo,
				"PRNumber": param.PRNumber,
				"SHA1":     param.SHA1,
			},
			"HideKey": param.HideKey,
			"Vars":    param.Vars,
		}
		maps.Copy(paramMap, paramExpr)

		c.Logger.Debug("judge whether an existing comment is hidden",
			"node_id", nodeID,
			"condition", param.Condition,
			"param", paramMap,
		)
		f, err := prg.Run(paramMap)
		if err != nil {
			slogerr.WithError(c.Logger, err).Error("judge whether an existing comment is hidden",
				"node_id", nodeID,
			)
			continue
		}
		if !f {
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, nil
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
