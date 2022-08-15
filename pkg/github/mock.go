package github

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

type Mock struct {
	Stderr   io.Writer
	Silent   bool
	Login    string
	PRNumber int
}

func (mock *Mock) CreateComment(ctx context.Context, cmt *Comment) error {
	if mock.Silent {
		return nil
	}
	msg := "[github-comment][DRYRUN] Comment to " + cmt.Org + "/" + cmt.Repo + " sha1:" + cmt.SHA1
	if cmt.PRNumber != 0 {
		msg += " issue:" + strconv.Itoa(cmt.PRNumber)
	}
	fmt.Fprintln(mock.Stderr, msg+"\n[github-comment][DRYRUN] "+cmt.Body)
	return nil
}

func (mock *Mock) HideComment(ctx context.Context, nodeID string) error {
	return nil
}

func (mock *Mock) ListComments(ctx context.Context, pr *PullRequest) ([]*IssueComment, error) {
	return nil, nil
}

func (mock *Mock) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return mock.Login, nil
}

func (mock *Mock) PRNumberWithSHA(ctx context.Context, owner, repo, sha string) (int, error) {
	return mock.PRNumber, nil
}
