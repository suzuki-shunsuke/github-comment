package platform

import (
	"fmt"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/option"
)

type Drone struct {
	getEnv func(string) string
}

func (drone Drone) Match() bool {
	return drone.getEnv("DRONE") != ""
}

func (drone Drone) ComplementPost(opts *option.PostOptions) error {
	if opts.Org == "" {
		opts.Org = drone.getEnv("DRONE_REPO_OWNER")
	}
	if opts.Repo == "" {
		opts.Repo = drone.getEnv("DRONE_REPO_NAME")
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	pr := drone.getEnv("DRONE_PULL_REQUEST")
	if pr == "" {
		opts.SHA1 = drone.getEnv("DRONE_COMMIT_SHA1")
		return nil
	}
	if b, err := strconv.Atoi(pr); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("DRONE_PULL_REQUEST is invalid. It is failed to parse DRONE_PULL_REQUEST as an integer: %w", err)
	}
	return nil
}

func (drone Drone) ComplementExec(opts *option.ExecOptions) error {
	if opts.Org == "" {
		opts.Org = drone.getEnv("DRONE_REPO_OWNER")
	}
	if opts.Repo == "" {
		opts.Repo = drone.getEnv("DRONE_REPO_NAME")
	}
	if opts.SHA1 != "" || opts.PRNumber != 0 {
		return nil
	}
	pr := drone.getEnv("DRONE_PULL_REQUEST")
	if pr == "" {
		opts.SHA1 = drone.getEnv("DRONE_COMMIT_SHA1")
		return nil
	}
	if b, err := strconv.Atoi(pr); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("DRONE_PULL_REQUEST is invalid. It is failed to parse DRONE_PULL_REQUEST as an integer: %w", err)
	}
	return nil
}
