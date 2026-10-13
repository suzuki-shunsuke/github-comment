package option

import (
	"errors"
)

type DeleteOptions struct {
	Options
	DeleteKey     string
	Condition     string
	StdinTemplate bool
}

func ValidateDelete(opts *DeleteOptions) error {
	if opts.PRNumber <= 0 {
		return errors.New("pull request or issue number is required")
	}
	if opts.DeleteKey == "" && opts.Condition == "" {
		return errors.New("delete-key or condition are required")
	}
	return validate(&opts.Options)
}
