# nightcap

[English](README.md) | 한국어 | [日本語](README-ja.md)

밤사이 Claude Code로 계획 문서를 위에서부터 커밋 하나씩, PR 하나씩 구현하고, 아침에 사람이 합칠 PR들을 stack으로 쌓아 둔다.

## 설치

```sh
curl -fsSL https://raw.githubusercontent.com/foreverfl/nightcap/main/install.sh | sh
```

최신 릴리스를 `~/.local/bin/nightcap`에 놓고, PATH에 그 폴더가 없으면 `~/.zshrc`나 `~/.bashrc`에 한 줄을 더한다. 특정 릴리스는 `curl … | NIGHTCAP_VERSION=v0.1.0 sh`처럼 sh 앞에 붙인다. Go가 있으면 클론한 폴더에서 `go install .`도 된다.

## 사용법

```sh
nightcap run ~/work/todolist/blog-v2/commit-bot.md     # 열린 커밋을 PR 순서대로 전부 돌린다. 다 끝나거나, 커밋이 실패하거나, 사용량이 차면 멈춘다
nightcap status ~/work/todolist/blog-v2/commit-bot.md  # 계획의 레포와 그 워크트리를 찍는다
nightcap lint ~/work/todolist/blog-v2/commit-bot.md    # 밤 실행이 쓸 수 없는 것을 찍는다
nightcap init                                          # 예시 설정을 ~/.config/nightcap/nightcap.yml에 쓴다
nightcap config                                        # 그 파일을 $VISUAL이나 $EDITOR로 연다
nightcap version
```

## 설정

### 계획 파일

계획 파일은 이렇게 생겼다. [examples/plan-ko.md](examples/plan-ko.md)를 복사해서 시작하고, 규칙은 [docs/plan-format-ko.md](docs/plan-format-ko.md)에 있다. 맨 위 네 키는 모두 꼭 있어야 하고, 비울 목록은 `[]`로 적는다:

```text
---
repo: ~/workspace/apps-web/notes
test: make check                                 # 커밋마다 통과해야 하는 테스트
protected: ["*.hurl", "*_test.go", "Makefile"]   # 에이전트가 추가는 해도 고치거나 지우지 못하는 계약 파일
dependencies: [go.mod, go.sum]                   # 에이전트가 건드리지 못하는 의존성 파일
---

# 메모 앱 — 커밋 단위

## 📄 feat: 메모 목록과 추가 (notes-api)

HTTP로 메모를 보고 더할 수 있게 해서, 웹 화면이 부를 곳을 만든다.

- [ ] 메모 목록 돌려주기
  - WHAT  메모 전부를 최신순 JSON으로 돌려준다
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  hurl hurl/notes/GET.hurl 초록

- [ ] 메모 추가하기
  - WHAT  JSON 본문으로 받은 메모를 저장하고 201로 답한다
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  go test ./internal/note 통과
```

`protected`와 `dependencies` 패턴이 파일을 맞추는 방법:

- `/`가 없으면: 어느 폴더에 있든 파일 이름만. `*_test.go`는 `internal/feed/feed_test.go`를 잡는다
- `/`가 있으면: 레포 루트부터 경로 전체
- `*`: 폴더 하나. `services/*/Cargo.toml`은 `services/rust/Cargo.toml`을 잡고 `services/rust/src/main.rs`는 안 잡는다
- `**`: 폴더 몇 개든. `services/**/*_test.go`는 `services/` 밑 어느 깊이의 테스트 파일이든 잡는다

### nightcap.yml

`nightcap init`이 [examples/nightcap.yml](examples/nightcap.yml)을 `~/.config/nightcap/nightcap.yml`에 쓰고, `nightcap config`가 그 파일을 연다. 이 기계에서 모든 계획이 같이 쓰는 설정만 둔다.

nightcap은 대상 레포에 브랜치 말고 아무것도 쓰지 않는다. 상태와 로그는 `~/.local/share/nightcap/<레포>`, worktree는 `~/.cache/nightcap/<레포>` 밑에 생긴다. `XDG_CONFIG_HOME` · `XDG_DATA_HOME` · `XDG_CACHE_HOME`이 있으면 그 밑으로 간다.

```yaml
work_start: "09:00"        # 비우면 출근 전 제한 없음. 적으면 이 시각 5시간 안에는 새 사용량 창을 열지 않는다
limit_session: 80          # 5시간 창 사용량이 이 % 넘으면 새 커밋을 시작하지 않는다
limit_weekly_all: 80       # 주간 사용량이 이 % 넘으면 새 커밋을 시작하지 않는다. 리셋 24시간 안이면 무시
max_turns: 30
max_rounds: 3
webhook: https://discord.com/api/webhooks/…   # Discord나 Slack(hooks.slack.com). 밤이 끝나면 메시지 하나, 비우면 안 보낸다
lang: [en, ko]             # PR 본문 언어, 기본 [ko]
merge_method: rebase       # 🤖 PR이 auto-merge되는 방식: rebase, 아니면 merge. 기본 merge
```

첫 🤖 PR 전에 레포에서 auto-merge를 한 번 켠다. `merge_method: rebase`면 rebase 머지도 같이 허용한다:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```
