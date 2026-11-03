package pr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// View is the PR whose head is a branch: State is OPEN, MERGED or CLOSED, and both fields are empty when there is none.
type View struct {
	State string `json:"state"`
	URL   string `json:"url"`
}

// ViewBranch asks gh, from repoDir, for the PR whose head is branch. It returns an empty View when the branch has no PR,
// or an error when gh fails for any other reason or prints something that is not the JSON it was asked for.
func ViewBranch(repoDir, branch string) (View, error) {
	command := exec.Command(ghBinary(), "pr", "view", branch, "--json", "state,url")
	command.Dir = repoDir
	output, err := command.Output()
	if exitErr, isExitErr := errors.AsType[*exec.ExitError](err); isExitErr {
		if strings.Contains(string(exitErr.Stderr), "no pull requests found") {
			return View{}, nil
		}
		return View{}, fmt.Errorf("gh pr view: %w\n%s", err, exitErr.Stderr)
	}
	if err != nil {
		return View{}, err
	}
	var view View
	if err := json.Unmarshal(output, &view); err != nil {
		return View{}, err
	}
	return view, nil
}
