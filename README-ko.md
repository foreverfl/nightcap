# nightcap

계획 문서의 제목 하나를 집어 밤사이 Claude Code로 유닛을 하나씩 구현하고, 아침에 사람이 합칠 PR을 연다.

- 계약 파일은 만들 수만 있다: 에이전트는 테스트 · hurl · 명세를 추가할 수는 있어도 고치거나 지우지 못한다.
- 의존성은 사람 몫이다: go.mod · Cargo.toml · package.json이 바뀌면 PR 없이 빨강으로 끝난다.
- 라벨은 제목의 표시가 정한다: `🤖`는 auto-merge(`nightcap:auto`), `📄`는 PR만 읽고(`nightcap:read-pr`), `👤`는 사람이 한다. 밤에 걸린 것은 라벨이 아니라 위험 목록에 적힌다.

## 설정

계획 파일 모양은 [docs/plan-format-ko.md](docs/plan-format-ko.md). [examples/plan-ko.md](examples/plan-ko.md)를 복사해서 시작한다.

계획 파일 맨 위에 이걸 적는다. 네 키 모두 꼭 있어야 하고, 비울 목록은 `[]`로 적는다:

```yaml
---
repo: ~/workspace/apps-web/blog-v2
test: make check                                         # 유닛마다 통과해야 하는 검사
protected: ["*.tsp", "*.hurl", "*_test.go", "Makefile"]  # 에이전트가 추가는 해도 고치거나 지우지 못하는 계약 파일
dependencies: [go.mod, go.sum]                           # 에이전트가 건드리지 못하는 의존성 파일
---
```

[examples/nightcap.yml](examples/nightcap.yml)을 바이너리 옆에 `nightcap.yml`로 복사한다(gitignore 대상). 이 기계에서 모든 계획이 같이 쓰는 설정만 둔다.

그 폴더에서 계획 파일을 주고 돌린다:

```sh
nightcap run ~/work/todolist/blog-v2/commit-bot.md
```

첫 🤖 제목 전에 레포에서 auto-merge를 한 번 켠다. `merge_method: rebase`면 rebase 머지도 같이 허용한다:

```sh
gh repo edit <owner>/<repo> --enable-auto-merge --enable-rebase-merge
```
