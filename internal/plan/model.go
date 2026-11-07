package plan

// Mark is the reading mark before a heading's type: how much a person reads before the PR is merged.
type Mark string

const (
	MarkNone     Mark = ""          // no mark yet, so a night run skips the heading and lint reports it
	MarkAuto     Mark = "auto"      // 🤖 merged without a person
	MarkReadPR   Mark = "read-pr"   // 📄 merged after reading the PR
	MarkReadCode Mark = "read-code" // 👀 merged after reading the code
	MarkPerson   Mark = "person"    // 👤 a person does this PR by hand
)

// Section is one "## mark type: title (slug)" heading; it becomes one branch and one PR.
type Section struct {
	Mark    Mark
	Type    string // feat, fix, test, docs, refactor or chore
	Title   string
	Slug    string // branch is Type + "/" + Slug
	Summary string // first plain line under the heading, used as the PR body's first sentence
	Units   []Unit // open "- [ ]" units only
}

// IsPerson reports whether a person does this PR by hand, so a night run never takes it.
func (section Section) IsPerson() bool {
	return section.Mark == MarkPerson
}

// Unit is one open checkbox under a Section; it becomes one commit.
type Unit struct {
	Title    string
	Body     string // the checkbox line plus its indented lines, fences included
	HasCheck bool   // the DONE line names a check, so a night run may take it
}
