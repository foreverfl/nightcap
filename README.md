# nightcap

Runs Claude Code overnight against one heading of a plan file, one unit at a time, and opens a PR that a person merges in the morning.

- Contract files are create-only: the agent may add a test, a hurl or a spec, never edit or delete one.
- Dependencies are yours: a change to go.mod, Cargo.toml or package.json is a red run, not a PR.
- Read the PR, not the code: `nightcap:read-pr` means the body and the risk list are enough, `nightcap:read-code` means read the diff.
