package cmd

import (
	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/fsys"
	"github.com/urfave/cli/v2"
)

func (runner Runner) initCommand() cli.Command {
	return cli.Command{
		Name:   "init",
		Usage:  "scaffold a configuration file if it doesn't exist",
		Action: runner.initAction,
	}
}

// initAction is an entrypoint of the subcommand "init".
func (runner Runner) initAction(c *cli.Context) error {
	ctrl := api.InitController{
		Fsys: fsys.Fsys{},
	}
	return ctrl.Run(c.Context)
}
