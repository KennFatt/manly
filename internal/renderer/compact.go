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
	case ShowCollectionView:
		return renderCompactShowCollection(w, value)
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
	case AnalyticsView:
		return renderCompactAnalytics(w, value)
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
	if !view.Recursive || view.ShowDirectories {
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
	if !view.HideUsage {
		fmt.Fprintln(w)
		if view.Recursive {
			fmt.Fprintln(w, "Details: manly show <ID>")
		} else {
			fmt.Fprintln(w, "List: manly list <PATH>")
		}
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

func renderCompactShowCollection(w io.Writer, view ShowCollectionView) error {
	for index, result := range view.Results {
		if index > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s\n%s\n", result.Concept.ID, strings.TrimSpace(result.Concept.Content))
	}
	return nil
}

func renderCompactSearch(w io.Writer, view SearchView) error {
	rows := make([][]string, 0, len(view.Results))
	for _, result := range view.Results {
		rows = append(rows, []string{
			fmt.Sprintf("%.2f", result.Score),
			result.Concept.ID,
			result.Concept.Title,
		})
	}
	if err := renderCompactTable(w, []string{"SCORE", "ID", "TITLE"}, rows); err != nil {
		return err
	}
	if notice := searchNotice(view); notice != "" {
		fmt.Fprintln(w, notice)
	}
	return nil
}

func renderCompactContext(w io.Writer, view ContextView) error {
	for _, result := range view.Results {
		fmt.Fprintf(w, "%s\n%s\n\n", result.Concept.ID, strings.TrimSpace(result.Concept.Content))
	}
	if notice := contextNotice(view); notice != "" {
		fmt.Fprintln(w, notice)
	}
	return nil
}

func renderCompactLinks(w io.Writer, view LinksView) error {
	rows := make([][]string, 0, len(view.Links))
	for _, link := range view.Links {
		rows = append(rows, []string{link.Label, compactLinkTarget(link)})
	}
	return renderCompactTable(w, []string{"LABEL", "TARGET"}, rows)
}

func renderCompactBacklinks(w io.Writer, view BacklinksView) error {
	rows := make([][]string, 0, len(view.Backlinks))
	for _, link := range view.Backlinks {
		rows = append(rows, []string{link.Target, link.Label})
	}
	return renderCompactTable(w, []string{"SOURCE", "LABEL"}, rows)
}

func renderCompactTable(w io.Writer, headers []string, rows [][]string) error {
	widths := make([]int, len(headers))
	for index, header := range headers {
		widths[index] = utf8.RuneCountInString(header)
	}
	for rowIndex, row := range rows {
		if len(row) != len(headers) {
			return fmt.Errorf("compact table row %d has %d columns, want %d", rowIndex, len(row), len(headers))
		}
		for index, value := range row {
			if width := utf8.RuneCountInString(value); width > widths[index] {
				widths[index] = width
			}
		}
	}

	if err := writeCompactTableRow(w, headers, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writeCompactTableRow(w, row, widths); err != nil {
			return err
		}
	}
	return nil
}

func writeCompactTableRow(w io.Writer, values []string, widths []int) error {
	for index, value := range values {
		if index > 0 {
			if _, err := io.WriteString(w, "  "); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, value); err != nil {
			return err
		}
		if index < len(values)-1 {
			padding := widths[index] - utf8.RuneCountInString(value)
			if _, err := io.WriteString(w, strings.Repeat(" ", padding)); err != nil {
				return err
			}
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
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
	if len(view.Errors) > 0 || len(view.Warnings) > 0 {
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "Root: %s\n", view.Root)
	fmt.Fprintf(w, "Mode: %s\n", view.Mode)
	fmt.Fprintf(w, "Bundles: %d\n", view.Stats.Bundles)
	fmt.Fprintf(w, "Markdown files: %d\n", view.Stats.MarkdownFiles)
	fmt.Fprintf(w, "Reserved files: %d\n", view.Stats.ReservedFiles)
	fmt.Fprintf(w, "Concept files: %d\n", view.Stats.ConceptFiles)
	fmt.Fprintf(w, "Loaded concepts: %d\n", view.Stats.LoadedConcepts)
	fmt.Fprintf(w, "Invalid concept files: %d\n", view.Stats.InvalidConceptFiles)
	fmt.Fprintf(w, "Links checked: %d\n", view.Stats.LinksChecked)
	fmt.Fprintf(w, "Broken links: %d\n", view.Stats.BrokenLinks)
	fmt.Fprintf(w, "Stale generated indexes: %d\n", view.Stats.StaleGeneratedIndexes)
	fmt.Fprintf(w, "Errors: %d\n", view.Stats.Errors)
	fmt.Fprintf(w, "Warnings: %d\n", view.Stats.Warnings)
	if view.Valid {
		fmt.Fprintf(w, "OKF validation passed\t%d warning(s)\n", len(view.Warnings))
	}
	return nil
}
