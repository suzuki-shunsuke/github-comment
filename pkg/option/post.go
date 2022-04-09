package option

import (
	"errors"
)

type Options struct {
	PRNumber           int
	Org                string
	Repo               string
	Token              string
	SHA1               string
	Template           string
	TemplateForTooLong string
	TemplateKey        string
	ConfigPath         string
	HideOldComment     string
	LogLevel           string
	Vars               map[string]string
	EmbeddedVarNames   []string
	DryRun             bool
	SkipNoToken        bool
	Silent             bool
}

func validate(opts *Options) error {
	if opts.Org == "" {
		return errors.New("org is required")
	}
	if opts.Repo == "" {
		return errors.New("repo is required")
	}
	if opts.Token == "" && !opts.SkipNoToken {
		return errors.New("token is required")
	}
	if opts.SHA1 == "" && opts.PRNumber <= 0 {
		return errors.New("sha1 or pr are required")
	}
	return nil
}

type PostOptions struct {
	Options
	StdinTemplate bool
}

func ValidatePost(opts *PostOptions) error {
	if err := validate(&opts.Options); err != nil {
		return err
	}
	if opts.Template == "" && opts.TemplateKey == "" {
		return errors.New("template or template-key are required")
	}
	return nil
}
