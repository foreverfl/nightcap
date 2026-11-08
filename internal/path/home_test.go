package path

import (
	"path/filepath"
	"testing"
)

func TestConfigFileFollowsXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/x")
	if got, _ := ConfigFile(); got != "/tmp/x/nightcap/nightcap.yml" {
		t.Fatalf("ConfigFile() = %q, want /tmp/x/nightcap/nightcap.yml", got)
	}
}

func TestConfigFileFallsBackToTheHomeConfigFolder(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/me")
	want := filepath.Join("/home/me", ".config", "nightcap", "nightcap.yml")
	if got, _ := ConfigFile(); got != want {
		t.Fatalf("ConfigFile() = %q, want %q", got, want)
	}
}
