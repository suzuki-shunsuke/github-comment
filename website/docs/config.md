---
sidebar_position: 500
---

# Configuration

## JSON Schema

You can use JSON Schema of github-comment's configuration file.

- https://github.com/suzuki-shunsuke/github-comment/blob/main/json-schema/github-comment.json
- https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/refs/heads/main/json-schema/github-comment.json

If you look for a CLI tool to validate configuration with JSON Schema, [ajv-cli](https://ajv.js.org/packages/ajv-cli.html) is useful.

```sh
ajv --spec=draft2020 -s json-schema/github-comment.json -d github-comment.yaml
```

### Input Complementation by YAML Language Server

[Please see the comment too.](https://github.com/szksh-lab/.github/issues/67#issuecomment-2564960491)

Version: `main`

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/main/json-schema/github-comment.json
```

Or pinning version:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/v6.3.1/json-schema/github-comment.json
```

## Configuration file path

The configuration file path can be specified with the `--config (-c)` option.
If the configuration file path isn't specified, the file named `^\.?github-comment\.ya?ml$` would be searched from the current directory to the root directory.

## Configuration file

You can scaffold a configuration file by `github-comment init` command.

```sh
github-comment init
```

e.g.

```yaml
hide:
  default: |
    Comment.HasMeta && Comment.Meta.SHA1 != Commit.SHA1 && ! (Comment.Meta.Program == "tfcmt" && Comment.Meta.Command == "apply")
post:
  tfmigrate-hcl-not-found:
    template: |
      ## :x: {{if .Vars.tfaction_target}}{{.Vars.tfaction_target}}: {{end}}.tfmigrate.hcl isn't found

      {{template "link" .}}

      To run `tfmigrate plan`, `.tfmigrate.hcl` is required.
    template_for_too_long: |
      comment is too long
exec:
  tfmigrate-plan:
    - when: true
      template: |
        ## {{template "status" .}} {{if .Vars.tfaction_target}}{{.Vars.tfaction_target}}: {{end}} tfmigrate plan

        {{template "link" .}}

        {{template "join_command" .}}

        {{template "hidden_combined_output" .}}
      template_for_too_long: |
        ## {{template "status" .}} {{if .Vars.tfaction_target}}{{.Vars.tfaction_target}}: {{end}} tfmigrate plan

        {{template "link" .}}

        {{template "join_command" .}}

        Command output is omitted as it is too long.
```

## Environment variables

- GITHUB_TOKEN, GITHUB_ACCESS_TOKEN
- GH_COMMENT_SKIP_NO_TOKEN, GITHUB_COMMENT_SKIP_NO_TOKEN
- GITHUB_COMMENT_SKIP
- GH_COMMENT_REPO_ORG
- GH_COMMENT_REPO_NAME
- GH_COMMENT_SHA1
- GH_COMMENT_CONFIG
- GH_COMMENT_PR_NUMBER, CI_INFO_PR_NUMBER
- GH_COMMENT_LOG_LEVEL
- `GH_COMMENT_VAR_*`

Please see [Complement](complement.md) too.

## Template Engine

Some fields such `template` and `template_for_too_long` are rendered by [html/template](https://pkg.go.dev/html/template).

## Template Functions

[sprig functions](http://masterminds.github.io/sprig/) except for the following functions are available.

- expandenv, env, getHostByName
- [os](http://masterminds.github.io/sprig/os.html)
- [network](http://masterminds.github.io/sprig/network.html)

And the following custom functions are also available.

- AvoidHTMLEscape: Skip escaping HTML

e.g.

```
{{.CombinedOutput | AvoidHTMLEscape}}
```

## Variables

### post command

- Org: GitHub Organization name
- Repo: GitHub Repository name
- PRNumber: Pull request number
- SHA1: Commit hash
- TemplateKey: Template key
- Vars: variables passed by `-var` and `-var-file`, the config `vars`, and the environment variables `GH_COMMENT_VAR_*`

### exec command

In addition to the variables of `post` command, the following variables are available in `when` and `template`.

- Stdout: the command standard output
- Stderr: the command standard error output
- CombinedOutput: Stdout + Stderr
- Command: https://golang.org/pkg/os/exec/#Cmd.String
- JoinCommand: the string which the command and arguments are joined with the space character ` `
- ExitCode: the command exit code

## exec

The each template is list which element has the attribute `when` and `template`, and `dont_comment`.
The attribute `when` is evaluated by the evaluation engine  https://github.com/antonmedv/expr , and the result should be `boolean`.
About expr, please see https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md too.
When the evaluation result is `false` the element is ignored, and the first matching element is used.
If `dont_comment` is `true`, the comment isn't created.
If no element matches, the comment isn't created without error.

## Define reusable template components

```yaml
templates:
  <template name>: <template content>
post:
  default:
    template: |
      {{template "<template name>" .}} ...
```

## Define variables

```yaml
vars:
  <variable name>: <variable value>
post:
  default:
    template: |
      {{.Vars.<variable name>}} ...
```

The variable can be passed with the option `-var <variable name>:<variable value>` too.

e.g.

```console
$ github-comment post -var name:foo
```

## See also

- [Builtin Templates](builtin-template.md)
- [GitHub Enterprise Support](github-enterprise.md)
- [Complement](complement.md)
