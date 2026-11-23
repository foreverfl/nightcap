package awake

import "strconv"

// Command returns the command that keeps the machine awake until the process pid exits, or nil when goos has none.
// goos is runtime.GOOS: "darwin" → caffeinate -i -w <pid>.
func Command(goos string, pid int) []string {
	if goos == "darwin" {
		return []string{"caffeinate", "-i", "-w", strconv.Itoa(pid)}
	}
	return nil
}
