package option

import (
	"errors"
)

type ExecOptions struct {
	Options
	Args        []string
	SkipComment bool
}

func (opts *ExecOptions) Skipped() bool {
	if opts.SkipComment {
		return true
	}
	return opts.Token == "" && opts.SkipNoToken
}

func ValidateExec(opts *ExecOptions) error {
	if err := validate(&opts.Options); err != nil {
		return err
	}
	if opts.TemplateKey == "" {
		return errors.New("template-key is required")
	}
	if len(opts.Args) == 0 {
		return errors.New("command is required")
	}
	return nil
}
