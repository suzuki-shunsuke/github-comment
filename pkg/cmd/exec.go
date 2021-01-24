package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
	"github.com/urfave/cli/v2"
)

func parseExecOptions(opts *option.ExecOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.Args = c.Args().Slice()
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
	opts.LogLevel = c.String("log-level")
	vars, err := parseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return err
	}
	opts.Vars = vars
	return nil
}

func existFile(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func getExecCommenter(ctx context.Context, opts option.ExecOptions) api.Commenter {
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

func (runner *Runner) execAction(c *cli.Context) error {
	opts := option.ExecOptions{}
	if err := parseExecOptions(&opts, c); err != nil {
		return err
	}
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		opts.SkipComment = skipComment
	}
	setLogLevel(opts.LogLevel)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	var pt api.Platform
	if p, f := platform.Get(); f {
		pt = p
	}

	cfgReader := config.Reader{
		ExistFile: existFile,
	}
	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken
	opts.Silent = opts.Silent || cfg.Silent

	ctrl := api.ExecController{
		Wd:        wd,
		Getenv:    os.Getenv,
		Stdin:     runner.Stdin,
		Stdout:    runner.Stdout,
		Stderr:    runner.Stderr,
		Commenter: getExecCommenter(c.Context, opts),
		Renderer: &template.Renderer{
			Getenv: os.Getenv,
		},
		Executor: execute.Executor{
			Stdout: runner.Stdout,
			Stderr: runner.Stderr,
			Env:    os.Environ(),
		},
		Expr:     &expr.Expr{},
		Platform: pt,
		Config:   cfg,
	}
	return ctrl.Exec(c.Context, opts)
}
