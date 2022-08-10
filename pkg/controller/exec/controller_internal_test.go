package exec

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
)

func TestController_getExecConfig(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title       string
		ctrl        *Controller
		execConfigs []*config.ExecConfig
		cmtParams   *CommentParams
		exp         *config.ExecConfig
		f           bool
		isErr       bool
	}{
		{
			title:       "no exec configs",
			ctrl:        &Controller{},
			execConfigs: []*config.ExecConfig{},
			exp:         nil,
		},
		{
			title: "no exec config matches",
			ctrl: &Controller{
				Expr: &expr.Expr{},
			},
			execConfigs: []*config.ExecConfig{
				{
					When: "false",
				},
			},
			exp: nil,
		},
		{
			title: "first matched config is returned",
			ctrl: &Controller{
				Expr: &expr.Expr{},
			},
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
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			execConfig, f, err := d.ctrl.getExecConfig(d.execConfigs, d.cmtParams)
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
