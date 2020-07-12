package main

import (
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/github-comment/pkg/cmd"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
)

func main() {
	runner := cmd.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := runner.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ecerror.GetExitCode(err))
	}
}
