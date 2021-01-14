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
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
)

// Commenter is API to post a comment to GitHub
type Commenter interface {
	Create(ctx context.Context, cmt comment.Comment) error
	List(ctx context.Context, pr comment.PullRequest) ([]comment.IssueComment, error)
	HideComment(ctx context.Context, nodeID string) error
	GetAuthenticatedUser(ctx context.Context) (string, error)
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
	Stderr    io.Writer
	Commenter Commenter
	Renderer  Renderer
	Platform  Platform
	Config    config.Config
	Expr      Expr
}

type Platform interface {
	ComplementPost(opts *option.PostOptions) error
	ComplementExec(opts *option.ExecOptions) error
	CI() string
}

func listHiddenComments( //nolint:funlen
	ctx context.Context,
	commenter Commenter, exp Expr,
	getEnv func(string) string,
	stderr io.Writer,
	cmt comment.Comment,
	paramExpr map[string]interface{},
) ([]string, error) {
	if cmt.Minimize == "" {
		return nil, nil
	}
	login, err := commenter.GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("get an authenticated user: %w", err)
	}

	comments, err := commenter.List(ctx, comment.PullRequest{
		Org:      cmt.Org,
		Repo:     cmt.Repo,
		PRNumber: cmt.PRNumber,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	nodeIDs := []string{}
	prg, err := exp.Compile(cmt.Minimize)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, comment := range comments {
		nodeID := comment.ID

		// TODO remove these filters
		if !comment.ViewerCanMinimize {
			continue
		}
		if comment.IsMinimized {
			continue
		}
		if comment.Author.Login != login {
			continue
		}

		param := map[string]interface{}{
			"Comment": map[string]interface{}{
				"Body": comment.Body,
				// "CreatedAt": comment.CreatedAt,
			},
			"Commit": map[string]interface{}{
				"Org":      cmt.Org,
				"Repo":     cmt.Repo,
				"PRNumber": cmt.PRNumber,
				"SHA":      cmt.SHA1,
			},
			"Vars": cmt.Vars,
			"PostedComment": map[string]interface{}{
				"Body":        cmt.Body,
				"TemplateKey": cmt.TemplateKey,
			},
			"Env": getEnv,
		}

		for k, v := range paramExpr {
			param[k] = v
		}

		f, err := prg.Run(param)
		if err != nil {
			fmt.Fprintf(stderr, "[ERROR] judge whether an existing comment is hidden %s: %v\n", nodeID, err)
			continue
		}
		if !f {
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
		if err := commenter.HideComment(ctx, nodeID); err != nil {
			fmt.Fprintf(stderr, "[ERROR] hide an old comment %s: %v\n", nodeID, err)
			continue
		}
	}
	return nodeIDs, nil
}

func (ctrl PostController) listHiddenComments(ctx context.Context, cmt comment.Comment) ([]string, error) {
	return listHiddenComments(
		ctx, ctrl.Commenter, ctrl.Expr, ctrl.Getenv, ctrl.Stderr, cmt, nil)
}

func hideComments(ctx context.Context, commenter Commenter, stderr io.Writer, nodeIDs []string) {
	for _, nodeID := range nodeIDs {
		if err := commenter.HideComment(ctx, nodeID); err != nil {
			fmt.Fprintf(stderr, "[ERROR] hide an old comment %s: %v\n", nodeID, err)
			continue
		}
	}
}

func (ctrl PostController) hideComments(ctx context.Context, nodeIDs []string) {
	hideComments(ctx, ctrl.Commenter, ctrl.Stderr, nodeIDs)
}

func (ctrl PostController) Post(ctx context.Context, opts option.PostOptions) error {
	cmt, err := ctrl.getCommentParams(opts)
	if err != nil {
		return err
	}
	nodeIDs, err := ctrl.listHiddenComments(ctx, cmt)
	if err != nil {
		return err
	}
	if err := ctrl.Commenter.Create(ctx, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	ctrl.hideComments(ctx, nodeIDs)
	return nil
}

func (ctrl PostController) getCommentParams(opts option.PostOptions) (comment.Comment, error) { //nolint:funlen
	cmt := comment.Comment{}
	if ctrl.Platform != nil {
		if err := ctrl.Platform.ComplementPost(&opts); err != nil {
			return cmt, fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
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
		opts.Minimize = tpl.Minimize
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

	return comment.Comment{
		PRNumber:       opts.PRNumber,
		Org:            opts.Org,
		Repo:           opts.Repo,
		Body:           tpl,
		BodyForTooLong: tplForTooLong,
		SHA1:           opts.SHA1,
		Minimize:       opts.Minimize,
		Vars:           cfg.Vars,
		TemplateKey:    opts.TemplateKey,
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

func (ctrl PostController) readTemplateFromConfig(cfg config.Config, key string) (config.PostConfig, error) {
	if t, ok := cfg.Post[key]; ok {
		return t, nil
	}
	return config.PostConfig{}, errors.New("the template " + key + " isn't found")
}
