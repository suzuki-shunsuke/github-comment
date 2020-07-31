package cmd

import (
	"context"
	"io"

	"github.com/suzuki-shunsuke/github-comment/pkg/constant"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner Runner) Run(ctx context.Context, args []string) error {
	postCommand := runner.postCommand()
	execCommand := runner.execCommand()
	initCommand := runner.initCommand()
	app := cli.App{
		Name:    "github-comment",
		Usage:   "post a comment to GitHub",
		Version: constant.Version,
		Commands: []*cli.Command{
			&postCommand,
			&execCommand,
			&initCommand,
		},
	}
	return app.RunContext(ctx, args)
}
