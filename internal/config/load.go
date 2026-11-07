package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const FileName = "nightcap.yml"

// defaultAllowedTools lets the agent read, edit and run local checks and commits, but not push.
var defaultAllowedTools = []string{
	"Read", "Edit", "Write", "Glob", "Grep",
	"Bash(go:*)", "Bash(make:*)", "Bash(hurl:*)", "Bash(cargo:*)", "Bash(npm test:*)",
	"Bash(git status:*)", "Bash(git diff:*)", "Bash(git log:*)", "Bash(git add:*)", "Bash(git commit:*)",
}

// defaultProtected are the contract files a target gets when nightcap.yml names none.
var defaultProtected = []string{"*_test.go", "*.hurl", "openapi/*", "Makefile"}

// defaultDependencies are the dependency files a target gets when nightcap.yml names none.
var defaultDependencies = []string{
	"go.mod", "go.sum",
	"Cargo.toml", "Cargo.lock",
	"package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
}

// Load reads <dir>/nightcap.yml, fills missing keys with defaults
// (max_turns 30, max_rounds 3, lang [ko], merge_method merge, protected and dependencies per target; an empty work_start turns the before-work limit off) and expands a leading "~/" in plan and repo.
// It returns an error when the file is unreadable or malformed, targets is empty,
// or a target lacks plan, repo or check.
func Load(dir string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		return Config{}, err
	}
	config := Config{MaxTurns: 30, MaxRounds: 3, AllowedTools: defaultAllowedTools, Lang: []string{"ko"}, MergeMethod: "merge"}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return Config{}, fmt.Errorf("%s: %w", FileName, err)
	}
	if len(config.Targets) == 0 {
		return Config{}, errors.New(FileName + ": targets is required")
	}
	for i := range config.Targets {
		target := &config.Targets[i]
		if target.Plan == "" || target.Repo == "" || target.Check == "" {
			return Config{}, fmt.Errorf("%s: targets[%d] needs plan, repo and check", FileName, i)
		}
		target.Plan = expandHome(target.Plan)
		target.Repo = expandHome(target.Repo)
		if len(target.Protected) == 0 {
			target.Protected = defaultProtected
		}
		if len(target.Dependencies) == 0 {
			target.Dependencies = defaultDependencies
		}
	}
	return config, nil
}

// expandHome replaces a leading "~/" with the home directory; any other path comes back unchanged.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
