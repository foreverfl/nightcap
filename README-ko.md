# nightcap

계획 문서의 제목 하나를 집어 밤사이 Claude Code로 유닛을 하나씩 구현하고, 아침에 사람이 합칠 PR을 연다.

- 계약 파일은 만들 수만 있다: 에이전트는 테스트 · hurl · 명세를 추가할 수는 있어도 고치거나 지우지 못한다.
- 의존성은 사람 몫이다: go.mod · Cargo.toml · package.json이 바뀌면 PR 없이 빨강으로 끝난다.
- 라벨은 제목의 표시가 정한다: `🤖`는 auto-merge(`nightcap:auto`), `📄`는 PR만 읽고(`nightcap:read-pr`), `👤`는 사람이 한다. 밤에 걸린 것은 라벨이 아니라 위험 목록에 적힌다.

## 설정

계획 파일 모양은 [docs/plan-format-ko.md](docs/plan-format-ko.md). [examples/plan-ko.md](examples/plan-ko.md)를 복사해서 시작한다.

바이너리 옆에 `nightcap.yml`을 쓰고(gitignore 대상) 그 폴더에서 `nightcap run`을 친다. 계획 문서 하나가 항목 하나:

```yaml
work_start: "09:00"        # 비우면 출근 전 제한 없음. 적으면 이 시각 5시간 전부터 새 창을 열지 않는다
limit_session: 80          # 5시간 창 %가 이 위면 새 유닛을 안 연다
limit_weekly_all: 80       # 주간 %가 이 위면 새 유닛을 안 연다. 리셋 24시간 안이면 무시
max_turns: 30
max_rounds: 3
webhook: https://discord.com/api/webhooks/…   # Discord나 Slack(hooks.slack.com). 밤이 끝나면 메시지 하나, 비우면 안 보낸다
targets:
  - plan: ~/work/todolist/blog-v2/github-rank-commit.md
    repo: ~/workspace/apps-web/blog-v2
    check: make check
    lang: [en, ko]            # PR 본문 언어, 기본값 [ko]
    merge_method: rebase      # 🤖 PR이 auto-merge될 때 방식: rebase면 rebase, 아니면 merge. 기본값 merge
```

첫 🤖 제목 전에 레포에서 auto-merge를 한 번 켠다. `merge_method: rebase`면 rebase 머지도 같이 허용한다:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```
