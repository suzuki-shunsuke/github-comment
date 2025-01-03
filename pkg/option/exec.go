package option

import (
	"errors"
)

type ExecOptions struct {
	Options
	Args            []string
	SkipComment     bool
	Outputs         []*Output
	UpdateCondition string
}

type Output struct {
	File   string
	GitHub bool
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
