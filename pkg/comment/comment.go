package comment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/go-httpclient/httpclient"
	"golang.org/x/oauth2"
)

type Comment struct {
	PRNumber       int
	Org            string
	Repo           string
	Body           string
	BodyForTooLong string
	SHA1           string
	Minimize       string
	TemplateKey    string
	Vars           map[string]interface{}
}

type Commenter struct {
	HTTPClient httpclient.Client
	Token      string

	V4Client *githubv4.Client
}

func New(ctx context.Context, token string) Commenter {
	return Commenter{
		Token:      token,
		HTTPClient: httpclient.New("https://api.github.com"),
		V4Client: githubv4.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))),
	}
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

type ValidationError struct {
	Message string `json:"message"`
}

func (commenter Commenter) getPath(cmt Comment) string {
	if cmt.PRNumber != 0 {
		return "/repos/" + cmt.Org + "/" + cmt.Repo + "/issues/" + strconv.Itoa(cmt.PRNumber) + "/comments"
	}
	return "/repos/" + cmt.Org + "/" + cmt.Repo + "/commits/" + cmt.SHA1 + "/comments"
}

func (commenter Commenter) create(ctx context.Context, cmt Comment, tooLong bool) error {
	body := cmt.Body
	if tooLong {
		body = cmt.BodyForTooLong
	}
	_, err := commenter.HTTPClient.Call(ctx, httpclient.CallParams{ //nolint:bodyclose
		Method: http.MethodPost,
		Path:   commenter.getPath(cmt),
		Header: http.Header{
			"Authorization": []string{"token " + commenter.Token},
		},
		RequestBody: map[string]string{
			"body": body,
		},
	})
	if err != nil {
		return fmt.Errorf("send a comment by GitHub API: %w", err)
	}
	return nil
}

func (commenter Commenter) Create(ctx context.Context, cmt Comment) error {
	err := commenter.create(ctx, cmt, false)
	if cmt.BodyForTooLong == "" {
		return err
	}
	e := &httpclient.Error{}
	if errors.As(err, &e) {
		validationErrors := ValidationErrors{}
		if err := json.Unmarshal(e.BodyByte(), &validationErrors); err == nil {
			for _, ve := range validationErrors.Errors {
				if strings.HasPrefix(ve.Message, "Body is too long") {
					return commenter.create(ctx, cmt, true)
				}
			}
		}
	}
	return err
}

type PullRequest struct {
	PRNumber int
	Org      string
	Repo     string
}

func (commenter Commenter) list(ctx context.Context, pr PullRequest, page int) ([]map[string]interface{}, error) {
	cmts := []map[string]interface{}{}
	_, err := commenter.HTTPClient.Call(ctx, httpclient.CallParams{ //nolint:bodyclose
		Method: http.MethodGet,
		Path:   "/repos/" + pr.Org + "/" + pr.Repo + "/issues/" + strconv.Itoa(pr.PRNumber) + "/comments",
		Query: url.Values{
			"per_page": []string{"100"},
			"page":     []string{strconv.Itoa(page)},
		},
		Header: http.Header{
			"Authorization": []string{"token " + commenter.Token},
		},
		ResponseBody: &cmts,
	})
	if err != nil {
		return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
	}
	return cmts, nil
}

const (
	paginationLimit = 100
	perPage         = 100
)

func (commenter Commenter) ListV3(ctx context.Context, pr PullRequest) ([]map[string]interface{}, error) {
	cmts := []map[string]interface{}{}
	for page := 1; page <= paginationLimit; page++ {
		comments, err := commenter.list(ctx, pr, page)
		if err != nil {
			return nil, err
		}
		cmts = append(cmts, comments...)
		if len(comments) < perPage {
			return cmts, nil
		}
	}
	return cmts, nil
}

// `graphql:"IssueComment(isMinimized: false, viewerCanMinimize: true)"`
type IssueComment struct {
	ID     string
	Body   string
	Author struct {
		Login string
	}
	CreatedAt string
	// TODO remove
	IsMinimized       bool
	ViewerCanMinimize bool
}

func (commenter Commenter) List(ctx context.Context, pr PullRequest) ([]IssueComment, error) {
	// https://github.com/shurcooL/githubv4#pagination
	var q struct {
		Repository struct {
			Issue struct {
				Comments struct {
					Nodes    []IssueComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(pr.Org),
		"repositoryName":  githubv4.String(pr.Repo),
		"issueNumber":     githubv4.Int(pr.PRNumber),
		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var allComments []IssueComment
	for {
		if err := commenter.V4Client.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
		}
		allComments = append(allComments, q.Repository.Issue.Comments.Nodes...)
		if !q.Repository.Issue.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.Issue.Comments.PageInfo.EndCursor)
	}
	return allComments, nil
}

func (commenter Commenter) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user := struct {
		Login string `json:"login"`
	}{}
	_, err := commenter.HTTPClient.Call(ctx, httpclient.CallParams{ //nolint:bodyclose
		Method: http.MethodGet,
		Path:   "/user",
		Header: http.Header{
			"Authorization": []string{"token " + commenter.Token},
		},
		ResponseBody: &user,
	})
	if err != nil {
		return "", fmt.Errorf("get an anuthenticated user by GitHub API: %w", err)
	}
	return user.Login, nil
}
