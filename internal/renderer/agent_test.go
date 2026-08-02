package renderer

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestAgentListContainsSelectionMetadataOnly(t *testing.T) {
	view := ListView{
		Root:      "/private/root",
		Path:      "/react",
		Recursive: true,
		Directories: []Directory{
			{Path: "/react/hooks", Count: 2},
		},
		Entries: []ListEntry{
			{
				Concept: Concept{
					ID:          "/react/hooks-as-controllers",
					Path:        "react/hooks-as-controllers.md",
					Type:        "Guideline",
					Title:       "Hooks as Controllers",
					Description: "Keep components thin.",
					Tags:        []string{"react", "hooks"},
					Content:     "This body must not be emitted.",
				},
				Actions: []Action{{Name: "edit", Command: "manly edit /react/hooks-as-controllers"}},
			},
		},
	}

	var output bytes.Buffer
	renderer, err := New(FormatAgent)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := renderer.Render(&output, view); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output.String())
	}
	if _, ok := decoded["root"]; ok {
		t.Fatal("agent output should omit root")
	}
	if _, ok := decoded["entries"]; ok {
		t.Fatal("agent output should use concepts instead of entries")
	}
	if _, ok := decoded["actions"]; ok {
		t.Fatal("agent output should omit actions")
	}
	if got := decoded["path"]; got != "/react" {
		t.Fatalf("path = %v, want /react", got)
	}
	if got := decoded["recursive"]; got != true {
		t.Fatalf("recursive = %v, want true", got)
	}
	if got := decoded["truncated"]; got != false {
		t.Fatalf("truncated = %v, want false", got)
	}

	directories, ok := decoded["directories"].([]any)
	if !ok || len(directories) != 1 || directories[0] != "/react/hooks" {
		t.Fatalf("directories = %v, want one path", decoded["directories"])
	}
	concepts, ok := decoded["concepts"].([]any)
	if !ok || len(concepts) != 1 {
		t.Fatalf("concepts = %v, want one concept", decoded["concepts"])
	}
	concept := concepts[0].(map[string]any)
	if concept["id"] != "/react/hooks-as-controllers" || concept["type"] != "Guideline" || concept["title"] != "Hooks as Controllers" || concept["description"] != "Keep components thin." {
		t.Fatalf("concept metadata = %v", concept)
	}
	if _, ok := concept["path"]; ok {
		t.Fatal("agent concept should omit duplicate path")
	}
	if _, ok := concept["content"]; ok {
		t.Fatal("agent concept should omit content")
	}
	if _, ok := concept["actions"]; ok {
		t.Fatal("agent concept should omit actions")
	}
}

func TestAgentRendererRejectsNonListViews(t *testing.T) {
	var output bytes.Buffer
	renderer, err := New(FormatAgent)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := renderer.Render(&output, ShowView{}); err == nil {
		t.Fatal("agent renderer accepted a non-list view")
	}
}
