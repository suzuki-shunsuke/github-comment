package option

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ExecOptions struct {
	PRNumber    int
	Org         string
	Repo        string
	Token       string
	SHA1        string
	Template    string
	TemplateKey string
	ConfigPath  string
	Args        []string
	Vars        map[string]string
}

func ValidateExec(opts ExecOptions) error {
	if opts.Org == "" {
		return errors.New("org is required")
	}
	if opts.Repo == "" {
		return errors.New("repo is required")
	}
	if opts.Token == "" {
		return errors.New("token is required")
	}
	if opts.TemplateKey == "" {
		return errors.New("template-key is required")
	}
	if opts.SHA1 == "" && opts.PRNumber == -1 {
		return errors.New("sha1 or pr are required")
	}
	if len(opts.Args) == 0 {
		return errors.New("command is required")
	}
	return nil
}

func ComplementExec(opts *ExecOptions, getEnv func(string) string) error {
	if !IsCircleCI(getEnv) {
		return nil
	}
	if opts.Org == "" {
		opts.Org = getEnv("CIRCLE_PROJECT_USERNAME")
	}
	if opts.Repo == "" {
		opts.Repo = getEnv("CIRCLE_PROJECT_REPONAME")
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	pr := getEnv("CIRCLE_PULL_REQUEST")
	if pr == "" {
		opts.SHA1 = getEnv("CIRCLE_SHA1")
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
