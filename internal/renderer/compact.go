package renderer

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type compactRenderer struct{}

var _ Renderer = compactRenderer{}

func (compactRenderer) Format() Format {
	return FormatCompact
}

func (compactRenderer) Render(w io.Writer, view View) error {
	switch value := view.(type) {
	case ListView:
		return renderCompactList(w, value)
	case ShowView:
		return renderCompactShow(w, value)
	case SearchView:
		return renderCompactSearch(w, value)
	case ContextView:
		return renderCompactContext(w, value)
	case LinksView:
		return renderCompactLinks(w, value)
	case BacklinksView:
		return renderCompactBacklinks(w, value)
	case GraphView:
		return renderCompactGraph(w, value)
	case CheckView:
		return renderCompactCheck(w, value)
	default:
		return unsupportedView(view)
	}
}

type compactListRow struct {
	Key   string
	Value string
}

func renderCompactList(w io.Writer, view ListView) error {
	rows := make([]compactListRow, 0, len(view.Directories)+len(view.Entries))
	if !view.Recursive {
		for _, directory := range view.Directories {
			rows = append(rows, compactListRow{
				Key:   directory.Path + "/",
				Value: fmt.Sprintf("%d concepts", directory.Count),
			})
		}
	}
	for _, entry := range view.Entries {
		rows = append(rows, compactListRow{Key: entry.Concept.ID, Value: entry.Concept.Title})
	}

	keyWidth := compactListKeyWidth(rows)
	firstHeader, secondHeader := "ID", "TITLE"
	if !view.Recursive {
		firstHeader, secondHeader = "PATH", "TITLE / CONCEPTS"
	}
	fmt.Fprintf(w, "%-*s  %s\n", keyWidth, firstHeader, secondHeader)
	for _, row := range rows {
		fmt.Fprintf(w, "%-*s  %s\n", keyWidth, row.Key, row.Value)
	}
	fmt.Fprintln(w)
	if view.Recursive {
		fmt.Fprintln(w, "Details: manly show <ID>")
	} else {
		fmt.Fprintln(w, "List: manly list <PATH>")
	}
	if view.Root != "" {
		fmt.Fprintf(w, "Root: %s\n", view.Root)
	}
	return nil
}

func compactListKeyWidth(rows []compactListRow) int {
	width := utf8.RuneCountInString("ID")
	for _, row := range rows {
		rowWidth := utf8.RuneCountInString(row.Key)
		if rowWidth > width {
			width = rowWidth
		}
	}
	return width
}

func renderCompactShow(w io.Writer, view ShowView) error {
	fmt.Fprintf(w, "%s\n%s\n", view.Concept.ID, strings.TrimSpace(view.Concept.Content))
	return nil
}

func renderCompactSearch(w io.Writer, view SearchView) error {
	for _, result := range view.Results {
		fmt.Fprintf(w, "%.2f\t%s\t%s\n", result.Score, result.Concept.ID, result.Concept.Title)
	}
	return nil
}

func renderCompactContext(w io.Writer, view ContextView) error {
	for _, result := range view.Results {
		fmt.Fprintf(w, "%s\n%s\n\n", result.Concept.ID, strings.TrimSpace(result.Concept.Content))
	}
	return nil
}

func renderCompactLinks(w io.Writer, view LinksView) error {
	for _, link := range view.Links {
		fmt.Fprintf(w, "%s\t%s\n", link.Label, compactLinkTarget(link))
	}
	return nil
}

func renderCompactBacklinks(w io.Writer, view BacklinksView) error {
	for _, link := range view.Backlinks {
		fmt.Fprintf(w, "%s\t%s\n", link.Target, link.Label)
	}
	return nil
}

func renderCompactGraph(w io.Writer, view GraphView) error {
	for _, node := range view.Nodes {
		fmt.Fprintf(w, "%d\t%s\t%s\n", node.Depth, node.ID, node.Title)
	}
	return nil
}

func renderCompactCheck(w io.Writer, view CheckView) error {
	for _, issue := range view.Errors {
		fmt.Fprintf(w, "ERROR\t%s\t%s\n", issue.Path, issue.Message)
	}
	for _, issue := range view.Warnings {
		fmt.Fprintf(w, "WARNING\t%s\t%s\n", issue.Path, issue.Message)
	}
	if view.Valid {
		fmt.Fprintf(w, "OKF validation passed\t%d warning(s)\n", len(view.Warnings))
	}
	return nil
}
