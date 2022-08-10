package github

import (
	"context"
	"fmt"
)

func (client *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := client.user.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("get an authenticated user by GitHub API: %w", err)
	}
	return user.GetLogin(), nil
}
