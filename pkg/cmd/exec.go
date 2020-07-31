package cmd

import (
	"io"
	"os"

	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
	"github.com/suzuki-shunsuke/go-httpclient/httpclient"
	"github.com/urfave/cli/v2"
)

func (runner Runner) execCommand() cli.Command {
	return cli.Command{
		Name:   "exec",
		Usage:  "execute a command and post the result as a comment",
		Action: runner.execAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Usage: "GitHub organization name",
			},
			&cli.StringFlag{
				Name:  "repo",
				Usage: "GitHub repository name",
			},
			&cli.StringFlag{
				Name:    "token",
				Usage:   "GitHub API token",
				EnvVars: []string{"GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"},
			},
			&cli.StringFlag{
				Name:  "sha1",
				Usage: "commit sha1",
			},
			&cli.StringFlag{
				Name:  "template",
				Usage: "comment template",
			},
			&cli.StringFlag{
				Name:    "template-key",
				Aliases: []string{"k"},
				Usage:   "comment template key",
				Value:   "default",
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "configuration file path",
			},
			&cli.IntFlag{
				Name:  "pr",
				Usage: "GitHub pull request number",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "template variable",
			},
		},
	}
}

func parseExecOptions(opts *option.ExecOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.Args = c.Args().Slice()
	vars, err := parseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return err
	}
	opts.Vars = vars
	return nil
}

func existFile(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (runner Runner) execAction(c *cli.Context) error {
	opts := option.ExecOptions{}
	if err := parseExecOptions(&opts, c); err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	pt := platform.Get(os.Getenv, func(p string) (io.ReadCloser, error) {
		return os.Open(p)
	})

	ctrl := api.ExecController{
		Wd:     wd,
		Getenv: os.Getenv,
		Reader: config.Reader{
			ExistFile: existFile,
		},
		Stdin:  runner.Stdin,
		Stdout: runner.Stdout,
		Stderr: runner.Stderr,
		Commenter: comment.Commenter{
			Token:      opts.Token,
			HTTPClient: httpclient.New("https://api.github.com"),
		},
		Renderer: template.Renderer{
			Getenv: os.Getenv,
		},
		Executor: execute.Executor{
			Stdout: runner.Stdout,
			Stderr: runner.Stderr,
			Env:    os.Environ(),
		},
		Expr:     expr.Expr{},
		Platform: pt,
	}
	return ctrl.Exec(c.Context, opts)
}
