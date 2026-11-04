# nightcap

Runs Claude Code overnight against one heading of a plan file, one unit at a time, and opens a PR that a person merges in the morning.

- Contract files are create-only: the agent may add a test, a hurl or a spec, never edit or delete one.
- Dependencies are yours: a change to go.mod, Cargo.toml or package.json is a red run, not a PR.
- Read the PR, not the code: `nightcap:read-pr` means the body and the risk list are enough, `nightcap:read-code` means read the diff.

## Setup

Plan files follow [docs/plan-format.md](docs/plan-format.md); copy [examples/plan.md](examples/plan.md) to start.

Write `nightcap.yml` next to the binary (it is gitignored) and run `nightcap run` from that folder. One target per plan file:

```yaml
work_start: "09:00"        # leave empty for no before-work limit; when set, no new usage window opens within 5h of it
usage_stop: 80             # no new unit starts once the usage window is past this %; 0 or missing means no cap
max_turns: 30
max_rounds: 3
targets:
  - plan: ~/work/todolist/blog-v2/github-rank-commit.md
    repo: ~/workspace/apps-web/blog-v2
    check: make check
    read_code: [services/go-api/internal/auth/*]
    lang: [en, ko]            # PR body languages, default [ko]
```
