package claude

import (
	"context"
	_ "embed"
)

//go:embed prompts/review-order.md
var reviewOrderInstructions string

// ReviewOrder asks claude, in one turn with no tools, to write the PR's 읽는 순서 section from the branch's diff:
// one block per data flow, then a 흐름 밖 list. It returns an error when claude fails or does not finish.
func ReviewOrder(ctx context.Context, diff string) (string, error) {
	return askOnce(ctx, "review order", reviewOrderInstructions+"\n"+diff)
}
