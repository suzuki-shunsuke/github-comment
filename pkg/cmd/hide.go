package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/controller"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/platform"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"golang.org/x/term"
)

// hideAction is an entrypoint of the subcommand "hide".
func (r *Runner) hideAction(ctx context.Context, logger *slogutil.Logger, args *HideArgs) error { //nolint:funlen
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		if skipComment {
			return nil
		}
	}

	vars, err := parseVars(args.Vars, args.VarFiles)
	if err != nil {
		return err
	}

	opts := &option.HideOptions{
		Options: option.Options{
			PRNumber:    args.PRNumber,
			Org:         args.Org,
			Repo:        args.Repo,
			Token:       args.Token,
			SHA1:        args.SHA1,
			ConfigPath:  args.ConfigPath,
			LogLevel:    args.LogLevel,
			Vars:        vars,
			DryRun:      args.DryRun,
			SkipNoToken: args.SkipNoToken,
			Silent:      args.Silent,
		},
		HideKey:   args.HideKey,
		Condition: args.Condition,
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
