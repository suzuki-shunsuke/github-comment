package platform

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type gitHubActionsPayload struct {
	PullRequest struct {
		Number int `json:"number"`
	} `json:"pull_request"`
}

type GitHubActions struct {
	read   func(string) (io.ReadCloser, error)
	getEnv func(string) string
}

func (ac GitHubActions) Match() bool {
	return ac.getEnv("GITHUB_ACTIONS") != ""
}

func (ac GitHubActions) getPRNumberFromPayload(body io.Reader) (int, error) {
	p := gitHubActionsPayload{}
	if err := json.NewDecoder(body).Decode(&p); err != nil {
		return 0, err
	}
	return p.PullRequest.Number, nil
}

func (ac GitHubActions) ComplementPost(opts *option.PostOptions) error {
	a := strings.SplitN(ac.getEnv("GITHUB_REPOSITORY"), "/", 2)
	if opts.Org == "" {
		opts.Org = a[0]
	}
	if opts.Repo == "" {
		if len(a) == 2 {
			opts.Repo = a[1]
		}
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	f, err := ac.read(ac.getEnv("GITHUB_EVENT_PATH"))
	if err != nil {
		return err
	}
	defer f.Close()
	pr, err := ac.getPRNumberFromPayload(f)
	if err != nil {
		return err
	}
	if pr == 0 {
		opts.SHA1 = ac.getEnv("GITHUB_SHA")
		return nil
	}
	opts.PRNumber = pr
	return nil
}

func (ac GitHubActions) ComplementExec(opts *option.ExecOptions) error {
	a := strings.SplitN(ac.getEnv("GITHUB_REPOSITORY"), "/", 2)
	if opts.Org == "" {
		opts.Org = a[0]
	}
	if opts.Repo == "" {
		if len(a) == 2 {
			opts.Repo = a[1]
		}
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	f, err := ac.read(ac.getEnv("GITHUB_EVENT_PATH"))
	if err != nil {
		return err
	}
	defer f.Close()
	pr, err := ac.getPRNumberFromPayload(f)
	if err != nil {
		return err
	}
	if pr == 0 {
		opts.SHA1 = ac.getEnv("GITHUB_SHA")
		return nil
	}
	opts.PRNumber = pr
	return nil
}
