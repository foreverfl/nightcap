# nightcap

Runs Claude Code overnight against one heading of a plan file, one unit at a time, and opens a PR that a person merges in the morning.

- Contract files are create-only: the agent may add a test, a hurl or a spec, never edit or delete one.
- Dependencies are yours: a change to go.mod, Cargo.toml or package.json is a red run, not a PR.
- The heading picks the label: `🤖` auto-merges (`nightcap:auto`), `📄` read the PR (`nightcap:read-pr`), `👀` read the code (`nightcap:read-code`), `👤` a person does it. What nightcap notices at night goes into the risk list, not the label.

## Setup

Plan files follow [docs/plan-format.md](docs/plan-format.md); copy [examples/plan.md](examples/plan.md) to start.

Write `nightcap.yml` next to the binary (it is gitignored) and run `nightcap run` from that folder. One target per plan file:

```yaml
work_start: "09:00"        # leave empty for no before-work limit; when set, no new usage window opens within 5h of it
limit_session: 80          # no new unit starts while the 5h window is above this %
limit_weekly_all: 80       # no new unit starts while the weekly % is above this; ignored within 24h of its reset
max_turns: 30
max_rounds: 3
webhook: https://discord.com/api/webhooks/…   # Discord or Slack (hooks.slack.com); one message when the night ends, empty for none
targets:
  - plan: ~/work/todolist/blog-v2/github-rank-commit.md
    repo: ~/workspace/apps-web/blog-v2
    check: make check
    read_code: [services/go-api/internal/auth/*]
    lang: [en, ko]            # PR body languages, default [ko]
    merge_method: rebase      # how a 🤖 PR auto-merges: rebase, otherwise merge. default merge
```
