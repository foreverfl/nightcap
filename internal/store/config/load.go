package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

const FileName = "nightcap.yml"

// defaultAllowedTools lets the agent read, edit and run local checks and commits, but not push.
var defaultAllowedTools = []string{
	"Read", "Edit", "Write", "Glob", "Grep",
	"Bash(go:*)", "Bash(make:*)", "Bash(hurl:*)", "Bash(cargo:*)", "Bash(npm test:*)",
	"Bash(git status:*)", "Bash(git diff:*)", "Bash(git log:*)", "Bash(git add:*)", "Bash(git commit:*)",
}

// Load reads the nightcap.yml at path and fills missing keys with defaults: max_turns 30, max_rounds 3, lang [ko], merge_method merge.
// An empty work_start turns the before-work limit off.
func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("config: %s not found, run nightcap init", path)
	}
	if err != nil {
		return Config{}, err
	}
	config := Config{MaxTurns: 30, MaxRounds: 3, AllowedTools: defaultAllowedTools, Lang: []string{"ko"}, MergeMethod: "merge"}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return Config{}, fmt.Errorf("%s: %w", FileName, err)
	}
	return config, nil
}
