package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/cmd/exec"
	"github.com/suzuki-shunsuke/github-comment/pkg/cmd/hide"
	"github.com/suzuki-shunsuke/github-comment/pkg/cmd/initcmd"
	"github.com/suzuki-shunsuke/github-comment/pkg/cmd/post"
	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	stdio   *domain.Stdio
	ldFlags *LDFlags
	logE    *logrus.Entry
	osEnv   osenv.OSEnv
}

func New(stdio *domain.Stdio, logE *logrus.Entry, osEnv osenv.OSEnv, ldFlags *LDFlags) *Runner {
	return &Runner{
		stdio:   stdio,
		logE:    logE,
		osEnv:   osEnv,
		ldFlags: ldFlags,
	}
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (flags *LDFlags) AppVersion() string {
	return flags.Version + " (" + flags.Commit + ")"
}

type command interface {
	Command() *cli.Command
}

func (runner *Runner) Run(ctx context.Context, args []string) error {
	app := cli.App{
		Name:    "github-comment",
		Usage:   "post a comment to GitHub",
		Version: runner.ldFlags.AppVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				EnvVars: []string{"GITHUB_COMMENT_LOG_LEVEL"},
			},
		},
	}
	cmds := []command{
		post.New(runner.stdio, runner.logE, runner.osEnv),
		exec.New(runner.stdio, runner.logE, runner.osEnv),
		initcmd.New(runner.stdio, runner.logE, runner.osEnv),
		hide.New(runner.stdio, runner.logE, runner.osEnv),
	}
	app.Commands = make([]*cli.Command, len(cmds))
	for i, cmd := range cmds {
		app.Commands[i] = cmd.Command()
	}
	runner.osEnv = osenv.New()
	return app.RunContext(ctx, args) //nolint:wrapcheck
}
