package comment

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

type Mock struct {
	Stderr io.Writer
}

func (mock Mock) Create(ctx context.Context, cmt Comment) error {
	msg := "[github-comment][DRYRUN] Comment to " + cmt.Org + "/" + cmt.Repo + " sha1:" + cmt.SHA1
	if cmt.PRNumber != 0 {
		msg += " issue:" + strconv.Itoa(cmt.PRNumber)
	}
	fmt.Fprintln(mock.Stderr, msg+"\n[github-comment][DRYRUN] "+cmt.Body)
	return nil
}
