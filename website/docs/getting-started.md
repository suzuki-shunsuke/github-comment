---
sidebar_position: 300
---

# Getting Started

Please prepare a GitHub access token. https://github.com/settings/tokens

`github-comment` provides the following subcommands.

* init: generate a configuration file template
* post: create a comment
* exec: execute a shell command and create a comment according to the command result
* hide: hide pull request or issue comments

## post

Let's create a simple comment. **Please change the parameter properly**.

```console
$ github-comment post -token <your GitHub personal access token> -org suzuki-shunsuke -repo github-comment -pr 1 -template test
```

https://github.com/suzuki-shunsuke/github-comment/pull/1#issuecomment-601501451

![post-test](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/post-test.png)

You can pass the API token from the environment variable `GITHUB_TOKEN` or `GITHUB_ACCESS_TOKEN` too.
Then you sent a comment `test` to the pull request https://github.com/suzuki-shunsuke/github-comment/pull/1 .
Instead of pull request, you can send a comment to a commit

```console
$ github-comment post -org suzuki-shunsuke -repo github-comment -sha1 36b1ade9740768f3645c240d487b53bee9e42702 -template test
```

https://github.com/suzuki-shunsuke/github-comment/commit/36b1ade9740768f3645c240d487b53bee9e42702#commitcomment-37933181

![comment-to-commit](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/comment-to-commit.png)

The template is rendered with [Go's text/template](https://golang.org/pkg/text/template/).

You can write the template in the configuration file.

github-comment.yaml

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
You can define multiple templates in the configuration file and specify the template by the argument `-template-key (-k)`.

```console
$ github-comment post -k hello
```

If `-template` and `-template-key` aren't set, the template `default` is used.

## exec

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

```console
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

```console
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
