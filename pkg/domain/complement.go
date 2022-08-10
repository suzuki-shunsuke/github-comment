package domain

import (
	"context"
	"io"

	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type ComplementEntry interface {
	Entry() (string, error)
	Type() string
}

type ComplementWithNameEntry interface {
	Entry() (string, error)
	Type() string
}

type Platform interface {
	ComplementPost(opts *option.PostOptions) error
	ComplementExec(opts *option.ExecOptions) error
	ComplementHide(opts *option.HideOptions) error
	CI() string
}

type Expr interface {
	Match(expression string, params interface{}) (bool, error)
	Compile(expression string) (expr.Program, error)
}

type Renderer interface {
	Render(tpl string, templates map[string]string, params interface{}) (string, error)
}

// GitHub is API to post a comment to GitHub
type GitHub interface {
	CreateComment(ctx context.Context, cmt *github.Comment) error
	ListComments(ctx context.Context, pr *github.PullRequest) ([]*github.IssueComment, error)
	HideComment(ctx context.Context, nodeID string) error
	GetAuthenticatedUser(ctx context.Context) (string, error)
	PRNumberWithSHA(ctx context.Context, owner, repo, sha string) (int, error)
}

type Stdio struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
