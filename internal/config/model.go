package config

// Config is nightcap.yml in the current folder: shared limits plus one Target per plan file.
type Config struct {
	MaxTurns       int      `yaml:"max_turns"`
	MaxRounds      int      `yaml:"max_rounds"`
	WorkStart      string   `yaml:"work_start"`       // "HH:MM" local time
	LimitSession   int      `yaml:"limit_session"`    // 5h window % above which no new unit starts; 0 means no cap
	LimitWeeklyAll int      `yaml:"limit_weekly_all"` // weekly all-models % above which no new unit starts; 0 means no cap
	AllowedTools   []string `yaml:"allowed_tools"`    // claude --allowedTools; read, edit, and only local commands
	Webhook        string   `yaml:"webhook"`          // Discord or Slack webhook URL for the end-of-night message; empty sends nothing
	Lang           []string `yaml:"lang"`             // PR body languages, e.g. [ko] or [en, ko]
	MergeMethod    string   `yaml:"merge_method"`     // how a 🤖 PR auto-merges: rebase, otherwise merge
	Targets        []Target `yaml:"targets"`
}

// Target pairs one plan file with the repo its units commit to.
type Target struct {
	Plan         string   `yaml:"plan"`
	Repo         string   `yaml:"repo"`
	Check        string   `yaml:"check"`
	ReadCode     []string `yaml:"read_code"`    // path globs that always send a PR to read-code
	Protected    []string `yaml:"protected"`    // contract file globs the agent may add but not edit or delete
	Dependencies []string `yaml:"dependencies"` // dependency file globs the agent may not touch
}
