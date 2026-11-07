package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = "nightcap.yml"

// defaultAllowedTools lets the agent read, edit and run local checks and commits, but not push.
var defaultAllowedTools = []string{
	"Read", "Edit", "Write", "Glob", "Grep",
	"Bash(go:*)", "Bash(make:*)", "Bash(hurl:*)", "Bash(cargo:*)", "Bash(npm test:*)",
	"Bash(git status:*)", "Bash(git diff:*)", "Bash(git log:*)", "Bash(git add:*)", "Bash(git commit:*)",
}

// DefaultProtected are the contract files a plan gets when its frontmatter names none.
var DefaultProtected = []string{"*_test.go", "*.hurl", "openapi/*", "Makefile"}

// DefaultDependencies are the dependency files a plan gets when its frontmatter names none.
var DefaultDependencies = []string{
	"go.mod", "go.sum",
	"Cargo.toml", "Cargo.lock",
	"package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
}

// Load reads <dir>/nightcap.yml, fills missing keys with defaults
// (max_turns 30, max_rounds 3, lang [ko], merge_method merge; an empty work_start turns the before-work limit off).
// It returns an error when the file is unreadable or malformed.
func Load(dir string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		return Config{}, err
	}
	config := Config{MaxTurns: 30, MaxRounds: 3, AllowedTools: defaultAllowedTools, Lang: []string{"ko"}, MergeMethod: "merge"}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return Config{}, fmt.Errorf("%s: %w", FileName, err)
	}
	return config, nil
}
