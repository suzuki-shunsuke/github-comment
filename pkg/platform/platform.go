package platform

import (
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
)

type Platform struct {
	platform cienv.Platform
}

func (pt *Platform) getRepoOrg() (string, error) { //nolint:unparam
	if pt.platform != nil {
		if org := pt.platform.RepoOwner(); org != "" {
			return org, nil
		}
	}
	return "", nil
}

func (pt *Platform) getRepoName() (string, error) { //nolint:unparam
	if pt.platform != nil {
		if repo := pt.platform.RepoName(); repo != "" {
			return repo, nil
		}
	}
	return "", nil
}

func (pt *Platform) getSHA1() (string, error) { //nolint:unparam
	if pt.platform != nil {
		if sha1 := pt.platform.SHA(); sha1 != "" {
			return sha1, nil
		}
	}
	return "", nil
}

func (pt *Platform) getPRNumber() (int, error) {
	if pt.platform != nil {
		pr, err := pt.platform.PRNumber()
		if err != nil {
			return 0, fmt.Errorf("get a pull request number from an environment variable: %w", err)
		}
		if pr > 0 {
			return pr, nil
		}
	}

	if prS := os.Getenv("CI_INFO_PR_NUMBER"); prS != "" {
		a, err := strconv.Atoi(prS)
		if err != nil {
			return 0, fmt.Errorf("get a pull request number from an environment variable: %w", err)
		}
		if a > 0 {
			return a, nil
		}
	}
	return 0, nil
}

func (pt *Platform) complement(opts *option.Options) error {
	if opts.Org == "" {
		org, err := pt.getRepoOrg()
		if err != nil {
			return err
		}
		opts.Org = org
	}
	if opts.Repo == "" {
		repo, err := pt.getRepoName()
		if err != nil {
			return err
		}
		opts.Repo = repo
	}
	if opts.SHA1 == "" {
		sha1, err := pt.getSHA1()
		if err != nil {
			return err
		}
		opts.SHA1 = sha1
	}
	if opts.PRNumber > 0 {
		return nil
	}
	pr, err := pt.getPRNumber()
	if err != nil {
		return err
	}
	opts.PRNumber = pr

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
	return pt.platform.ID()
}

func (pt *Platform) ComplementExec(opts *option.ExecOptions) error {
	return pt.complement(&opts.Options)
}

func Get() *Platform {
	cienv.Add(func(param *cienv.Param) cienv.Platform {
		return NewGoogleCloudBuild(param)
	})
	return &Platform{
		platform: cienv.Get(nil),
	}
}
