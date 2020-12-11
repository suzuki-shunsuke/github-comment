package option

import (
	"errors"
)

type PostOptions struct {
	PRNumber           int
	Org                string
	Repo               string
	Token              string
	SHA1               string
	Template           string
	TemplateForTooLong string
	TemplateKey        string
	ConfigPath         string
	Vars               map[string]string
	DryRun             bool
	SkipNoToken        bool
	Silent             bool
	StdinTemplate      bool
}

func ValidatePost(opts PostOptions) error {
	if opts.Org == "" {
		return errors.New("org is required")
	}
	if opts.Repo == "" {
		return errors.New("repo is required")
	}
	if opts.Token == "" && !opts.SkipNoToken {
		return errors.New("token is required")
	}
	if opts.Template == "" && opts.TemplateKey == "" {
		return errors.New("template or template-key are required")
	}
	if opts.SHA1 == "" && opts.PRNumber == -1 {
		return errors.New("sha1 or pr are required")
	}
	return nil
}
