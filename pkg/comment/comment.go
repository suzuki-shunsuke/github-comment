package comment

import (
	"context"
	"net/http"
	"strconv"

	"github.com/suzuki-shunsuke/go-httpclient/httpclient"
)

type Comment struct {
	PRNumber int
	Org      string
	Repo     string
	Body     string
	SHA1     string
}

type Commenter struct {
	HTTPClient *httpclient.Client
	Token      string
}

func (commenter Commenter) getPath(cmt Comment) string {
	if cmt.SHA1 == "" {
		return "/repos/" + cmt.Org + "/" + cmt.Repo + "/issues/" + strconv.Itoa(cmt.PRNumber) + "/comments"
	}
	return "/repos/" + cmt.Org + "/" + cmt.Repo + "/commits/" + cmt.SHA1 + "/comments"
}

func (commenter Commenter) Create(ctx context.Context, cmt Comment) error {
	_, err := commenter.HTTPClient.Call(ctx, httpclient.CallParams{
		Method: http.MethodPost,
		Path:   commenter.getPath(cmt),
		Header: http.Header{
			"Authorization": []string{"token " + commenter.Token},
		},
		RequestBody: map[string]string{
			"body": cmt.Body,
		},
	})
	return err
}
