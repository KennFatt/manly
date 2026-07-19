package main

import (
	"io"
	"os"

	"github.com/KennFatt/manly/internal/knowledge"
	"github.com/KennFatt/manly/internal/renderer"
)

func renderConceptList(root string, bundle *knowledge.Bundle, concepts []*knowledge.Concept, format outputFormat, heading string) error {
	view := renderer.ListView{
		Root:      root,
		Path:      heading,
		Heading:   heading,
		Recursive: true,
		Entries:   conceptEntries(concepts),
	}
	return renderOutput(os.Stdout, format, view)
}

func renderJSONRecursiveDirectory(w io.Writer, root string, prefix string, directories []string, concepts []*knowledge.Concept) error {
	view := renderer.ListView{
		Root:        root,
		Path:        directoryDisplay(prefix),
		Recursive:   true,
		Directories: directoryEntries(directories, nil),
		Entries:     conceptEntries(concepts),
	}
	return renderOutput(w, formatJSON, view)
}

func renderDirectoryContents(root string, bundle *knowledge.Bundle, prefix string, directories []string, concepts []*knowledge.Concept, format outputFormat) error {
	view := renderer.ListView{
		Root:        root,
		Path:        directoryDisplay(prefix),
		Heading:     bundleDirectoryTitle(bundle, prefix),
		Directories: directoryEntries(directories, bundle),
		Entries:     conceptEntries(concepts),
		Count:       countConceptsUnder(bundle, prefix),
	}
	return renderOutput(os.Stdout, format, view)
}

func conceptEntries(concepts []*knowledge.Concept) []renderer.ListEntry {
	entries := make([]renderer.ListEntry, 0, len(concepts))
	for _, concept := range concepts {
		entries = append(entries, renderer.ListEntry{
			Concept: viewConcept(concept, false),
			Actions: actionViews(concept.ID),
		})
	}
	return entries
}

func directoryEntries(directories []string, bundle *knowledge.Bundle) []renderer.Directory {
	entries := make([]renderer.Directory, 0, len(directories))
	for _, directory := range directories {
		count := 0
		if bundle != nil {
			count = countConceptsUnder(bundle, trimDirectoryPrefix(directory))
		}
		entries = append(entries, renderer.Directory{Path: directory, Count: count})
	}
	return entries
}

func trimDirectoryPrefix(directory string) string {
	if len(directory) == 0 || directory == "/" {
		return ""
	}
	return directory[1:]
}

func renderShow(w io.Writer, concept *knowledge.Concept, outgoing []linkView, backlinks []knowledge.Backlink, format outputFormat) error {
	backlinkViews := make([]linkView, 0, len(backlinks))
	for _, backlink := range backlinks {
		backlinkViews = append(backlinkViews, backlinkView(backlink))
	}
	view := renderer.ShowView{
		Concept:   viewConcept(concept, true),
		Links:     outgoing,
		Backlinks: backlinkViews,
		Actions:   actionViews(concept.ID),
	}
	return renderOutput(w, format, view)
}

func renderShowCollection(w io.Writer, bundle *knowledge.Bundle, concepts []*knowledge.Concept, format outputFormat) error {
	results := make([]renderer.ShowResult, 0, len(concepts))
	for _, concept := range concepts {
		backlinks, err := bundle.Backlinks(concept.ID)
		if err != nil {
			return err
		}
		backlinkViews := make([]linkView, 0, len(backlinks))
		for _, backlink := range backlinks {
			backlinkViews = append(backlinkViews, backlinkView(backlink))
		}
		results = append(results, renderer.ShowResult{
			Concept:   viewConcept(concept, true),
			Links:     linkViews(concept.Links),
			Backlinks: backlinkViews,
			Actions:   actionViews(concept.ID),
		})
	}
	return renderOutput(w, format, renderer.ShowCollectionView{Results: results})
}

func renderSearchResults(w io.Writer, results []knowledge.SearchResult, query string, format outputFormat) error {
	view := renderer.SearchView{Query: query, Results: searchResults(results)}
	return renderOutput(w, format, view)
}

func renderContextResults(w io.Writer, results []knowledge.SearchResult, query string, format outputFormat) error {
	view := renderer.ContextView{Query: query, Results: contextResults(results)}
	return renderOutput(w, format, view)
}

func searchResults(results []knowledge.SearchResult) []renderer.SearchResult {
	views := make([]renderer.SearchResult, 0, len(results))
	for _, result := range results {
		views = append(views, renderer.SearchResult{
			Concept: viewConcept(result.Concept, false),
			Score:   result.Score,
			Actions: actionViews(result.Concept.ID),
		})
	}
	return views
}

