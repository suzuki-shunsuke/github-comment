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

func Post(
	ctx context.Context, wd string, opts *option.PostOptions,
	getEnv func(string) string, isTerminal func() bool, stdin io.Reader,
	httpClient *http.Client, existFile func(string) bool,
	readConfig func(string, *config.Config) error,
) error {
	if option.IsCircleCI(getEnv) {
		if err := option.ComplementPost(opts, getEnv); err != nil {
			return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if opts.Template == "" && !isTerminal() {
		if b, err := ioutil.ReadAll(stdin); err == nil {
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
			p, b, err := config.Find(wd, existFile)
			if err != nil {
				return err
			}
			if !b {
				return errors.New("configuration file isn't found")
			}
			opts.ConfigPath = p
		}
		if err := readConfig(opts.ConfigPath, cfg); err != nil {
			return err
		}
		if t, ok := cfg.Post[opts.TemplateKey]; ok {
			opts.Template = t
		} else {
			return errors.New("the template " + opts.TemplateKey + " isn't found")
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
	}

	cmt := &comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     opts.Template,
		SHA1:     opts.SHA1,
	}
	if err := comment.Create(ctx, httpClient, opts.Token, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}
