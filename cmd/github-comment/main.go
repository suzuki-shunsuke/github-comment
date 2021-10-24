package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/suzuki-shunsuke/github-comment/pkg/cmd"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
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
	runner := cmd.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &cmd.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
	return runner.Run(ctx, os.Args) //nolint:wrapcheck
}
