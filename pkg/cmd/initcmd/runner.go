package initcmd

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

func (runner *Runner) Command() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "scaffold a configuration file if it doesn't exist",
		Action: runner.initAction,
	}
}
