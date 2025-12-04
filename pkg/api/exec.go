package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type ExecController struct {
	Wd       string
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Getenv   func(string) string
	Reader   Reader
	GitHub   GitHub
	Renderer Renderer
	Executor Executor
	Expr     Expr
	Platform Platform
	Config   *config.Config
	Fs       afero.Fs
}

func (c *ExecController) Exec(ctx context.Context, logger *slog.Logger, opts *option.ExecOptions) error { //nolint:funlen,cyclop
	cfg := c.Config

	if cfg.Base != nil {
		if opts.Org == "" {
			opts.Org = cfg.Base.Org
		}
		if opts.Repo == "" {
			opts.Repo = cfg.Base.Repo
		}
	}

	if c.Platform != nil {
		if err := c.Platform.ComplementExec(opts); err != nil {
			return fmt.Errorf("complement opts with CI built in environment variables: %w", err)
		}
	}

	if opts.PRNumber == 0 && opts.SHA1 != "" {
		prNum, err := c.GitHub.PRNumberWithSHA(ctx, opts.Org, opts.Repo, opts.SHA1)
		if err != nil {
			slogerr.WithError(logger, err).Warn("list associated prs",
				"org", opts.Org,
				"repo", opts.Repo,
				"sha", opts.SHA1,
			)
		}
		if prNum > 0 {
			opts.PRNumber = prNum
		}
	}

	if len(opts.Args) == 0 {
		return errors.New("command is required")
	}

	result, execErr := c.Executor.Run(ctx, &execute.Params{
		Cmd:   opts.Args[0],
		Args:  opts.Args[1:],
		Stdin: c.Stdin,
	})

	if opts.SkipComment {
		if execErr != nil {
			return ecerror.Wrap(execErr, result.ExitCode)
		}
		return nil
	}

	execConfigs, err := c.getExecConfigs(cfg, opts)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	if err := option.ValidateExec(opts); err != nil {
		return fmt.Errorf("validate command options: %w", err)
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]any, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	ci := ""
	if c.Platform != nil {
		ci = c.Platform.CI()
	}
	joinCommand := strings.Join(opts.Args, " ")
	templates := template.GetTemplates(&template.ParamGetTemplates{
		Templates:      cfg.Templates,
		CI:             ci,
		JoinCommand:    joinCommand,
		CombinedOutput: result.CombinedOutput,
	})
	if err := c.post(ctx, logger, execConfigs, &ExecCommentParams{
		ExitCode:       result.ExitCode,
		Command:        result.Cmd,
		JoinCommand:    joinCommand,
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
		Outputs:        opts.Outputs,
	}, templates); err != nil {
		if !opts.Silent {
			fmt.Fprintf(c.Stderr, "github-comment error: %+v\n", err)
		}
	}
	if execErr != nil {
		return ecerror.Wrap(execErr, result.ExitCode)
	}
	return nil
}

type ExecCommentParams struct {
	Stdout         string
	Stderr         string
	CombinedOutput string
	Command        string
	JoinCommand    string
	ExitCode       int
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
	Vars        map[string]any
	Outputs     []*option.Output
}

type Executor interface {
	Run(ctx context.Context, params *execute.Params) (*execute.Result, error)
}

type Expr interface {
	Match(expression string, params any) (bool, error)
	Compile(expression string) (expr.Program, error)
}

func (c *ExecController) getExecConfigs(cfg *config.Config, opts *option.ExecOptions) ([]*config.ExecConfig, error) {
	var execConfigs []*config.ExecConfig
	if opts.Template == "" && opts.TemplateKey != "" {
		a, ok := cfg.Exec[opts.TemplateKey]
		if !ok {
			if opts.TemplateKey != "default" {
				return nil, errors.New("template isn't found: " + opts.TemplateKey)
			}
			execConfigs = []*config.ExecConfig{
				{
					When: "ExitCode != 0",
					Template: `{{template "status" .}} {{template "link" .}}

{{template "join_command" .}}

{{template "hidden_combined_output" .}}`,
				},
			}
		} else {
			execConfigs = a
		}
	}
	return execConfigs, nil
}

// getExecConfig returns matched ExecConfig.
// If no ExecConfig matches, the second returned value is false.
func (c *ExecController) getExecConfig(
	execConfigs []*config.ExecConfig, cmtParams *ExecCommentParams,
) (*config.ExecConfig, bool, error) {
	for _, execConfig := range execConfigs {
		f, err := c.Expr.Match(execConfig.When, cmtParams)
		if err != nil {
			return nil, false, fmt.Errorf("test a condition is matched: %w", err)
		}
		if !f {
			continue
		}
		return execConfig, true, nil
	}
	return nil, false, nil
}

