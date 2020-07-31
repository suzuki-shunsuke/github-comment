package platform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type CircleCI struct {
	getEnv func(string) string
}

func (cc CircleCI) Match() bool {
	return cc.getEnv("CIRCLECI") != ""
}

func (cc CircleCI) ComplementPost(opts *option.PostOptions) error {
	if opts.Org == "" {
		opts.Org = cc.getEnv("CIRCLE_PROJECT_USERNAME")
	}
	if opts.Repo == "" {
		opts.Repo = cc.getEnv("CIRCLE_PROJECT_REPONAME")
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	pr := cc.getEnv("CIRCLE_PULL_REQUEST")
	if pr == "" {
		opts.SHA1 = cc.getEnv("CIRCLE_SHA1")
		return nil
	}
	a := strings.LastIndex(pr, "/")
	if a == -1 {
		return nil
	}
	prNum := pr[a+1:]
	if b, err := strconv.Atoi(prNum); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("failed to extract a pull request number from the environment variable CIRCLE_PULL_REQUEST: %w", err)
	}
	return nil
}

func (cc CircleCI) ComplementExec(opts *option.ExecOptions) error {
	if opts.Org == "" {
		opts.Org = cc.getEnv("CIRCLE_PROJECT_USERNAME")
	}
	if opts.Repo == "" {
		opts.Repo = cc.getEnv("CIRCLE_PROJECT_REPONAME")
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	pr := cc.getEnv("CIRCLE_PULL_REQUEST")
	if pr == "" {
		opts.SHA1 = cc.getEnv("CIRCLE_SHA1")
		return nil
	}
	a := strings.LastIndex(pr, "/")
	if a == -1 {
		return nil
	}
	prNum := pr[a+1:]
	if b, err := strconv.Atoi(prNum); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("failed to extract a pull request number from the environment variable CIRCLE_PULL_REQUEST: %w", err)
	}
	return nil
}
