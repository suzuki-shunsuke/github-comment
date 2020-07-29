package template

import (
	"bytes"
	"html/template"
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
	}).Parse(tpl)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, params); err != nil {
		return "", err
	}
	return buf.String(), nil
}
