---
sidebar_position: 900
---

# Builtin Templates

Some default templates are provided.
They are overwritten if the same name templates are defined in the configuration file.

* [status](#status)
* [join_command](#join_command)
* [hidden_combined_output](#hidden_combined_output)
* [link](#link)
* [`exec`'s default template](#execs-default-template)

## status

Usage:

```
{{template "status" .}}
```

Content of the template:

```
:{{if eq .ExitCode 0}}white_check_mark{{else}}x{{end}}:
```

## join_command

Usage:

```
{{template "join_command" .}}
```

Content of the template:

If `.JoinCommand` includes \`\`\`,

```
<pre><code>$ {{.JoinCommand | AvoidHTMLEscape}}</pre></code>
```

Otherwise,

<pre><code>
```
<br/>
$ {'{{'}.JoinCommand | AvoidHTMLEscape{'}}'}
<br/>
```
</code></pre>

## hidden_combined_output

Usage:

```
{{template "hidden_combined_output" .}}
```

Content of the template:

If `.CombinedOutput` includes \`\`\` ,

```
<details><pre><code>{{.CombinedOutput | AvoidHTMLEscape}}</code></pre></details>
```

Otherwise,

<pre><code>
{'<details>'}
<br/>
```
<br/>
{'{{'}.CombinedOutput | AvoidHTMLEscape{'}}'}
<br/>
```
<br/>
{'</details>'}
</code></pre>

## link

Usage:

```
{{template "link" .}}
```

Content of the template:

`link` is different per CI service.

### CircleCI

```
[workflow](https://circleci.com/workflow-run/{{env "CIRCLE_WORKFLOW_ID" }}) [job]({{env "CIRCLE_BUILD_URL"}}) (job: {{env "CIRCLE_JOB"}})
```

### CodeBuild

```
[Build link]({{env "CODEBUILD_BUILD_URL"}})
```

### Drone

```
[build]({{env "DRONE_BUILD_LINK"}}) [step]({{env "DRONE_BUILD_LINK"}}/{{env "DRONE_STAGE_NUMBER"}}/{{env "DRONE_STEP_NUMBER"}})
```

### GitHub Actions

```
[Build link](https://github.com/{{env "GITHUB_REPOSITORY"}}/actions/runs/{{env "GITHUB_RUN_ID"}})
```

## `exec`'s default template

```yaml
when: ExitCode != 0
template: |
  {{template "status" .}} {{template "link" .}}

  {{template "join_command" .}}

  {{template "hidden_combined_output" .}}
```
