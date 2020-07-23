package main

import (
	"context"
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/github-comment/pkg/cmd"
	"github.com/suzuki-shunsuke/github-comment/pkg/signal"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go signal.Handle(os.Stderr, cancel)
	runner := cmd.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := runner.Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ecerror.GetExitCode(err))
	}
}
