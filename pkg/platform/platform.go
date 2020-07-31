package platform

import (
	"io"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type Platform interface {
	ComplementPost(opts *option.PostOptions) error
	ComplementExec(opts *option.ExecOptions) error
	Match() bool
}

func Get(getEnv func(string) string, read func(string) (io.ReadCloser, error)) Platform {
	platforms := []Platform{
		GitHubActions{
			read:   read,
			getEnv: getEnv,
		},
		Drone{
			getEnv: getEnv,
		},
		CircleCI{
			getEnv: getEnv,
		},
	}
	for _, platform := range platforms {
		if platform.Match() {
			return platform
		}
	}
	return nil
}
