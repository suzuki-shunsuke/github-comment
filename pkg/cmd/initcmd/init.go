package initcmd

import (
	"github.com/suzuki-shunsuke/github-comment/pkg/controller/initcmd"
	"github.com/suzuki-shunsuke/github-comment/pkg/fsys"
	"github.com/urfave/cli/v2"
)

// initAction is an entrypoint of the subcommand "init".
func (runner *Runner) initAction(c *cli.Context) error {
	ctrl := &initcmd.Controller{
		Fsys: &fsys.Fsys{},
	}
	return ctrl.Run(c.Context) //nolint:wrapcheck
}
