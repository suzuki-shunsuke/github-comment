package comment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/suzuki-shunsuke/go-httpclient/httpclient"
)

type Comment struct {
	PRNumber       int
	Org            string
	Repo           string
	Body           string
	BodyForTooLong string
	SHA1           string
}

type Commenter struct {
	HTTPClient httpclient.Client
	Token      string
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
