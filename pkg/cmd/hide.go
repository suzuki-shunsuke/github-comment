package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

// parseHideOptions parses the command line arguments of the subcommand "hide".
func parseHideOptions(opts *option.HideOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
	opts.LogLevel = c.String("log-level")
	opts.HideKey = c.String("hide-key")
	opts.SHA1 = c.String("sha1")
	return nil
}

func getHideCommenter(ctx context.Context, opts option.HideOptions) api.Commenter {
	if opts.DryRun {
		return comment.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}
	}
	if opts.SkipNoToken && opts.Token == "" {
		return comment.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}
	}

	return comment.New(ctx, opts.Token)
}

// hideAction is an entrypoint of the subcommand "hide".
func (runner *Runner) hideAction(c *cli.Context) error {
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		if skipComment {
			return nil
		}
	}
	opts := option.HideOptions{}
	if err := parseHideOptions(&opts, c); err != nil {
		return err
	}

	setLogLevel(opts.LogLevel)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	var pt api.Platform
	if p, f := platform.Get(); f {
		pt = &p
	}

	cfgReader := config.Reader{
		ExistFile: existFile,
	}

	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken

	ctrl := api.HideController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !terminal.IsTerminal(0)
		},
		Stderr:    runner.Stderr,
		Commenter: getHideCommenter(c.Context, opts),
		Platform:  pt,
		Config:    cfg,
		Expr:      &expr.Expr{},
	}
	return ctrl.Hide(c.Context, opts)
}
