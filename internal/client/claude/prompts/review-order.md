아래 git diff를 읽고 GitHub PR 본문의 「## 읽는 순서」 칸에 들어갈 내용만 한국어로 써라. 제목 줄(`## 읽는 순서`)은 쓰지 않는다.

- 이 diff가 더하거나 바꾼 데이터 흐름마다 블록 하나. 블록 제목은 `흐름 N. <언제 도는지>: <무엇을 하는지>`.
- 블록 본문은 text 펜스. 맨 바깥 입구(입력한 명령, 들어온 라우트)에서 맨 아래(실행한 외부 명령, 쓴 파일이나 테이블)까지.
- 한 줄에 한 단계. 아래로 부르면 `→ `, 돌아온 결과를 처리하면 `← `.
- 단계마다 `경로 함수: 하는 일`. 함수만 있거나 경로만 있는 단계는 쓰지 않는다.
- 시작하는 때가 다른 흐름은 블록을 나눈다.
- 블록들 밑에 `흐름 밖` 목록. 어느 흐름에도 안 보이는 것만 한 줄씩: 실패 때의 대체 동작, 이름 변경, 테스트 준비.
- 파일별 목록과 커밋 해시 목록은 쓰지 않는다.

예시:

흐름 1. 실행을 시작할 때: 오늘 밤 할 제목과 그 base를 정한다

```text
nightcap run                                   터미널에서 실행
→ main.go main: 첫 인자가 run이면 run.Run
→ internal/cmd/run/run.go Run: 제목을 위에서부터 훑는다
→ internal/pr/view.go ViewBranch → gh pr view <branch> --json state,url
← internal/state/store.go SavePRStatus: 받은 상태와 URL을 .nightcap/state.json에 적는다
```

흐름 밖
- `dir`를 전부 `worktreeDir`로 바꿨다. 동작은 안 바뀐다

diff:
