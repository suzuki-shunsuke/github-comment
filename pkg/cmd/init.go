package cmd

import (
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/fsys"
	"github.com/urfave/cli/v2"
)

// initAction is an entrypoint of the subcommand "init".
func (r *Runner) initAction(c *cli.Context) error {
	ctrl := api.InitController{
		Fsys: &fsys.Fsys{},
	}
	return ctrl.Run(c.Context) //nolint:wrapcheck
}
