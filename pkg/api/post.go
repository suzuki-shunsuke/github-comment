package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
)

type PostController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If thre is the standard input, it is treated as the comment template
	HasStdin  func() bool
	Stdin     io.Reader
	Stderr    io.Writer
	Commenter Commenter
	Renderer  Renderer
	Platform  Platform
	Config    config.Config
	Expr      Expr
}

func (ctrl *PostController) Post(ctx context.Context, opts option.PostOptions) error {
	cmt, err := ctrl.getCommentParams(opts)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"org":       cmt.Org,
		"repo":      cmt.Repo,
		"pr_number": cmt.PRNumber,
		"sha":       cmt.SHA1,
	}).Debug("comment meta data")

	cmtCtrl := CommentController{
		Commenter: ctrl.Commenter,
		Expr:      ctrl.Expr,
		Getenv:    ctrl.Getenv,
	}
	return cmtCtrl.Post(ctx, cmt, nil)
}

// Reader is API to find and read the configuration file of github-comment
type Reader interface {
	FindAndRead(cfgPath, wd string) (config.Config, error)
}

type Renderer interface {
	Render(tpl string, templates map[string]string, params interface{}) (string, error)
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
	Vars        map[string]interface{}
}

type Platform interface {
	ComplementPost(opts *option.PostOptions) error
	ComplementExec(opts *option.ExecOptions) error
	ComplementHide(opts *option.HideOptions) error
	CI() string
}

func (ctrl *PostController) getCommentParams(opts option.PostOptions) (comment.Comment, error) { //nolint:funlen,cyclop
	cmt := comment.Comment{}
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementPost(&opts); err != nil {
			return cmt, fmt.Errorf("failed to complement opts with platform built in environment variables: %w", err)
		}
	}
	if opts.Template == "" && opts.StdinTemplate {
		tpl, err := ctrl.readTemplateFromStdin()
		if err != nil {
			return cmt, err
		}
		opts.Template = tpl
	}

	cfg := ctrl.Config

	if opts.Org == "" {
		opts.Org = cfg.Base.Org
	}
	if opts.Repo == "" {
		opts.Repo = cfg.Base.Repo
	}

	if err := option.ValidatePost(opts); err != nil {
		return cmt, fmt.Errorf("opts is invalid: %w", err)
	}

	if opts.Template == "" {
		tpl, err := ctrl.readTemplateFromConfig(cfg, opts.TemplateKey)
		if err != nil {
			return cmt, err
		}
		opts.Template = tpl.Template
		opts.TemplateForTooLong = tpl.TemplateForTooLong
		opts.EmbeddedVarNames = tpl.EmbeddedVarNames
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]interface{}, len(opts.Vars))
	}
	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	ci := ""
	if ctrl.Platform != nil {
		ci = ctrl.Platform.CI()
	}
	templates := template.GetTemplates(template.ParamGetTemplates{
		Templates: cfg.Templates,
		CI:        ci,
	})
	tpl, err := ctrl.Renderer.Render(opts.Template, templates, PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return cmt, fmt.Errorf("render a template for post: %w", err)
	}
	tplForTooLong, err := ctrl.Renderer.Render(opts.TemplateForTooLong, templates, PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return cmt, fmt.Errorf("render a template template_for_too_long for post: %w", err)
	}

	cmtCtrl := CommentController{
		Commenter: ctrl.Commenter,
		Expr:      ctrl.Expr,
		Getenv:    ctrl.Getenv,
		Platform:  ctrl.Platform,
	}
	embeddedMetadata := make(map[string]interface{}, len(opts.EmbeddedVarNames))
	for _, name := range opts.EmbeddedVarNames {
		if v, ok := cfg.Vars[name]; ok {
			embeddedMetadata[name] = v
		}
	}
	embeddedComment, err := cmtCtrl.getEmbeddedComment(map[string]interface{}{
		"SHA1":        opts.SHA1,
		"TemplateKey": opts.TemplateKey,
		"Vars":        embeddedMetadata,
	})
	if err != nil {
		return cmt, err
	}

	tpl += embeddedComment
	tplForTooLong += embeddedComment

	return comment.Comment{
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		Body:           tpl,
		BodyForTooLong: tplForTooLong,
		SHA1:           opts.SHA1,
		HideOldComment: opts.HideOldComment,
		Vars:           cfg.Vars,
		TemplateKey:    opts.TemplateKey,
	}, nil
}

func (ctrl *PostController) readTemplateFromStdin() (string, error) {
	if !ctrl.HasStdin() {
		return "", nil
	}
	b, err := ioutil.ReadAll(ctrl.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read standard input: %w", err)
	}
	return string(b), nil
}

func (ctrl *PostController) readTemplateFromConfig(cfg config.Config, key string) (config.PostConfig, error) {
	if t, ok := cfg.Post[key]; ok {
		return t, nil
	}
	return config.PostConfig{}, errors.New("the template " + key + " isn't found")
}
