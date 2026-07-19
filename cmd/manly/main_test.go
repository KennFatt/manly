package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KennFatt/manly/internal/knowledge"
)

func TestCLIWorkflow(t *testing.T) {
	root := t.TempDir()
	if output, err := runCommand(t, root, "init"); err != nil || !strings.Contains(output, "Initialized OKF bundle") {
		t.Fatalf("init = %q, %v", output, err)
	}
	if _, err := runCommand(t, root, "add", "/a", "--type", "Guideline", "--title", "Boundary Rules", "--description", "External data", "--tag", "go,safety"); err != nil {
		t.Fatalf("add /a error = %v", err)
	}
	if _, err := runCommand(t, root, "add", "/group/b", "--type=Note", "--title=B"); err != nil {
		t.Fatalf("add /group/b error = %v", err)
	}
	writeCLIFile(t, root, "a.md", "---\ntype: Guideline\ntitle: Boundary Rules\ntags: [go, safety]\n---\n\nExternal data. See [B](group/b.md), [Web](https://example.com), and [Missing](missing.md).\n")
	writeCLIFile(t, root, "group/b.md", "---\ntype: Note\ntitle: B\n---\n\nLinked body.\n")
	writeCLIFile(t, root, "group/nested/c.md", "---\ntype: Note\ntitle: C\n---\n\nNested body.\n")
	writeCLIFile(t, root, "index.md", "---\nokf_version: \"0.1\"\n---\n\n<!-- manly:generated:start -->\nold\n<!-- manly:generated:end -->\n")
	writeCLIFile(t, root, "group/index.md", "<!-- manly:generated:start -->\n<!-- manly:generated:end -->\n")

	assertCommandContains(t, root, []string{"list"}, "/a", "group")
	assertCommandContains(t, root, []string{"list", "--recursive"}, "/a", "/group/b")
	assertCommandContains(t, root, []string{"list", "--recursive", "--format", "markdown"}, "# Knowledge Bundle", "/group/b.md")
	assertCommandContains(t, root, []string{"list", "--recursive", "--format", "json"}, `"recursive": true`, `"/group/b"`)
	assertCommandContains(t, root, []string{"list", "/group", "--format", "markdown"}, "# Group", "/group/b.md")
	assertCommandContains(t, root, []string{"show", "/group", "--format", "compact"}, "/group/b", "/group/nested/c")
	assertCommandContains(t, root, []string{"show", "/a", "/group/nested/c", "--format", "json"}, `"results"`, `"/a"`, `"/group/nested/c"`)
	for _, format := range []string{"fancy", "json", "compact", "markdown"} {
		assertCommandSucceeds(t, root, []string{"show", "/group", "--format", format})
	}

	for _, format := range []string{"fancy", "json", "compact", "markdown"} {
		assertCommandSucceeds(t, root, []string{"show", "/a", "--format", format})
		assertCommandSucceeds(t, root, []string{"search", "external data", "--format", format})
		assertCommandSucceeds(t, root, []string{"context", "/a", "--format", format})
		assertCommandSucceeds(t, root, []string{"links", "/a", "--format", format})
		assertCommandSucceeds(t, root, []string{"backlinks", "/group/b", "--format", format})
		assertCommandSucceeds(t, root, []string{"graph", "/a", "--depth", "1", "--format", format})
	}
	assertCommandContains(t, root, []string{"context", "external data", "--limit", "1", "--format", "json"}, `"results"`)
	assertCommandContains(t, root, []string{"search", "external", "--tag", "safety", "--type", "guide", "--path", "/", "--limit", "1", "--format", "json"}, `"/a"`)

	if output, err := runCommand(t, root, "index", "--check"); err == nil || !strings.Contains(output, "stale:") {
		t.Fatalf("stale index check = %q, %v", output, err)
	}
	assertCommandContains(t, root, []string{"index"}, "Updated index.md", "Updated group/index.md")
	assertCommandSucceeds(t, root, []string{"index", "--check"})
	assertCommandContains(t, root, []string{"check", "--strict"}, "Root:", "Mode: single-bundle", "Bundles: 1", "Loaded concepts:")
	assertCommandContains(t, root, []string{"check", "--format", "json"}, `"root"`, `"mode": "single-bundle"`, `"stats"`, `"bundles"`, `"valid": true`)

	t.Setenv("EDITOR", "true")
	assertCommandSucceeds(t, root, []string{"edit", "/a"})
	assertCommandContains(t, root, []string{"move", "/group/b", "/group/renamed"}, "updated 3 link(s)")
	data, err := os.ReadFile(filepath.Join(root, "a.md"))
	if err != nil || !strings.Contains(string(data), "group/renamed.md") {
		t.Fatalf("moved link = %q, %v", data, err)
	}
}

func TestWorkspaceDirectoryEntriesSeparateBundleAndDirectory(t *testing.T) {
	bundle := &knowledge.Bundle{Concepts: []*knowledge.Concept{{RelPath: "general/naming.md"}}}

	entries := workspaceDirectoryEntries(&knowledge.Workspace{}, "engineering-preferences", bundle, []string{"/general"})

	if len(entries) != 1 || entries[0].Path != "/engineering-preferences/general" {
		t.Fatalf("workspaceDirectoryEntries() = %#v", entries)
	}
}

