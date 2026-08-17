package knowledge

import (
	"strings"
	"testing"
)

func TestLoadWorkspaceResolvesLinksWithinEachBundle(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "---\ndescription: Workspace description.\n---\n")
	writeTestFile(t, root, "engineering-preferences/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Engineering\n")
	writeTestConcept(t, root, "engineering-preferences/typescript/type-safety.md", "Type Safety", "Safe types.")
	writeTestConcept(t, root, "engineering-preferences/react/handler-naming.md", "Handler Naming", "See [Type Safety](/typescript/type-safety.md).")
	writeTestFile(t, root, "personal/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Personal\n")
	writeTestConcept(t, root, "personal/typescript/type-safety.md", "Other Type Safety", "Other safe types.")

	workspace, err := LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace() error = %v", err)
	}
	if workspace.SingleRoot || len(workspace.Bundles) != 2 {
		t.Fatalf("workspace = %#v", workspace)
	}
	if workspace.Description != "Workspace description." {
		t.Fatalf("workspace description = %q", workspace.Description)
	}
	ref, err := workspace.ResolveConcept("/engineering-preferences/react/handler-naming")
	if err != nil {
		t.Fatalf("ResolveConcept() error = %v", err)
	}
	if ref.Concept.Links[0].TargetID != "/typescript/type-safety" || ref.Concept.Links[0].Broken {
		t.Fatalf("bundle-local link = %+v", ref.Concept.Links[0])
	}
	if got := workspace.QualifyID(ref.BundleName, ref.Concept.Links[0].TargetID); got != "/engineering-preferences/typescript/type-safety" {
		t.Fatalf("QualifyID() = %q", got)
	}
	results, err := workspace.Search("safe types", SearchOptions{Limit: 10})
	if err != nil || len(results) < 2 {
		t.Fatalf("workspace search = %#v, %v", results, err)
	}
}

func TestLoadWorkspaceRejectsMalformedBundleCandidate(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "broken/index.md", "# Not a bundle\n")
	if _, err := LoadWorkspace(root); err == nil || !strings.Contains(err.Error(), "bundle candidate") {
		t.Fatalf("LoadWorkspace() error = %v", err)
	}
}

func TestLoadWorkspacePreservesSingleBundleMode(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Bundle\n")
	writeTestConcept(t, root, "local.md", "Local", "Local concept.")
	workspace, err := LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace() error = %v", err)
	}
	if !workspace.SingleRoot {
		t.Fatal("LoadWorkspace() did not preserve single-bundle mode")
	}
	if _, err := workspace.ResolveConcept("/local"); err != nil {
		t.Fatalf("ResolveConcept() error = %v", err)
	}
}

func TestWorkspaceMode(t *testing.T) {
	singleRoot := t.TempDir()
	writeTestFile(t, singleRoot, "index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Bundle\n")
	writeTestConcept(t, singleRoot, "local.md", "Local", "Local concept.")
	single, err := LoadWorkspace(singleRoot)
	if err != nil {
		t.Fatalf("LoadWorkspace(single) error = %v", err)
	}
	if got := single.Mode(); got != ModeSingle {
		t.Fatalf("single root Mode() = %v, want ModeSingle", got)
	}

	multiRoot := t.TempDir()
	writeTestFile(t, multiRoot, "prefs/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Prefs\n")
	writeTestConcept(t, multiRoot, "prefs/local.md", "Local", "Local concept.")
	multi, err := LoadWorkspace(multiRoot)
	if err != nil {
		t.Fatalf("LoadWorkspace(multi) error = %v", err)
	}
	if got := multi.Mode(); got != ModeWorkspace {
		t.Fatalf("multi-bundle root Mode() = %v, want ModeWorkspace", got)
	}
}

func TestWorkspaceModeString(t *testing.T) {
	tests := []struct {
		mode WorkspaceMode
		want string
	}{
		{ModeWorkspace, "workspace"},
		{ModeSingle, "single"},
	}
	for _, test := range tests {
		if got := test.mode.String(); got != test.want {
			t.Fatalf("mode %v String() = %q, want %q", test.mode, got, test.want)
		}
	}
}
