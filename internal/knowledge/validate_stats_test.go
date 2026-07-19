package knowledge

import (
	"path/filepath"
	"testing"
)

func TestValidateReportsScanStatistics(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestConcept(t, root, "valid.md", "Valid", "See [Missing](missing.md).")
	writeTestFile(t, root, "invalid.md", "not frontmatter\n")

	report, err := Validate(root, false)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report.Stats.Root != root || report.Stats.Mode != "single-bundle" {
		t.Fatalf("root stats = %#v", report.Stats)
	}
	if report.Stats.Bundles != 1 || report.Stats.MarkdownFiles != 3 || report.Stats.ReservedFiles != 1 {
		t.Fatalf("file stats = %#v", report.Stats)
	}
	if report.Stats.ConceptFiles != 2 || report.Stats.LoadedConcepts != 1 || report.Stats.InvalidConceptFiles != 1 {
		t.Fatalf("concept stats = %#v", report.Stats)
	}
	if report.Stats.LinksChecked != 1 || report.Stats.BrokenLinks != 1 {
		t.Fatalf("link stats = %#v", report.Stats)
	}
}

func TestValidateWorkspaceRootAggregatesBundleStatistics(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "first/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# First\n")
	writeTestConcept(t, root, "first/a.md", "A", "A body.")
	writeTestFile(t, root, "second/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Second\n")
	writeTestConcept(t, root, "second/b.md", "B", "B body.")

	report, err := ValidateWorkspaceRoot(root, false)
	if err != nil {
		t.Fatalf("ValidateWorkspaceRoot() error = %v", err)
	}
	if report.Stats.Root != root || report.Stats.Mode != "workspace" || report.Stats.Bundles != 2 {
		t.Fatalf("workspace stats = %#v", report.Stats)
	}
	if report.Stats.LoadedConcepts != 2 || len(report.Bundles) != 2 {
		t.Fatalf("concept stats = %#v, bundles = %#v", report.Stats, report.Bundles)
	}
	if report.Bundles[0].Root != filepath.Join(root, "first") || report.Bundles[1].Root != filepath.Join(root, "second") {
		t.Fatalf("bundle roots = %#v", report.Bundles)
	}
}
