# github-comment-cli

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/github-comment-cli/status.svg)](https://cloud.drone.io/suzuki-shunsuke/github-comment-cli)
[![codecov](https://codecov.io/gh/suzuki-shunsuke/github-comment-cli/branch/master/graph/badge.svg)](https://codecov.io/gh/suzuki-shunsuke/github-comment-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/github-comment-cli)](https://goreportcard.com/report/github.com/suzuki-shunsuke/github-comment-cli)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/github-comment-cli.svg)](https://github.com/suzuki-shunsuke/github-comment-cli)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/github-comment-cli/master/LICENSE)

CLI to create a GitHub comment with GitHub REST API

## Usage

```
$ github-comment -version
$ github-comment -help
# comment to a commit
$ github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-revision <revision>] [-template <template>]
# comment to a pull request
$ github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>] [-template <template>]
$ echo "<comment>" | github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>]
```

## Options

* token: GitHub API token to create a comment
* org: GitHub organization name
* repo: GitHub repository name
* revision: commit SHA
* pr: pull request number
* template: comment text

## Support standard input to pass a template

Instead of `-template`, we can pass a template from a standard input.

```
$ echo hello | github-comment
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
