package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
)

func TestExecController_getExecConfig(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title       string
		ctrl        *ExecController
		execConfigs []*config.ExecConfig
		cmtParams   *ExecCommentParams
		exp         *config.ExecConfig
		f           bool
		isErr       bool
	}{
		{
			title:       "no exec configs",
			ctrl:        &ExecController{},
			execConfigs: []*config.ExecConfig{},
			exp:         nil,
		},
		{
			title: "no exec config matches",
			ctrl: &ExecController{
				Expr: &expr.Expr{},
			},
			cmtParams: &ExecCommentParams{},
			execConfigs: []*config.ExecConfig{
				{
					When: "false",
				},
			},
			exp: nil,
		},
		{
			title: "first matched config is returned",
			ctrl: &ExecController{
				Expr: &expr.Expr{},
			},
			cmtParams: &ExecCommentParams{},
			execConfigs: []*config.ExecConfig{
				{
					When:        "true",
					Template:    "foo",
					DontComment: true,
				},
				{
					When:     "true",
					Template: "bar",
				},
			},
			exp: &config.ExecConfig{
				When:        "true",
				Template:    "foo",
				DontComment: true,
			},
			f: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			execConfig, f, err := d.ctrl.getExecConfig(d.execConfigs, d.cmtParams)
			if d.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, d.exp, execConfig)
			require.Equal(t, d.f, f)
		})
	}
}
