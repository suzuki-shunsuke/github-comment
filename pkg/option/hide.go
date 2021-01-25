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
	if opts.PRNumber == 0 {
		return errors.New("pull request or issue number is required")
	}
	if opts.HideKey == "" {
		return errors.New("hide-key is required")
	}
	if err := validate(opts.Options); err != nil {
		return err
	}
	return nil
}
