---
sidebar_position: 860
---

# Delete comments

When github-comment is used at CI, github-comment posts a comment at every builds.
So outdated comments would remain.
Unlike the [`hide`](hide) command which minimizes outdated comments, the `delete` command removes them completely.

By the subcommand `delete`, you can delete outdated comments.
github-comment injects meta data like SHA1 into comments as HTML comment, the same as the `hide` command.

e.g.

```
<!-- github-comment: {"JobID":"xxx","JobName":"plan","SHA1":"79acc0778da6660712a65fd41a48b72cb7ad16c0","TemplateKey":"default","Vars":{}} -->
```

In `delete` command, github-comment does the following things.

1. gets the list of pull request (issue) comments
1. extracts the injected meta data from comments
1. deletes comments which match the [expr](https://github.com/expr-lang/expr/blob/master/docs/language-definition.md) expression

Unlike the `hide` command, the `delete` command also targets comments that are already minimized.

:::caution
Unlike `hide` (minimize), deletion is irreversible.
Please configure the condition carefully so that only intended comments are deleted.
:::

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
* DeleteKey
* Vars
* Env: `func(string) string`

The default condition is `Comment.HasMeta && Comment.Meta.SHA1 != Commit.SHA1`.
you can configure the condition in the configuration file.

```yaml
delete:
  default: "true"
  hello: 'Comment.HasMeta && (Comment.Meta.SHA1 != Commit.SHA1 && Comment.Meta.Vars.target == "hello")'
```

you can specify the condition key with `--delete-key (-k)` option.

```console
$ github-comment delete -k hello
```

If the key isn't specified, the key `default` is used.

you can specify the condition with `-condition` option.

```console
$ github-comment delete -condition 'Comment.Body contains "foo"'
```
