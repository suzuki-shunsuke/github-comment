package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/controller"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"golang.org/x/term"
)

func getGitHub(ctx context.Context, logger *slog.Logger, opts *option.Options, cfg *config.Config) (controller.GitHub, error) {
	if opts.DryRun {
		return &github.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}, nil
	}
	if opts.SkipNoToken && opts.Token == "" {
		return &github.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}, nil
	}

	// https://github.com/suzuki-shunsuke/github-comment/issues/1489
	if cfg.GHEBaseURL == "" {
		cfg.GHEBaseURL = os.Getenv("GITHUB_API_URL")
	}
	if cfg.GHEGraphQLEndpoint == "" {
		cfg.GHEGraphQLEndpoint = os.Getenv("GITHUB_GRAPHQL_URL")
	}

	return github.New(ctx, &github.ParamNew{ //nolint:wrapcheck
		Token:              opts.Token,
		GHEBaseURL:         cfg.GHEBaseURL,
		GHEGraphQLEndpoint: cfg.GHEGraphQLEndpoint,
		Logger:             logger,
	})
}

// postAction is an entrypoint of the subcommand "post".
func (r *Runner) postAction(ctx context.Context, logger *slogutil.Logger, args *PostArgs) error { //nolint:funlen
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

	opts := &option.PostOptions{
		Options: option.Options{
			PRNumber:    args.PRNumber,
			Org:         args.Org,
			Repo:        args.Repo,
			Token:       args.Token,
			SHA1:        args.SHA1,
			Template:    args.Template,
			TemplateKey: args.TemplateKey,
			ConfigPath:  args.ConfigPath,
			LogLevel:    args.LogLevel,
			Vars:        vars,
			DryRun:      args.DryRun,
			SkipNoToken: args.SkipNoToken,
			Silent:      args.Silent,
		},
		StdinTemplate:   args.StdinTemplate,
		UpdateCondition: args.UpdateCondition,
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

	ctrl := controller.PostController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !term.IsTerminal(0)
		},
		Stdin:  r.Stdin,
		Stderr: r.Stderr,
		GitHub: gh,
		Renderer: &template.Renderer{
			Getenv: os.Getenv,
		},
		Platform: pt,
		Config:   cfg,
		Expr:     &expr.Expr{},
	}
	return ctrl.Post(ctx, logger.Logger, opts) //nolint:wrapcheck
}
