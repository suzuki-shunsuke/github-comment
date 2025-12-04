package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type PostController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If there is the standard input, it is treated as the comment template
	HasStdin func() bool
	Stdin    io.Reader
	Stderr   io.Writer
	GitHub   GitHub
	Renderer Renderer
	Platform Platform
	Config   *config.Config
	Expr     Expr
	Logger   *slog.Logger
}

func (c *PostController) Post(ctx context.Context, opts *option.PostOptions) error {
	cmt, err := c.getCommentParams(ctx, opts)
	if err != nil {
		return err
	}
	c.Logger.Debug("comment meta data",
		"org", cmt.Org,
		"repo", cmt.Repo,
		"pr_number", cmt.PRNumber,
		"sha", cmt.SHA1,
	)

	cmtCtrl := CommentController{
		GitHub: c.GitHub,
		Expr:   c.Expr,
		Getenv: c.Getenv,
	}
	return cmtCtrl.Post(ctx, cmt)
}

func (c *PostController) setUpdatedCommentID(ctx context.Context, cmt *github.Comment, updateCondition string) error { //nolint:funlen
	prg, err := c.Expr.Compile(updateCondition)
	if err != nil {
		return err //nolint:wrapcheck
	}

	login, err := c.GitHub.GetAuthenticatedUser(ctx)
	if err != nil {
		slogerr.WithError(c.Logger, err).Warn("get an authenticated user")
	}

	comments, err := c.GitHub.ListComments(ctx, &github.PullRequest{
		Org:      cmt.Org,
		Repo:     cmt.Repo,
		PRNumber: cmt.PRNumber,
	})
	if err != nil {
		return fmt.Errorf("list issue or pull request comments: %w", err)
	}
	c.Logger.Debug("get comments",
		"org", cmt.Org,
		"repo", cmt.Repo,
		"pr_number", cmt.PRNumber,
	)

	for _, comnt := range comments {
		if comnt.IsMinimized {
			// ignore minimized comments
			continue
		}
		if login != "" && comnt.Author.Login != login {
			// ignore other users' comments
			continue
		}

		metadata := map[string]any{}
		hasMeta := extractMetaFromComment(comnt.Body, &metadata)
		paramMap := map[string]any{
			"Comment": map[string]any{
				"Body":    comnt.Body,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]any{
				"Org":      cmt.Org,
				"Repo":     cmt.Repo,
				"PRNumber": cmt.PRNumber,
				"SHA1":     cmt.SHA1,
			},
			"Vars": cmt.Vars,
		}

		c.Logger.Debug("judge whether an existing comment is ready for editing",
			"node_id", comnt.ID,
			"condition", updateCondition,
			"param", paramMap,
		)
		f, err := prg.Run(paramMap)
		if err != nil {
			slogerr.WithError(c.Logger, err).Error("judge whether an existing comment is hidden",
				"node_id", comnt.ID,
			)
			continue
		}
		if !f {
			continue
		}
		cmt.CommentID = comnt.DatabaseID
	}
	return nil
}

// Reader is API to find and read the configuration file of github-comment
type Reader interface {
	FindAndRead(cfgPath, wd string) (config.Config, error)
}

type Renderer interface {
	Render(tpl string, templates map[string]string, params any) (string, error)
}

type PostTemplateParams struct {
	// PRNumber is the pull request number where the comment is posted
	PRNumber int
	// Org is the GitHub Organization or User name
	Org string
	// Repo is the GitHub Repository name
	Repo string
	// SHA1 is the commit SHA1
	SHA1        string
	TemplateKey string
	Vars        map[string]any
}

type Platform interface {
	ComplementPost(opts *option.PostOptions) error
	ComplementExec(opts *option.ExecOptions) error
	ComplementHide(opts *option.HideOptions) error
	CI() string
}

func (c *PostController) getCommentParams(ctx context.Context, opts *option.PostOptions) (*github.Comment, error) { //nolint:funlen,cyclop,gocognit
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
		if err := c.Platform.ComplementPost(opts); err != nil {
			return nil, fmt.Errorf("failed to complement opts with platform built in environment variables: %w", err)
		}
	}

	if opts.PRNumber == 0 && opts.SHA1 != "" {
		prNum, err := c.GitHub.PRNumberWithSHA(ctx, opts.Org, opts.Repo, opts.SHA1)
		if err != nil {
			slogerr.WithError(c.Logger, err).Warn("list associated prs",
				"org", opts.Org,
				"repo", opts.Repo,
				"sha", opts.SHA1,
			)
		}
		if prNum > 0 {
			opts.PRNumber = prNum
		}
	}

	if opts.Template == "" && opts.StdinTemplate {
		tpl, err := c.readTemplateFromStdin()
		if err != nil {
			return nil, err
		}
		opts.Template = tpl
	}

	if err := option.ValidatePost(opts); err != nil {
		return nil, fmt.Errorf("opts is invalid: %w", err)
	}

	if opts.Template == "" {
		tpl, err := c.readTemplateFromConfig(cfg, opts.TemplateKey)
		if err != nil {
			return nil, err
		}
		opts.Template = tpl.Template
		opts.TemplateForTooLong = tpl.TemplateForTooLong
		opts.EmbeddedVarNames = tpl.EmbeddedVarNames
		if opts.UpdateCondition == "" {
			opts.UpdateCondition = tpl.UpdateCondition
		}
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
	templates := template.GetTemplates(&template.ParamGetTemplates{
		Templates: cfg.Templates,
		CI:        ci,
	})
	tpl, err := c.Renderer.Render(opts.Template, templates, PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return nil, fmt.Errorf("render a template for post: %w", err)
	}
	tplForTooLong, err := c.Renderer.Render(opts.TemplateForTooLong, templates, PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return nil, fmt.Errorf("render a template template_for_too_long for post: %w", err)
	}

	cmtCtrl := CommentController{
		GitHub:   c.GitHub,
		Expr:     c.Expr,
		Getenv:   c.Getenv,
		Platform: c.Platform,
	}
	embeddedMetadata := make(map[string]any, len(opts.EmbeddedVarNames))
	for _, name := range opts.EmbeddedVarNames {
		if v, ok := cfg.Vars[name]; ok {
			embeddedMetadata[name] = v
		}
	}
	embeddedComment, err := cmtCtrl.getEmbeddedComment(map[string]any{
		"SHA1":        opts.SHA1,
		"TemplateKey": opts.TemplateKey,
		"Vars":        embeddedMetadata,
	})
	if err != nil {
		return nil, err
	}

	tpl += embeddedComment
	tplForTooLong += embeddedComment

	cmt := &github.Comment{
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		Body:           tpl,
		BodyForTooLong: tplForTooLong,
		SHA1:           opts.SHA1,
		Vars:           cfg.Vars,
		TemplateKey:    opts.TemplateKey,
	}
	if opts.UpdateCondition != "" && opts.PRNumber != 0 {
		if err := c.setUpdatedCommentID(ctx, cmt, opts.UpdateCondition); err != nil {
			return nil, err
		}
	}
	return cmt, nil
}

func (c *PostController) readTemplateFromStdin() (string, error) {
	if !c.HasStdin() {
		return "", nil
	}
	b, err := io.ReadAll(c.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read standard input: %w", err)
	}
	return string(b), nil
}

func (c *PostController) readTemplateFromConfig(cfg *config.Config, key string) (*config.PostConfig, error) {
	if t, ok := cfg.Post[key]; ok {
		return t, nil
	}
	return nil, errors.New("the template " + key + " isn't found")
}
