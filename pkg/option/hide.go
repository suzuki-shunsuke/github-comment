package option

import (
	"errors"
)

type HideOptions struct {
	Options
	HideKey       string
	StdinTemplate bool
}

func ValidateHide(opts HideOptions) error {
	if err := validate(opts.Options); err != nil {
		return err
	}
	if opts.HideKey == "" {
		return errors.New("hide-key is required")
	}
	return nil
}
