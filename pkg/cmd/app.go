package cmd

import (
	"io"

	"github.com/suzuki-shunsuke/github-comment/pkg/constant"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner Runner) Run(args []string) error {
	postCommand := runner.postCommand()
	execCommand := runner.execCommand()
	app := cli.App{
		Version: constant.Version,
		Commands: []*cli.Command{
			&postCommand,
			&execCommand,
		},
	}
	return app.Run(args)
}
