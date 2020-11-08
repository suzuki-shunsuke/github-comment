package platform

import (
	"fmt"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/go-ci-env/cienv"
)

type Platform struct {
	Platform cienv.Platform
}

func (pt Platform) ComplementPost(opts *option.PostOptions) error {
	if opts.Org == "" {
		opts.Org = pt.Platform.RepoOwner()
	}
	if opts.Repo == "" {
		opts.Repo = pt.Platform.RepoName()
	}
	if opts.SHA1 == "" {
		opts.SHA1 = pt.Platform.SHA()
	}
	if opts.PRNumber != 0 {
		return nil
	}
	pr, err := pt.Platform.PRNumber()
	if err != nil {
		return fmt.Errorf("get a pull request number from an environment variable: %w", err)
	}
	if pr > 0 {
		opts.PRNumber = pr
	}
	return nil
}

func (pt Platform) CI() string {
	if pt.Platform == nil {
		return ""
	}
	return pt.Platform.CI()
}

func (pt Platform) ComplementExec(opts *option.ExecOptions) error {
	if opts.Org == "" {
		opts.Org = pt.Platform.RepoOwner()
	}
	if opts.Repo == "" {
		opts.Repo = pt.Platform.RepoName()
	}
	if opts.SHA1 == "" {
		opts.SHA1 = pt.Platform.SHA()
	}
	if opts.PRNumber != 0 {
		return nil
	}
	pr, err := pt.Platform.PRNumber()
	if err != nil {
		return fmt.Errorf("get a pull request number from an environment variable: %w", err)
	}
	if pr > 0 {
		opts.PRNumber = pr
	}
	return nil
}

func Get() (Platform, bool) {
	pt := Platform{
		Platform: cienv.Get(),
	}
	if pt.Platform == nil {
		return Platform{}, false
	}
	return pt, true
}
