package platform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type CodeBuild struct {
	getEnv func(string) string
}

func (cb CodeBuild) Match() bool {
	return cb.getEnv("CODEBUILD_BUILD_ID") != ""
}

func (cb CodeBuild) ComplementPost(opts *option.PostOptions) error { //nolint:dupl
	url := cb.getEnv("CODEBUILD_SOURCE_REPO_URL")
	if opts.Org == "" {
		if strings.HasPrefix(url, "https://github.com") {
			a := strings.Split(url, "/")
			opts.Org = a[len(a)-2]
		}
	}
	if opts.Repo == "" {
		if strings.HasPrefix(url, "https://github.com") {
			a := strings.Split(url, "/")
			opts.Repo = strings.TrimSuffix(a[len(a)-1], ".git")
		}
	}
	if opts.SHA1 == "" {
		opts.SHA1 = cb.getEnv("CODEBUILD_RESOLVED_SOURCE_VERSION")
	}
	if opts.PRNumber != 0 {
		return nil
	}
	pr := cb.getEnv("CODEBUILD_SOURCE_VERSION")
	if !strings.HasPrefix(pr, "pr/") {
		return nil
	}
	i := strings.Index(pr, "/")
	if i == -1 {
		return nil
	}
	if b, err := strconv.Atoi(pr[i+1:]); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("CODEBUILD_SOURCE_VERSION is invalid. It is failed to parse DRONE_PULL_REQUEST as an integer: %w", err)
	}
	return nil
}

func (cb CodeBuild) ComplementExec(opts *option.ExecOptions) error { //nolint:dupl
	url := cb.getEnv("CODEBUILD_SOURCE_REPO_URL")
	if opts.Org == "" {
		if strings.HasPrefix(url, "https://github.com") {
			a := strings.Split(url, "/")
			opts.Org = a[len(a)-2]
		}
	}
	if opts.Repo == "" {
		if strings.HasPrefix(url, "https://github.com") {
			a := strings.Split(url, "/")
			opts.Repo = strings.TrimSuffix(a[len(a)-1], ".git")
		}
	}
	if opts.SHA1 == "" {
		opts.SHA1 = cb.getEnv("CODEBUILD_RESOLVED_SOURCE_VERSION")
	}
	if opts.PRNumber != 0 {
		return nil
	}
	pr := cb.getEnv("CODEBUILD_SOURCE_VERSION")
	if !strings.HasPrefix(pr, "pr/") {
		return nil
	}
	i := strings.Index(pr, "/")
	if i == -1 {
		return nil
	}
	if b, err := strconv.Atoi(pr[i+1:]); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("CODEBUILD_SOURCE_VERSION is invalid. It is failed to parse DRONE_PULL_REQUEST as an integer: %w", err)
	}
	return nil
}
