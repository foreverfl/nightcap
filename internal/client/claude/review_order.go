package claude

import (
	"context"
	_ "embed"
)

//go:embed prompts/review-order.md
var reviewOrderInstructions string

// ReviewOrder asks claude to write the PR's review order from the branch's diff: one block per data flow, then what sits outside them.
func ReviewOrder(ctx context.Context, diff string) (string, error) {
	return askOnce(ctx, "review order", reviewOrderInstructions+"\n"+diff)
}
