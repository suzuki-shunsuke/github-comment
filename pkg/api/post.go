package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

type PostController struct {
	Wd         string
	Getenv     func(string) string
	IsTerminal func() bool
	Stdin      io.Reader
	HTTPClient *http.Client
	ExistFile  func(string) bool
	ReadConfig func(string, *config.Config) error
}

func (ctrl PostController) Post(
	ctx context.Context, opts *option.PostOptions,
) error {
	if option.IsCircleCI(ctrl.Getenv) {
		if err := option.ComplementPost(opts, ctrl.Getenv); err != nil {
			return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if opts.Template == "" && !ctrl.IsTerminal() {
		if b, err := ioutil.ReadAll(ctrl.Stdin); err == nil {
			opts.Template = string(b)
		} else {
			return fmt.Errorf("failed to read standard input: %w", err)
		}
	}

	if err := option.ValidatePost(opts); err != nil {
		return fmt.Errorf("opts is invalid: %w", err)
	}

	if opts.Template == "" {
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
		} else {
			return errors.New("the template " + opts.TemplateKey + " isn't found")
		}
	}
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"Env": os.Getenv,
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

	cmt := &comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     opts.Template,
		SHA1:     opts.SHA1,
	}
	if err := comment.Create(ctx, ctrl.HTTPClient, opts.Token, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}
