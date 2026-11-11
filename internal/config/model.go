package config

// Config is nightcap.yml in the current folder: the settings every plan shares on this machine.
type Config struct {
	MaxTurns       int      `yaml:"max_turns"`
	MaxRounds      int      `yaml:"max_rounds"`
	WorkStart      string   `yaml:"work_start"`       // "HH:MM" local time
	LimitSession   int      `yaml:"limit_session"`    // 5h window % above which no new commit starts; 0 means no cap
	LimitWeeklyAll int      `yaml:"limit_weekly_all"` // weekly all-models % above which no new commit starts; 0 means no cap
	AllowedTools   []string `yaml:"allowed_tools"`    // claude --allowedTools; read, edit, and only local commands
	Webhook        string   `yaml:"webhook"`          // Discord or Slack webhook URL for the end-of-night message; empty sends nothing
	Lang           []string `yaml:"lang"`             // PR body languages, e.g. [ko] or [en, ko]
	MergeMethod    string   `yaml:"merge_method"`     // how a 🤖 PR auto-merges: rebase, otherwise merge
	AuthorName     string   `yaml:"author_name"`      // author and committer of night commits, set only with author_email; empty keeps the gitconfig
	AuthorEmail    string   `yaml:"author_email"`
}
