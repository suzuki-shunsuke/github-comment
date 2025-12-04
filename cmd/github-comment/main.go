package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/cmd"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	if err := core(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ecerror.GetExitCode(err))
	}
}

func core() error {
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
	return runner.Run(ctx, os.Args) //nolint:wrapcheck
}
