---
sidebar_position: 850
---

# Hide comments

[#210](https://github.com/suzuki-shunsuke/github-comment/pull/210)

When github-comment is used at CI, github-comment posts a comment at every builds.
So outdated comments would remain.
You would like to hide outdated comments.

By the subcommand `hide`, you can hide outdated comments.
From github-comment v3, github-comments injects meta data like SHA1 into comments as HTML comment.

e.g.

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
1. hide comments which match the [expr](https://github.com/expr-lang/expr/blob/master/docs/language-definition.md) expression

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
you can configure the condition in the configuration file.

```yaml
hide:
  default: "true"
  hello: 'Comment.HasMeta && (Comment.Meta.SHA1 != Commit.SHA1 && Comment.Meta.Vars.target == "hello")'
```

you can specify the template with `--hide-key (-k)` option.

```console
$ github-comment hide -k hello
```

If the template isn't specified, the template `default` is used.

you can specify the condition with `-condition` option.

```console
$ github-comment hide -condition 'Comment.Body contains "foo"'
```
