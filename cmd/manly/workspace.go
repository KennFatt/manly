package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KennFatt/manly/internal/config"
	"github.com/KennFatt/manly/internal/knowledge"
	"github.com/KennFatt/manly/internal/renderer"
)

func loadWorkspace(root string) (*knowledge.Workspace, error) {
	return knowledge.LoadWorkspace(root)
}

func displayConcept(workspace *knowledge.Workspace, ref knowledge.ConceptRef) *knowledge.Concept {
	concept := *ref.Concept
	concept.ID = workspace.QualifyID(ref.BundleName, concept.ID)
	if ref.BundleName != "" {
		concept.RelPath = filepath.ToSlash(filepath.Join(ref.BundleName, ref.Concept.RelPath))
	}
	concept.Links = displayLinks(workspace, ref.BundleName, ref.Concept.Links)
	return &concept
}

func displayLinks(workspace *knowledge.Workspace, bundleName string, links []knowledge.Link) []knowledge.Link {
	result := make([]knowledge.Link, 0, len(links))
	for _, link := range links {
		copy := link
		if copy.TargetID != "" {
			copy.TargetID = workspace.QualifyID(bundleName, copy.TargetID)
		}
		if bundleName != "" && copy.TargetPath != "" {
			copy.TargetPath = filepath.ToSlash(filepath.Join(bundleName, copy.TargetPath))
		}
		result = append(result, copy)
	}
	return result
}

func displayRefs(workspace *knowledge.Workspace, refs []knowledge.ConceptRef) []*knowledge.Concept {
	concepts := make([]*knowledge.Concept, 0, len(refs))
	for _, ref := range refs {
		concepts = append(concepts, displayConcept(workspace, ref))
	}
	return concepts
}

func refsForBundle(workspace *knowledge.Workspace, bundleName string, bundle *knowledge.Bundle, concepts []*knowledge.Concept) []knowledge.ConceptRef {
	refs := make([]knowledge.ConceptRef, 0, len(concepts))
	for _, concept := range concepts {
		refs = append(refs, knowledge.ConceptRef{BundleName: bundleName, Bundle: bundle, Concept: concept})
	}
	return refs
}

func allWorkspaceRefs(workspace *knowledge.Workspace) []knowledge.ConceptRef {
	var refs []knowledge.ConceptRef
	for _, bundle := range workspace.Bundles {
		refs = append(refs, refsForBundle(workspace, workspaceName(workspace, bundle), bundle, bundle.Concepts)...)
	}
	return refs
}

func workspaceName(workspace *knowledge.Workspace, bundle *knowledge.Bundle) string {
	if workspace.SingleRoot {
		return ""
	}
	for name, candidate := range workspace.ByName {
		if candidate == bundle {
			return name
		}
	}
	return ""
}

func resolveWorkspaceConcepts(workspace *knowledge.Workspace, arguments []string) ([]knowledge.ConceptRef, bool, error) {
	collection := len(arguments) > 1
	selected := make([]knowledge.ConceptRef, 0, len(arguments))
	seen := make(map[string]bool)
	for _, argument := range arguments {
		if ref, err := workspace.ResolveConcept(argument); err == nil {
			id := workspace.QualifyID(ref.BundleName, ref.Concept.ID)
			if !seen[id] {
				selected = append(selected, *ref)
				seen[id] = true
			}
			continue
		}
		bundle, prefix, name, err := workspace.ResolveDirectory(argument)
		if err != nil {
			return nil, false, err
		}
		if bundle == nil {
			return nil, false, fmt.Errorf("concept or directory not found: %s", argument)
		}
		concepts := bundle.ConceptsUnder(prefix, true)
		if len(concepts) == 0 {
			return nil, false, fmt.Errorf("concept or directory not found: %s", argument)
		}
		collection = true
		for _, concept := range concepts {
			id := workspace.QualifyID(name, concept.ID)
			if seen[id] {
				continue
			}
			selected = append(selected, knowledge.ConceptRef{BundleName: name, Bundle: bundle, Concept: concept})
			seen[id] = true
		}
	}
	return selected, collection, nil
}

func workspaceDirectoryEntries(workspace *knowledge.Workspace, name string, bundle *knowledge.Bundle, directories []string) []renderer.Directory {
	entries := make([]renderer.Directory, 0, len(directories))
	for _, directory := range directories {
		path := directory
		if name != "" {
			path = "/" + name + "/" + strings.TrimPrefix(directory, "/")
		}
		entries = append(entries, renderer.Directory{Path: path, Count: countConceptsUnder(bundle, trimDirectoryPrefix(directory))})
	}
	return entries
}

func renderWorkspaceRecursiveList(workspace *knowledge.Workspace, format outputFormat, display config.Display) error {
	refs := allWorkspaceRefs(workspace)
	return renderOutput(os.Stdout, format, renderer.ListView{
		Root:        workspace.Root,
		Path:        "/",
		Heading:     "Knowledge Workspace",
		Recursive:   true,
		Entries:     conceptEntries(displayRefs(workspace, refs), display.Actions),
		HideActions: !display.Actions,
		HideUsage:   !display.Usage,
	})
}

