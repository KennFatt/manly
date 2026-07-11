package renderer

import (
	"fmt"
	"io"
	"strings"
)

type fancyRenderer struct{}

var _ Renderer = fancyRenderer{}

func (fancyRenderer) Format() Format {
	return FormatFancy
}

func (fancyRenderer) Render(w io.Writer, view View) error {
	switch value := view.(type) {
	case ListView:
		return renderFancyList(w, value)
	case ShowView:
		return renderFancyShow(w, value)
	case SearchView:
		return renderFancySearch(w, value)
	case ContextView:
		return renderFancyContext(w, value)
	case LinksView:
		return renderFancyLinks(w, value)
	case BacklinksView:
		return renderFancyBacklinks(w, value)
	case GraphView:
		return renderFancyGraph(w, value)
	case CheckView:
		return renderFancyCheck(w, value)
	default:
		return unsupportedView(view)
	}
}

func renderFancyList(w io.Writer, view ListView) error {
	if view.Heading != "" {
		fmt.Fprintf(w, "%s\n\n", view.Heading)
	}
	if !view.Recursive {
		for _, directory := range view.Directories {
			fmt.Fprintf(w, "  %-24s %d concepts\n", directory.Path, directory.Count)
		}
	}
	for _, entry := range view.Entries {
		fmt.Fprintf(w, "  %s\n", entry.Concept.ID)
		if entry.Concept.Description != "" {
			fmt.Fprintf(w, "      %s\n", entry.Concept.Description)
		}
		fmt.Fprintf(w, "      Open: manly show %s\n", entry.Concept.ID)
		if view.Recursive {
			fmt.Fprintln(w)
		}
	}
	if !view.Recursive {
		fmt.Fprintf(w, "\n%d concept(s)\n", view.Count)
	}
	return nil
}

func renderFancyShow(w io.Writer, view ShowView) error {
	fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(view.Concept.Content))
	if len(view.Links) > 0 {
		fmt.Fprintln(w, "Links:")
		for index, link := range view.Links {
			if link.Target != "" {
				fmt.Fprintf(w, "[%d] %s\n    %s\n    manly show %s\n\n", index+1, link.Label, link.Target, link.Target)
			} else if command := linkNavigationCommand(link); command != "" {
				fmt.Fprintf(w, "[%d] %s\n    %s\n    %s\n\n", index+1, link.Label, link.TargetPath, command)
			} else if link.URL != "" {
				fmt.Fprintf(w, "[%d] %s\n    %s\n\n", index+1, link.Label, link.URL)
			} else {
				fmt.Fprintf(w, "[%d] %s\n    broken: %s\n\n", index+1, link.Label, link.URL)
			}
		}
	}
	if len(view.Backlinks) > 0 {
		fmt.Fprintln(w, "Backlinks:")
		for _, backlink := range view.Backlinks {
			fmt.Fprintf(w, "  %s (%s)\n", backlink.Target, backlink.Label)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "Actions:")
	renderAction(w, "Open", "manly show "+view.Concept.ID)
	renderAction(w, "Context", "manly context "+view.Concept.ID)
	renderAction(w, "Edit", "manly edit "+view.Concept.ID)
	renderAction(w, "Backlinks", "manly backlinks "+view.Concept.ID)
	return nil
}

func renderFancySearch(w io.Writer, view SearchView) error {
	fmt.Fprintf(w, "Search results for %q\n\n", view.Query)
	for index, result := range view.Results {
		fmt.Fprintf(w, "[%d] %s\n", index+1, result.Concept.Title)
		fmt.Fprintf(w, "    %s\n", result.Concept.ID)
		if result.Concept.Description != "" {
			fmt.Fprintf(w, "    %s\n", result.Concept.Description)
		}
		fmt.Fprintf(w, "    Open:    manly show %s\n", result.Concept.ID)
		fmt.Fprintf(w, "    Context: manly context %s\n\n", result.Concept.ID)
	}
	if len(view.Results) == 0 {
		fmt.Fprintln(w, "No matching concepts.")
	}
	return nil
}

func renderFancyContext(w io.Writer, view ContextView) error {
	fmt.Fprintf(w, "Context for %q\n\n", view.Query)
	for _, result := range view.Results {
		fmt.Fprintf(w, "## %s\n", result.Concept.Title)
		fmt.Fprintf(w, "ID: %s\n\n%s\n\n", result.Concept.ID, strings.TrimSpace(result.Concept.Content))
		fmt.Fprintf(w, "Open: manly show %s\n\n", result.Concept.ID)
	}
	if len(view.Results) == 0 {
		fmt.Fprintln(w, "No matching concepts.")
	}
	return nil
}

func renderFancyLinks(w io.Writer, view LinksView) error {
	fmt.Fprintf(w, "Links from %s\n\n", view.Source)
	for index, link := range view.Links {
		if link.Target != "" {
			fmt.Fprintf(w, "[%d] %s\n    %s\n    manly show %s\n\n", index+1, link.Label, link.Target, link.Target)
		} else if command := linkNavigationCommand(link); command != "" {
			fmt.Fprintf(w, "[%d] %s\n    %s\n    %s\n\n", index+1, link.Label, link.TargetPath, command)
		} else {
			fmt.Fprintf(w, "[%d] %s\n    %s\n\n", index+1, link.Label, link.URL)
		}
	}
	if len(view.Links) == 0 {
		fmt.Fprintln(w, "No links.")
	}
	return nil
}

func renderFancyBacklinks(w io.Writer, view BacklinksView) error {
	fmt.Fprintf(w, "Backlinks to %s\n\n", view.Target)
	for index, link := range view.Backlinks {
		title := link.Title
		if title == "" {
			title = link.Target
		}
		fmt.Fprintf(w, "[%d] %s\n    %s (%s)\n    manly show %s\n\n", index+1, title, link.Target, link.Label, link.Target)
	}
	if len(view.Backlinks) == 0 {
		fmt.Fprintln(w, "No backlinks.")
	}
	return nil
}

func renderFancyGraph(w io.Writer, view GraphView) error {
	fmt.Fprintln(w, "Concept graph")
	for _, node := range view.Nodes {
		fmt.Fprintf(w, "%*s%s  %s\n", node.Depth*2, "", node.ID, node.Title)
	}
	return nil
}

func renderFancyCheck(w io.Writer, view CheckView) error {
	for _, issue := range view.Errors {
		fmt.Fprintf(w, "ERROR: %s: %s\n", issue.Path, issue.Message)
	}
	for _, issue := range view.Warnings {
		fmt.Fprintf(w, "WARNING: %s: %s\n", issue.Path, issue.Message)
	}
	if view.Valid {
		fmt.Fprintf(w, "OKF validation passed: %d warning(s)\n", len(view.Warnings))
	}
	return nil
}

func renderAction(w io.Writer, name, command string) {
	fmt.Fprintf(w, "  %-9s %s\n", name+":", command)
}
