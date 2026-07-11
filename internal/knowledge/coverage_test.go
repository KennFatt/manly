package knowledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsInvalidRootsAndConcepts(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("Load() accepted a missing root")
	}

	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(file); err == nil {
		t.Fatal("Load() accepted a file as a root")
	}

	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestFile(t, root, "invalid.md", "# Missing frontmatter\n")
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "invalid.md") {
		t.Fatalf("Load() error = %v, want invalid concept path", err)
	}
}

func TestFrontmatterMetadataAndDisplayFallbacks(t *testing.T) {
	metadata, body, err := parseFrontmatter("---\r\ntags: [go, 2]\r\nvalue: 42\r\n---\r\n\r\nbody\r\n")
	if err != nil {
		t.Fatalf("parseFrontmatter() error = %v", err)
	}
	if body != "\nbody\n" {
		t.Fatalf("body = %q", body)
	}
	if got := metadataString(metadata, "value"); got != "42" {
		t.Fatalf("metadataString() = %q, want 42", got)
	}
	if got := metadataStrings(metadata, "tags"); len(got) != 2 || got[1] != "2" {
		t.Fatalf("metadataStrings() = %#v", got)
	}
	if _, _, err := parseFrontmatter("---\ninvalid: [\n---\n"); err == nil {
		t.Fatal("parseFrontmatter() accepted invalid YAML")
	}

	concept := &Concept{RelPath: "type-safe-data.md", Body: "# Heading\n\nFirst useful sentence.\n"}
	if got := displayTitle(concept); got != "Type Safe Data" {
		t.Fatalf("displayTitle() = %q", got)
	}
	if got := displayDescription(concept); got != "First useful sentence." {
		t.Fatalf("displayDescription() = %q", got)
	}
	concept.Body = "# Heading\n"
	if got := displayDescription(concept); got != "" {
		t.Fatalf("empty displayDescription() = %q", got)
	}
}

func TestSearchFiltersAndDefaults(t *testing.T) {
	bundle := &Bundle{Concepts: []*Concept{
		{ID: "/go/boundaries", RelPath: "go/boundaries.md", Type: "Guideline", Title: "Boundary Rules", Description: "External data safety", Tags: []string{"Go", "safety"}, Body: "Treat inputs as partial."},
		{ID: "/go/testing", RelPath: "go/testing.md", Type: "Procedure", Title: "Testing", Tags: []string{"go"}, Body: "Run focused tests."},
		{ID: "/other/notes", RelPath: "other/notes.md", Type: "Note", Title: "Other Notes", Body: "Unrelated."},
	}}

	if got := Search(bundle, "", SearchOptions{}); got != nil {
		t.Fatalf("empty search = %#v, want nil", got)
	}
	results := Search(bundle, "external data", SearchOptions{Tag: "SAFETY", Type: "guide", Path: "/go", Limit: 1})
	if len(results) != 1 || results[0].Concept.ID != "/go/boundaries" {
		t.Fatalf("filtered search = %#v", results)
	}
	results = Search(bundle, "go", SearchOptions{Limit: 0})
	if len(results) != 2 {
		t.Fatalf("default-limit search returned %d results", len(results))
	}
	if got := Search(bundle, "go", SearchOptions{Path: "/missing"}); len(got) != 0 {
		t.Fatalf("missing path search = %#v", got)
	}
}

func TestLinkResolutionCoversLocalAndIgnoredLinks(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestFile(t, root, "docs/index.md", "# Docs\n")
	writeTestConcept(t, root, "a.md", "A", "[B](/b.md#section) [Docs](docs) [Index](/index.md) [Missing](../outside.md) [Anchor](#heading) [Web](https://example.com) ![Image](image.png)\n\n```md\n[Code](b.md)\n```\n")
	writeTestConcept(t, root, "b.md", "B", "B body")

	bundle, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	links, err := bundle.Outgoing("/a")
	if err != nil {
		t.Fatalf("Outgoing() error = %v", err)
	}
	if len(links) != 6 {
		t.Fatalf("got %d links, want 6: %#v", len(links), links)
	}
	if links[0].TargetID != "/b" || links[0].TargetPath != "b.md" {
		t.Fatalf("root link = %#v", links[0])
	}
	if links[1].TargetPath != "docs/index.md" || links[1].TargetID != "" {
		t.Fatalf("directory link = %#v", links[1])
	}
	if links[2].TargetPath != "index.md" || links[2].Broken {
		t.Fatalf("reserved link = %#v", links[2])
	}
	if !links[3].Broken || !links[4].External || !links[5].External {
		t.Fatalf("special links = %#v", links[3:])
	}
}

