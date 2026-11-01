package plan

// Section is one "## type: title (slug)" heading; it becomes one branch and one PR.
type Section struct {
	Type    string // feat, fix, test, docs, refactor or chore
	Title   string
	Slug    string // branch is Type + "/" + Slug
	Summary string // first plain line under the heading, used as the PR body's first sentence
	Units   []Unit // open "- [ ]" units only
}

// Unit is one open checkbox under a Section; it becomes one commit.
type Unit struct {
	Title    string
	Body     string // the checkbox line plus its indented lines, fences included
	HasCheck bool   // the DONE line names a check, so a night run may take it
}
