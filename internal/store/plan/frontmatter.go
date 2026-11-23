package plan

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseMeta returns the frontmatter between the plan's leading "---" line and the next one; a file without one has an empty Meta.
// It fails when the closing "---" is missing or the yaml is malformed.
func ParseMeta(text string) (Meta, error) {
	var meta Meta
	rest, hasFrontmatter := strings.CutPrefix(text, "---\n")
	if !hasFrontmatter {
		return meta, nil
	}
	frontmatter, _, isClosed := strings.Cut(rest, "\n---\n")
	if !isClosed {
		return meta, errors.New("plan: frontmatter not closed")
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return meta, err
	}
	if repo, isHome := strings.CutPrefix(meta.Repo, "~/"); isHome {
		if home, err := os.UserHomeDir(); err == nil {
			meta.Repo = filepath.Join(home, repo)
		}
	}
	return meta, nil
}

// UnmarshalYAML reads test: as one command when it is a string, or as an area-to-command map when it is a mapping.
func (commands *TestCommands) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&commands.All)
	}
	return node.Decode(&commands.ByArea)
}

// CommandFor returns the test command for a unit whose tag has area, e.g. "back" → "cargo test".
// ok is false when test: has no line for that area.
func (commands TestCommands) CommandFor(area string) (command string, ok bool) {
	if commands.All != "" {
		return commands.All, true
	}
	command, ok = commands.ByArea[area]
	return command, ok
}

// IsEmpty reports whether test: was left out or holds nothing.
func (commands TestCommands) IsEmpty() bool {
	return commands.All == "" && len(commands.ByArea) == 0
}
