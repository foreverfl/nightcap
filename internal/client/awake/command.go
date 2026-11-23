package awake

import "strconv"

// Command returns the command that keeps the machine awake until the process pid exits, or nil when goos has none.
// goos is runtime.GOOS: "darwin" → caffeinate, "linux" → systemd-inhibit holding a tail that ends with pid.
func Command(goos string, pid int) []string {
	switch goos {
	case "darwin":
		return []string{"caffeinate", "-i", "-w", strconv.Itoa(pid)}
	case "linux":
		// systemd-inhibit holds the lock only while its command runs, so it wraps a tail that exits with pid
		return []string{"systemd-inhibit", "--what=sleep:idle", "--who=nightcap", "--why=nightcap run",
			"tail", "--pid=" + strconv.Itoa(pid), "-f", "/dev/null"}
	}
	return nil
}
