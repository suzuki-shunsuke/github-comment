---
sidebar_position: 800
---

# Complement

## Complement options with Platform's built-in Environment variables

The following platforms are supported.

* CircleCI
* GitHub Actions
* Drone
* AWS CodeBuild
* Google Cloud Build

To complement, [suzuki-shunske/go-ci-env](https://github.com/suzuki-shunsuke/go-ci-env) is used.

## Google Cloud Build Support

[#521](https://github.com/suzuki-shunsuke/github-comment/pull/521), github-comment >= [v4.4.0](https://github.com/suzuki-shunsuke/github-comment/releases/tag/v4.4.0)

Set the environment variable `GOOGLE_CLOUD_BUILD`.

```sh
GOOGLE_CLOUD_BUILD=true
```

Set the following environment variables using [substitutions](https://cloud.google.com/cloud-build/docs/configuring-builds/substitute-variable-values).

* `COMMIT_SHA`
* `BUILD_ID`
* `PROJECT_ID`
* `_PR_NUMBER`
* `_REGION`

Specify the repository owner and name in `github-comment.yaml`.

e.g.

github-comment.yaml

```yaml
base:
  org: suzuki-shunsuke
  repo: github-comment
```

## Complement the pull request number from CI_INFO_PR_NUMBER

The environment variable `CI_INFO_PR_NUMBER` is set by [ci-info](https://github.com/suzuki-shunsuke/ci-info) by default. 
If the pull request number can't be gotten from platform's built-in environment variables but `CI_INFO_PR_NUMBER` is set, github-comment uses `CI_INFO_PR_NUMBER`.

## Complement options with any environment variables

:::caution
This feature was removed from [v5.0.0](https://github.com/suzuki-shunsuke/github-comment/releases/tag/v5.0.0) for security reason.
:::
