package template

import (
	"bytes"
	"html/template"
)

type Renderer struct {
	Getenv func(string) string
}

func (renderer Renderer) Render(tpl string, params interface{}) (string, error) {
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
