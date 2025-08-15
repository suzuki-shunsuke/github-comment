---
sidebar_position: 1000
---

# GitHub Enterprise Support

:::warning
I ([@suzuki-shunsuke](http://github.com/suzuki-shunsuke)) don't use GitHub Enterprise, so I can't confirm if github-comment works well for GitHub Enterprise.
:::

github-comment >= [v4.2.0](https://github.com/suzuki-shunsuke/github-comment/releases/tag/v4.2.0)

[#462](https://github.com/suzuki-shunsuke/github-comment/issues/462) [#464](https://github.com/suzuki-shunsuke/github-comment/issues/464)

:::note
From github-comment [v6.2.0](https://github.com/suzuki-shunsuke/github-comment/releases/tag/v6.2.0),
github-comment gets GitHub API endpoints from environment variables `GITHUB_API_URL` and `GITHUB_GRAPHQL_URL`, which are built-in variables of GitHub Actions.
:::

Please set the following fields in configuration file `github-comment.yaml`.

GitHub Enterprise Server
```yaml
ghe_base_url: http(s)://<your_enterprise_hostname> # CHANGE
ghe_graphql_endpoint: http(s)://<your_enterprise_hostname>/api/graphql # CHANGE
```

See. https://docs.github.com/en/enterprise-server/graphql/guides/forming-calls-with-graphql#the-graphql-endpoint

GitHub Enterprise Cloud
```yaml
ghe_base_url: https://api.github.com
ghe_graphql_endpoint: https://api.github.com/graphql
```

See. https://docs.github.com/en/enterprise-cloud@latest/graphql/guides/forming-calls-with-graphql#the-graphql-endpoint

- https://docs.github.com/en/enterprise-server@3.5/rest/overview/resources-in-the-rest-api#current-version
- https://docs.github.com/en/enterprise-server@2.20/graphql/guides/forming-calls-with-graphql#the-graphql-endpoint
