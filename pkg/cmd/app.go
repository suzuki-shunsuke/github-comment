package cmd

import (
	"context"
	"io"
	"log/slog"

	"github.com/suzuki-shunsuke/urfave-cli-v3-util/helpall"
	"github.com/urfave/cli/v3"
)

type Runner struct {
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Logger      *slog.Logger
	LogLevelVar *slog.LevelVar
	LDFlags     *LDFlags
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (f *LDFlags) AppVersion() string {
	return f.Version + " (" + f.Commit + ")"
}

func (r *Runner) Run(ctx context.Context, args []string) error { //nolint:funlen
	return helpall.With(&cli.Command{ //nolint:wrapcheck
		Name:    "github-comment",
		Usage:   "post a comment to GitHub",
		Version: r.LDFlags.AppVersion(),
		Commands: []*cli.Command{
			{
				Name:   "post",
				Usage:  "post a comment",
				Action: r.postAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "org",
						Usage:   "GitHub organization name",
						Sources: cli.EnvVars("GH_COMMENT_REPO_ORG"),
					},
					&cli.StringFlag{
						Name:    "repo",
						Usage:   "GitHub repository name",
						Sources: cli.EnvVars("GH_COMMENT_REPO_NAME"),
					},
					&cli.StringFlag{
						Name:    "token",
						Usage:   "GitHub API token",
						Sources: cli.EnvVars("GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"),
					},
					&cli.StringFlag{
						Name:    "sha1",
						Usage:   "commit sha1",
						Sources: cli.EnvVars("GH_COMMENT_SHA1"),
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
						Name:    "config",
						Usage:   "configuration file path",
						Sources: cli.EnvVars("GH_COMMENT_CONFIG"),
					},
					&cli.IntFlag{
						Name:    "pr",
						Usage:   "GitHub pull request number",
						Sources: cli.EnvVars("GH_COMMENT_PR_NUMBER"),
					},
					&cli.StringSliceFlag{
						Name:  "var",
						Usage: "template variable",
					},
					&cli.StringSliceFlag{
						Name:  "var-file",
						Usage: "template variable name and file path",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "output a comment to standard error output instead of posting to GitHub",
					},
					&cli.BoolFlag{
						Name:    "skip-no-token",
						Aliases: []string{"n"},
						Usage:   "works like dry-run if the GitHub Access Token isn't set",
						Sources: cli.EnvVars("GH_COMMENT_SKIP_NO_TOKEN", "GITHUB_COMMENT_SKIP_NO_TOKEN"),
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
					&cli.StringFlag{
						Name:    "update-condition",
						Aliases: []string{"u"},
						Usage:   "update the comment that matches with the condition",
					},
				},
			},
			{
				Name:   "exec",
				Usage:  "execute a command and post the result as a comment",
				Action: r.execAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "org",
						Usage:   "GitHub organization name",
						Sources: cli.EnvVars("GH_COMMENT_REPO_ORG"),
					},
					&cli.StringFlag{
						Name:    "repo",
						Usage:   "GitHub repository name",
						Sources: cli.EnvVars("GH_COMMENT_REPO_NAME"),
					},
					&cli.StringFlag{
						Name:    "token",
						Usage:   "GitHub API token",
						Sources: cli.EnvVars("GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"),
					},
					&cli.StringFlag{
						Name:    "sha1",
						Usage:   "commit sha1",
						Sources: cli.EnvVars("GH_COMMENT_SHA1"),
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
						Name:    "config",
						Usage:   "configuration file path",
						Sources: cli.EnvVars("GH_COMMENT_CONFIG"),
					},
					&cli.IntFlag{
						Name:    "pr",
						Usage:   "GitHub pull request number",
						Sources: cli.EnvVars("GH_COMMENT_PR_NUMBER"),
					},
					&cli.StringSliceFlag{
						Name:  "out",
						Usage: "output destination",
					},
					&cli.StringSliceFlag{
						Name:  "var",
						Usage: "template variable",
					},
					&cli.StringSliceFlag{
						Name:  "var-file",
						Usage: "template variable name and file path",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "output a comment to standard error output instead of posting to GitHub",
					},
					&cli.BoolFlag{
						Name:    "skip-no-token",
						Aliases: []string{"n"},
						Usage:   "works like dry-run if the GitHub Access Token isn't set",
						Sources: cli.EnvVars("GH_COMMENT_SKIP_NO_TOKEN", "GITHUB_COMMENT_SKIP_NO_TOKEN"),
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
				Action: r.initAction,
			},
			{
				Name:   "hide",
				Usage:  "hide issue or pull request comments",
				Action: r.hideAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "org",
						Usage:   "GitHub organization name",
						Sources: cli.EnvVars("GH_COMMENT_REPO_ORG"),
					},
					&cli.StringFlag{
						Name:    "repo",
						Usage:   "GitHub repository name",
						Sources: cli.EnvVars("GH_COMMENT_REPO_NAME"),
					},
					&cli.StringFlag{
						Name:    "token",
						Usage:   "GitHub API token",
						Sources: cli.EnvVars("GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"),
					},
					&cli.StringFlag{
						Name:    "config",
						Usage:   "configuration file path",
						Sources: cli.EnvVars("GH_COMMENT_CONFIG"),
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
						Name:    "pr",
						Usage:   "GitHub pull request number",
						Sources: cli.EnvVars("GH_COMMENT_PR_NUMBER"),
					},
					&cli.StringFlag{
						Name:  "sha1",
						Usage: "commit sha1",
					},
					&cli.StringSliceFlag{
						Name:  "var",
						Usage: "template variable",
					},
					&cli.StringSliceFlag{
						Name:  "var-file",
						Usage: "template variable name and file path",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "output a comment to standard error output instead of posting to GitHub",
					},
					&cli.BoolFlag{
						Name:    "skip-no-token",
						Aliases: []string{"n"},
						Usage:   "works like dry-run if the GitHub Access Token isn't set",
						Sources: cli.EnvVars("GH_COMMENT_SKIP_NO_TOKEN", "GITHUB_COMMENT_SKIP_NO_TOKEN"),
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
				Sources: cli.EnvVars("GH_COMMENT_LOG_LEVEL"),
			},
		},
	}, nil).Run(ctx, args)
}
