---
sidebar_position: 860
---

# Output the result to a text file

[#1566](https://github.com/suzuki-shunsuke/github-comment/pull/1566) [v6.3.0](https://github.com/suzuki-shunsuke/github-comment/releases/tag/v6.3.0)

Instead of posting a comment to a GitHub Issue or Pull Request, you can output the result to a text file using `github-comment exec`'s `-out` option.
This is useful to output the result of GitHub Actions `workflow_dispatch` or `schedule` events to `$GITHUB_STEP_SUMMARY`.

e.g.

```sh
github-comment exec -out "file:$GITHUB_STEP_SUMMARY" -- npm test
```

You can post both GitHub and a file.

e.g.

```sh
github-comment exec -out github -out "file:$GITHUB_STEP_SUMMARY" -- npm test
```

The value of `-out` must be either `github` or `file:<file path>`.
