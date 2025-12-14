---
sidebar_position: 100
---

# github-comment

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/github-comment/status.svg)](https://cloud.drone.io/suzuki-shunsuke/github-comment)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/github-comment)](https://goreportcard.com/report/github.com/suzuki-shunsuke/github-comment)
[![DeepWiki](https://img.shields.io/badge/DeepWiki-suzuki--shunsuke%2Fgithub--comment-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACwAAAAyCAYAAAAnWDnqAAAAAXNSR0IArs4c6QAAA05JREFUaEPtmUtyEzEQhtWTQyQLHNak2AB7ZnyXZMEjXMGeK/AIi+QuHrMnbChYY7MIh8g01fJoopFb0uhhEqqcbWTp06/uv1saEDv4O3n3dV60RfP947Mm9/SQc0ICFQgzfc4CYZoTPAswgSJCCUJUnAAoRHOAUOcATwbmVLWdGoH//PB8mnKqScAhsD0kYP3j/Yt5LPQe2KvcXmGvRHcDnpxfL2zOYJ1mFwrryWTz0advv1Ut4CJgf5uhDuDj5eUcAUoahrdY/56ebRWeraTjMt/00Sh3UDtjgHtQNHwcRGOC98BJEAEymycmYcWwOprTgcB6VZ5JK5TAJ+fXGLBm3FDAmn6oPPjR4rKCAoJCal2eAiQp2x0vxTPB3ALO2CRkwmDy5WohzBDwSEFKRwPbknEggCPB/imwrycgxX2NzoMCHhPkDwqYMr9tRcP5qNrMZHkVnOjRMWwLCcr8ohBVb1OMjxLwGCvjTikrsBOiA6fNyCrm8V1rP93iVPpwaE+gO0SsWmPiXB+jikdf6SizrT5qKasx5j8ABbHpFTx+vFXp9EnYQmLx02h1QTTrl6eDqxLnGjporxl3NL3agEvXdT0WmEost648sQOYAeJS9Q7bfUVoMGnjo4AZdUMQku50McDcMWcBPvr0SzbTAFDfvJqwLzgxwATnCgnp4wDl6Aa+Ax283gghmj+vj7feE2KBBRMW3FzOpLOADl0Isb5587h/U4gGvkt5v60Z1VLG8BhYjbzRwyQZemwAd6cCR5/XFWLYZRIMpX39AR0tjaGGiGzLVyhse5C9RKC6ai42ppWPKiBagOvaYk8lO7DajerabOZP46Lby5wKjw1HCRx7p9sVMOWGzb/vA1hwiWc6jm3MvQDTogQkiqIhJV0nBQBTU+3okKCFDy9WwferkHjtxib7t3xIUQtHxnIwtx4mpg26/HfwVNVDb4oI9RHmx5WGelRVlrtiw43zboCLaxv46AZeB3IlTkwouebTr1y2NjSpHz68WNFjHvupy3q8TFn3Hos2IAk4Ju5dCo8B3wP7VPr/FGaKiG+T+v+TQqIrOqMTL1VdWV1DdmcbO8KXBz6esmYWYKPwDL5b5FA1a0hwapHiom0r/cKaoqr+27/XcrS5UwSMbQAAAABJRU5ErkJggg==)](https://deepwiki.com/suzuki-shunsuke/github-comment)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/github-comment.svg)](https://github.com/suzuki-shunsuke/github-comment)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/github-comment/main/LICENSE)

[github-comment](https://github.com/suzuki-shunsuke/github-comment) is a CLI to create and hide GitHub comments by GitHub REST API.

```console
$ github-comment post -k test
```

![post-test](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/post-test.png)

```console
$ github-comment exec -- go test ./...
```

![exec-go-test](https://cdn.jsdelivr.net/gh/suzuki-shunsuke/artifact@master/github-comment/exec-go-test.png)

```console
$ github-comment hide
```

![image](https://user-images.githubusercontent.com/13323303/161659056-998b566d-f5c4-4f7e-8a60-0191d6543763.png)

## Motivation

github-comment's motivation is to improve CI's user experience.
Using github-comment, you can notify the result of CI and
users can understand what the problem is and what they should do quickly.

It's bothersome to check CI log and investigate what command failed and why it failed.
If you run a shell script, it may be difficult to understand what command failed.

Please see the following screenshot. _[ref](https://github.com/suzuki-shunsuke/github-action-validate-envoy-proxy)_

You would be able to understand the failed command and its log quickly.

![image](https://user-images.githubusercontent.com/13323303/146356131-27d9ae75-1c61-4ec0-9f1f-f4f6f15b6b05.png)

You can also post a guide. _[ref](https://suzuki-shunsuke.github.io/tfaction/docs/feature/follow-up-pr)_

![image](https://user-images.githubusercontent.com/13323303/161662650-264959f8-fee7-46fb-bf9b-865b25ba70f9.png)

By the guide, users would be able to solve the problem by themselves.
This improves the productivity of both Developer and DevOps Engineer.

github-comment also supports hiding old comments, which makes pull requests clean.

## Features

* Easy to install
  * github-comment is a single executable binary written in Go. No dependency is needed
* Post comments according to the result of command
  * For example, you can post comment only when the command failed or the specific phrase is outputted
* Hide old comments conditionally
* Get CI information automatically from CI Environment variables
* Manage configuration with YAML
  * You can separate CI script and github-comment configuration
* Provide built-in templates

## Blog

Written in Japanese. 

* [2021-12-28 github-comment で PR にコメントをして CI の結果を分かりやすくする](https://zenn.dev/shunsuke_suzuki/articles/improve-cicd-with-github-comment)
* [2020-07-31 github-comment - GitHub にコメントを投稿する CLI](https://techblog.szksh.cloud/github-comment/)

## Release Notes

[GitHub Releases](https://github.com/suzuki-shunsuke/github-comment/releases)
