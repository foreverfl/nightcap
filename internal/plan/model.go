package plan

// Mark is the reading mark before a heading's type: how much a person reads before the PR is merged.
type Mark string

const (
	MarkNone   Mark = ""        // no mark yet, so a night run skips the heading and lint reports it
	MarkAuto   Mark = "auto"    // 🤖 merged without a person
	MarkReadPR Mark = "read-pr" // 📄 merged after reading the PR
	MarkPerson Mark = "person"  // 👤 a person does this PR by hand
)

// Emoji returns the mark as a plan writes it, such as "🤖"; MarkNone returns "".
func (mark Mark) Emoji() string {
	for emoji, known := range marks {
		if known == mark {
			return emoji
		}
	}
	return ""
}

// Section is one "## mark type: title (slug)" heading; it becomes one branch and one PR.
type PR struct {
	Mark    Mark
	Type    string // feat, fix, test, docs, refactor or chore
	Title   string
	Slug    string // branch is Type + "/" + Slug
	Summary string // first plain line under the heading, used as the PR body's first sentence
	Units   []Unit // open "- [ ]" units only
}

// IsPerson reports whether a person does this PR by hand, so a night run never takes it.
func (section PR) IsPerson() bool {
	return section.Mark == MarkPerson
}

// Unit is one open checkbox under a Section; it becomes one commit.
type Unit struct {
	Title    string
	Body     string // the checkbox line plus its indented lines, fences included
	HasCheck bool   // the DONE line names a check, so a night run may take it
}

// Meta is the plan file's frontmatter: which repo the plan runs in and how its units are tested.
type Meta struct {
	Repo         string   `yaml:"repo"`         // a leading "~/" is expanded to the home folder
	Test         string   `yaml:"test"`         // shell command that proves a unit, such as make check
	Protected    []string `yaml:"protected"`    // contract file globs the agent may add but not edit or delete
	Dependencies []string `yaml:"dependencies"` // dependency file globs the agent may not touch
}
