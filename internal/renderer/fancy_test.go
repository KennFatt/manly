package renderer

import (
	"strings"
	"testing"
)

func TestFancySearchNotice(t *testing.T) {
	tests := []struct {
		name string
		view SearchView
		want string
	}{
		{
			name: "weak",
			view: SearchView{
				Query:     "decisions",
				Confident: false,
				Results: []SearchResult{
					{Score: 4, Confidence: "low", Concept: Concept{ID: "/weak", Title: "Weak match", Description: "Body only."}},
				},
			},
			want: "Search results for \"decisions\"\n\n" +
				"[1] Weak match\n" +
				"    /weak\n" +
				"    Body only.\n" +
				"    Open:    manly show /weak\n" +
				"    Context: manly context /weak\n\n" +
				"No confident match: strongest result is low confidence.\n",
		},
		{
			name: "empty",
			view: SearchView{Query: "pasta"},
			want: "Search results for \"pasta\"\n\n" +
				"No matching concepts.\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output strings.Builder
			if err := (fancyRenderer{}).Render(&output, test.view); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestFancyContextNotice(t *testing.T) {
	view := ContextView{
		Query:     "decisions",
		Confident: false,
		Results: []ContextResult{
			{Confidence: "low", Concept: Concept{ID: "/weak", Title: "Weak", Content: "Body."}},
		},
	}
	var output strings.Builder
	if err := (fancyRenderer{}).Render(&output, view); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := "Context for \"decisions\"\n\n" +
		"## Weak\n" +
		"ID: /weak\n\nBody.\n\n" +
		"Open: manly show /weak\n\n" +
		"No confident match: strongest result is low confidence.\n"
	if got := output.String(); got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}
