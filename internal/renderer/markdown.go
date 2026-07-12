package renderer

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type markdownRenderer struct{}

var _ Renderer = markdownRenderer{}

func (markdownRenderer) Format() Format {
	return FormatMarkdown
}

func (markdownRenderer) Render(w io.Writer, view View) error {
	switch value := view.(type) {
	case ListView:
		return renderMarkdownList(w, value)
	case ShowView:
		return renderMarkdownShow(w, value)
	case SearchView:
		return renderMarkdownSearch(w, value)
	case ContextView:
		return renderMarkdownContext(w, value)
	case LinksView:
		return renderMarkdownLinks(w, value)
	case BacklinksView:
		return renderMarkdownBacklinks(w, value)
	case GraphView:
		return renderMarkdownGraph(w, value)
	case CheckView:
		return renderMarkdownCheck(w, value)
	default:
		return unsupportedView(view)
	}
}

func renderMarkdownList(w io.Writer, view ListView) error {
	if view.Heading != "" {
		fmt.Fprintf(w, "# %s\n\n", view.Heading)
	}
	for _, directory := range view.Directories {
		fmt.Fprintf(w, "* [%s](%s/)\n", filepath.Base(directory.Path), directory.Path)
	}
	for _, entry := range view.Entries {
		concept := entry.Concept
		fmt.Fprintf(w, "* [%s](%s.md) - %s\n", concept.Title, concept.ID, concept.Description)
	}
	if view.Root != "" {
		fmt.Fprintf(w, "\n**Root:** %s\n", view.Root)
	}
	return nil
}

func renderMarkdownShow(w io.Writer, view ShowView) error {
	fmt.Fprintf(w, "# %s\n\n%s\n", view.Concept.Title, strings.TrimSpace(view.Concept.Content))
	return nil
}

func renderMarkdownSearch(w io.Writer, view SearchView) error {
	fmt.Fprintf(w, "# Search results for %q\n\n", view.Query)
	for _, result := range view.Results {
		fmt.Fprintf(w, "* [%s](%s.md) - %s\n", result.Concept.Title, result.Concept.ID, result.Concept.Description)
	}
	return nil
}

func renderMarkdownContext(w io.Writer, view ContextView) error {
	for _, result := range view.Results {
		fmt.Fprintf(w, "## %s\n\n%s\n\n", result.Concept.Title, strings.TrimSpace(result.Concept.Content))
	}
	return nil
}

func renderMarkdownLinks(w io.Writer, view LinksView) error {
	fmt.Fprintf(w, "# Links from %s\n\n", view.Source)
	for _, link := range view.Links {
		fmt.Fprintf(w, "* [%s](%s)\n", link.Label, link.URL)
	}
	return nil
}

func renderMarkdownBacklinks(w io.Writer, view BacklinksView) error {
	fmt.Fprintf(w, "# Backlinks to %s\n\n", view.Target)
	for _, link := range view.Backlinks {
		fmt.Fprintf(w, "* [%s](%s.md)\n", link.Label, link.Target)
	}
	return nil
}

func renderMarkdownGraph(w io.Writer, view GraphView) error {
	fmt.Fprintln(w, "# Concept graph")
	for _, node := range view.Nodes {
		fmt.Fprintf(w, "* [%s](%s.md) - depth %d\n", node.Title, node.ID, node.Depth)
	}
	return nil
}

func renderMarkdownCheck(w io.Writer, view CheckView) error {
	for _, issue := range view.Errors {
		fmt.Fprintf(w, "* ERROR: %s: %s\n", issue.Path, issue.Message)
	}
	for _, issue := range view.Warnings {
		fmt.Fprintf(w, "* WARNING: %s: %s\n", issue.Path, issue.Message)
	}
	if view.Valid {
		fmt.Fprintf(w, "* OKF validation passed: %d warning(s)\n", len(view.Warnings))
	}
	return nil
}
