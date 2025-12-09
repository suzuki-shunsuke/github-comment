package cmd

import (
	"context"
	"io"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, logger *slogutil.Logger, env *urfave.Env) error { //nolint:funlen,maintidx
	r := &Runner{
		Stdin:  env.Stdin,
		Stdout: env.Stdout,
		Stderr: env.Stderr,
	}

	globalFlags := &GlobalFlags{}
	postArgs := &PostArgs{GlobalFlags: globalFlags}
	execArgs := &ExecArgs{GlobalFlags: globalFlags}
	hideArgs := &HideArgs{GlobalFlags: globalFlags}

	return urfave.Command(env, &cli.Command{ //nolint:wrapcheck
		Name:  "github-comment",
		Usage: "post a comment to GitHub",
		Commands: []*cli.Command{
			{
				Name:  "post",
				Usage: "post a comment",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return r.postAction(ctx, logger, postArgs)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "org",
						Usage:       "GitHub organization name",
						Sources:     cli.EnvVars("GH_COMMENT_REPO_ORG"),
						Destination: &postArgs.Org,
					},
					&cli.StringFlag{
						Name:        "repo",
						Usage:       "GitHub repository name",
						Sources:     cli.EnvVars("GH_COMMENT_REPO_NAME"),
						Destination: &postArgs.Repo,
					},
					&cli.StringFlag{
						Name:        "token",
						Usage:       "GitHub API token",
						Sources:     cli.EnvVars("GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"),
						Destination: &postArgs.Token,
					},
					&cli.StringFlag{
						Name:        "sha1",
						Usage:       "commit sha1",
						Sources:     cli.EnvVars("GH_COMMENT_SHA1"),
						Destination: &postArgs.SHA1,
					},
					&cli.StringFlag{
						Name:        "template",
						Usage:       "comment template",
						Destination: &postArgs.Template,
					},
					&cli.StringFlag{
						Name:        "template-key",
						Aliases:     []string{"k"},
						Usage:       "comment template key",
						Value:       "default",
						Destination: &postArgs.TemplateKey,
					},
					&cli.StringFlag{
						Name:        "config",
						Usage:       "configuration file path",
						Sources:     cli.EnvVars("GH_COMMENT_CONFIG"),
						Destination: &postArgs.ConfigPath,
					},
					&cli.IntFlag{
						Name:        "pr",
						Usage:       "GitHub pull request number",
						Sources:     cli.EnvVars("GH_COMMENT_PR_NUMBER"),
						Destination: &postArgs.PRNumber,
					},
					&cli.StringSliceFlag{
						Name:        "var",
						Usage:       "template variable",
						Destination: &postArgs.Vars,
					},
					&cli.StringSliceFlag{
						Name:        "var-file",
						Usage:       "template variable name and file path",
						Destination: &postArgs.VarFiles,
					},
					&cli.BoolFlag{
						Name:        "dry-run",
						Usage:       "output a comment to standard error output instead of posting to GitHub",
						Destination: &postArgs.DryRun,
					},
					&cli.BoolFlag{
						Name:        "skip-no-token",
						Aliases:     []string{"n"},
						Usage:       "works like dry-run if the GitHub Access Token isn't set",
						Sources:     cli.EnvVars("GH_COMMENT_SKIP_NO_TOKEN", "GITHUB_COMMENT_SKIP_NO_TOKEN"),
						Destination: &postArgs.SkipNoToken,
					},
					&cli.BoolFlag{
						Name:        "silent",
						Aliases:     []string{"s"},
						Usage:       "suppress the output of dry-run and skip-no-token",
						Destination: &postArgs.Silent,
					},
					&cli.BoolFlag{
						Name:        "stdin-template",
						Usage:       "read standard input as the template",
						Destination: &postArgs.StdinTemplate,
					},
					&cli.StringFlag{
						Name:        "update-condition",
						Aliases:     []string{"u"},
						Usage:       "update the comment that matches with the condition",
						Destination: &postArgs.UpdateCondition,
					},
				},
			},
			{
				Name:  "exec",
				Usage: "execute a command and post the result as a comment",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return r.execAction(ctx, logger, execArgs)
				},
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name:        "args",
						Max:         -1,
						Destination: &execArgs.Args,
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "org",
						Usage:       "GitHub organization name",
						Sources:     cli.EnvVars("GH_COMMENT_REPO_ORG"),
						Destination: &execArgs.Org,
					},
					&cli.StringFlag{
						Name:        "repo",
						Usage:       "GitHub repository name",
						Sources:     cli.EnvVars("GH_COMMENT_REPO_NAME"),
						Destination: &execArgs.Repo,
					},
					&cli.StringFlag{
						Name:        "token",
						Usage:       "GitHub API token",
						Sources:     cli.EnvVars("GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"),
						Destination: &execArgs.Token,
					},
					&cli.StringFlag{
						Name:        "sha1",
						Usage:       "commit sha1",
						Sources:     cli.EnvVars("GH_COMMENT_SHA1"),
						Destination: &execArgs.SHA1,
					},
					&cli.StringFlag{
						Name:        "template",
						Usage:       "comment template",
						Destination: &execArgs.Template,
					},
					&cli.StringFlag{
						Name:        "template-key",
						Aliases:     []string{"k"},
						Usage:       "comment template key",
						Value:       "default",
						Destination: &execArgs.TemplateKey,
					},
					&cli.StringFlag{
						Name:        "config",
						Usage:       "configuration file path",
						Sources:     cli.EnvVars("GH_COMMENT_CONFIG"),
						Destination: &execArgs.ConfigPath,
					},
					&cli.IntFlag{
						Name:        "pr",
						Usage:       "GitHub pull request number",
						Sources:     cli.EnvVars("GH_COMMENT_PR_NUMBER"),
						Destination: &execArgs.PRNumber,
					},
					&cli.StringSliceFlag{
						Name:        "out",
						Usage:       "output destination",
						Destination: &execArgs.Outputs,
					},
					&cli.StringSliceFlag{
						Name:        "var",
						Usage:       "template variable",
						Destination: &execArgs.Vars,
					},
					&cli.StringSliceFlag{
						Name:        "var-file",
						Usage:       "template variable name and file path",
						Destination: &execArgs.VarFiles,
					},
					&cli.BoolFlag{
						Name:        "dry-run",
						Usage:       "output a comment to standard error output instead of posting to GitHub",
						Destination: &execArgs.DryRun,
					},
					&cli.BoolFlag{
						Name:        "skip-no-token",
						Aliases:     []string{"n"},
						Usage:       "works like dry-run if the GitHub Access Token isn't set",
						Sources:     cli.EnvVars("GH_COMMENT_SKIP_NO_TOKEN", "GITHUB_COMMENT_SKIP_NO_TOKEN"),
						Destination: &execArgs.SkipNoToken,
					},
					&cli.BoolFlag{
						Name:        "silent",
						Aliases:     []string{"s"},
						Usage:       "suppress the output of dry-run and skip-no-token",
						Destination: &execArgs.Silent,
					},
				},
			},
			{
				Name:   "init",
				Usage:  "scaffold a configuration file if it doesn't exist",
				Action: r.initAction,
			},
			{
				Name:  "hide",
				Usage: "hide issue or pull request comments",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return r.hideAction(ctx, logger, hideArgs)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "org",
						Usage:       "GitHub organization name",
						Sources:     cli.EnvVars("GH_COMMENT_REPO_ORG"),
						Destination: &hideArgs.Org,
					},
					&cli.StringFlag{
						Name:        "repo",
						Usage:       "GitHub repository name",
						Sources:     cli.EnvVars("GH_COMMENT_REPO_NAME"),
						Destination: &hideArgs.Repo,
					},
					&cli.StringFlag{
						Name:        "token",
						Usage:       "GitHub API token",
						Sources:     cli.EnvVars("GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"),
						Destination: &hideArgs.Token,
					},
					&cli.StringFlag{
						Name:        "config",
						Usage:       "configuration file path",
						Sources:     cli.EnvVars("GH_COMMENT_CONFIG"),
						Destination: &hideArgs.ConfigPath,
					},
					&cli.StringFlag{
						Name:        "condition",
						Usage:       "hide condition",
						Destination: &hideArgs.Condition,
					},
					&cli.StringFlag{
						Name:        "hide-key",
						Aliases:     []string{"k"},
						Usage:       "hide condition key",
						Value:       "default",
						Destination: &hideArgs.HideKey,
					},
					&cli.IntFlag{
						Name:        "pr",
						Usage:       "GitHub pull request number",
						Sources:     cli.EnvVars("GH_COMMENT_PR_NUMBER"),
						Destination: &hideArgs.PRNumber,
					},
					&cli.StringFlag{
						Name:        "sha1",
						Usage:       "commit sha1",
						Destination: &hideArgs.SHA1,
					},
					&cli.StringSliceFlag{
						Name:        "var",
						Usage:       "template variable",
						Destination: &hideArgs.Vars,
					},
					&cli.StringSliceFlag{
						Name:        "var-file",
						Usage:       "template variable name and file path",
						Destination: &hideArgs.VarFiles,
					},
					&cli.BoolFlag{
						Name:        "dry-run",
						Usage:       "output a comment to standard error output instead of posting to GitHub",
						Destination: &hideArgs.DryRun,
					},
					&cli.BoolFlag{
						Name:        "skip-no-token",
						Aliases:     []string{"n"},
						Usage:       "works like dry-run if the GitHub Access Token isn't set",
						Sources:     cli.EnvVars("GH_COMMENT_SKIP_NO_TOKEN", "GITHUB_COMMENT_SKIP_NO_TOKEN"),
						Destination: &hideArgs.SkipNoToken,
					},
					&cli.BoolFlag{
						Name:        "silent",
						Aliases:     []string{"s"},
						Usage:       "suppress the output of dry-run and skip-no-token",
						Destination: &hideArgs.Silent,
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "log level",
				Sources:     cli.EnvVars("GH_COMMENT_LOG_LEVEL"),
				Destination: &globalFlags.LogLevel,
			},
		},
	}).Run(ctx, env.Args)
}

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
