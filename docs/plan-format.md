# Plan format

nightcap reads the one markdown file you pass to `nightcap run`. It starts with a frontmatter naming repo, test, protected and dependencies, as the README shows. See `examples/plan.md` for a full one.

## `##` = one PR

```markdown
## 📄 feat: list and add notes (notes-api)

Notes can be listed and added over HTTP, so the web page has something to call.
```

- A PR is the line `## mark type: title (slug)`. A `###` line is read the same way, but `##` is the convention.
- The mark says how much a person reads before merging, and the PR label follows it as written:
  - `🤖`: `nightcap:auto`. GitHub auto-merge is turned on, so the PR merges once the required checks pass. For PRs that only add contract and test files
  - `📄`: `nightcap:read-pr`. The body and the risk list are enough
  - `👤`: a person does this PR by hand. nightcap never picks it
- What nightcap notices at night (outside WHERE, a migration) goes into the risk list, never into the label.
- A PR without a mark is skipped, and `nightcap lint` prints it as `no mark:`.
- `type` is one of `feat`, `fix`, `test`, `docs`, `refactor`, `chore`.
- The branch is `type/slug` and the PR title is `type: title`.
- The first plain line under the `##` line becomes the first sentence of the PR body.
- A `##` line without a known type or a `(slug)` is skipped, together with the commits under it.

## `- [ ]` = one commit

```markdown
- [ ] add a note
  - WHAT  save a note from a JSON body and answer 201
  - WHERE internal/note/handler.go, internal/note/store.go
  - DONE  go test ./internal/note passes
```

- A commit is a `- [ ]` line plus every indented line under it, fences included.
- The whole commit text is handed to the agent as it is.
- `- [x]` commits are done and skipped.
- `DONE` names the test that proves the commit: `hurl`, `go test`, `make check`, `cargo test` or `npm test`.
- A commit whose `DONE` names no test is left for a person; a night run skips it.
- `WHERE` lists the files the commit may touch; a change outside them goes into the PR's risk list.
