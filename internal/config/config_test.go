package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadBootstrapsDefaults(t *testing.T) {
	home := t.TempDir()

	result, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if result.Root != filepath.Join(home, ".okf") {
		t.Fatalf("root = %q", result.Root)
	}
	if result.Defaults.Format != "compact" || result.Defaults.List.Recursive {
		t.Fatalf("defaults = %#v", result.Defaults)
	}
	if !result.Display.Actions || !result.Display.Usage {
		t.Fatalf("display = %#v", result.Display)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "manly", "config.yml")); err != nil {
		t.Fatalf("bootstrapped config: %v", err)
	}
}

func TestLoadMergesPartialConfigWithoutRewriting(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "manly", "config.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "root: ~/notes\ndefaults:\n  format: json\ndisplay:\n  actions: false\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if result.Root != filepath.Join(home, "notes") || result.Defaults.Format != "json" {
		t.Fatalf("result = %#v", result)
	}
	if result.Defaults.List.Recursive || result.Display.Actions || !result.Display.Usage {
		t.Fatalf("merged defaults/display = %#v / %#v", result.Defaults, result.Display)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Fatalf("existing config was rewritten: %q", got)
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"malformed yaml", "display: [", "load"},
		{"unknown key", "display:\n  actions: true\n  typo: false\n", "decode"},
		{"invalid format", "defaults:\n  format: terminal\n", "defaults.format"},
		{"empty root", "root: ''\n", "root"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, ".config", "manly", "config.yml")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(home)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want %q", err, test.want)
			}
		})
	}
}
