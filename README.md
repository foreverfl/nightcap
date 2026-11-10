# nightcap

English | [한국어](README-ko.md) | [日本語](README-ja.md)

Runs Claude Code overnight down a plan file, one commit at a time and one PR after another, and leaves the PRs stacked for a person to merge in the morning.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/foreverfl/nightcap/main/install.sh | sh
```

It puts the latest release in `~/.local/bin/nightcap` and adds that folder to PATH in `~/.zshrc` or `~/.bashrc` when it is missing. `curl … | NIGHTCAP_VERSION=v0.1.0 sh` installs one release instead. With Go, `go install .` in a clone works too.

## Usage

```sh
nightcap run ~/work/todolist/blog-v2/commit-bot.md     # runs every open commit, PR by PR, until nothing is left, a commit fails or usage runs out
nightcap status ~/work/todolist/blog-v2/commit-bot.md  # prints the plan's repo and its worktrees
nightcap lint ~/work/todolist/blog-v2/commit-bot.md    # prints what a night run cannot use
nightcap init                                          # writes the example config to ~/.config/nightcap/nightcap.yml
nightcap config                                        # opens that file in $VISUAL or $EDITOR
nightcap version
```

## Setup

### Plan file

A plan file looks like this; copy [examples/plan.md](examples/plan.md) to start, and see [docs/plan-format.md](docs/plan-format.md) for the rules. All four frontmatter keys are required; write `[]` for a list you want empty:

```text
---
repo: ~/workspace/apps-web/notes
test: make check                                 # the test every commit must pass
protected: ["*.hurl", "*_test.go", "Makefile"]   # contract files the agent may add but not edit or delete
dependencies: [go.mod, go.sum]                   # dependency files the agent may not touch
---

# notes app — commits

## 📄 feat: list and add notes (notes-api)

Notes can be listed and added over HTTP, so the web page has something to call.

- [ ] list notes
  - WHAT  return every note as JSON, newest first
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  hurl hurl/notes/GET.hurl green

- [ ] add a note
  - WHAT  save a note from a JSON body and answer 201
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  go test ./internal/note passes
```

How a `protected` or `dependencies` pattern matches a file:

- Without `/`: the file name in any folder. `*_test.go` catches `internal/feed/feed_test.go`
- With `/`: the whole path from the repo root
- `*`: one folder. `services/*/Cargo.toml` catches `services/rust/Cargo.toml`, not `services/rust/src/main.rs`
- `**`: any number of folders. `services/**/*_test.go` catches a test file at any depth under `services/`

### nightcap.yml

`nightcap init` writes [examples/nightcap.yml](examples/nightcap.yml) to `~/.config/nightcap/nightcap.yml`, and `nightcap config` opens it. It holds the settings every plan shares on this machine.

nightcap writes nothing into your repo but the branch: state and logs go under `~/.local/share/nightcap/<repo>`, worktrees under `~/.cache/nightcap/<repo>`. `XDG_CONFIG_HOME`, `XDG_DATA_HOME` and `XDG_CACHE_HOME` move them when set.

```yaml
work_start: "09:00"        # leave empty for no before-work limit; when set, no new usage window opens within 5h of it
limit_session: 80          # no new commit starts while the 5h window is above this %
limit_weekly_all: 80       # no new commit starts while the weekly % is above this; ignored within 24h of its reset
max_turns: 30
max_rounds: 3
webhook: https://discord.com/api/webhooks/…   # Discord or Slack (hooks.slack.com); one message when the night ends, empty for none
lang: [en, ko]             # PR body languages, default [ko]
merge_method: rebase       # how a 🤖 PR auto-merges: rebase, otherwise merge. default merge
```

Before the first 🤖 PR, turn auto-merge on in the repo once. With `merge_method: rebase`, allow rebase merges too:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```
