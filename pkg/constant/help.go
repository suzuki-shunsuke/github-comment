package constant

const Help = `
github-comment - Post a comment to GitHub Issue or Github Commit with GitHub REST API

## USAGE

  $ github-comment -version
  $ github-comment -help
  # comment to a commit
  $ github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-revision <revision>] [-template <template>]
  # comment to a pull request
  $ github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>] [-template <template>]
	$ echo "<comment>" | github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>]

## Options

* token: GitHub API token to create a comment
* org: GitHub organization name
* repo: GitHub repository name
* revision: commit SHA
* pr: pull request number
* template: comment text

## Support standard input to pass a template

Instead of "-template", we can pass a template from a standard input.

$ echo hello | github-comment

## Environment variables

* GITHUB_TOKEN: complement the option "token"

## Support to complement options with CircleCI built-in Environment variables
  
* org: CIRCLE_PROJECT_USERNAME
* repo: CIRCLE_PROJECT_REPONAME
* pr: CIRCLE_PULL_REQUEST
* revision: CIRCLE_SHA`
