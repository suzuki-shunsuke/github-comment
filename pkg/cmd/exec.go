package cmd

import (
	"fmt"
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

func (runner Runner) execCommand() cli.Command { //nolint:funlen,dupl
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
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "output a comment to standard error output instead of posting to GitHub",
			},
			&cli.BoolFlag{
				Name:    "skip-no-token",
				Aliases: []string{"n"},
				Usage:   "works like dry-run if the GitHub Access Token isn't set",
				EnvVars: []string{"GITHUB_COMMENT_SKIP_NO_TOKEN"},
			},
			&cli.BoolFlag{
				Name:    "silent",
				Aliases: []string{"s"},
				Usage:   "suppress the output of dry-run and skip-no-token",
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
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
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

func getExecCommenter(opts option.ExecOptions) api.Commenter {
	if opts.DryRun {
		return comment.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}
	}
	if opts.SkipNoToken && opts.Token == "" {
		return comment.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}
	}
	return comment.Commenter{
		Token:      opts.Token,
		HTTPClient: httpclient.New("https://api.github.com"),
	}
}

func (runner Runner) execAction(c *cli.Context) error {
	opts := option.ExecOptions{}
	if err := parseExecOptions(&opts, c); err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	var pt api.Platform
	if p, f := platform.Get(); f {
		pt = p
	}

	cfgReader := config.Reader{
		ExistFile: existFile,
	}
	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken
	opts.Silent = opts.Silent || cfg.Silent

	ctrl := api.ExecController{
		Wd:        wd,
		Getenv:    os.Getenv,
		Stdin:     runner.Stdin,
		Stdout:    runner.Stdout,
		Stderr:    runner.Stderr,
		Commenter: getExecCommenter(opts),
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
		Config:   cfg,
	}
	return ctrl.Exec(c.Context, opts)
}
