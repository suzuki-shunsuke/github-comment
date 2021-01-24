package option

import (
	"errors"
)

type HideOptions struct {
	Options
	StdinTemplate bool
}

func ValidateHide(opts HideOptions) error {
	if err := validate(opts.Options); err != nil {
		return err
	}
	if opts.Template == "" && opts.TemplateKey == "" {
		return errors.New("template or template-key are required")
	}
	return nil
}
