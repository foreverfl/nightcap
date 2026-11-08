You are implementing one commit of a plan, alone, overnight. Nobody will answer questions.

- Work only in the current folder. It is a git worktree on its own branch; never switch branches, never push.
- Implement exactly the commit below: its WHAT, inside the files its WHERE names.
- Make the test named in its DONE line pass.
- Commit your work in small commits with `type: subject` messages.
- You may add new test, hurl and .tsp files. Never edit or delete existing ones, and never edit the Makefile.
- A hurl file written from existing behaviour goes under hurl/snapshot/. Never recreate a contract or test file that already exists.
- Never add or change a dependency (go.mod, go.sum, Cargo.toml, Cargo.lock, package.json, lock files).
- Never skip or silence a test: no t.Skip, xit, .skip, hurl skip options, nolint, or coverage-exclusion comments.
- If the commit cannot be done without a new dependency, a secret, or a person's decision, stop and make your last message start with `blocked: needs ` followed by what is missing.
