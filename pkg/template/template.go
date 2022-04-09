package template

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig/v3"
)

type ParamGetTemplates struct {
	Templates      map[string]string
	CI             string
	JoinCommand    string
	CombinedOutput string
}

func GetTemplates(param *ParamGetTemplates) map[string]string {
	buildLinks := map[string]string{
		"circleci":       `[workflow](https://circleci.com/workflow-run/{{env "CIRCLE_WORKFLOW_ID" }}) [job]({{env "CIRCLE_BUILD_URL"}}) (job: {{env "CIRCLE_JOB"}})`,
		"codebuild":      `[Build link]({{env "CODEBUILD_BUILD_URL"}})`,
		"drone":          `[build]({{env "DRONE_BUILD_LINK"}}) [step]({{env "DRONE_BUILD_LINK"}}/{{env "DRONE_STAGE_NUMBER"}}/{{env "DRONE_STEP_NUMBER"}})`,
		"github-actions": `[Build link](https://github.com/{{env "GITHUB_REPOSITORY"}}/actions/runs/{{env "GITHUB_RUN_ID"}})`,
	}

	builtinTemplates := map[string]string{
		"status":                 `:{{if eq .ExitCode 0}}white_check_mark{{else}}x{{end}}:`,
		"join_command":           "```\n$ {{.JoinCommand | AvoidHTMLEscape}}\n```",
		"hidden_combined_output": "<details>\n\n```\n{{.CombinedOutput | AvoidHTMLEscape}}\n```\n\n</details>",
	}
	if strings.Contains(param.JoinCommand, "```") {
		builtinTemplates["join_command"] = "<pre><code>$ {{.JoinCommand | AvoidHTMLEscape}}</pre></code>"
	}
	if strings.Contains(param.CombinedOutput, "```") {
		builtinTemplates["hidden_combined_output"] = "<details><pre><code>{{.CombinedOutput | AvoidHTMLEscape}}</code></pre></details>"
	}

	ret := map[string]string{
		"link": "",
	}
	if param.CI != "" {
		if link, ok := buildLinks[param.CI]; ok {
			ret["link"] = link
		}
	}
	for k, v := range builtinTemplates {
		ret[k] = v
	}
	for k, v := range param.Templates {
		ret[k] = v
	}
	return ret
}

type Renderer struct {
	Getenv func(string) string
}

func addTemplates(tpl string, templates map[string]string) string {
	for k, v := range templates {
		tpl += `{{define "` + k + `"}}` + v + "{{end}}"
	}
	return tpl
}

func avoidHTMLEscape(text string) template.HTML {
	return template.HTML(text) //nolint:gosec
}

func (renderer *Renderer) Render(tpl string, templates map[string]string, params interface{}) (string, error) {
	tpl = addTemplates(tpl, templates)
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"Env":             renderer.Getenv,
		"AvoidHTMLEscape": avoidHTMLEscape,
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
