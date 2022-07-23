package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
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
		"circleci": fmt.Sprintf(
			`[workflow](https://circleci.com/workflow-run/%s) [job](%s) (job: %s)`,
			os.Getenv("CIRCLE_WORKFLOW_ID"),
			os.Getenv("CIRCLE_BUILD_URL"),
			os.Getenv("CIRCLE_JOB"),
		),
		"codebuild": fmt.Sprintf(`[Build link](%s)`, os.Getenv("CODEBUILD_BUILD_URL")),
		"drone": fmt.Sprintf(
			`[build](%s) [step](%s/%s/%s)`,
			os.Getenv("DRONE_BUILD_LINK"),
			os.Getenv("DRONE_BUILD_LINK"),
			os.Getenv("DRONE_STAGE_NUMBER"),
			os.Getenv("DRONE_STEP_NUMBER"),
		),
		"github-actions": fmt.Sprintf(
			`[Build link](https://github.com/%s/actions/runs/%s)`,
			os.Getenv("GITHUB_REPOSITORY"),
			os.Getenv("GITHUB_RUN_ID"),
		),
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

	// delete some functions for security reason
	funcs := sprig.FuncMap()
	delete(funcs, "env")
	delete(funcs, "expandenv")
	delete(funcs, "getHostByName")
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"AvoidHTMLEscape": avoidHTMLEscape,
	}).Funcs(funcs).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, params); err != nil {
		return "", fmt.Errorf("render a template with params: %w", err)
	}
	return buf.String(), nil
}
