package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIProcessWorkflow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MANLY_ROOT", "")
	binary := filepath.Join(t.TempDir(), "manly")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	root := t.TempDir()
	result := runBinary(t, binary, root, "init")
	if result.err != nil || !strings.Contains(result.output, "Initialized OKF bundle") {
		t.Fatalf("binary init = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "add", "/first-note", "--type", "Note", "--title", "First Note")
	if result.err != nil || !strings.Contains(result.output, "Created /first-note") {
		t.Fatalf("binary add = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "search", "first", "--format", "json")
	if result.err != nil || !strings.Contains(result.output, `"id": "/first-note"`) {
		t.Fatalf("binary search = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "context", "/first-note", "--format", "json")
	if result.err != nil || !strings.Contains(result.output, `"confident": true`) || !strings.Contains(result.output, `"confidence": "high"`) {
		t.Fatalf("binary context exact id = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "search", "/first-note", "--format", "json")
	if result.err != nil || !strings.Contains(result.output, `"confident": true`) {
		t.Fatalf("binary search exact id = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "context", "first", "--type", "Guideline", "--format", "json")
	if result.err != nil || !strings.Contains(result.output, `"results": []`) || !strings.Contains(result.output, `"confident": false`) {
		t.Fatalf("binary context filter = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "context", "first", "--type", "Note", "--format", "json")
	if result.err != nil || !strings.Contains(result.output, `"id": "/first-note"`) {
		t.Fatalf("binary context type filter = %q, %v", result.output, result.err)
	}
	result = runBinary(t, binary, root, "check")
	if result.err != nil || !strings.Contains(result.output, "OKF validation passed") {
		t.Fatalf("binary check = %q, %v", result.output, result.err)
	}
}

type binaryResult struct {
	output string
	err    error
}

func runBinary(t *testing.T, binary, root string, args ...string) binaryResult {
	t.Helper()
	commandArgs := append([]string{"--root", root}, args...)
	command := exec.Command(binary, commandArgs...)
	command.Env = append(os.Environ(), "MANLY_ROOT=")
	output, err := command.CombinedOutput()
	return binaryResult{output: string(output), err: err}
}
