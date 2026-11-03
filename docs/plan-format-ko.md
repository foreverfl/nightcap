# 계획 형식

nightcap은 항목마다 마크다운 파일 하나를 읽는다. 전체 예시는 `examples/plan-ko.md`.

## 제목 = PR 하나

```markdown
## feat: 메모 목록과 추가 (notes-api)

HTTP로 메모를 보고 더할 수 있게 해서, 웹 화면이 부를 곳을 만든다.
```

- 제목은 `## type: 제목 (slug)` 꼴이다. `###`도 같게 읽지만 `##`가 기본이다.
- `👤`로 시작하는 제목(`## 👤 feat: …`)은 사람이 손으로 하는 PR이다. nightcap은 그 제목을 읽지 않고, 고르지도 않는다.
- `type`은 `feat` · `fix` · `test` · `docs` · `refactor` · `chore` 중 하나다.
- 브랜치는 `type/slug`, PR 제목은 `type: 제목`이 된다.
- 제목 밑 첫 일반 문장이 PR 본문의 첫 문장이 된다.
- 알려진 type이나 `(slug)`가 없는 제목은 그 밑의 유닛과 함께 건너뛴다.

## 유닛 = 커밋 하나

```markdown
- [ ] 메모 추가하기
  - WHAT  JSON 본문으로 받은 메모를 저장하고 201로 답한다
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  go test ./internal/note 통과
```

- 유닛은 `- [ ]` 줄과 그 밑의 들여쓴 줄 전부다. 펜스도 포함한다.
- 유닛 글 전체가 그대로 에이전트에게 넘어간다.
- `- [x]` 유닛은 끝난 것이라 건너뛴다.
- `DONE`에는 유닛을 증명하는 검사 이름을 쓴다: `hurl` · `go test` · `make check` · `cargo test` · `npm test`.
- `DONE`에 검사 이름이 없는 유닛은 사람 몫이다. 야간 실행은 건너뛴다.
- `WHERE`는 유닛이 건드려도 되는 파일 목록이다. 그 밖의 파일을 고치면 코드까지 읽어야 하는 PR이 된다.
