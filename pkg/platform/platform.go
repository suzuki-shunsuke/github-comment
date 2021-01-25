package platform

import (
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/go-ci-env/cienv"
)

type Platform struct {
	platform cienv.Platform
}

func (pt *Platform) complement(opts *option.Options) error {
	if opts.Org == "" {
		opts.Org = pt.platform.RepoOwner()
	}
	if opts.Repo == "" {
		opts.Repo = pt.platform.RepoName()
	}
	if opts.SHA1 == "" {
		opts.SHA1 = pt.platform.SHA()
	}
	if opts.PRNumber != 0 {
		return nil
	}
	pr, err := pt.platform.PRNumber()
	if err != nil {
		return fmt.Errorf("get a pull request number from an environment variable: %w", err)
	}
	if pr > 0 {
		opts.PRNumber = pr
	} else if prS := os.Getenv("CI_INFO_PR_NUMBER"); prS != "" {
		a, err := strconv.Atoi(prS)
		if err != nil {
			return fmt.Errorf("get a pull request number from an environment variable: %w", err)
		}
		opts.PRNumber = a
	}
	return nil
}

func (pt *Platform) ComplementPost(opts *option.PostOptions) error {
	return pt.complement(&opts.Options)
}

func (pt *Platform) ComplementHide(opts *option.HideOptions) error {
	return pt.complement(&opts.Options)
}

func (pt *Platform) CI() string {
	if pt.platform == nil {
		return ""
	}
	return pt.platform.CI()
}

func (pt *Platform) ComplementExec(opts *option.ExecOptions) error {
	return pt.complement(&opts.Options)
}

func Get() (Platform, bool) {
	pt := Platform{
		platform: cienv.Get(),
	}
	if pt.platform == nil {
		return Platform{}, false
	}
	return pt, true
}
