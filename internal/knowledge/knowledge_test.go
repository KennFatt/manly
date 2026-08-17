package knowledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadResolvesLinksAndBacklinks(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestConcept(t, root, "a.md", "A", "[B](b.md) and [C](/c.md) and [External](https://example.com).")
	writeTestConcept(t, root, "b.md", "B", "B body.")
	writeTestConcept(t, root, "c.md", "C", "C body.")

	bundle, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	concept, err := bundle.Get("/a")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(concept.Links) != 3 {
		t.Fatalf("got %d links, want 3", len(concept.Links))
	}
	if concept.Links[0].TargetID != "/b" || concept.Links[1].TargetID != "/c" || !concept.Links[2].External {
		t.Fatalf("unexpected links: %+v", concept.Links)
	}

	backlinks, err := bundle.Backlinks("/b")
	if err != nil {
		t.Fatalf("Backlinks() error = %v", err)
	}
	if len(backlinks) != 1 || backlinks[0].Concept.ID != "/a" {
		t.Fatalf("unexpected backlinks: %+v", backlinks)
	}
}

func TestLoadReadsBundleMetadata(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "---\ntitle: Engineering Preferences\ndescription: Shared engineering conventions.\n---\n")

	bundle, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if bundle.Title != "Engineering Preferences" {
		t.Fatalf("bundle title = %q", bundle.Title)
	}
	if bundle.Description != "Shared engineering conventions." {
		t.Fatalf("bundle description = %q", bundle.Description)
	}

	rootWithoutDescription := t.TempDir()
	writeTestFile(t, rootWithoutDescription, "index.md", "# Knowledge Bundle\n")
	bundleWithoutDescription, err := Load(rootWithoutDescription)
	if err != nil {
		t.Fatalf("Load() without description error = %v", err)
	}
	if bundleWithoutDescription.Description != "" {
		t.Fatalf("bundle description without metadata = %q", bundleWithoutDescription.Description)
	}
}

func TestSearchAndGraph(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestConcept(t, root, "a.md", "Boundary Rules", "External data must be treated as partial.")
	writeTestConcept(t, root, "b.md", "Type Safety", "Use safe types.")
	writeTestConcept(t, root, "c.md", "Unrelated", "Different topic.")
	writeTestFile(t, root, "a.md", "---\ntype: Guideline\ntitle: Boundary Rules\ntags: [safety]\n---\n\nExternal data must be treated as partial. See [Type Safety](b.md).\n")

	bundle, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	results := Search(bundle, "external data", SearchOptions{Limit: 5})
	if len(results) == 0 || results[0].Concept.ID != "/a" {
		t.Fatalf("unexpected search results: %+v", results)
	}
	nodes, err := bundle.Graph("/a", 1)
	if err != nil {
		t.Fatalf("Graph() error = %v", err)
	}
	if len(nodes) != 2 || nodes[1].Concept.ID != "/b" || nodes[1].Depth != 1 {
		t.Fatalf("unexpected graph: %+v", nodes)
	}
}

func TestAddAndMoveUpdatesLinks(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Bundle\n")
	writeTestConcept(t, root, "a.md", "A", "See [B](b.md).")
	writeTestConcept(t, root, "b.md", "B", "B body.")

	id, err := Add(root, "/new-concept", NewConcept{Type: "Guideline", Title: "New Concept"}, false)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if id != "/new-concept" {
		t.Fatalf("Add() ID = %q", id)
	}
	if _, err := os.Stat(filepath.Join(root, "new-concept.md")); err != nil {
		t.Fatalf("created concept missing: %v", err)
	}

	changed, err := Move(root, "/b", "/renamed")
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	if changed != 1 {
		t.Fatalf("Move() changed %d links, want 1", changed)
	}
	data, err := os.ReadFile(filepath.Join(root, "a.md"))
	if err != nil {
		t.Fatalf("read moved link: %v", err)
	}
	if !strings.Contains(string(data), "renamed.md") {
		t.Fatalf("moved link was not updated: %s", data)
	}
}

func TestCanonicalIDRejectsReservedFiles(t *testing.T) {
	if _, err := CanonicalID("/index"); err == nil {
		t.Fatal("CanonicalID() accepted reserved index")
	}
	if _, err := CanonicalID("../outside"); err == nil {
		t.Fatal("CanonicalID() accepted path traversal")
	}
}

func writeTestConcept(t *testing.T, root, name, title, body string) {
	t.Helper()
	writeTestFile(t, root, name, "---\ntype: Guideline\ntitle: "+title+"\n---\n\n"+body+"\n")
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
