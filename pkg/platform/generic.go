package platform

import (
	"fmt"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
)

type Param struct {
	RepoOwner []domain.ComplementEntry
	RepoName  []domain.ComplementEntry
	SHA       []domain.ComplementEntry
	PRNumber  []domain.ComplementEntry
}

type Client struct {
	param Param
}

func New(param Param) Client {
	return Client{
		param: param,
	}
}

func (client *Client) render(entries []domain.ComplementEntry) (string, error) {
	for _, entry := range entries {
		a, err := entry.Entry()
		if err != nil {
			return "", err //nolint:wrapcheck
		}
		if a != "" {
			return a, nil
		}
	}
	return "", nil
}

func (client *Client) returnString(entries []domain.ComplementEntry) string {
	s, err := client.render(entries)
	if err != nil {
		return ""
	}
	return s
}

func (client *Client) RepoOwner() string {
	return client.returnString(client.param.RepoOwner)
}

func (client *Client) RepoName() string {
	return client.returnString(client.param.RepoName)
}

func (client *Client) SHA() string {
	return client.returnString(client.param.SHA)
}

func (client *Client) IsPR() bool {
	return client.returnString(client.param.PRNumber) != ""
}

func (client *Client) PRNumber() (int, error) {
	s, err := client.render(client.param.PRNumber)
	if err != nil {
		return 0, err
	}
	if s == "" {
		return 0, nil
	}
	b, err := strconv.Atoi(s)
	if err == nil {
		return b, nil
	}
	return 0, fmt.Errorf("parse pull request number as int: %w", err)
}
