package option

import (
	"errors"
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
	DryRun      bool
	SkipNoToken bool
}

func ValidateExec(opts ExecOptions) error {
	if opts.Org == "" {
		return errors.New("org is required")
	}
	if opts.Repo == "" {
		return errors.New("repo is required")
	}
	if opts.Token == "" && !opts.SkipNoToken {
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
