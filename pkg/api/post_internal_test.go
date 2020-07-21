package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostController_readTemplateFromStdin(t *testing.T) {
	data := []struct {
		title string
		ctrl  PostController
		exp   string
		isErr bool
	}{
		{
			title: "no standard input",
			ctrl: PostController{
				HasStdin: func() bool {
					return false
				},
			},
		},
		{
			title: "standard input",
			ctrl: PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
			},
			exp: "hello",
		},
	}
	for _, d := range data {
		d := d
		tpl, err := d.ctrl.readTemplateFromStdin()
		if d.isErr {
			require.NotNil(t, err)
			return
		}
		require.Nil(t, err)
		require.Equal(t, d.exp, tpl)
	}
}
