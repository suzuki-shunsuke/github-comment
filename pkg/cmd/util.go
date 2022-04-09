package cmd

import (
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
)

func getPlatformParam(comp *config.Complement) *platform.Param {
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
