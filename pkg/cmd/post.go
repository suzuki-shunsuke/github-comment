package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func (runner Runner) postCommand() cli.Command { //nolint:funlen
	return cli.Command{
		Name:   "post",
		Usage:  "post a comment",
		Action: runner.postAction,
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
			&cli.BoolFlag{
				Name:  "stdin-template",
				Usage: "read standard input as the template",
			},
		},
	}
}

func parseVarsFlag(varsSlice []string) (map[string]string, error) {
	vars := make(map[string]string, len(varsSlice))
	for _, v := range varsSlice {
		a := strings.SplitN(v, ":", 2)
		if len(a) < 2 { //nolint:gomnd
			return nil, errors.New("invalid var flag. The format should be '--var <key>:<value>")
		}
		vars[a[0]] = a[1]
	}
	return vars, nil
}

// parsePostOptions parses the command line arguments of the subcommand "post".
func parsePostOptions(opts *option.PostOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
	opts.StdinTemplate = c.Bool("stdin-template")
	vars, err := parseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return err
	}
	opts.Vars = vars
	return nil
}

func getPostCommenter(ctx context.Context, opts option.PostOptions) api.Commenter {
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

	return comment.New(ctx, opts.Token)
}

// postAction is an entrypoint of the subcommand "post".
func (runner Runner) postAction(c *cli.Context) error {
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		if skipComment {
			return nil
		}
	}
	opts := option.PostOptions{}
	if err := parsePostOptions(&opts, c); err != nil {
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

	ctrl := api.PostController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !terminal.IsTerminal(0)
		},
		Stdin:     runner.Stdin,
		Stderr:    runner.Stderr,
		Commenter: getPostCommenter(c.Context, opts),
		Renderer: template.Renderer{
			Getenv: os.Getenv,
		},
		Platform: pt,
		Config:   cfg,
		Expr:     expr.Expr{},
	}
	return ctrl.Post(c.Context, opts)
}
