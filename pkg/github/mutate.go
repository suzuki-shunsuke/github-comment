package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

func (c *Client) HideComment(ctx context.Context, nodeID string) error {
	var m struct {
		MinimizeComment struct {
			MinimizedComment struct {
				MinimizedReason   githubv4.String
				IsMinimized       githubv4.Boolean
				ViewerCanMinimize githubv4.Boolean
			}
		} `graphql:"minimizeComment(input:$input)"`
	}
	input := githubv4.MinimizeCommentInput{
		Classifier: githubv4.ReportedContentClassifiersOutdated,
		SubjectID:  nodeID,
	}
	if err := c.ghV4.Mutate(ctx, &m, input, nil); err != nil {
		return fmt.Errorf("hide an old comment: %w", err)
	}
	return nil
}
