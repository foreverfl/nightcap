# nightcap

[English](README.md) | [한국어](README-ko.md) | 日本語

夜のあいだに Claude Code で計画ファイルを上から、コミットを一つずつ、PR を一つずつ実装し、朝に人がマージする PR を stack にして残す。

## インストール

```sh
curl -fsSL https://raw.githubusercontent.com/foreverfl/nightcap/main/install.sh | sh
```

最新リリースを `~/.local/bin/nightcap` に置き、そのフォルダが PATH になければ `~/.zshrc` か `~/.bashrc` に一行足す。特定のリリースは `curl … | NIGHTCAP_VERSION=v0.1.0 sh` のように sh の前に付ける。Go があれば、クローンしたフォルダで `go install .` でもいい。

## 使い方

```sh
nightcap run ~/work/todolist/blog-v2/commit-bot.md     # 残っているコミットを PR の順にすべて実行する。終わるか、コミットが失敗するか、使用量が上限に達すると止まる
nightcap status ~/work/todolist/blog-v2/commit-bot.md  # 計画のリポジトリとその worktree を表示する
nightcap lint ~/work/todolist/blog-v2/commit-bot.md    # 夜の実行で使えないものを表示する
nightcap clean                                         # nightcap が ~/.cache/nightcap に残した worktree をすべて削除する
nightcap init                                          # 設定の例を ~/.config/nightcap/nightcap.yml に書く
nightcap config                                        # そのファイルを $VISUAL か $EDITOR で開く
nightcap update                                        # install.sh と同じ方法で ~/.local/bin/nightcap を最新リリースに置き換える
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

`nightcap init` が [examples/nightcap.yml](examples/nightcap.yml) を `~/.config/nightcap/nightcap.yml` に書き、`nightcap config` がそのファイルを開く。このマシンのすべての計画が共有する設定だけを置く。

```yaml
work_start: "09:00"        # 空なら出勤前の制限なし。書くと、この時刻の 5 時間以内には新しい使用量ウィンドウを開かない
limit_session: 80          # 5 時間ウィンドウの使用量がこの % を超えると、新しいコミットを始めない
limit_weekly_all: 80       # 週間の使用量がこの % を超えると、新しいコミットを始めない。リセットの 24 時間以内なら無視
max_turns: 30
max_rounds: 3
webhook: https://discord.com/api/webhooks/…   # Discord か Slack（hooks.slack.com）。夜が終わるとメッセージを一つ、空なら送らない
lang: [en, ko]             # PR 本文の言語、既定は [ko]
merge_method: rebase       # 🤖 PR の auto-merge の方法: rebase、それ以外は merge。既定は merge
# author_name: "nightcap/<自分の GitHub ID>"                         # 夜のコミットの作成者名、git blame に出る。author_email と一緒に書く
# author_email: "<ボット ID>+<アプリ名>[bot]@users.noreply.github.com"  # ボット ID: gh api 'users/<アプリ名>[bot]' --jq .id
```

`author_name` と `author_email` は夜のコミットの作成者を決める。名前を `nightcap/<自分の GitHub ID>` にすると、`git blame` とエディタに夜に回した人が出る。メールはボットのものを使うので、GitHub では引き続きボットのコミットになる。両方書くか、両方空にする。片方だけだと両方無視され、gitconfig の作成者が使われる。

最初の 🤖 PR の前に、リポジトリで auto-merge を一度オンにする。`merge_method: rebase` なら rebase マージも許可する:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```

### ファイルの場所

nightcap は対象リポジトリにブランチ以外何も書かない。それ以外はすべてここにある:

```text
~/.local/bin/nightcap                       実行ファイル。install.sh と nightcap update がここに置く
~/.config/nightcap/nightcap.yml             設定。nightcap init が作り、nightcap config が開く
~/.local/share/nightcap/<リポジトリのフォルダ名>/
  state.json                                コミットごとの結果。nightcap clean が消す
  runs/261111/231635/                       nightcap run 1 回につき 1 フォルダ、開始時刻が名前
    run.log                                 実行が画面に出した行。朝に読むのはこのファイル
    events.jsonl                            いつ何が起きたか、1 行に JSON 1 つ
    <slug>/commit1-round1.jsonl             コミット 1 つの 1 ラウンドで claude が出した生の出力
~/.cache/nightcap/<リポジトリのフォルダ名>/worktrees/<slug>   PR を作る worktree。nightcap clean が消す
```

`XDG_CONFIG_HOME` · `XDG_DATA_HOME` · `XDG_CACHE_HOME` があれば、設定・記録・worktree のフォルダはその下に置く。旧バージョンが記録フォルダに残した `logs/` と `runs/<日付>/<時刻>.jsonl` はそのまま残す。
