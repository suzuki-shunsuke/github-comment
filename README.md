# github-comment

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/github-comment/status.svg)](https://cloud.drone.io/suzuki-shunsuke/github-comment)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/github-comment)](https://goreportcard.com/report/github.com/suzuki-shunsuke/github-comment)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/github-comment.svg)](https://github.com/suzuki-shunsuke/github-comment)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/main/LICENSE)

CLI to create and hide GitHub comments by GitHub REST API

```
$ github-comment post -template test
```

![post-test](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/post-test.png)

```
$ github-comment exec -- go test ./...
```

![exec-go-test](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/exec-go-test.png)

## Note

* Currently, `github-comment` doesn't support GitHub Enterprise

## Install

Please download a binary from the [release page](https://github.com/suzuki-shunsuke/github-comment/releases).

Or you can install github-comment with [Homebrew](https://brew.sh/).

```
$ brew install suzuki-shunsuke/github-comment/github-comment
```

```
$ github-comment --version
$ github-comment --help
```

## Getting Started

Please prepare a GitHub access token. https://github.com/settings/tokens

`github-comment` provides the following subcommands.

* init: generate a configuration file template
* post: create a comment
* exec: execute a shell command and create a comment according to the command result
* hide: hide pull request or issue comments

### post

Let's create a simple comment. **Please change the parameter properly**.

```
$ github-comment post -token <your GitHub personal access token> -org suzuki-shunsuke -repo github-comment -pr 1 -template test
```

https://github.com/suzuki-shunsuke/github-comment/pull/1#issuecomment-601501451

![post-test](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/post-test.png)

You can pass the API token from the environment variable `GITHUB_TOKEN` or `GITHUB_ACCESS_TOKEN` too.
Then we sent a comment `test` to the pull request https://github.com/suzuki-shunsuke/github-comment/pull/1 .
Instead of pull request, we can send a comment to a commit

```
$ github-comment post -org suzuki-shunsuke -repo github-comment -sha1 36b1ade9740768f3645c240d487b53bee9e42702 -template test
```

https://github.com/suzuki-shunsuke/github-comment/commit/36b1ade9740768f3645c240d487b53bee9e42702#commitcomment-37933181

![comment-to-commit](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/comment-to-commit.png)

The template is rendered with [Go's text/template](https://golang.org/pkg/text/template/).

You can write the template in the configuration file.

.github-comment.yml

```yaml
post:
  default:
    template: |
      {{.Org}}/{{.Repo}} test
  hello:
    template: |
      hello world
```

If the argument `-template` is given, the configuration file is ignored.
We can define multiple templates in the configuration file and specify the template by the argument `-template-key (-k)`.

```
$ github-comment post -k hello
```

If `-template` and `-template-key` aren't set, the template `default` is used.

### exec

Let's add the following configuration in the configuration file.

```yaml
exec:
  hello:
    - when: true
      template: |
        exit code: {{.ExitCode}}

        ```
        $ {{.Command}}
        ```

        Stdout:

        ```
        {{.Stdout}}
        ```

        Stderr:

        ```
        {{.Stderr}}
        ```

        CombinedOutput:

        ```
        {{.CombinedOutput}}
        ```
```

Then run a command and send the result as a comment.

```
$ github-comment exec -org suzuki-shunsuke -repo github-comment -pr 1 -k hello -- bash -c "echo foo; echo bar >&2; echo zoo"
bar
foo
zoo
```

https://github.com/suzuki-shunsuke/github-comment/pull/1#issuecomment-601503124

![exec-1](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/exec-1.png)

Let's send the comment only if the command failed.
Update the above configuration.

```yaml
exec:
  hello:
    - when: ExitCode != 0
      template: |
        ...
```

Run the above command again, then the command wouldn't be created.

If the command failed, then the comment is created.

```
$ github-comment exec -org suzuki-shunsuke -repo github-comment -pr 1 -k hello -- curl -f https://github.com/suzuki-shunsuke/not_found
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
curl: (22) The requested URL returned error: 404 Not Found
exit status 22
```

https://github.com/suzuki-shunsuke/github-comment/pull/1#issuecomment-601505610

You can change the template by the command result.

```yaml
exec:
  hello:
    - when: ExitCode != 0
      template: |
        command failed
    - when: ExitCode == 0
      template: |
        command is succeeded
```

By `dont_comment`, you can define the condition which the message isn't created.

```yaml
exec:
  hello:
    - when: ExitCode != 0
      dont_comment: true
    - when: true
      template: |
        Hello, world
```

### hide

[#210](https://github.com/suzuki-shunsuke/github-comment/pull/210)

When github-comment is used at CI, github-comment posts a comment at every builds.
So outdated comments would remain.
We want to hide outdated comments.

By the subcommand `hide`, we can hide outdated comments.
From github-comment v3, github-comments injects meta data like SHA1 into comments as HTML comment.

ex.

```
<!-- github-comment: {"JobID":"xxx","JobName":"plan","SHA1":"79acc0778da6660712a65fd41a48b72cb7ad16c0","TemplateKey":"default","Vars":{}} -->
```

The following meta data is injected.

* JobName (support only some CI platform)
* JobID (support only some CI platform)
* WorkflowName (support only some CI platform)
* TemplateKey
* Vars
* SHA1

From github-comment v4, only variables specified by `embedded_var_names` are embedded into the comment.

In `hide` command, github-comment does the following things.

1. gets the list of pull request (issue) comments
1. extracts the injected meta data from comments
1. hide comments which match the [expr](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md) expression

The following variable is passed to the expression.

* Commit:
  * Org
  * Repo
  * PRNumber
  * SHA1
* Comment
  * Body
  * HasMeta
  * Meta
    * SHA1
    * TemplateKey
    * Vars
* HideKey
* Vars
* Env: `func(string) string`

The default condition is `Comment.HasMeta && Comment.Meta.SHA1 != Commit.SHA1`.
We can configure the condition in the configuration file.

```yaml
hide:
  default: "true"
  hello: 'Comment.HasMeta && (Comment.Meta.SHA1 != Commit.SHA1 && Comment.Meta.Vars.target == "hello")'
```

We can specify the template with `--hide-key (-k)` option.

```
$ github-comment hide -k hello
```

If the template isn't specified, the template `default` is used.

We can specify the condition with `-condition` option.

```
$ github-comment hide -condition 'Comment.Body contains "foo"'
```

## Usage

```console
$ github-comment help
NAME:
   github-comment - post a comment to GitHub

USAGE:
   github-comment [global options] command [command options] [arguments...]

VERSION:
   3.1.0

COMMANDS:
   post     post a comment
   exec     execute a command and post the result as a comment
   init     scaffold a configuration file if it doesn't exist
   hide     hide issue or pull request comments
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value  log level [$GITHUB_COMMENT_LOG_LEVEL]
   --help, -h         show help (default: false)
   --version, -v      print the version (default: false)
```

```console
$ github-comment help post
NAME:
   github-comment post - post a comment

USAGE:
   github-comment post [command options] [arguments...]

OPTIONS:
   --org value                     GitHub organization name
   --repo value                    GitHub repository name
   --token value                   GitHub API token [$GITHUB_TOKEN, $GITHUB_ACCESS_TOKEN]
   --sha1 value                    commit sha1
   --template value                comment template
   --template-key value, -k value  comment template key (default: "default")
   --config value                  configuration file path
   --pr value                      GitHub pull request number (default: 0)
   --var value                     template variable
   --var-file value                template variable name and file path
   --dry-run                       output a comment to standard error output instead of posting to GitHub (default: false)
   --skip-no-token, -n             works like dry-run if the GitHub Access Token isn't set (default: false) [$GITHUB_COMMENT_SKIP_NO_TOKEN]
   --silent, -s                    suppress the output of dry-run and skip-no-token (default: false)
   --stdin-template                read standard input as the template (default: false)
```

```console
$ github-comment help exec
NAME:
   github-comment exec - execute a command and post the result as a comment

USAGE:
   github-comment exec [command options] [arguments...]

OPTIONS:
   --org value                     GitHub organization name
   --repo value                    GitHub repository name
   --token value                   GitHub API token [$GITHUB_TOKEN, $GITHUB_ACCESS_TOKEN]
   --sha1 value                    commit sha1
   --template value                comment template
   --template-key value, -k value  comment template key (default: "default")
   --config value                  configuration file path
   --pr value                      GitHub pull request number (default: 0)
   --var value                     template variable
   --var-file value                template variable name and file path
   --dry-run                       output a comment to standard error output instead of posting to GitHub (default: false)
   --skip-no-token, -n             works like dry-run if the GitHub Access Token isn't set (default: false) [$GITHUB_COMMENT_SKIP_NO_TOKEN]
   --silent, -s                    suppress the output of dry-run and skip-no-token (default: false)
```

```console
$ github-comment help hide
NAME:
   github-comment hide - hide issue or pull request comments

USAGE:
   github-comment hide [command options] [arguments...]

OPTIONS:
   --org value                 GitHub organization name
   --repo value                GitHub repository name
   --token value               GitHub API token [$GITHUB_TOKEN, $GITHUB_ACCESS_TOKEN]
   --config value              configuration file path
   --condition value           hide condition
   --hide-key value, -k value  hide condition key (default: "default")
   --pr value                  GitHub pull request number (default: 0)
   --sha1 value                commit sha1
   --var value                 template variable
   --dry-run                   output a comment to standard error output instead of posting to GitHub (default: false)
   --skip-no-token, -n         works like dry-run if the GitHub Access Token isn't set (default: false) [$GITHUB_COMMENT_SKIP_NO_TOKEN]
   --silent, -s                suppress the output of dry-run and skip-no-token (default: false)
```

## Configuration

### post: variables which can be used in template

* Org
* Repo
* PRNumber
* SHA1
* TemplateKey
* Vars
* Env: the function to get the environment variable https://golang.org/pkg/os/#Getenv
* AvoidHTMLEscape: the function to post a comment without HTML escape by [Go's html/template](https://golang.org/pkg/html/template/)
* Sprig Function: http://masterminds.github.io/sprig/

### exec

The configuration of `exec` is little more difficult than `post`, but the template key and `template` is same as `post`.
The each template is list which element has the attribute `when` and `template`, and `dont_comment`.
The attribute `when` is evaluated by the evaluation engine  https://github.com/antonmedv/expr , and the result should be `boolean`.
About expr, please see https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md too.
When the evaluation result is `false` the element is ignored, and the first matching element is used.
If `dont_comment` is `true`, the comment isn't created.
If no element matches, the comment isn't created without error.

In addition to the variables of `post` command, the following variables can be used in `when` and `template`

* Stdout: the command standard output
* Stderr: the command standard error output
* CombinedOutput: Stdout + Stderr
* Command: https://golang.org/pkg/os/exec/#Cmd.String
* JoinCommand: the string which the command and arguments are joined with the space character ` `
* ExitCode: the command exit code

### Define reusable template components

```yaml
templates:
  <template name>: <template content>
post:
  default:
    template: |
      {{template "<template name>" .}} ...
```

### Define variables

```yaml
vars:
  <variable name>: <variable value>
post:
  default:
    template: |
      {{.Vars.<variable name>}} ...
```

The variable can be passed with the option `-var <variable name>:<variable value>` too.

ex.

```
$ github-comment post -var name:foo
```

## post command supports standard input to pass a template

Instead of `-template`, we can pass a template from a standard input with `-stdin-template`.

```
$ echo hello | github-comment post -stdin-template
```

## post a substitute comment when it failed to post a too long comment

When the comment is too long, it failed to post a comment due to GitHub API's validation.

```json
{
  "message": "Validation Failed",
  "errors": [
    {
      "resource": "IssueComment",
      "code": "unprocessable",
      "field": "data",
      "message": "Body is too long (maximum is 65536 characters)"
    }
  ],
  "documentation_url": "https://docs.github.com/rest/reference/issues#create-an-issue-comment"
}
```

If a comment includes the long command standard output, you may encounter the error.

github-comment supports to post a substitute comment in that case.

When it failed to post a comment of `template`, github-comment posts a comment of `template_for_too_long` instead of `template`.

ex.

```yaml
post:
  hello:
    template: too long comment
    template_for_too_long: comment is too long
exec:
  hello:
    - when: ExitCode != 0
      template: |
        exit code: {{.ExitCode}}
        combined output: {{.CombinedOutput}}
      template_for_too_long: |
        comment is too long so the command output is omitted
        exit code: {{.ExitCode}}
```

## skip-no-token

https://github.com/suzuki-shunsuke/github-comment/issues/115

In some situation, the GitHub Access Token isn't exposed to the environment variable for the security.

For example, on Drone Secrets are not exposed to pull requests by default.

https://docs.drone.io/secret/repository/

> Secrets are not exposed to pull requests by default.
> This prevents a bad actor from sending a pull request and attempting to expose your secrets.

`github-comment` requires the GitHub Access Token, so it fails to run `github-comment post` and `github-comment exec`.

We can avoid the error by the command line option `--skip-no-token` or the configuration `skip_no_token: true`.
If the GitHub Access Token is set, this option is ignored.
If the GitHub Access Token isn't set, this option works like `--dry-run`.

## Skip to send a comment with Environment variable

https://github.com/suzuki-shunsuke/github-comment/issues/143

When we try to run shell scripts for CI on local for testing, in many case we don't want to send a comment.
So github-comment supports to skip to send a comment with an environment variable.

Set the environment variable `GITHUB_COMMENT_SKIP` to `true`.

```
$ export GITHUB_COMMENT_SKIP=true
$ github-comment post -template test # Do nothing
$ github-comment exec -- echo hello # a command is run but a comment isn't sent
hello
```

## Complement options with Platform's built-in Environment variables

The following platforms are supported.

* CircleCI
* GitHub Actions
* Drone
* AWS CodeBuild

To complement, [suzuki-shunske/go-ci-env](https://github.com/suzuki-shunsuke/go-ci-env) is used.

## Complement the pull request number from CI_INFO_PR_NUMBER

The environment variable `CI_INFO_PR_NUMBER` is set by [ci-info](https://github.com/suzuki-shunsuke/ci-info) by default. 
If the pull request number can't be gotten from platform's built-in environment variables but `CI_INFO_PR_NUMBER` is set, github-comment uses `CI_INFO_PR_NUMBER`.

## Complement options with any environment variables

In addition to the support of the above CI platforms, we can support other platforms with `complement` configuration.

```yaml
complement:
  pr:
  - type: envsubst
    value: '${GOOGLE_CLOUD_BUILD_PR_NUMBER}'
  org:
  - type: envsubst
    value: 'suzuki-shunsuke'
  repo:
  - type: envsubst
    value: '${GOOGLE_CLOUD_BUILD_REPO_NAME}'
  sha1:
  - type: envsubst
    value: '${GOOGLE_CLOUD_BUILD_COMMIT_SHA}'
  vars:
    yoo: # the variable "yoo" is added to "vars"
    - type: template
      value: '{{env "YOO"}}'
```

The following types are supported.

* `envsubst`: [drone/envsubst#EvalEnv](https://pkg.go.dev/github.com/drone/envsubst#EvalEnv)
* `template`: Go's [text/template](https://golang.org/pkg/text/template/) with [sprig functions](http://masterminds.github.io/sprig/)

## Builtin Templates

Some default templates are provided.
They are overwritten if the same name templates are defined in the configuration file.

* templates.status
* templates.join_command
* templates.hidden_combined_output
* templates.link
* exec.default

### templates.status

```
:{{if eq .ExitCode 0}}white_check_mark{{else}}x{{end}}:
```

### templates.join_command

If `.JoinCommand` includes \`\`\`,

```
<pre><code>$ {{.JoinCommand | AvoidHTMLEscape}}</pre></code>
```

Otherwise,

<pre><code>```
$ {{.JoinCommand | AvoidHTMLEscape}}
```</pre></code>

### templates.hidden_combined_output

If `.CombinedOutput` includes \`\`\` ,

```
<details><pre><code>{{.CombinedOutput | AvoidHTMLEscape}}</code></pre></details>
```

Otherwise,

<pre><code>&lt;details&gt;

```
{{.CombinedOutput | AvoidHTMLEscape}}
```

&lt;/details&gt;</pre></code>

### templates.link

`link` is different per CI service.

#### CircleCI

```
[workflow](https://circleci.com/workflow-run/{{env "CIRCLE_WORKFLOW_ID" }}) [job]({{env "CIRCLE_BUILD_URL"}}) (job: {{env "CIRCLE_JOB"}})
```

#### CodeBuild

```
[Build link]({{env "CODEBUILD_BUILD_URL"}})
```

#### Drone

```
[build]({{env "DRONE_BUILD_LINK"}}) [step]({{env "DRONE_BUILD_LINK"}}/{{env "DRONE_STAGE_NUMBER"}}/{{env "DRONE_STEP_NUMBER"}})
```

#### GitHub Actions

```
[Build link](https://github.com/{{env "GITHUB_REPOSITORY"}}/actions/runs/{{env "GITHUB_RUN_ID"}})
```

### exec.default

```yaml
when: ExitCode != 0
template: |
  {{template "status" .}} {{template "link" .}}

  {{template "join_command" .}}

  {{template "hidden_combined_output" .}}
```

## Configuration file path

The configuration file path can be specified with the `--config (-c)` option.
If the confgiuration file path isn't specified, the file named `.github-comment.yml` or `.github-comment.yaml` would be searched from the current directory to the root directory.

## Blog

Written in Japanese. https://techblog.szksh.cloud/github-comment/

## License

[MIT](LICENSE)
