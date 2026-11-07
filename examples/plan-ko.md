---
repo: ~/workspace/apps-web/notes
test: make check
protected: ["*.hurl", "*_test.go", "Makefile"]
dependencies: [go.mod, go.sum]
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

- [ ] 한 쪽에 보일 개수 정하기
  - WHAT  한 쪽에 메모를 몇 개 보일지 정한다
  - WHERE 이 파일
  - DONE  숫자 하나를 여기 적는다. 사람이 한다
