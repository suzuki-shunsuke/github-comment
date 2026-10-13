package controller

import (
	"context"
	"log/slog"
	"maps"

	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type ParamListComments struct {
	Condition string
	Org       string
	Repo      string
	SHA1      string
	PRNumber  int
	Vars      map[string]any
	// ExprParams are extra parameters merged into the expr parameter map.
	// e.g. {"HideKey": hideKey} for hide, {"DeleteKey": deleteKey} for delete.
	ExprParams map[string]any
}

// listCommentsByCondition lists the node IDs of comments that match the condition.
// exclude is a predicate to skip comments that must not be processed.
// It is shared by the hide and delete commands.
func listCommentsByCondition( //nolint:funlen
	ctx context.Context,
	logger *slog.Logger,
	gh GitHub,
	expr Expr,
	param *ParamListComments,
	exclude func(*github.IssueComment, string) bool,
) ([]string, error) {
	if param.Condition == "" {
		logger.Debug("the condition to list comments isn't set")
		return nil, nil
	}
	login, err := gh.GetAuthenticatedUser(ctx)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get an authenticated user")
	}

	comments, err := gh.ListComments(ctx, &github.PullRequest{
		Org:      param.Org,
		Repo:     param.Repo,
		PRNumber: param.PRNumber,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	logger.Debug("get comments",
		"count", len(comments),
		"org", param.Org,
		"repo", param.Repo,
		"pr_number", param.PRNumber,
	)

	nodeIDs := []string{}
	prg, err := expr.Compile(param.Condition)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, comment := range comments {
		nodeID := comment.ID
		if exclude(comment, login) {
			logger.Debug("exclude a comment",
				"node_id", nodeID,
				"login", login,
			)
			continue
		}

		metadata := map[string]any{}
		hasMeta := extractMetaFromComment(comment.Body, &metadata)
		paramMap := map[string]any{
			"Comment": map[string]any{
				"Body": comment.Body,
				// "CreatedAt": comment.CreatedAt,
				"Meta":    metadata,
				"HasMeta": hasMeta,
			},
			"Commit": map[string]any{
				"Org":      param.Org,
				"Repo":     param.Repo,
				"PRNumber": param.PRNumber,
				"SHA1":     param.SHA1,
			},
			"Vars": param.Vars,
		}
		maps.Copy(paramMap, param.ExprParams)

		logger.Debug("judge whether an existing comment matches the condition",
			"node_id", nodeID,
			"condition", param.Condition,
			"param", paramMap,
		)
		f, err := prg.Run(paramMap)
		if err != nil {
			slogerr.WithError(logger, err).Error("judge whether an existing comment matches the condition",
				"node_id", nodeID,
			)
			continue
		}
		if !f {
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, nil
}
