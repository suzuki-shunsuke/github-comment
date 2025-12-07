package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/controller"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/platform"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// parseHideOptions parses the command line arguments of the subcommand "hide".
func parseHideOptions(opts *option.HideOptions, c *cli.Command) error {
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
	opts.Condition = c.String("condition")
	opts.SHA1 = c.String("sha1")

	vars, err := parseVars(c)
	if err != nil {
		return err
	}
	opts.Vars = vars

	return nil
}

// hideAction is an entrypoint of the subcommand "hide".
func (r *Runner) hideAction(ctx context.Context, c *cli.Command, logger *slogutil.Logger) error {
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		if skipComment {
			return nil
		}
	}
	opts := &option.HideOptions{}
	if err := parseHideOptions(opts, c); err != nil {
		return err
	}

	if err := logger.SetLevel(opts.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	cfgReader := config.Reader{
		ExistFile: existFile,
	}

	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken

	var pt controller.Platform = platform.Get()

	gh, err := getGitHub(ctx, logger.Logger, &opts.Options, cfg)
	if err != nil {
		return fmt.Errorf("initialize commenter: %w", err)
	}

	ctrl := controller.HideController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !term.IsTerminal(0)
		},
		Stderr:   r.Stderr,
		GitHub:   gh,
		Platform: pt,
		Config:   cfg,
		Expr:     &expr.Expr{},
	}
	return ctrl.Hide(ctx, logger.Logger, opts) //nolint:wrapcheck
}
