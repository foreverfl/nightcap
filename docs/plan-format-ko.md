# 계획 형식

nightcap은 `nightcap run`에 준 마크다운 파일 하나를 읽는다. 파일 맨 위 프론트매터가 repo · test · protected · dependencies를 말한다(README 참고). 전체 예시는 `examples/plan-ko.md`.

## `##` = PR 하나

```markdown
## 📄 feat: 메모 목록과 추가 (notes-api)

HTTP로 메모를 보고 더할 수 있게 해서, 웹 화면이 부를 곳을 만든다.
```

- PR은 `## 표시 type: 제목 (slug)` 꼴의 줄이다. `###`도 같게 읽지만 `##`가 기본이다.
- 표시는 합치기 전에 사람이 얼마나 읽을지다. PR 라벨은 표시 그대로 붙는다:
  - `🤖`: `nightcap:auto`. GitHub auto-merge를 켜서, 필수 검사가 초록이 되면 사람 없이 합쳐진다. 계약 · 테스트 파일만 더하는 PR
  - `📄`: `nightcap:read-pr`. 본문과 위험 목록만 보면 된다
  - `👤`: 사람이 손으로 하는 PR. nightcap은 고르지 않는다
- 밤에 걸린 것(WHERE 밖 · 마이그레이션)은 라벨이 아니라 위험 목록에 적힌다.
- 표시가 없는 PR은 건너뛰고, `nightcap lint`가 `no mark:`로 찍는다.
- `type`은 `feat` · `fix` · `test` · `docs` · `refactor` · `chore` 중 하나다.
- 브랜치는 `type/slug`, PR 제목은 `type: 제목`이 된다.
- `##` 줄 밑 첫 일반 문장이 PR 본문의 첫 문장이 된다.
- 알려진 type이나 `(slug)`가 없는 `##` 줄은 그 밑의 커밋과 함께 건너뛴다.

## `- [ ]` = 커밋 하나

```markdown
- [ ] 메모 추가하기
  - WHAT  JSON 본문으로 받은 메모를 저장하고 201로 답한다
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  go test ./internal/note 통과
```

- 커밋은 `- [ ]` 줄과 그 밑의 들여쓴 줄 전부다. 펜스도 포함한다.
- 커밋 글 전체가 그대로 에이전트에게 넘어간다.
- `- [x]` 커밋은 끝난 것이라 건너뛴다.
- `DONE`에는 커밋을 증명하는 테스트 이름을 쓴다: `hurl` · `go test` · `make check` · `cargo test` · `npm test`.
- `DONE`에 테스트 이름이 없는 커밋은 사람 몫이다. 야간 실행은 건너뛴다.
- `WHERE`는 커밋이 건드려도 되는 파일 목록이다. 그 밖의 파일을 고치면 PR 위험 칸에 적힌다.
