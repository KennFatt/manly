package renderer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func treeTestView() TreeView {
	return TreeView{
		Root:         "/root",
		Path:         "/bundle",
		Heading:      "Bundle",
		Description:  "Bundle description.",
		Depth:        1,
		ConceptCount: 2,
		Nodes: []TreeNode{
			{
				Kind:         "directory",
				Path:         "/bundle/general",
				Title:        "General",
				Description:  "General description.",
				ConceptCount: 2,
				Hidden:       &TreeCut{Directories: 1, Concepts: 2},
			},
			{
				Kind:        "concept",
				Path:        "/bundle/note",
				Title:       "Note",
				Description: strings.Repeat("x", 101),
				Type:        "Note",
				Tags:        []string{"go"},
			},
			{
				Kind:  "concept",
				Path:  "/bundle/bare",
				Title: "Bare",
			},
		},
	}
}

func TestCompactTree(t *testing.T) {
	view := treeTestView()
	view.Description = strings.Repeat("y", 101)
	var output strings.Builder
	if err := (compactRenderer{}).Render(&output, view); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := "Bundle  2 concepts\n" +
		strings.Repeat("y", 100) + "...\n" +
		"/bundle\n" +
		"\n" +
		"├── /bundle/general/  General  2 concepts  General description.  (... 1 subdirectory, 2 concepts below)\n" +
		"├── /bundle/note  Note  " + strings.Repeat("x", 100) + "...\n" +
		"└── /bundle/bare  Bare\n" +
		"\n" +
		"Expand: manly tree /bundle --depth 1\n" +
		"Root: /root\n"
	if got := output.String(); got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestFancyTree(t *testing.T) {
	var output strings.Builder
	if err := (fancyRenderer{}).Render(&output, treeTestView()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	got := output.String()
	for _, fragment := range []string{
		"Bundle  2 concepts\n",
		"Bundle description.\n",
		"├── /bundle/general/  General  2 concepts  (... 1 subdirectory, 2 concepts below)\n",
		"General description.\n",
		"/bundle/note  Note  (Note)\n",
		strings.Repeat("x", 101) + "\n",
		"└── /bundle/bare  Bare\n",
		"Expand: manly tree /bundle --depth 1\n",
		"Root: /root\n",
	} {
		if !strings.Contains(got, fragment) {
			t.Errorf("fancy tree %q does not contain %q", got, fragment)
		}
	}
}

func TestMarkdownTree(t *testing.T) {
	var output strings.Builder
	if err := (markdownRenderer{}).Render(&output, treeTestView()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	got := output.String()
	for _, fragment := range []string{
		"# Bundle\n\nBundle description.\n\n",
		"* [General](/bundle/general/) (2 concepts) - General description.  (... 1 subdirectory, 2 concepts below)\n",
		"* [Note](/bundle/note.md) - Note - " + strings.Repeat("x", 101) + "\n",
		"* [Bare](/bundle/bare.md)\n",
		"\n**Root:** /root\n",
	} {
		if !strings.Contains(got, fragment) {
			t.Errorf("markdown tree %q does not contain %q", got, fragment)
		}
	}
}

func TestJSONTree(t *testing.T) {
	var output bytes.Buffer
	if err := (jsonRenderer{}).Render(&output, treeTestView()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output.String())
	}
	if decoded["root"] != "/root" || decoded["path"] != "/bundle" || decoded["depth"] != float64(1) {
		t.Fatalf("decoded tree = %v", decoded)
	}
	nodes, ok := decoded["nodes"].([]any)
	if !ok || len(nodes) != 3 {
		t.Fatalf("nodes = %v, want 3 nodes", decoded["nodes"])
	}
	first := nodes[0].(map[string]any)
	if first["kind"] != "directory" || first["concept_count"] != float64(2) {
		t.Fatalf("first node = %v", first)
	}
	hidden, ok := first["hidden"].(map[string]any)
	if !ok || hidden["directories"] != float64(1) || hidden["concepts"] != float64(2) {
		t.Fatalf("hidden = %v", first["hidden"])
	}
}

func TestAgentTreeOmitsRoot(t *testing.T) {
	var output bytes.Buffer
	if err := (agentRenderer{}).Render(&output, treeTestView()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output.String())
	}
	if _, ok := decoded["root"]; ok {
		t.Fatal("agent tree should omit root")
	}
	if decoded["path"] != "/bundle" || decoded["heading"] != "Bundle" {
		t.Fatalf("agent tree = %v", decoded)
	}
	nodes, ok := decoded["nodes"].([]any)
	if !ok || len(nodes) != 3 {
		t.Fatalf("agent nodes = %v, want 3 nodes", decoded["nodes"])
	}
	if nodes[2].(map[string]any)["description"] != nil {
		t.Fatalf("bare node should omit description: %v", nodes[2])
	}
}
