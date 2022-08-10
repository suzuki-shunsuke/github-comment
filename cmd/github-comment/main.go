package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/cmd"
	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
	"github.com/suzuki-shunsuke/github-comment/pkg/log"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

type HasExitCode interface {
	ExitCode() int
}

func main() {
	logE := log.New(version)
	if err := core(logE); err != nil {
		var hasExitCode HasExitCode
		if errors.As(err, &hasExitCode) {
			code := hasExitCode.ExitCode()
			logerr.WithError(logE.WithField("exit_code", code), err).Debug("command failed")
			os.Exit(code)
		}
		logerr.WithError(logE, err).Fatal("aqua failed")
	}
}

func core(logE *logrus.Entry) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	runner := cmd.New(
		&domain.Stdio{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
		logE,
		osenv.New(),
		&cmd.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	)
	return runner.Run(ctx, os.Args) //nolint:wrapcheck
}
