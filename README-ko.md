# nightcap

계획 문서의 제목 하나를 집어 밤사이 Claude Code로 유닛을 하나씩 구현하고, 아침에 사람이 합칠 PR을 연다.

- 계약 파일은 만들 수만 있다: 에이전트는 테스트 · hurl · 명세를 추가할 수는 있어도 고치거나 지우지 못한다.
- 의존성은 사람 몫이다: go.mod · Cargo.toml · package.json이 바뀌면 PR 없이 빨강으로 끝난다.
- 코드가 아니라 PR을 읽는다: `nightcap:read-pr`은 본문과 위험 목록만 보면 되고, `nightcap:read-code`는 diff까지 읽는다.
