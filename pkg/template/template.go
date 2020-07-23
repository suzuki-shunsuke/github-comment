package template

import (
	"bytes"
	"html/template"
)

type Renderer struct {
	Getenv func(string) string
}

type Params struct {
	// PRNumber is the pull request number where the comment is posted
	PRNumber int
	// Org is the GitHub Organization or User name
	Org string
	// Repo is the GitHub Repository name
	Repo string
	// SHA1 is the commit SHA1
	SHA1        string
	TemplateKey string
}

func (renderer Renderer) Render(tpl string, params Params) (string, error) {
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