func TestNavigationAndMutationErrorCases(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestConcept(t, root, "a.md", "A", "A body")
	writeTestConcept(t, root, "b.md", "B", "B body")
	bundle, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if links, err := bundle.Outgoing("/a"); err != nil || len(links) != 0 {
		t.Fatalf("Outgoing() = %#v, %v", links, err)
	}
	if _, err := bundle.Outgoing("/missing"); err == nil {
		t.Fatal("Outgoing() accepted a missing concept")
	}
	if _, err := bundle.Backlinks("/missing"); err == nil {
		t.Fatal("Backlinks() accepted a missing concept")
	}
	if _, err := bundle.Graph("/a", -1); err == nil {
		t.Fatal("Graph() accepted negative depth")
	}
	nodes, err := bundle.Graph("/a", 0)
	if err != nil || len(nodes) != 1 || nodes[0].Depth != 0 {
		t.Fatalf("depth-zero graph = %#v, %v", nodes, err)
	}
	if _, err := Add(root, "/new", NewConcept{}, false); err == nil {
		t.Fatal("Add() accepted a missing type")
	}
	if _, err := Add(root, "/a", NewConcept{Type: "Note"}, false); err == nil {
		t.Fatal("Add() overwrote an existing concept without force")
	}
	id, err := Add(root, "/new-concept", NewConcept{Type: "Note"}, false)
	if err != nil || id != "/new-concept" {
		t.Fatalf("Add() = %q, %v", id, err)
	}
	if _, err := Add(root, "/new-concept", NewConcept{Type: "Note"}, true); err != nil {
		t.Fatalf("forced Add() error = %v", err)
	}
}

func TestUpdateIndexesAndValidateBundle(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "---\nokf_version: \"0.1\"\n---\n\n<!-- manly:generated:start -->\nold\n<!-- manly:generated:end -->\n")
	writeTestFile(t, root, "group/index.md", "<!-- manly:generated:start -->\n<!-- manly:generated:end -->\n")
	writeTestConcept(t, root, "root-note.md", "Root Note", "Root content")
	writeTestConcept(t, root, "group/child.md", "Child", "Child content")

	stale, err := UpdateIndexes(root, true)
	if err != nil || len(stale) != 2 {
		t.Fatalf("UpdateIndexes(check) = %#v, %v", stale, err)
	}
	changed, err := UpdateIndexes(root, false)
	if err != nil || len(changed) != 2 {
		t.Fatalf("UpdateIndexes() = %#v, %v", changed, err)
	}
	stale, err = UpdateIndexes(root, true)
	if err != nil || len(stale) != 0 {
		t.Fatalf("second UpdateIndexes(check) = %#v, %v", stale, err)
	}
	data, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil || !strings.Contains(string(data), "Root Note") {
		t.Fatalf("updated root index = %q, %v", data, err)
	}

	writeTestFile(t, root, "log.md", "---\ntype: invalid\n---\n## not-a-date\n")
	writeTestFile(t, root, "nested/index.md", "---\ntitle: Invalid nested index\n---\n")
	writeTestFile(t, root, "bad.md", "---\ntimestamp: not-a-date\n---\n")
	writeTestFile(t, root, "broken.md", "---\ntype: Note\n---\n[Missing](missing.md)\n")
	report, err := Validate(root, true)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report.Valid() || len(report.Errors) < 3 || len(report.Warnings) < 1 {
		t.Fatalf("validation report = %#v", report)
	}
	if !hasIssue(report.Errors, "log.md must not contain frontmatter") || !hasIssue(report.Errors, "timestamp is not ISO 8601") {
		t.Fatalf("validation errors = %#v", report.Errors)
	}
	if !hasIssue(report.Warnings, "link target not found") {
		t.Fatalf("validation warnings = %#v", report.Warnings)
	}
}

func hasIssue(issues []Issue, fragment string) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Message, fragment) {
			return true
		}
	}
	return false
}