func renderWorkspaceRootList(workspace *knowledge.Workspace, format outputFormat, display config.Display) error {
	directories := make([]renderer.Directory, 0, len(workspace.Bundles))
	for _, bundle := range workspace.Bundles {
		name := workspaceName(workspace, bundle)
		directories = append(directories, renderer.Directory{Path: "/" + name, Count: len(bundle.Concepts)})
	}
	sort.Slice(directories, func(i, j int) bool { return directories[i].Path < directories[j].Path })
	return renderOutput(os.Stdout, format, renderer.ListView{
		Root:        workspace.Root,
		Path:        "/",
		Heading:     "Knowledge Workspace",
		Directories: directories,
		HideActions: !display.Actions,
		HideUsage:   !display.Usage,
	})
}

func renderWorkspaceDirectory(workspace *knowledge.Workspace, bundle *knowledge.Bundle, name, prefix string, recursive bool, format outputFormat, display config.Display) error {
	concepts := bundle.ConceptsUnder(prefix, recursive)
	refs := refsForBundle(workspace, name, bundle, concepts)
	directories := childDirectories(bundle, prefix)
	displayPath := "/" + name
	if prefix != "" {
		displayPath += "/" + prefix
	}
	view := renderer.ListView{
		Root:        workspace.Root,
		Path:        displayPath,
		Heading:     bundleDirectoryTitle(bundle, prefix),
		Recursive:   recursive,
		Entries:     conceptEntries(displayRefs(workspace, refs), display.Actions),
		HideActions: !display.Actions,
		HideUsage:   !display.Usage,
	}
	if recursive {
		view.Directories = workspaceDirectoryEntries(workspace, name, bundle, directories)
	} else {
		view.Directories = workspaceDirectoryEntries(workspace, name, bundle, directories)
		view.Count = countConceptsUnder(bundle, prefix)
	}
	return renderOutput(os.Stdout, format, view)
}

func renderWorkspaceShow(workspace *knowledge.Workspace, ref knowledge.ConceptRef, format outputFormat, display config.Display) error {
	backlinks, err := ref.Bundle.Backlinks(ref.Concept.ID)
	if err != nil {
		return err
	}
	displayed := displayConcept(workspace, ref)
	outgoing := displayLinks(workspace, ref.BundleName, ref.Concept.Links)
	displayedBacklinks := make([]knowledge.Backlink, 0, len(backlinks))
	for _, backlink := range backlinks {
		copy := backlink
		concept := knowledge.ConceptRef{BundleName: ref.BundleName, Bundle: ref.Bundle, Concept: backlink.Concept}
		copy.Concept = displayConcept(workspace, concept)
		copy.Link = displayLinks(workspace, ref.BundleName, []knowledge.Link{backlink.Link})[0]
		displayedBacklinks = append(displayedBacklinks, copy)
	}
	return renderShow(os.Stdout, displayed, linkViews(outgoing), displayedBacklinks, format, display)
}

func renderWorkspaceShowCollection(workspace *knowledge.Workspace, refs []knowledge.ConceptRef, format outputFormat, display config.Display) error {
	concepts := displayRefs(workspace, refs)
	// Relationships are bundle-local, so calculate them before replacing IDs for display.
	results := make([]renderer.ShowResult, 0, len(refs))
	for index, ref := range refs {
		backlinks, err := ref.Bundle.Backlinks(ref.Concept.ID)
		if err != nil {
			return err
		}
		displayedBacklinks := make([]knowledge.Backlink, 0, len(backlinks))
		for _, backlink := range backlinks {
			copy := backlink
			copy.Concept = displayConcept(workspace, knowledge.ConceptRef{BundleName: ref.BundleName, Bundle: ref.Bundle, Concept: backlink.Concept})
			copy.Link = displayLinks(workspace, ref.BundleName, []knowledge.Link{backlink.Link})[0]
			displayedBacklinks = append(displayedBacklinks, copy)
		}
		results = append(results, renderer.ShowResult{
			Concept:   viewConcept(concepts[index], true),
			Links:     linkViews(displayLinks(workspace, ref.BundleName, ref.Concept.Links)),
			Backlinks: linkViewsFromBacklinks(displayedBacklinks),
			Actions:   actionViews(concepts[index].ID, display.Actions),
			HideUsage: !display.Usage,
		})
	}
	return renderOutput(os.Stdout, format, renderer.ShowCollectionView{Results: results})
}

func linkViewsFromBacklinks(backlinks []knowledge.Backlink) []linkView {
	views := make([]linkView, 0, len(backlinks))
	for _, backlink := range backlinks {
		views = append(views, backlinkView(backlink))
	}
	return views
}

func renderWorkspaceSearch(workspace *knowledge.Workspace, results []knowledge.SearchResultRef, query string, format outputFormat) error {
	local := make([]knowledge.SearchResult, 0, len(results))
	for _, result := range results {
		concept := displayConcept(workspace, knowledge.ConceptRef{BundleName: result.BundleName, Bundle: workspace.ByName[result.BundleName], Concept: result.Concept})
		local = append(local, knowledge.SearchResult{Concept: concept, Score: result.Score})
	}
	return renderSearchResults(os.Stdout, local, query, format)
}
