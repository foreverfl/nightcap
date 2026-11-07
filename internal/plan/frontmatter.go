package plan

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseMeta reads the frontmatter between a leading "---" line and the next "---" line of a plan file.
// A file that does not start with "---" has an empty Meta. It returns an error when the closing "---" is missing or the yaml is malformed.
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
