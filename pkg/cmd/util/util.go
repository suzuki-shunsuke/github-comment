package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
	"github.com/suzuki-shunsuke/github-comment/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
)

func GetPlatformParam(comp *config.Complement) *platform.Param {
	if comp == nil {
		return &platform.Param{}
	}
	return &platform.Param{
		PRNumber:  comp.PR,
		RepoName:  comp.Repo,
		RepoOwner: comp.Org,
		SHA:       comp.SHA1,
		Vars:      comp.Vars,
	}
}

func ExistFile(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func ParseVarsFlag(varsSlice []string) (map[string]string, error) {
	vars := make(map[string]string, len(varsSlice))
	for _, v := range varsSlice {
		a := strings.SplitN(v, ":", 2) //nolint:gomnd
		if len(a) < 2 {                //nolint:gomnd
			return nil, errors.New("invalid var flag. The format should be '--var <key>:<value>")
		}
		vars[a[0]] = a[1]
	}
	return vars, nil
}

func ParseVarFilesFlag(varsSlice []string) (map[string]string, error) {
	vars := make(map[string]string, len(varsSlice))
	for _, v := range varsSlice {
		a := strings.SplitN(v, ":", 2) //nolint:gomnd
		if len(a) < 2 {                //nolint:gomnd
			return nil, errors.New("invalid var flag. The format should be '--var <key>:<value>")
		}
		name := a[0]
		filePath := a[1]
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read the value of the variable %s from the file %s: %w", name, filePath, err)
		}
		vars[name] = string(b)
	}
	return vars, nil
}

func GetGitHub(ctx context.Context, opts *option.Options, cfg *config.Config) (domain.GitHub, error) {
	if opts.DryRun {
		return &github.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}, nil
	}
	if opts.SkipNoToken && opts.Token == "" {
		return &github.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}, nil
	}

	return github.New(ctx, &github.ParamNew{ //nolint:wrapcheck
		Token:              opts.Token,
		GHEBaseURL:         cfg.GHEBaseURL,
		GHEGraphQLEndpoint: cfg.GHEGraphQLEndpoint,
	})
}
