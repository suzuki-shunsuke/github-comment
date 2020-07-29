package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

// Commenter is API to post a comment to GitHub
type Commenter interface {
	Create(ctx context.Context, cmt comment.Comment) error
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

type PostController struct {
	// Wd is a path to the working directory
	Wd string
	// Getenv returns the environment variable. os.Getenv
	Getenv func(string) string
	// HasStdin returns true if there is the standard input
	// If thre is the standard input, it is treated as the comment template
	HasStdin  func() bool
	Stdin     io.Reader
	Reader    Reader
	Commenter Commenter
	Renderer  Renderer
}

func (ctrl PostController) Post(ctx context.Context, opts option.PostOptions) error {
	cmt, err := ctrl.getCommentParams(ctx, opts)
	if err != nil {
		return err
	}
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}

func (ctrl PostController) getCommentParams(
	ctx context.Context, opts option.PostOptions,
) (comment.Comment, error) {
	cmt := comment.Comment{}
	if option.IsCircleCI(ctrl.Getenv) {
		if err := option.ComplementPost(&opts, ctrl.Getenv); err != nil {
			return cmt, fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if opts.Template == "" {
		tpl, err := ctrl.readTemplateFromStdin()
		if err != nil {
			return cmt, err
		}
		opts.Template = tpl
	}

	cfg, err := ctrl.Reader.FindAndRead(opts.ConfigPath, ctrl.Wd)
	if err != nil {
		return cmt, err
	}
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
		opts.Template = tpl
	}

	for k, v := range opts.Vars {
		cfg.Vars[k] = v
	}

	tpl, err := ctrl.Renderer.Render(opts.Template, cfg.Templates, PostTemplateParams{
		PRNumber:    opts.PRNumber,
		Org:         opts.Org,
		Repo:        opts.Repo,
		SHA1:        opts.SHA1,
		TemplateKey: opts.TemplateKey,
		Vars:        cfg.Vars,
	})
	if err != nil {
		return cmt, err
	}

	return comment.Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     tpl,
		SHA1:     opts.SHA1,
	}, nil
}

func (ctrl PostController) readTemplateFromStdin() (string, error) {
	if !ctrl.HasStdin() {
		return "", nil
	}
	b, err := ioutil.ReadAll(ctrl.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read standard input: %w", err)
	}
	return string(b), nil
}

func (ctrl PostController) readTemplateFromConfig(cfg config.Config, key string) (string, error) {
	if t, ok := cfg.Post[key]; ok {
		return t, nil
	}
	return "", errors.New("the template " + key + " isn't found")
}
