package option

import (
	"errors"
)

type HideOptions struct {
	Options
	HideKey       string
	Condition     string
	StdinTemplate bool
}

func ValidateHide(opts HideOptions) error {
	if opts.PRNumber == 0 {
		return errors.New("pull request or issue number is required")
	}
	if opts.HideKey == "" || opts.Condition == "" {
		return errors.New("hide-key or condition are required")
	}
	return validate(opts.Options)
}
