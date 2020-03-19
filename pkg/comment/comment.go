package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type Comment struct {
	PRNumber int
	Org      string
	Repo     string
	Body     string
	SHA1     string
}

func Create(ctx context.Context, client *http.Client, token string, cmt *Comment) error {
	endpoint := "https://api.github.com/repos/" + cmt.Org + "/" + cmt.Repo + "/issues/" + strconv.Itoa(cmt.PRNumber) + "/comments"
	if cmt.SHA1 != "" {
		endpoint = "https://api.github.com/repos/" + cmt.Org + "/" + cmt.Repo + "/commits/" + cmt.SHA1 + "/comments"
	}
	m := map[string]string{
		"body": cmt.Body,
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&m); err != nil {
		return fmt.Errorf("failed to create a request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buf)
	if err != nil {
		return fmt.Errorf("failed to create a request: %w", err)
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request is failure: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return errors.New("failed to create a comment: status code " + strconv.Itoa(resp.StatusCode) + " >= 400: " + endpoint)
	}
	return nil
}
