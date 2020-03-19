package cmd

import (
	"github.com/suzuki-shunsuke/github-comment/pkg/constant"
	"github.com/urfave/cli/v2"
)

func Run(args []string) error {
	app := &cli.App{
		Version: constant.Version,
		Commands: []*cli.Command{
			postCommand,
			execCommand,
		},
	}
	return app.Run(args)
}
