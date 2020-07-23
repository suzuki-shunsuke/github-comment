package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

func TestExecController_getExecConfig(t *testing.T) {
	data := []struct {
		title       string
		ctrl        ExecController
		opts        option.ExecOptions
		execConfigs []config.ExecConfig
		env         Env
		exp         config.ExecConfig
		f           bool
		isErr       bool
	}{
		{
			title: "no exec configs",
			ctrl:  ExecController{},
			opts: option.ExecOptions{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				Token:    "xxx",
				PRNumber: 1,
			},
			execConfigs: []config.ExecConfig{},
			exp:         config.ExecConfig{},
		},
		{
			title: "no exec config matches",
			ctrl: ExecController{
				Expr: expr.Expr{},
			},
			opts: option.ExecOptions{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				Token:    "xxx",
				PRNumber: 1,
			},
			execConfigs: []config.ExecConfig{
				{
					When: "false",
				},
			},
			exp: config.ExecConfig{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			execConfig, f, err := d.ctrl.getExecConfig(d.opts, d.execConfigs, d.env)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, execConfig)
			require.Equal(t, d.f, f)
		})
	}
}
