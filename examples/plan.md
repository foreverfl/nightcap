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

- [ ] pick the page size
  - WHAT  decide how many notes one page shows
  - WHERE this file
  - DONE  a number is written here; done by a person
