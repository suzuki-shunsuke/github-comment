package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type PostTemplateParams struct {
	PRNumber    int
	Org         string
	Repo        string
	SHA1        string
	TemplateKey string
}

type Commenter interface {
	Create(ctx context.Context, cmt comment.Comment) error
}

type PostController struct {
	Wd         string
	Getenv     func(string) string
	IsTerminal func() bool
	Stdin      io.Reader
	ExistFile  func(string) bool
	ReadConfig func(string, *config.Config) error
	Commenter  Commenter
}

func (ctrl PostController) Post(ctx context.Context, opts *option.PostOptions) error {
	cmt, err := ctrl.getCommentParams(ctx, opts)
	if err != nil {
		return err
	}
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}

func (ctrl PostController) getCommentParams(ctx context.Context, opts *option.PostOptions) (comment.Comment, error) {
	cmt := comment.Comment{}
	if option.IsCircleCI(ctrl.Getenv) {
		if err := option.ComplementPost(opts, ctrl.Getenv); err != nil {
			return cmt, fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if err := ctrl.readTemplateFromStdin(opts); err != nil {
		return cmt, err
	}

	if err := option.ValidatePost(opts); err != nil {
		return cmt, fmt.Errorf("opts is invalid: %w", err)
	}

	if opts.Template == "" {
		if err := ctrl.readTemplateFromConfig(opts); err != nil {
			return cmt, err
		}
	}

	if err := ctrl.render(opts); err != nil {
		return cmt, err
	}

	return comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     opts.Template,
		SHA1:     opts.SHA1,
	}, nil
}

func (ctrl PostController) readTemplateFromStdin(opts *option.PostOptions) error {
	if opts.Template != "" || ctrl.IsTerminal() {
		return nil
	}
	b, err := ioutil.ReadAll(ctrl.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read standard input: %w", err)
	}
	opts.Template = string(b)
	return nil
}

func (ctrl PostController) readTemplateFromConfig(opts *option.PostOptions) error {
	cfg := &config.Config{}
	if opts.ConfigPath == "" {
		p, b, err := config.Find(ctrl.Wd, ctrl.ExistFile)
		if err != nil {
			return err
		}
		if !b {
			return errors.New("configuration file isn't found")
		}
		opts.ConfigPath = p
	}
	if err := ctrl.ReadConfig(opts.ConfigPath, cfg); err != nil {
		return err
	}
	if t, ok := cfg.Post[opts.TemplateKey]; ok {
		opts.Template = t
		return nil
	}
	return errors.New("the template " + opts.TemplateKey + " isn't found")
}

func (ctrl PostController) render(opts *option.PostOptions) error {
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"Env": ctrl.Getenv,
	}).Parse(opts.Template)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, &PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
	}); err != nil {
		return err
	}
	opts.Template = buf.String()
	return nil
}
