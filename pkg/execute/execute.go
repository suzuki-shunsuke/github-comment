package execute

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/mattn/go-colorable"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

type Executor struct {
	Stdout io.Writer
	Stderr io.Writer
	Env    []string
}

type Result struct {
	ExitCode       int
	Cmd            string
	Stdout         string
	Stderr         string
	CombinedOutput string
}

type Params struct {
	Cmd   string
	Args  []string
	Stdin io.Reader
}

func (executor Executor) Run(ctx context.Context, params Params) (Result, error) {
	cmd := exec.Command(params.Cmd, params.Args...) //nolint:gosec
	cmd.Stdin = params.Stdin
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	combinedOutput := &bytes.Buffer{}
	uncolorizedStdout := colorable.NewNonColorable(stdout)
	uncolorizedStderr := colorable.NewNonColorable(stderr)
	uncolorizedCombinedOutput := colorable.NewNonColorable(combinedOutput)
	cmd.Stdout = io.MultiWriter(executor.Stdout, uncolorizedStdout, uncolorizedCombinedOutput)
	cmd.Stderr = io.MultiWriter(executor.Stderr, uncolorizedStderr, uncolorizedCombinedOutput)
	cmd.Env = executor.Env

	runner := timeout.NewRunner(0)
	err := runner.Run(ctx, cmd)
	ec := cmd.ProcessState.ExitCode()
	result := Result{
		ExitCode:       ec,
		Cmd:            cmd.String(),
		Stdout:         stdout.String(),
		Stderr:         stderr.String(),
		CombinedOutput: combinedOutput.String(),
	}
	if err == nil {
		return result, nil
	}
	return result, fmt.Errorf("run a command: %w", err)
}
