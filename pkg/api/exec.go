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

type ExecCommentParams struct {
	Stdout         string
	Stderr         string
	CombinedOutput string
	Command        string
	ExitCode       int
	Env            func(string) string
	// PRNumber is the pull request number where the comment is posted
	PRNumber int
	// Org is the GitHub Organization or User name
	Org string
	// Repo is the GitHub Repository name
	Repo string
	// SHA1 is the commit SHA1
	SHA1        string
	TemplateKey string
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
}

func (ctrl ExecController) Exec(ctx context.Context, opts option.ExecOptions) error {
	if err := option.ComplementExec(&opts, ctrl.Getenv); err != nil {
		return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
	}
	if err := option.ValidateExec(opts); err != nil {
		return err
	}

	cfg, err := ctrl.Reader.FindAndRead(opts.ConfigPath, ctrl.Wd)
	if err != nil {
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

	ctrl.post(ctx, opts, execConfigs, ExecCommentParams{
		ExitCode:       result.ExitCode,
		Command:        result.Cmd,
		Stdout:         result.Stdout,
		Stderr:         result.Stderr,
		CombinedOutput: result.CombinedOutput,
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		SHA1:           opts.SHA1,
		TemplateKey:    opts.TemplateKey,
	})
	if err != nil {
		return ecerror.Wrap(err, result.ExitCode)
	}
	return nil
}

// getExecConfig returns matched ExecConfig.
// If no ExecConfig matches, the second returned value is false.
func (ctrl ExecController) getExecConfig(
	opts option.ExecOptions, execConfigs []config.ExecConfig, cmtParams ExecCommentParams,
) (config.ExecConfig, bool, error) {
	for _, execConfig := range execConfigs {
		f, err := ctrl.Expr.Match(execConfig.When, cmtParams)
		if err != nil {
			return execConfig, false, err
		}
		if !f {
			continue
		}
		return execConfig, true, nil
	}
	return config.ExecConfig{}, false, nil
}

// getComment returns Comment.
// If the second returned value is false, no comment is posted.
func (ctrl ExecController) getComment(
	opts option.ExecOptions, execConfigs []config.ExecConfig, cmtParams ExecCommentParams,
) (comment.Comment, bool, error) {
	cmt := comment.Comment{}
	execConfig, f, err := ctrl.getExecConfig(opts, execConfigs, cmtParams)
	if err != nil {
		return cmt, false, err
	}
	if !f {
		return cmt, false, nil
	}
	if execConfig.DontComment {
		return cmt, false, nil
	}

	tpl, err := ctrl.Renderer.Render(execConfig.Template, cmtParams)
	if err != nil {
		return cmt, false, err
	}
	return comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     tpl,
		SHA1:     opts.SHA1,
	}, true, nil
}

func (ctrl ExecController) post(
	ctx context.Context, opts option.ExecOptions, execConfigs []config.ExecConfig, cmtParams ExecCommentParams,
) error {
	cmt, f, err := ctrl.getComment(opts, execConfigs, cmtParams)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}
