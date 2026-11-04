# Plan format

nightcap reads one markdown file per target. See `examples/plan.md` for a full one.

## Heading = one PR

```markdown
## feat: list and add notes (notes-api)

Notes can be listed and added over HTTP, so the web page has something to call.
```

- The heading is `## type: title (slug)`. A `###` heading is read the same way, but `##` is the convention.
- A heading that starts with `👤` (`## 👤 feat: …`) is a PR a person does by hand. nightcap does not read it and never picks it.
- `type` is one of `feat`, `fix`, `test`, `docs`, `refactor`, `chore`.
- The branch is `type/slug` and the PR title is `type: title`.
- The first plain line under the heading becomes the first sentence of the PR body.
- A heading without a known type or a `(slug)` is skipped, together with its units.

## Unit = one commit

```markdown
- [ ] add a note
  - WHAT  save a note from a JSON body and answer 201
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  go test ./internal/note passes
```

- A unit is a `- [ ]` line plus every indented line under it, fences included.
- The whole unit text is handed to the agent as it is.
- `- [x]` units are done and skipped.
- `DONE` names the check that proves the unit: `hurl`, `go test`, `make check`, `cargo test` or `npm test`.
- A unit whose `DONE` names no check is left for a person; a night run skips it.
- `WHERE` lists the files the unit may touch; a change outside them makes the PR one to read in full.
