<!-- unit.md의 한국어판. 사람이 읽는 용도이고 에이전트에게는 unit.md(영어)가 간다. 고칠 때는 둘을 같이 고친다. -->

계획의 유닛 하나를 밤에 혼자 구현한다. 질문에 답해 줄 사람은 없다.

- 지금 폴더 안에서만 일한다. 이 폴더는 자기 브랜치를 가진 git worktree다. 브랜치를 바꾸지 않고, push하지 않는다.
- 아래 유닛을 정확히 그대로 구현한다: WHAT을, WHERE가 적은 파일 안에서.
- DONE 줄에 적힌 검사를 통과시킨다.
- 작업은 `type: subject` 메시지로 작게 나눠 커밋한다.
- 새 테스트 · hurl · OpenAPI 파일은 더해도 된다. 기존 파일은 고치거나 지우지 않고, Makefile도 고치지 않는다.
- 의존성은 더하거나 바꾸지 않는다(go.mod, go.sum, Cargo.toml, Cargo.lock, package.json, lock 파일).
- 검사를 건너뛰거나 끄지 않는다: t.Skip, xit, .skip, hurl의 skip 옵션, nolint, 커버리지 제외 주석 금지.
- 새 의존성 · 비밀 값 · 사람의 결정 없이는 끝낼 수 없으면, 멈추고 마지막 메시지를 `blocked: needs `로 시작해 무엇이 없는지 적는다.
