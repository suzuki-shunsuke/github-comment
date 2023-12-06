package api

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed config.yaml
var cfgTemplate []byte

type Fsys interface {
	Exist(path string) bool
	Write(path string, content []byte) error
}

type InitController struct {
	Fsys Fsys
}

func (ctrl InitController) Run(_ context.Context) error {
	dst := "github-comment.yaml"
	if ctrl.Fsys.Exist(dst) {
		return nil
	}
	return ctrl.Fsys.Write(dst, []byte(strings.TrimSpace(string(cfgTemplate))+"\n")) //nolint:wrapcheck
}
