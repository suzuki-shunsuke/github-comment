# github-comment

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/github-comment/status.svg)](https://cloud.drone.io/suzuki-shunsuke/github-comment)
[![codecov](https://codecov.io/gh/suzuki-shunsuke/github-comment/branch/master/graph/badge.svg)](https://codecov.io/gh/suzuki-shunsuke/github-comment)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/github-comment)](https://goreportcard.com/report/github.com/suzuki-shunsuke/github-comment)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/github-comment.svg)](https://github.com/suzuki-shunsuke/github-comment)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/master/LICENSE)

CLI to create a GitHub comment with GitHub REST API

## Usage

```
$ github-comment version
$ github-comment help
# comment to a commit
$ github-comment post [-token <token>] [-org <org>] [-repo <repo>] [-revision <revision>] [-template <template>]
# comment to a pull request
$ github-comment post [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>] [-template <template>]
$ echo "<comment>" | github-comment post [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>]
$ github-comment exec [-token <token>] [-org <org>] [-repo <repo>] [-revision <revision>] [-template-key <template key>] -- <command>
$ github-comment exec -- echo hello
```

## Configuration

```yaml
---
post:
  # <template key>:
  default: |
    {{.Org}}/{{.Repo}}
    {{.PRNumber}}
    {{.SHA1}}
    {{.TemplateKey}}
    CIRCLE_PULL_REQUEST: {{Env "CIRCLE_PULL_REQUEST"}}
exec:
  # <template key>:
  default: # default configuration
    # https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md
    - when: ExitCode != 0
      # https://golang.org/pkg/text/template/
      template: |
        command: {{.Command}}
        exit code: {{.ExitCode}}
        stdout: {{.Stdout}}
        stderr: {{.Stderr}}
        combined output: {{.CombinedOutput}}
        CIRCLE_PULL_REQUEST: {{Env "CIRCLE_PULL_REQUEST"}}
    - when: ExitCode == 0
      dont_comment: true
  hello:
    - when: true
      template: hello
```

### post: variables which can be used in template

* Org
* Repo
* PRNumber
* SHA1
* TemplateKey

### exec: variables which can be used in `when` and `template`

* Stdout: the command standard output
* Stderr: the command standard error output
* CombinedOutput: Stdout + Stderr
* Command: https://golang.org/pkg/os/exec/#Cmd.String
* ExitCode: the command exit code
* Env: the function to get the environment variable https://golang.org/pkg/os/#Getenv

## Options

* token: GitHub API token to create a comment
* org: GitHub organization name
* repo: GitHub repository name
* revision: commit SHA
* pr: pull request number
* template: comment text
* template key: template key

## Support standard input to pass a template

Instead of `-template`, we can pass a template from a standard input.

```
$ echo hello | github-comment post
```

## Environment variables

* GITHUB_TOKEN: complement the option `token`

## Support to complement options with CircleCI built-in Environment variables

https://circleci.com/docs/2.0/env-vars/#built-in-environment-variables

* org: CIRCLE_PROJECT_USERNAME
* repo: CIRCLE_PROJECT_REPONAME
* pr: CIRCLE_PULL_REQUEST
* revision: CIRCLE_SHA

## License

[MIT](LICENSE)