// getComment returns Comment.
// If the second returned value is false, no comment is posted.
func (c *ExecController) getComment(execConfigs []*config.ExecConfig, cmtParams *ExecCommentParams, templates map[string]string) (*github.Comment, bool, error) { //nolint:funlen
	tpl := cmtParams.Template
	tplForTooLong := ""
	var embeddedVarNames []string
	if tpl == "" {
		execConfig, f, err := c.getExecConfig(execConfigs, cmtParams)
		if err != nil {
			return nil, false, err
		}
		if !f {
			return nil, false, nil
		}
		if execConfig.DontComment {
			return nil, false, nil
		}
		tpl = execConfig.Template
		tplForTooLong = execConfig.TemplateForTooLong
		embeddedVarNames = execConfig.EmbeddedVarNames
	}

	body, err := c.Renderer.Render(tpl, templates, cmtParams)
	if err != nil {
		return nil, false, fmt.Errorf("render a comment template: %w", err)
	}
	bodyForTooLong, err := c.Renderer.Render(tplForTooLong, templates, cmtParams)
	if err != nil {
		return nil, false, fmt.Errorf("render a comment template_for_too_long: %w", err)
	}

	cmtCtrl := CommentController{
		GitHub:   c.GitHub,
		Expr:     c.Expr,
		Getenv:   c.Getenv,
		Platform: c.Platform,
	}

	embeddedMetadata := make(map[string]any, len(embeddedVarNames))
	for _, name := range embeddedVarNames {
		if v, ok := cmtParams.Vars[name]; ok {
			embeddedMetadata[name] = v
		}
	}

	embeddedComment, err := cmtCtrl.getEmbeddedComment(map[string]any{
		"SHA1":        cmtParams.SHA1,
		"TemplateKey": cmtParams.TemplateKey,
		"Vars":        embeddedMetadata,
	})
	if err != nil {
		return nil, false, err
	}

	body += embeddedComment
	bodyForTooLong += embeddedComment

	return &github.Comment{
		PRNumber:       cmtParams.PRNumber,
		Org:            cmtParams.Org,
		Repo:           cmtParams.Repo,
		Body:           body,
		BodyForTooLong: bodyForTooLong,
		SHA1:           cmtParams.SHA1,
		Vars:           cmtParams.Vars,
		TemplateKey:    cmtParams.TemplateKey,
	}, true, nil
}

func (c *ExecController) post(
	ctx context.Context, logger *slog.Logger, execConfigs []*config.ExecConfig, cmtParams *ExecCommentParams,
	templates map[string]string,
) error {
	cmt, f, err := c.getComment(execConfigs, cmtParams, templates)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}
	logger.Debug("comment meta data",
		"org", cmt.Org,
		"repo", cmt.Repo,
		"pr_number", cmt.PRNumber,
		"sha", cmt.SHA1,
	)

	for _, out := range cmtParams.Outputs {
		if err := c.handleOutput(ctx, cmt, out); err != nil {
			return err
		}
	}
	return nil
}

func (c *ExecController) handleOutput(ctx context.Context, cmt *github.Comment, out *option.Output) error {
	if out.GitHub {
		cmtCtrl := CommentController{
			GitHub: c.GitHub,
			Expr:   c.Expr,
			Getenv: c.Getenv,
		}
		if err := cmtCtrl.Post(ctx, cmt); err != nil {
			return fmt.Errorf("post a comment to GitHub: %w", err)
		}
		return nil
	}
	if out.File == "" {
		return nil
	}
	f, err := c.Fs.Stat(out.File)
	if err != nil {
		if !errors.Is(err, afero.ErrFileNotFound) {
			return fmt.Errorf("check if the output file exists: %w", err)
		}
		f, err := c.Fs.Create(out.File)
		if err != nil {
			return fmt.Errorf("create a file: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString(cmt.Body); err != nil {
			return fmt.Errorf("write a comment to a file: %w", err)
		}
		return nil
	}
	file, err := c.Fs.OpenFile(out.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, f.Mode())
	if err != nil {
		return fmt.Errorf("open a file to write a comment: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(cmt.Body); err != nil {
		return fmt.Errorf("write a comment to a file: %w", err)
	}
	return nil
}
