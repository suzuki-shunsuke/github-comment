package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/controller"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
)

func existFile(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (r *Runner) execAction(ctx context.Context, logger *slogutil.Logger, args *ExecArgs) error { //nolint:cyclop,funlen
	vars, err := parseVars(args.Vars, args.VarFiles)
	if err != nil {
		return err
	}

	outputs := args.Outputs
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

	opts := &option.ExecOptions{
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
		Args:    args.Args,
		Outputs: outs,
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
