package api

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
)

type Env struct {
	Stdout         string
	Stderr         string
	CombinedOutput string
	Command        string
	ExitCode       int
	Env            func(string) string
}

type Executor interface {
	Run(ctx context.Context, params execute.Params) (execute.Result, error)
}

type Expr interface {
	Match(expression string, params interface{}) (bool, error)
}

type ExecController struct {
	Wd        string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Getenv    func(string) string
	Reader    Reader
	Commenter Commenter
	Renderer  Renderer
	Executor  Executor
	Expr      Expr
	Env       []string
}

func (ctrl ExecController) Exec(ctx context.Context, opts option.ExecOptions) error {
	if err := option.ComplementExec(&opts, ctrl.Getenv); err != nil {
		return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
	}
	if err := option.ValidateExec(opts); err != nil {
		return err
	}

	cfg := config.Config{}
	if opts.ConfigPath == "" {
		p, b, err := ctrl.Reader.Find(ctrl.Wd)
		if err != nil {
			return err
		}
		if !b {
			return errors.New("configuration file isn't found")
		}
		opts.ConfigPath = p
	}

	if err := ctrl.Reader.Read(opts.ConfigPath, &cfg); err != nil {
		return err
	}

	execConfigs, ok := cfg.Exec[opts.TemplateKey]
	if !ok {
		return errors.New("template isn't found: " + opts.TemplateKey)
	}

	result, err := ctrl.Executor.Run(ctx, execute.Params{
		Cmd:   opts.Args[0],
		Args:  opts.Args[1:],
		Stdin: ctrl.Stdin,
	})

	ctrl.execPost(ctx, opts, execConfigs, Env{
		ExitCode:       result.ExitCode,
		Command:        result.Cmd,
		Stdout:         result.Stdout,
		Stderr:         result.Stderr,
		CombinedOutput: result.CombinedOutput,
	})
	if err != nil {
		return ecerror.Wrap(err, result.ExitCode)
	}
	return nil
}

func (ctrl ExecController) execPostConfig(
	ctx context.Context, opts option.ExecOptions, execConfig config.ExecConfig, env Env,
) (bool, error) {
	f, err := ctrl.Expr.Match(execConfig.When, env)
	if err != nil {
		return false, err
	}
	if !f {
		return false, nil
	}
	if execConfig.DontComment {
		return true, nil
	}
	tpl, err := ctrl.Renderer.Render(execConfig.Template, env)
	if err != nil {
		return true, err
	}
	cmt := comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     tpl,
		SHA1:     opts.SHA1,
	}
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return true, fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return true, nil
}

func (ctrl ExecController) execPost(
	ctx context.Context, opts option.ExecOptions, execConfigs []config.ExecConfig, env Env,
) error {
	for _, execConfig := range execConfigs {
		f, err := ctrl.execPostConfig(ctx, opts, execConfig, env)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return nil
}
