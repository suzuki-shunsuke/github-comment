package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

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
	JoinCommand    string
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
	Template    string
	Vars        map[string]interface{}
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
	Platform  Platform
}

func (ctrl ExecController) Exec(ctx context.Context, opts option.ExecOptions) error { //nolint:funlen
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementExec(&opts); err != nil {
			return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}

	cfg, err := ctrl.Reader.FindAndRead(opts.ConfigPath, ctrl.Wd)
	if err != nil {
		return err
	}
	if opts.Org == "" {
		opts.Org = cfg.Base.Org
	}
	if opts.Repo == "" {
		opts.Repo = cfg.Base.Repo
	}

	if err := option.ValidateExec(opts); err != nil {
		return err
	}

	var execConfigs []config.ExecConfig
	if opts.Template == "" && opts.TemplateKey != "" {
		a, ok := cfg.Exec[opts.TemplateKey]
		if !ok {
			return errors.New("template isn't found: " + opts.TemplateKey)
		}
		execConfigs = a
	}

	result, err := ctrl.Executor.Run(ctx, execute.Params{
		Cmd:   opts.Args[0],
		Args:  opts.Args[1:],
		Stdin: ctrl.Stdin,
	})

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]interface{}, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	if err := ctrl.post(ctx, execConfigs, ExecCommentParams{
		ExitCode:       result.ExitCode,
		Command:        result.Cmd,
		JoinCommand:    strings.Join(opts.Args, " "),
		Stdout:         result.Stdout,
		Stderr:         result.Stderr,
		CombinedOutput: result.CombinedOutput,
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		SHA1:           opts.SHA1,
		TemplateKey:    opts.TemplateKey,
		Template:       opts.Template,
		Vars:           cfg.Vars,
	}, cfg.Templates); err != nil {
		fmt.Fprintf(ctrl.Stderr, "github-comment error: %+v\n", err)
	}
	if err != nil {
		return ecerror.Wrap(err, result.ExitCode)
	}
	return nil
}

// getExecConfig returns matched ExecConfig.
// If no ExecConfig matches, the second returned value is false.
func (ctrl ExecController) getExecConfig(
	execConfigs []config.ExecConfig, cmtParams ExecCommentParams,
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
	execConfigs []config.ExecConfig, cmtParams ExecCommentParams, templates map[string]string,
) (comment.Comment, bool, error) {
	cmt := comment.Comment{}
	tpl := cmtParams.Template
	tplForTooLong := ""
	if tpl == "" {
		execConfig, f, err := ctrl.getExecConfig(execConfigs, cmtParams)
		if err != nil {
			return cmt, false, err
		}
		if !f {
			return cmt, false, nil
		}
		if execConfig.DontComment {
			return cmt, false, nil
		}
		tpl = execConfig.Template
		tplForTooLong = execConfig.TemplateForTooLong
	}

	body, err := ctrl.Renderer.Render(tpl, templates, cmtParams)
	if err != nil {
		return cmt, false, err
	}
	bodyForTooLong, err := ctrl.Renderer.Render(tplForTooLong, templates, cmtParams)
	if err != nil {
		return cmt, false, err
	}
	return comment.Comment{
		PRNumber:       cmtParams.PRNumber,
		Org:            cmtParams.Org,
		Repo:           cmtParams.Repo,
		Body:           body,
		BodyForTooLong: bodyForTooLong,
		SHA1:           cmtParams.SHA1,
	}, true, nil
}

func (ctrl ExecController) post(
	ctx context.Context, execConfigs []config.ExecConfig, cmtParams ExecCommentParams,
	templates map[string]string,
) error {
	cmt, f, err := ctrl.getComment(execConfigs, cmtParams, templates)
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
