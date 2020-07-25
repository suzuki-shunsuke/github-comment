package api

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
)

type mockReader struct {
	cfg config.Config
	err error
}

func (m mockReader) FindAndRead(cfgPath, wd string) (config.Config, error) {
	return m.cfg, m.err
}

func TestPostController_getCommentParams(t *testing.T) {
	data := []struct {
		title string
		ctrl  PostController
		exp   comment.Comment
		isErr bool
		opts  option.PostOptions
	}{
		{
			title: "if there is a standard input, treat it as the template",
			ctrl: PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(k string) string {
					return ""
				},
				Reader:   mockReader{},
				Renderer: template.Renderer{},
			},
			opts: option.PostOptions{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				Token:    "xxx",
				PRNumber: 1,
			},
			exp: comment.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Body:     "hello",
			},
		},
		{
			title: "if template is passed as argument, standard input is ignored",
			ctrl: PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(k string) string {
					return ""
				},
				Reader:   mockReader{},
				Renderer: template.Renderer{},
			},
			opts: option.PostOptions{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				Token:    "xxx",
				PRNumber: 1,
				Template: "foo",
			},
			exp: comment.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Body:     "foo",
			},
		},
		{
			title: "read template from config",
			ctrl: PostController{
				HasStdin: func() bool {
					return false
				},
				Getenv: func(k string) string {
					return ""
				},
				Reader: mockReader{
					cfg: config.Config{
						Post: map[string]string{
							"default": "hello",
						},
					},
				},
				Renderer: template.Renderer{
					Getenv: func(k string) string {
						return ""
					},
				},
			},
			opts: option.PostOptions{
				Org:         "suzuki-shunsuke",
				Repo:        "github-comment",
				Token:       "xxx",
				TemplateKey: "default",
				PRNumber:    1,
			},
			exp: comment.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Body:     "hello",
			},
		},
		{
			title: "template is rendered properly",
			ctrl: PostController{
				HasStdin: func() bool {
					return false
				},
				Getenv: func(k string) string {
					return ""
				},
				Reader: mockReader{},
				Renderer: template.Renderer{
					Getenv: func(k string) string {
						if k == "FOO" {
							return "BAR"
						}
						return ""
					},
				},
			},
			opts: option.PostOptions{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				Token:    "xxx",
				PRNumber: 1,
				Template: `{{Env "FOO"}} {{.Org}} {{.Repo}} {{.PRNumber}}`,
			},
			exp: comment.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Body:     "BAR suzuki-shunsuke github-comment 1",
			},
		},
	}
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			cmt, err := d.ctrl.getCommentParams(ctx, d.opts)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, cmt)
		})
	}
}

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
		t.Run(d.title, func(t *testing.T) {
			tpl, err := d.ctrl.readTemplateFromStdin()
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, tpl)
		})
	}
}
