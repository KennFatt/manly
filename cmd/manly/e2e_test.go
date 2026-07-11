package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIProcessWorkflow(t *testing.T) {
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