func TestCLIArgumentValidation(t *testing.T) {
	output, err := captureOutput(t, func() error { return run([]string{"help"}) })
	if err != nil || !strings.Contains(output, "Commands:") || !strings.Contains(output, "init") {
		t.Fatalf("help = %q, %v", output, err)
	}
	if err := run([]string{"unknown"}); err == nil {
		t.Fatal("unknown command did not fail")
	}
	if err := run([]string{"--root"}); err == nil {
		t.Fatal("parser accepted a missing root value")
	}
	if err := run([]string{"show"}); err == nil {
		t.Fatal("parser accepted a missing concept")
	}
	if err := run([]string{"add", "/new-concept"}); err == nil {
		t.Fatal("parser accepted a missing concept type")
	}
	if err := run([]string{"--root", t.TempDir(), "list", "--format", "invalid"}); err == nil {
		t.Fatal("CLI accepted an invalid output format")
	}
	if _, err := directoryPrefix("../outside"); err == nil {
		t.Fatal("directoryPrefix accepted traversal")
	}
	if got := splitTags(" go, ,safety "); len(got) != 2 || got[0] != "go" || got[1] != "safety" {
		t.Fatalf("splitTags() = %#v", got)
	}
}

func TestCLIRootPrecedence(t *testing.T) {
	environmentRoot := t.TempDir()
	explicitRoot := t.TempDir()
	t.Setenv("MANLY_ROOT", environmentRoot)
	if _, err := captureOutput(t, func() error { return run([]string{"init"}) }); err != nil {
		t.Fatalf("environment root init = %v", err)
	}
	if _, err := os.Stat(filepath.Join(environmentRoot, "index.md")); err != nil {
		t.Fatalf("environment root index: %v", err)
	}
	if _, err := captureOutput(t, func() error {
		return run([]string{"--root", explicitRoot, "init"})
	}); err != nil {
		t.Fatalf("explicit root init = %v", err)
	}
	if _, err := os.Stat(filepath.Join(explicitRoot, "index.md")); err != nil {
		t.Fatalf("explicit root index: %v", err)
	}
}

func TestCLIConfigDefaultsAndOverrides(t *testing.T) {
	home := t.TempDir()
	root := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MANLY_ROOT", "")
	configPath := filepath.Join(home, ".config", "manly", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	configYAML := "root: " + root + "\ndefaults:\n  format: json\n  list:\n    recursive: true\ndisplay:\n  actions: false\n  usage: false\n"
	if err := os.WriteFile(configPath, []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := captureOutput(t, func() error { return run([]string{"--root", root, "init"}) }); err != nil {
		t.Fatal(err)
	}
	writeCLIFile(t, root, "nested/note.md", "---\ntype: Note\ntitle: Nested\n---\n\nBody.\n")

	output, err := captureOutput(t, func() error { return run([]string{"list"}) })
	if err != nil || !strings.Contains(output, `"recursive": true`) || strings.Contains(output, `"actions"`) {
		t.Fatalf("configured list = %q, %v", output, err)
	}
	output, err = captureOutput(t, func() error {
		return run([]string{"list", "--no-recursive", "--format", "compact"})
	})
	if err != nil || !strings.Contains(output, "nested") || strings.Contains(output, "List: manly list") || strings.Contains(output, "Details: manly show") {
		t.Fatalf("CLI list overrides = %q, %v", output, err)
	}
	output, err = captureOutput(t, func() error { return run([]string{"show", "/nested/note"}) })
	if err != nil || strings.Contains(output, `"actions"`) {
		t.Fatalf("configured show = %q, %v", output, err)
	}
}

func TestCLIVersion(t *testing.T) {
	output, err := captureOutput(t, func() error { return run([]string{"version"}) })
	if err != nil {
		t.Fatalf("version = %q, %v", output, err)
	}
	// Direct go build reports "dev" since no linker stamp is applied.
	if !strings.HasPrefix(output, "manly ") {
		t.Fatalf("version output %q does not start with 'manly '", output)
	}
	if strings.TrimSpace(output) == "manly" {
		t.Fatalf("version output %q is missing a version value", output)
	}
}

func TestCLIVersionNoBundle(t *testing.T) {
	// Version should succeed without a valid root.
	result := run([]string{"version"})
	if result != nil {
		t.Fatalf("version without bundle = %v", result)
	}
}

func TestCLIParsingHelpers(t *testing.T) {
	if got, err := resolveRoot("/explicit"); err != nil || got != "/explicit" {
		t.Fatalf("resolveRoot() = %q, %v", got, err)
	}
	if got := directoryTitle("type-safe"); got != "Type safe" {
		t.Fatalf("directoryTitle() = %q", got)
	}
	if got := directoryDisplay(""); got != "/" {
		t.Fatalf("directoryDisplay() = %q", got)
	}
}

func runCommand(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	return captureOutput(t, func() error {
		return run(append([]string{"--root", root}, args...))
	})
}

func assertCommandSucceeds(t *testing.T, root string, args []string) {
	t.Helper()
	if output, err := runCommand(t, root, args...); err != nil {
		t.Fatalf("%v = %q, %v", args, output, err)
	}
}

func assertCommandContains(t *testing.T, root string, args []string, fragments ...string) {
	t.Helper()
	output, err := runCommand(t, root, args...)
	if err != nil {
		t.Fatalf("%v = %q, %v", args, output, err)
	}
	for _, fragment := range fragments {
		if !strings.Contains(output, fragment) {
			t.Fatalf("%v output %q does not contain %q", args, output, fragment)
		}
	}
}

func captureOutput(t *testing.T, action func() error) (string, error) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	result := action()
	_ = writer.Close()
	os.Stdout = original
	data, readErr := io.ReadAll(reader)
	_ = reader.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(data), result
}

func writeCLIFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
