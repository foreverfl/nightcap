# nightcap

Runs Claude Code overnight against one heading of a plan file, one unit at a time, and opens a PR that a person merges in the morning.

- Contract files are create-only: the agent may add a test, a hurl or a spec, never edit or delete one.
- Dependencies are yours: a change to go.mod, Cargo.toml or package.json is a red run, not a PR.
- The heading picks the label: `🤖` auto-merges (`nightcap:auto`), `📄` read the PR (`nightcap:read-pr`), `👤` a person does it. What nightcap notices at night goes into the risk list, not the label.

## Setup

Plan files follow [docs/plan-format.md](docs/plan-format.md); copy [examples/plan.md](examples/plan.md) to start.

Put this at the top of your plan file. All four keys are required; write `[]` for a list you want empty:

```yaml
---
repo: ~/workspace/apps-web/blog-v2
test: make check                                         # the check every unit must pass
protected: ["*.tsp", "*.hurl", "*_test.go", "Makefile"]  # contract files the agent may add but not edit or delete
dependencies: [go.mod, go.sum]                           # dependency files the agent may not touch
---
```

Copy [examples/nightcap.yml](examples/nightcap.yml) next to the binary as `nightcap.yml` (it is gitignored). It holds the settings every plan shares on this machine.

Then run it from that folder with the plan file:

```sh
nightcap run ~/work/todolist/blog-v2/commit-bot.md
```

Before the first 🤖 heading, turn auto-merge on in the repo once. With `merge_method: rebase`, allow rebase merges too:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```
