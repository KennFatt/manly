package renderer

import (
	"strings"
	"testing"
)

func renderForTest(t *testing.T, view View) string {
	t.Helper()
	var output strings.Builder
	if err := (markdownRenderer{}).Render(&output, view); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}

func TestMarkdownListBundleDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "present",
			description: "Shared engineering conventions.",
			want:        "# Engineering Preferences\n\n" + "Shared engineering conventions.\n\n" + "* [Naming](/general/naming.md) - Use clear names.\n",
		},
		{
			name: "absent",
			want: "# Engineering Preferences\n\n" + "* [Naming](/general/naming.md) - Use clear names.\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := ListView{
				Heading:     "Engineering Preferences",
				Description: test.description,
				Entries: []ListEntry{{Concept: Concept{
					ID:          "/general/naming",
					Title:       "Naming",
					Description: "Use clear names.",
				}}},
			}
			if got := renderForTest(t, view); got != test.want {
				t.Errorf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestMarkdownSearchNotice(t *testing.T) {
	tests := []struct {
		name string
		view SearchView
		want string
	}{
		{
			name: "confident",
			view: SearchView{
				Query:     "ask",
				Confident: true,
				Results: []SearchResult{
					{Confidence: "high", Concept: Concept{ID: "/ask", Title: "Ask", Description: "Ask questions."}},
				},
			},
			want: "# Search results for \"ask\"\n\n" +
				"* [Ask](/ask.md) - Ask questions.\n",
		},
		{
			name: "weak",
			view: SearchView{
				Query:     "decisions",
				Confident: false,
				Results: []SearchResult{
					{Confidence: "low", Concept: Concept{ID: "/flowcharts", Title: "Flowcharts", Description: "Model processes."}},
				},
			},
			want: "# Search results for \"decisions\"\n\n" +
				"* [Flowcharts](/flowcharts.md) - Model processes.\n" +
				"\n> no confident match: strongest result is low confidence\n",
		},
		{
			name: "empty",
			view: SearchView{Query: "pasta"},
			want: "# Search results for \"pasta\"\n\n" +
				"> no concepts matched\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := renderForTest(t, test.view); got != test.want {
				t.Errorf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestMarkdownContextNotice(t *testing.T) {
	tests := []struct {
		name string
		view ContextView
		want string
	}{
		{
			name: "weak",
			view: ContextView{
				Query:     "decisions",
				Confident: false,
				Results: []ContextResult{
					{Confidence: "low", Concept: Concept{ID: "/flowcharts", Title: "Flowcharts", Content: "Model processes."}},
				},
			},
			want: "## Flowcharts\n\nModel processes.\n\n" +
				"> no confident match: strongest result is low confidence\n",
		},
		{
			name: "empty",
			view: ContextView{Query: "pasta"},
			want: "> no concepts matched\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := renderForTest(t, test.view); got != test.want {
				t.Errorf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}
