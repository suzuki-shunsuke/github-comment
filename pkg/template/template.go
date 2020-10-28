package template

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
)

type Renderer struct {
	Getenv func(string) string
}

func addTemplates(tpl string, templates map[string]string) string {
	for k, v := range templates {
		tpl += `{{define "` + k + `"}}` + v + "{{end}}"
	}
	return tpl
}

func (renderer Renderer) Render(tpl string, templates map[string]string, params interface{}) (string, error) {
	tpl = addTemplates(tpl, templates)
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"Env": renderer.Getenv,
	}).Funcs(sprig.FuncMap()).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, params); err != nil {
		return "", fmt.Errorf("render a template with params: %w", err)
	}
	return buf.String(), nil
}
