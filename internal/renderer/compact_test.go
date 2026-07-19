package renderer

import (
	"strings"
	"testing"
)

func TestCompactTables(t *testing.T) {
	tests := []struct {
		name string
		view View
		want string
	}{
		{
			name: "search",
			view: SearchView{
				Results: []SearchResult{
					{Score: 12, Concept: Concept{ID: "/short", Title: "Short"}},
					{Score: 3.5, Concept: Concept{ID: "/engineering/preferences", Title: "Longer title"}},
				},
			},
			want: "SCORE  ID                        TITLE\n" +
				"12.00  /short                    Short\n" +
				"3.50   /engineering/preferences  Longer title\n",
		},
		{
			name: "links",
			view: LinksView{
				Links: []Link{
					{Label: "short", Target: "/internal/topic"},
					{Label: "long label", TargetPath: "docs/file.md"},
					{Label: "external", URL: "https://example.com"},
					{Label: "broken", Broken: true},
				},
			},
			want: "LABEL       TARGET\n" +
				"short       /internal/topic\n" +
				"long label  docs/file.md\n" +
				"external    https://example.com\n" +
				"broken      broken\n",
		},
		{
			name: "backlinks",
			view: BacklinksView{
				Backlinks: []Link{
					{Target: "/source/short", Label: "mentioned"},
					{Target: "/source/longer-concept", Label: "a longer label"},
				},
			},
			want: "SOURCE                  LABEL\n" +
				"/source/short           mentioned\n" +
				"/source/longer-concept  a longer label\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output strings.Builder
			if err := (compactRenderer{}).Render(&output, test.view); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCompactTablesEmpty(t *testing.T) {
	tests := []struct {
		name string
		view View
		want string
	}{
		{name: "search", view: SearchView{}, want: "SCORE  ID  TITLE\n"},
		{name: "links", view: LinksView{}, want: "LABEL  TARGET\n"},
		{name: "backlinks", view: BacklinksView{}, want: "SOURCE  LABEL\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output strings.Builder
			if err := (compactRenderer{}).Render(&output, test.view); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}
