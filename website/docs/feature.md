---
sidebar_position: 700
---

# Feature

## post command supports standard input to pass a template

Instead of `-template`, you can pass a template from a standard input with `-stdin-template`.

```console
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

e.g.

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

You can avoid the error by the command line option `--skip-no-token` or the configuration `skip_no_token: true`.
If the GitHub Access Token is set, this option is ignored.
If the GitHub Access Token isn't set, this option works like `--dry-run`.

## Skip to send a comment with Environment variable

https://github.com/suzuki-shunsuke/github-comment/issues/143

When you try to run shell scripts for CI on local for testing, in many case you don't want to send a comment.
So github-comment supports to skip to send a comment with an environment variable.

Set the environment variable `GITHUB_COMMENT_SKIP` to `true`.

```console
$ export GITHUB_COMMENT_SKIP=true
$ github-comment post -template test # Do nothing
$ github-comment exec -- echo hello # a command is run but a comment isn't sent
hello
```
