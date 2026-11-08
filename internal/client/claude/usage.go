package claude

import (
	"context"
	"os/exec"
)

// Usage runs `claude -p /usage` and returns its stream-json output for usage.Read.
func Usage(ctx context.Context) ([]byte, error) {
	return exec.CommandContext(ctx, claudeBinary(), "-p", "/usage", "--output-format", "stream-json", "--verbose").Output()
}
