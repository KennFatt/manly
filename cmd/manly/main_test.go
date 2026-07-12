package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	assertCommandSucceeds(t, root, []string{"check", "--strict"})
	assertCommandSucceeds(t, root, []string{"check", "--format", "json"})

	t.Setenv("EDITOR", "true")
	assertCommandSucceeds(t, root, []string{"edit", "/a"})
	assertCommandContains(t, root, []string{"move", "/group/b", "/group/renamed"}, "updated 3 link(s)")
	data, err := os.ReadFile(filepath.Join(root, "a.md"))
	if err != nil || !strings.Contains(string(data), "group/renamed.md") {
		t.Fatalf("moved link = %q, %v", data, err)
	}
}

func TestCLIArgumentValidation(t *testing.T) {
	if err := run([]string{"help"}); err != nil {
		t.Fatalf("help error = %v", err)
	}
	if err := run([]string{"unknown"}); err == nil {
		t.Fatal("unknown command did not fail")
	}
	if _, err := parseFormat("invalid"); err == nil {
		t.Fatal("parseFormat accepted an invalid format")
	}
	if _, err := parseFormat("human"); err == nil || !strings.Contains(err.Error(), "compact, fancy, json, markdown") {
		t.Fatalf("parseFormat accepted removed format or omitted available formats: %v", err)
	}
	if _, _, err := parseGlobalArgs([]string{"--root"}); err == nil {
		t.Fatal("parseGlobalArgs accepted a missing root value")
	}
	if _, err := directoryPrefix("../outside"); err == nil {
		t.Fatal("directoryPrefix accepted traversal")
	}
	if got := splitTags(" go, ,safety "); len(got) != 2 || got[0] != "go" || got[1] != "safety" {
		t.Fatalf("splitTags() = %#v", got)
	}
}

func TestCLIParsingHelpers(t *testing.T) {
	t.Setenv("MANLY_ROOT", "/environment-root")
	root, args, err := parseGlobalArgs([]string{"show", "/concept"})
	if err != nil || root != "/environment-root" || len(args) != 2 {
		t.Fatalf("parseGlobalArgs() = %q, %#v, %v", root, args, err)
	}
	root, args, err = parseGlobalArgs([]string{"--root=/explicit", "check"})
	if err != nil || root != "/explicit" || len(args) != 1 || args[0] != "check" {
		t.Fatalf("explicit parseGlobalArgs() = %q, %#v, %v", root, args, err)
	}
	got := normalizeFlagArgs([]string{"query", "--format", "json", "--limit=2"}, map[string]bool{"--format": true, "--limit": true})
	if strings.Join(got, " ") != "--format json --limit=2 query" {
		t.Fatalf("normalizeFlagArgs() = %#v", got)
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
