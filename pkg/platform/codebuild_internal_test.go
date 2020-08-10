package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

func TestCodeBuild_Match(t *testing.T) {
	data := []struct {
		title  string
		getEnv func(string) string
		exp    bool
	}{
		{
			title: "match",
			getEnv: func(k string) string {
				if k == "CODEBUILD_BUILD_ID" {
					return "xxx"
				}
				return ""
			},
			exp: true,
		},
		{
			title: "doesn't match",
			getEnv: func(k string) string {
				return ""
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			cb := CodeBuild{
				getEnv: d.getEnv,
			}
			if d.exp {
				require.True(t, cb.Match())
				return
			}
			require.False(t, cb.Match())
		})
	}
}

func TestCodeBuild_ComplementPost(t *testing.T) {
	data := []struct {
		title  string
		getEnv func(string) string
		opts   *option.PostOptions
		exp    *option.PostOptions
		isErr  bool
	}{
		{
			title: "normal",
			getEnv: func(k string) string {
				switch k {
				case "CODEBUILD_BUILD_ID":
					return "xxx"
				case "CODEBUILD_SOURCE_REPO_URL":
					return "https://github.com/suzuki-shunsuke/github-comment.git"
				case "CODEBUILD_RESOLVED_SOURCE_VERSION":
					return "8cf666d2c390ed7fc4030c4fe7bc1b3f4903730d"
				case "CODEBUILD_SOURCE_VERSION":
					return "pr/10"
				}
				return ""
			},
			opts: &option.PostOptions{},
			exp: &option.PostOptions{
				Org:      "suzuki-shunsuke",
				Repo:     "github-comment",
				SHA1:     "8cf666d2c390ed7fc4030c4fe7bc1b3f4903730d",
				PRNumber: 10,
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			cb := CodeBuild{
				getEnv: d.getEnv,
			}
			err := cb.ComplementPost(d.opts)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.opts)
		})
	}
}
