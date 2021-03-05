package platform

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/go-ci-env/cienv"
)

type Platform struct {
	platform cienv.Platform
	compl    Complement
}

func (pt *Platform) getRepoOrg() (string, error) {
	if pt.platform != nil {
		if org := pt.platform.RepoOwner(); org != "" {
			return org, nil
		}
	}
	return complement(pt.compl.Org)
}

func (pt *Platform) getRepoName() (string, error) {
	if pt.platform != nil {
		if repo := pt.platform.RepoName(); repo != "" {
			return repo, nil
		}
	}
	return complement(pt.compl.Repo)
}

func (pt *Platform) getSHA1() (string, error) {
	if pt.platform != nil {
		if sha1 := pt.platform.SHA(); sha1 != "" {
			return sha1, nil
		}
	}
	return complement(pt.compl.SHA1)
}

func (pt *Platform) getPRNumber() (int, error) {
	if pt.platform != nil {
		pr, err := pt.platform.PRNumber()
		if err != nil {
			return 0, fmt.Errorf("get a pull request number from an environment variable: %w", err)
		}
		if pr != 0 {
			return pr, nil
		}
	}

	if prS := os.Getenv("CI_INFO_PR_NUMBER"); prS != "" {
		a, err := strconv.Atoi(prS)
		if err != nil {
			return 0, fmt.Errorf("get a pull request number from an environment variable: %w", err)
		}
		if a != 0 {
			return a, nil
		}
	}
	prS, err := complement(pt.compl.PR)
	if err != nil {
		return 0, err
	}
	if prS != "" {
		a, err := strconv.Atoi(prS)
		if err != nil {
			return 0, fmt.Errorf("get a pull request number from an environment variable: %w", err)
		}
		return a, nil
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
	if opts.PRNumber != 0 {
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
	return pt.platform.CI()
}

func (pt *Platform) ComplementExec(opts *option.ExecOptions) error {
	return pt.complement(&opts.Options)
}

type Complement struct {
	PR   []string
	Org  []string
	Repo []string
	SHA1 []string
}

func complement(tpls []string) (string, error) {
	for _, tpl := range tpls {
		tmpl, err := template.New("_").Funcs(sprig.TxtFuncMap()).Parse(tpl)
		if err != nil {
			return "", fmt.Errorf("compile complement template: %w", err)
		}
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, nil); err != nil {
			logrus.WithFields(logrus.Fields{
				"template": tpl,
			}).WithError(err).Debug("failed to parse complement template")
			continue
		}
		if s := strings.TrimSpace(buf.String()); s != "" {
			return s, nil
		}
	}
	return "", nil
}

func Get(cpl Complement) Platform {
	return Platform{
		platform: cienv.Get(),
		compl:    cpl,
	}
}
