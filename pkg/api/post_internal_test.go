package api

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
)

func TestPostController_getCommentParams(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		ctrl  *PostController
		exp   *github.Comment
		isErr bool
		opts  *option.PostOptions
	}{
		{
			title: "if there is a standard input, treat it as the template",
			ctrl: &PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(_ string) string {
					return ""
				},
				Renderer: &template.Renderer{},
				Config:   &config.Config{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:      "suzuki-shunsuke",
					Repo:     "github-comment",
					Token:    "xxx",
					PRNumber: 1,
				},
				StdinTemplate: true,
			},
			exp: &github.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Vars:     map[string]any{},
			},
		},
		{
			title: "if template is passed as argument, standard input is ignored",
			ctrl: &PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(_ string) string {
					return ""
				},
				Renderer: &template.Renderer{},
				Config:   &config.Config{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:      "suzuki-shunsuke",
					Repo:     "github-comment",
					Token:    "xxx",
					PRNumber: 1,
					Template: "foo",
				},
			},
			exp: &github.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Vars:     map[string]any{},
			},
		},
		{
			title: "read template from config",
			ctrl: &PostController{
				HasStdin: func() bool {
					return false
				},
				Getenv: func(_ string) string {
					return ""
				},
				Config: &config.Config{
					Post: map[string]*config.PostConfig{
						"default": {
							Template: "hello",
						},
					},
				},
				Renderer: &template.Renderer{
					Getenv: func(_ string) string {
						return ""
					},
				},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:         "suzuki-shunsuke",
					Repo:        "github-comment",
					Token:       "xxx",
					TemplateKey: "default",
					PRNumber:    1,
				},
			},
			exp: &github.Comment{
				Org:         "suzuki-shunsuke",
				Repo:        "github-comment",
				PRNumber:    1,
				TemplateKey: "default",
				Vars:        map[string]any{},
			},
		},
		{
			title: "template is rendered properly",
			ctrl: &PostController{
				HasStdin: func() bool {
					return false
				},
				Getenv: func(_ string) string {
					return ""
				},
				Renderer: &template.Renderer{
					Getenv: func(k string) string {
						if k == "FOO" {
							return "BAR"
						}
						return ""
					},
				},
				Config: &config.Config{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Org:      "suzuki-shunsuke",
					Repo:     "github-comment",
					Token:    "xxx",
					PRNumber: 1,
					Template: `{{.Org}} {{.Repo}} {{.PRNumber}}`,
				},
			},
			exp: &github.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Vars:     map[string]any{},
			},
		},
		{
			title: "config.base",
			ctrl: &PostController{
				HasStdin: func() bool {
					return true
				},
				Stdin: strings.NewReader("hello"),
				Getenv: func(_ string) string {
					return ""
				},
				Config: &config.Config{
					Base: &config.Base{
						Org:  "suzuki-shunsuke",
						Repo: "github-comment",
					},
				},
				Renderer: &template.Renderer{},
			},
			opts: &option.PostOptions{
				Options: option.Options{
					Token:    "xxx",
					PRNumber: 1,
				},
				StdinTemplate: true,
			},
			exp: &github.Comment{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				PRNumber: 1,
				Vars:     map[string]any{},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			cmt, err := d.ctrl.getCommentParams(ctx, slog.Default(), d.opts)
			if d.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			cmt.Body = ""
			cmt.BodyForTooLong = ""
			require.Equal(t, d.exp, cmt)
		})
	}
}

func TestPostController_readTemplateFromStdin(t *testing.T) {
	t.Parallel()
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
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			tpl, err := d.ctrl.readTemplateFromStdin()
			if d.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, d.exp, tpl)
		})
	}
}
