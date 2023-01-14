package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	issue IssuesService
	pr    PullRequestsService
	repo  RepositoriesService
	user  UsersService
	ghV4  V4Client
}

type ParamNew struct {
	Token              string
	GHEBaseURL         string
	GHEGraphQLEndpoint string
}

func New(ctx context.Context, param *ParamNew) (*Client, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: param.Token},
	))
	client := &Client{}
	if param.GHEBaseURL == "" {
		gh := github.NewClient(httpClient)
		client.issue = gh.Issues
		client.repo = gh.Repositories
		client.user = gh.Users
		client.pr = gh.PullRequests
	} else {
		gh, err := github.NewEnterpriseClient(param.GHEBaseURL, param.GHEBaseURL, httpClient)
		if err != nil {
			return nil, fmt.Errorf("initialize GitHub Enterprise API Client: %w", err)
		}
		client.issue = gh.Issues
		client.repo = gh.Repositories
		client.user = gh.Users
		client.pr = gh.PullRequests
	}
	if param.GHEGraphQLEndpoint == "" {
		client.ghV4 = githubv4.NewClient(httpClient)
	} else {
		client.ghV4 = githubv4.NewEnterpriseClient(param.GHEGraphQLEndpoint, httpClient)
	}

	return client, nil
}

type V4Client interface {
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

type IssuesService interface {
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	EditComment(ctx context.Context, owner string, repo string, commentID int64, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

type RepositoriesService interface {
	CreateComment(ctx context.Context, owner, repo, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error)
	UpdateComment(ctx context.Context, owner, repo string, id int64, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error)
}

type UsersService interface {
	Get(ctx context.Context, user string) (*github.User, *github.Response, error)
}

type PullRequestsService interface {
	ListPullRequestsWithCommit(ctx context.Context, owner, repo, sha string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}
