package api

import (
	"context"
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

	return comment.Comment{
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		SHA1:           opts.SHA1,
		HideOldComment: "Comment.HasMeta && Comment.Meta.SHA1 != Commit.SHA1",
	}, nil
}
