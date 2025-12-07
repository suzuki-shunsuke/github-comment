package controller

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

func (c InitController) Run(_ context.Context) error {
	dst := "github-comment.yaml"
	if c.Fsys.Exist(dst) {
		return nil
	}
	return c.Fsys.Write(dst, []byte(strings.TrimSpace(string(cfgTemplate))+"\n")) //nolint:wrapcheck
}
