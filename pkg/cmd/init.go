package cmd

import (
	"context"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/fsys"
	"github.com/urfave/cli/v3"
)

// initAction is an entrypoint of the subcommand "init".
func (r *Runner) initAction(ctx context.Context, _ *cli.Command) error {
	ctrl := api.InitController{
		Fsys: &fsys.Fsys{},
	}
	return ctrl.Run(ctx) //nolint:wrapcheck
}
