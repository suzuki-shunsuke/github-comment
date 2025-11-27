package template

import (
	"bytes"
	"fmt"
	"html/template"
	"maps"
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
	cloudBuildRegion := os.Getenv("_REGION")
	if cloudBuildRegion == "" {
		cloudBuildRegion = "global"
	}
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
			`[Build link](%s/%s/actions/runs/%s)`,
			os.Getenv("GITHUB_SERVER_URL"),
			os.Getenv("GITHUB_REPOSITORY"),
			os.Getenv("GITHUB_RUN_ID"),
		),
		"cloud-build": fmt.Sprintf(
			"https://console.cloud.google.com/cloud-build/builds;region=%s/%s?project=%s",
			cloudBuildRegion,
			os.Getenv("BUILD_ID"),
			os.Getenv("PROJECT_ID"),
		),
	}

	builtinTemplates := map[string]string{
		"status":       `:{{if eq .ExitCode 0}}white_check_mark{{else}}x{{end}}:`,
		"join_command": "```\n$ {{.JoinCommand | AvoidHTMLEscape}}\n```",
		"hidden_combined_output": `<details>

{{WrapCode .CombinedOutput}}

</details>`,
	}

	ret := map[string]string{
		"link": "",
	}
	if param.CI != "" {
		if link, ok := buildLinks[param.CI]; ok {
			ret["link"] = link
		}
	}
	maps.Copy(ret, builtinTemplates)
	maps.Copy(ret, param.Templates)
	return ret
}

type Renderer struct {
	Getenv func(string) string
}

func addTemplates(tpl string, templates map[string]string) string {
	var tplSb86 strings.Builder
	for k, v := range templates {
		tplSb86.WriteString(`{{define "` + k + `"}}` + v + "{{end}}")
	}
	tpl += tplSb86.String()
	return tpl
}

func avoidHTMLEscape(text string) template.HTML {
	return template.HTML(text) //nolint:gosec
}

func wrapCode(text string) any {
	if len(text) > 60000 { //nolint:mnd
		text = text[:20000] + `

# ...
# ... The maximum length of GitHub Comment is 65536, so the content is omitted by github-comment.
# ...

` + text[len(text)-20000:]
	}
	if strings.Contains(text, "```") {
		return template.HTML("<pre><code>" + template.HTMLEscapeString(text) + "</code></pre>") //nolint:gosec
	}
	return template.HTML("\n```\n" + text + "\n```\n") //nolint:gosec
}

func (r *Renderer) Render(tpl string, templates map[string]string, params any) (string, error) {
	tpl = addTemplates(tpl, templates)

	// delete some functions for security reason
	funcs := sprig.FuncMap()
	delete(funcs, "env")
	delete(funcs, "expandenv")
	delete(funcs, "getHostByName")
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"AvoidHTMLEscape": avoidHTMLEscape,
		"WrapCode":        wrapCode,
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
