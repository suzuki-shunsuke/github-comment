package hide

import (
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	stdio *domain.Stdio
	logE  *logrus.Entry
	osEnv osenv.OSEnv
}

func New(stdio *domain.Stdio, logE *logrus.Entry, osEnv osenv.OSEnv) *Runner {
	return &Runner{
		stdio: stdio,
		logE:  logE,
		osEnv: osEnv,
	}
}

func (runner *Runner) Command() *cli.Command { //nolint:funlen
	return &cli.Command{
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
	}
}
