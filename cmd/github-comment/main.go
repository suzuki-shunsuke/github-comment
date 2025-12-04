package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/cmd"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	os.Exit(core())
}

func core() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	logLevelVar := &slog.LevelVar{}
	logger := slogutil.New(&slogutil.InputNew{
		Name:    "github-comment",
		Version: version,
		Out:     os.Stderr,
		Level:   logLevelVar,
	})
	runner := cmd.Runner{
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Logger:      logger,
		LogLevelVar: logLevelVar,
		LDFlags: &cmd.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
	if err := runner.Run(ctx, os.Args); err != nil {
		slogerr.WithError(logger, err).Error("github-comment failed")
		return ecerror.GetExitCode(err)
	}
	return 0
}
