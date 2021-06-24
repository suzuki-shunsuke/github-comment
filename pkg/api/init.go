package api

import (
	"context"
	"strings"
)

const cfgTemplate = `---
# skip_no_token: true
# base:
#   org:
#   repo:
# vars:
#   foo: bar
#   zoo:
#     foo: hello
# templates:
#   header: "# {{.Org}}/{{.Repo}}"
# post:
#   default:
#     template: |
#       {{template "header" .}}
#       {{.Vars.foo}} {{.Vars.zoo.foo}}
#       {{.Org}} {{.Repo}} {{.PRNumber}} {{.SHA1}} {{.TemplateKey}}
#   hello:
#     template: hello
# exec:
#   hello:
#     - when: true
#       template: |
#         {{template "header" .}}
#         {{.Vars.foo}} {{.Vars.zoo.foo}}
#         {{.Org}} {{.Repo}} {{.PRNumber}} {{.SHA1}} {{.TemplateKey}}
#         exit code: {{.ExitCode}}
#
#         ` + "```" + `
#         $ {{.Command}}
#         ` + "```" + `
#
#         Stdout:
#
#         ` + "```" + `
#         {{.Stdout}}
#         ` + "```" + `
#
#         Stderr:
#
#         ` + "```" + `
#         {{.Stderr}}
#         ` + "```" + `
#
#         CombinedOutput:
#
#         ` + "```" + `
#         {{.CombinedOutput}}
#         ` + "```" + `
#       template_for_too_long: |
#         {{template "header" .}}
#         {{.Vars.foo}} {{.Vars.zoo.foo}}
#         {{.Org}} {{.Repo}} {{.PRNumber}} {{.SHA1}} {{.TemplateKey}}
#         exit code: {{.ExitCode}}
#
#         ` + "```" + `
#         $ {{.Command}}
#         ` + "```" + `
`

type Fsys interface {
	Exist(string) bool
	Write(path string, content []byte) error
}

type InitController struct {
	Fsys Fsys
}

func (ctrl InitController) Run(ctx context.Context) error {
	dst := ".github-comment.yml"
	if ctrl.Fsys.Exist(dst) {
		return nil
	}
	return ctrl.Fsys.Write(dst, []byte(strings.Trim(cfgTemplate, "\n"))) //nolint:wrapcheck
}
