package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/controller"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/urfave/cli/v3"
)

func parseExecOptions(opts *option.ExecOptions, c *cli.Command) error {
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

	outputs := c.StringSlice("out")
	outs := make([]*option.Output, len(outputs))
	for i, o := range outputs {
		if o == "github" {
			outs[i] = &option.Output{GitHub: true}
			continue
		}
		if f, ok := strings.CutPrefix(o, "file:"); ok {
			outs[i] = &option.Output{File: f}
			continue
		}
		return errors.New("invalid the value of -out. -out must be either github or file:<file path>")
	}
	if len(outputs) == 0 {
		outs = []*option.Output{{GitHub: true}}
	}
	opts.Outputs = outs

	vars, err := parseVars(c)
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

func (r *Runner) execAction(ctx context.Context, c *cli.Command, logger *slogutil.Logger) error {
	opts := &option.ExecOptions{}
	if err := parseExecOptions(opts, c); err != nil {
		return err
	}
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		opts.SkipComment = skipComment
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
	opts.Silent = opts.Silent || cfg.Silent

	var pt controller.Platform = platform.Get()

	gh, err := getGitHub(ctx, logger.Logger, &opts.Options, cfg)
	if err != nil {
		return fmt.Errorf("initialize commenter: %w", err)
	}

	ctrl := controller.ExecController{
		Wd:     wd,
		Getenv: os.Getenv,
		Stdin:  r.Stdin,
		Stdout: r.Stdout,
		Stderr: r.Stderr,
		GitHub: gh,
		Renderer: &template.Renderer{
			Getenv: os.Getenv,
		},
		Executor: &execute.Executor{
			Stdout: r.Stdout,
			Stderr: r.Stderr,
			Env:    os.Environ(),
		},
		Expr:     &expr.Expr{},
		Platform: pt,
		Config:   cfg,
		Fs:       afero.NewOsFs(),
	}
	return ctrl.Exec(ctx, logger.Logger, opts) //nolint:wrapcheck
}
