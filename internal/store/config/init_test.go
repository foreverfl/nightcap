package config

import (
	"os"
	"testing"
)

func TestInitWritesTheExampleOnceAndLeavesItAfter(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	configPath, created, err := Init([]byte("max_rounds: 3\n"))
	if err != nil || !created {
		t.Fatalf("first Init: created %v, err %v", created, err)
	}
	if _, created, err := Init([]byte("max_rounds: 9\n")); err != nil || created {
		t.Fatalf("second Init: created %v, err %v; want false, nil", created, err)
	}
	if text, _ := os.ReadFile(configPath); string(text) != "max_rounds: 3\n" {
		t.Fatalf("config = %q, want the first example left as it was", text)
	}
}
