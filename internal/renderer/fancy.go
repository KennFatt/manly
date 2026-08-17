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
	case ShowCollectionView:
		return renderFancyShowCollection(w, value)
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
	case AnalyticsView:
		return renderFancyAnalytics(w, value)
	default:
		return unsupportedView(view)
	}
}

func renderFancyList(w io.Writer, view ListView) error {
	if view.Heading != "" {
		fmt.Fprintf(w, "%s\n\n", view.Heading)
	}
	if !view.Recursive || view.ShowDirectories {
		for _, directory := range view.Directories {
			fmt.Fprintf(w, "  %-24s %d concepts\n", directory.Path, directory.Count)
		}
	}
	for _, entry := range view.Entries {
		fmt.Fprintf(w, "  %s\n", entry.Concept.ID)
		if entry.Concept.Description != "" {
			fmt.Fprintf(w, "      %s\n", entry.Concept.Description)
		}
		if !view.HideActions && !view.HideUsage {
			fmt.Fprintf(w, "      Open: manly show %s\n", entry.Concept.ID)
		}
		if view.Recursive {
			fmt.Fprintln(w)
		}
	}
	if !view.Recursive {
		fmt.Fprintf(w, "\n%d concept(s)\n", view.Count)
	}
	if view.Root != "" {
		fmt.Fprintf(w, "Root: %s\n", view.Root)
	}
	return nil
}

func renderFancyShowCollection(w io.Writer, view ShowCollectionView) error {
	for index, result := range view.Results {
		if index > 0 {
			fmt.Fprintln(w)
		}
		if err := renderFancyShow(w, ShowView{
			Concept:     result.Concept,
			Links:       result.Links,
			Backlinks:   result.Backlinks,
			Actions:     result.Actions,
			HideUsage:   result.HideUsage,
			HideActions: len(result.Actions) == 0,
		}); err != nil {
			return err
		}
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
	if len(view.Actions) > 0 && !view.HideActions && !view.HideUsage {
		fmt.Fprintln(w, "Actions:")
		for _, action := range view.Actions {
			renderAction(w, strings.Title(action.Name), action.Command)
		}
	}
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
	} else if notice := searchNotice(view); notice != "" {
		fmt.Fprintf(w, "%s.\n", strings.ToUpper(notice[:1])+notice[1:])
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
	} else if notice := contextNotice(view); notice != "" {
		fmt.Fprintf(w, "%s.\n", strings.ToUpper(notice[:1])+notice[1:])
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
	if len(view.Errors) > 0 || len(view.Warnings) > 0 {
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "Validation summary")
	fmt.Fprintf(w, "\nRoot: %s\n", view.Root)
	fmt.Fprintf(w, "Mode: %s\n", view.Mode)
	fmt.Fprintln(w, "\nBundles")
	for _, bundle := range view.Bundles {
		fmt.Fprintf(w, "  %-28s %d Markdown files, %d concepts, %d loaded, %d invalid\n", bundle.Name, bundle.MarkdownFiles, bundle.ConceptFiles, bundle.LoadedConcepts, bundle.InvalidConceptFiles)
	}
	fmt.Fprintln(w, "\nTotals")
	fmt.Fprintf(w, "  Markdown files:          %d\n", view.Stats.MarkdownFiles)
	fmt.Fprintf(w, "  Reserved files:          %d\n", view.Stats.ReservedFiles)
	fmt.Fprintf(w, "  Concept files:           %d\n", view.Stats.ConceptFiles)
	fmt.Fprintf(w, "  Loaded concepts:         %d\n", view.Stats.LoadedConcepts)
	fmt.Fprintf(w, "  Invalid concept files:   %d\n", view.Stats.InvalidConceptFiles)
	fmt.Fprintf(w, "  Links checked:           %d\n", view.Stats.LinksChecked)
	fmt.Fprintf(w, "  Broken links:            %d\n", view.Stats.BrokenLinks)
	fmt.Fprintf(w, "  Stale generated indexes: %d\n", view.Stats.StaleGeneratedIndexes)
	fmt.Fprintf(w, "  Errors:                  %d\n", view.Stats.Errors)
	fmt.Fprintf(w, "  Warnings:                %d\n", view.Stats.Warnings)
	if view.Valid {
		fmt.Fprintf(w, "\nOKF validation passed: %d warning(s)\n", len(view.Warnings))
	}
	return nil
}

func renderAction(w io.Writer, name, command string) {
	fmt.Fprintf(w, "  %-9s %s\n", name+":", command)
}
