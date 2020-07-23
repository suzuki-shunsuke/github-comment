# github-comment

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/github-comment/status.svg)](https://cloud.drone.io/suzuki-shunsuke/github-comment)
[![codecov](https://codecov.io/gh/suzuki-shunsuke/github-comment/branch/master/graph/badge.svg)](https://codecov.io/gh/suzuki-shunsuke/github-comment)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/github-comment)](https://goreportcard.com/report/github.com/suzuki-shunsuke/github-comment)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/github-comment.svg)](https://github.com/suzuki-shunsuke/github-comment)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/master/LICENSE)

CLI to create a GitHub comment with GitHub REST API

```
$ github-comment post -template test
```

<p align="center">
  <img src="https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/post-test.png">
</p>

```
$ github-comment exec -- go test ./...
```

<p align="center">
  <img src="https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/exec-go-test.png">
</p>

## Note

* Currently, `github-comment` doesn't support GitHub Enterprise

## Install

Please download a binary from the [release page](https://github.com/suzuki-shunsuke/github-comment/releases).

```
$ github-comment --version
$ github-comment --help
```

## Getting Started

Please prepare a GitHub access token. https://github.com/settings/tokens

`github-comment` provides two subcommands.

* post: create a comment
* exec: execute a shell command and create a comment according to the command result

### post

Let's create a simple comment. **Please change the parameter properly**.

```
$ github-comment post -token <your GitHub personal access token> -org suzuki-shunsuke -repo github-comment -pr 1 -template test
```

https://github.com/suzuki-shunsuke/github-comment/pull/1#issuecomment-601501451

<p align="center">
  <img src="https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/post-test.png">
</p>

You can pass the API token from the environment variable `GITHUB_TOKEN` or `GITHUB_ACCESS_TOKEN` too.
Then we sent a comment `test` to the pull request https://github.com/suzuki-shunsuke/github-comment/pull/1 .
Instead of pull request, we can send a comment to a commit

```
$ github-comment post -org suzuki-shunsuke -repo github-comment -sha1 36b1ade9740768f3645c240d487b53bee9e42702 -template test
```

https://github.com/suzuki-shunsuke/github-comment/commit/36b1ade9740768f3645c240d487b53bee9e42702#commitcomment-37933181

<p align="center">
  <img src="https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/comment-to-commit.png">
</p>

The template is rendered with [Go's text/template](https://golang.org/pkg/text/template/).

You can write the template in the configuration file.

.github-comment.yml

```yaml
post:
  default: |
    {{.Org}}/{{.Repo}} test
  hello: |
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

<p align="center">
  <img src="https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/exec-1.png">
</p>

Let's send the comment only if the command is failed.
Update the above configuration.

```yaml
exec:
  hello:
    - when: ExitCode != 0
      template: |
        ...
```

Run the above command again, then the command wouldn't be created.

If the command is failed, then the comment is created.

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
        command is failed
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

## Configuration

### post: variables which can be used in template

* Org
* Repo
* PRNumber
* SHA1
* TemplateKey

### exec

The configuration of `exec` is little more difficult than `post`, but the template key and `template` is same as `post`.
The each template is list which element has the attribute `when` and `template`, and `dont_comment`.
The attribute `when` is evaluated by the evaluation engine  https://github.com/antonmedv/expr , and the result should be `boolean`.
About expr, please see https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md too.
When the evaluation result is `false` the element is ignored, and the first matching element is used.
If `dont_comment` is `true`, the comment isn't created.
If no element matches, the comment isn't created without error.

The following variables can be used in `when` and `template`

* Stdout: the command standard output
* Stderr: the command standard error output
* CombinedOutput: Stdout + Stderr
* Command: https://golang.org/pkg/os/exec/#Cmd.String
* ExitCode: the command exit code
* Env: the function to get the environment variable https://golang.org/pkg/os/#Getenv
* Org
* Repo
* PRNumber
* SHA1
* TemplateKey

## Options

* token: GitHub API token to create a comment
* org: GitHub organization name
* repo: GitHub repository name
* revision: commit SHA
* pr: pull request number
* template: comment text
* template-key: template key

## post command supports standard input to pass a template

Instead of `-template`, we can pass a template from a standard input.

```
$ echo hello | github-comment post
```

## Environment variables

* GITHUB_TOKEN, GITHUB_ACCESS_TOKEN: complement the option `token`

## Support to complement options with CircleCI built-in Environment variables

https://circleci.com/docs/2.0/env-vars/#built-in-environment-variables

* org: CIRCLE_PROJECT_USERNAME
* repo: CIRCLE_PROJECT_REPONAME
* pr: CIRCLE_PULL_REQUEST
* revision: CIRCLE_SHA

## License

[MIT](LICENSE)
