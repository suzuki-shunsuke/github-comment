package cmd

import (
	"context"
	"io"

	"github.com/suzuki-shunsuke/github-comment/pkg/constant"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner *Runner) Run(ctx context.Context, args []string) error { //nolint:funlen
	app := cli.App{
		Name:    "github-comment",
		Usage:   "post a comment to GitHub",
		Version: constant.Version,
		Commands: []*cli.Command{
			{
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
			},
			{
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
			},
			{
				Name:   "init",
				Usage:  "scaffold a configuration file if it doesn't exist",
				Action: runner.initAction,
			},
			{
				Name:   "hide",
				Usage:  "hide issue or pull request comments",
				Action: runner.hideAction,
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
						Name:  "config",
						Usage: "configuration file path",
					},
					&cli.StringFlag{
						Name:  "condition",
						Usage: "hide condition",
					},
					&cli.StringFlag{
						Name:    "hide-key",
						Aliases: []string{"k"},
						Usage:   "hide condition key",
						Value:   "default",
					},
					&cli.IntFlag{
						Name:  "pr",
						Usage: "GitHub pull request number",
					},
					&cli.StringFlag{
						Name:  "sha1",
						Usage: "commit sha1",
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
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				EnvVars: []string{"GITHUB_COMMENT_LOG_LEVEL"},
			},
		},
	}
	return app.RunContext(ctx, args) //nolint:wrapcheck
}
