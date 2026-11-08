# nightcap

[English](README.md) | [한국어](README-ko.md) | 日本語

計画ファイルの PR を一つ選び、夜のあいだに Claude Code でコミットを一つずつ実装して、朝に人がマージする PR を開く。

## インストール

```sh
curl -fsSL https://raw.githubusercontent.com/foreverfl/nightcap/main/install.sh | sh
```

最新リリースを `~/.local/bin/nightcap` に置き、そのフォルダが PATH になければ `~/.zshrc` か `~/.bashrc` に一行足す。特定のリリースは `curl … | NIGHTCAP_VERSION=v0.1.0 sh` のように sh の前に付ける。Go があれば、クローンしたフォルダで `go install .` でもいい。

## 使い方

```sh
nightcap run ~/work/todolist/blog-v2/commit-bot.md     # コミットが残っている一番上の PR の、次のコミットを実行する
nightcap status ~/work/todolist/blog-v2/commit-bot.md  # 計画のリポジトリとその worktree を表示する
nightcap lint ~/work/todolist/blog-v2/commit-bot.md    # 夜の実行で使えないものを表示する
nightcap version
```

## 設定

### 計画ファイル

計画ファイルはこういう形をしている。[examples/plan.md](examples/plan.md) をコピーして始め、ルールは [docs/plan-format.md](docs/plan-format.md) にある。frontmatter の四つのキーはすべて必須で、空にするリストは `[]` と書く:

```text
---
repo: ~/workspace/apps-web/notes
test: make check                                 # どのコミットも通らなければならないテスト
protected: ["*.hurl", "*_test.go", "Makefile"]   # エージェントが追加はできても、編集も削除もできない契約ファイル
dependencies: [go.mod, go.sum]                   # エージェントが触れない依存ファイル
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

### nightcap.yml

[examples/nightcap.yml](examples/nightcap.yml) を `~/.config/nightcap/nightcap.yml` にコピーする。このマシンのすべての計画が共有する設定だけを置く。

nightcap は対象リポジトリにブランチ以外何も書かない。状態とログは `~/.local/share/nightcap/<リポジトリ>`、worktree は `~/.cache/nightcap/<リポジトリ>` の下にできる。`XDG_CONFIG_HOME` · `XDG_DATA_HOME` · `XDG_CACHE_HOME` があればその下に置く。

```yaml
work_start: "09:00"        # 空なら出勤前の制限なし。書くと、この時刻の 5 時間以内には新しい使用量ウィンドウを開かない
limit_session: 80          # 5 時間ウィンドウの使用量がこの % を超えると、新しいコミットを始めない
limit_weekly_all: 80       # 週間の使用量がこの % を超えると、新しいコミットを始めない。リセットの 24 時間以内なら無視
max_turns: 30
max_rounds: 3
webhook: https://discord.com/api/webhooks/…   # Discord か Slack（hooks.slack.com）。夜が終わるとメッセージを一つ、空なら送らない
lang: [en, ko]             # PR 本文の言語、既定は [ko]
merge_method: rebase       # 🤖 PR の auto-merge の方法: rebase、それ以外は merge。既定は merge
```

最初の 🤖 PR の前に、リポジトリで auto-merge を一度オンにする。`merge_method: rebase` なら rebase マージも許可する:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```
