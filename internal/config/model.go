package config

// Config is nightcap.yml in the current folder: shared limits plus one Target per plan file.
type Config struct {
	MaxTurns     int      `yaml:"max_turns"`
	MaxRounds    int      `yaml:"max_rounds"`
	WorkStart    string   `yaml:"work_start"`    // "HH:MM" local time
	AllowedTools []string `yaml:"allowed_tools"` // claude --allowedTools; read, edit, and only local commands
	Targets      []Target `yaml:"targets"`
}

// Target pairs one plan file with the repo its units commit to.
type Target struct {
	Plan     string   `yaml:"plan"`
	Repo     string   `yaml:"repo"`
	Check    string   `yaml:"check"`
	ReadCode []string `yaml:"read_code"` // path globs that always send a PR to read-code
	Lang     []string `yaml:"lang"`      // PR body languages, e.g. [ko] or [en, ko]
}