func contextResults(results []knowledge.SearchResult) []renderer.ContextResult {
	views := make([]renderer.ContextResult, 0, len(results))
	for _, result := range results {
		links := make([]linkView, 0, len(result.Concept.Links))
		for _, link := range result.Concept.Links {
			if link.External || link.TargetID != "" || link.TargetPath != "" || link.Broken {
				links = append(links, renderLink(link))
			}
		}
		views = append(views, renderer.ContextResult{
			Concept: viewConcept(result.Concept, true),
			Score:   result.Score,
			Links:   links,
			Actions: actionViews(result.Concept.ID),
		})
	}
	return views
}

func renderLinks(concept *knowledge.Concept, links []knowledge.Link, format outputFormat) error {
	view := renderer.LinksView{Source: concept.ID, Links: linkViews(links)}
	return renderOutput(os.Stdout, format, view)
}

func renderBacklinks(concept *knowledge.Concept, backlinks []knowledge.Backlink, format outputFormat) error {
	views := make([]linkView, 0, len(backlinks))
	for _, backlink := range backlinks {
		views = append(views, backlinkView(backlink))
	}
	view := renderer.BacklinksView{Target: concept.ID, Backlinks: views}
	return renderOutput(os.Stdout, format, view)
}

func linkViews(links []knowledge.Link) []linkView {
	views := make([]linkView, 0, len(links))
	for _, link := range links {
		if link.External || link.TargetID != "" || link.TargetPath != "" || link.Broken {
			views = append(views, renderLink(link))
		}
	}
	return views
}

func renderGraph(nodes []knowledge.GraphNode, format outputFormat) error {
	views := make([]renderer.GraphNode, 0, len(nodes))
	for _, node := range nodes {
		views = append(views, renderer.GraphNode{
			ID:    node.Concept.ID,
			Title: conceptTitle(node.Concept),
			Depth: node.Depth,
		})
	}
	return renderOutput(os.Stdout, format, renderer.GraphView{Nodes: views})
}

func renderCheck(report knowledge.ValidationReport, format outputFormat) error {
	view := renderer.CheckView{
		Root:     report.Stats.Root,
		Mode:     report.Stats.Mode,
		Stats:    checkStats(report),
		Bundles:  checkBundles(report.Bundles),
		Errors:   checkIssues(report.Errors),
		Warnings: checkIssues(report.Warnings),
		Valid:    report.Valid(),
	}
	return renderOutput(os.Stdout, format, view)
}

func checkStats(report knowledge.ValidationReport) renderer.CheckStats {
	return renderer.CheckStats{
		Bundles:               report.Stats.Bundles,
		MarkdownFiles:         report.Stats.MarkdownFiles,
		ReservedFiles:         report.Stats.ReservedFiles,
		ConceptFiles:          report.Stats.ConceptFiles,
		LoadedConcepts:        report.Stats.LoadedConcepts,
		InvalidConceptFiles:   report.Stats.InvalidConceptFiles,
		LinksChecked:          report.Stats.LinksChecked,
		BrokenLinks:           report.Stats.BrokenLinks,
		StaleGeneratedIndexes: report.Stats.StaleGeneratedIndexes,
		Errors:                len(report.Errors),
		Warnings:              len(report.Warnings),
	}
}

func checkBundles(bundles []knowledge.BundleValidationStats) []renderer.CheckBundle {
	result := make([]renderer.CheckBundle, 0, len(bundles))
	for _, bundle := range bundles {
		result = append(result, renderer.CheckBundle{
			Name:                  bundle.Name,
			Root:                  bundle.Root,
			MarkdownFiles:         bundle.MarkdownFiles,
			ReservedFiles:         bundle.ReservedFiles,
			ConceptFiles:          bundle.ConceptFiles,
			LoadedConcepts:        bundle.LoadedConcepts,
			InvalidConceptFiles:   bundle.InvalidConceptFiles,
			LinksChecked:          bundle.LinksChecked,
			BrokenLinks:           bundle.BrokenLinks,
			StaleGeneratedIndexes: bundle.StaleGeneratedIndexes,
		})
	}
	return result
}

func checkIssues(issues []knowledge.Issue) []renderer.Issue {
	result := make([]renderer.Issue, 0, len(issues))
	for _, issue := range issues {
		result = append(result, renderer.Issue{Path: issue.Path, Message: issue.Message})
	}
	return result
}
