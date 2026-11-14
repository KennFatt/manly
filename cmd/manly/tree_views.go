package main

import (
	"github.com/KennFatt/manly/internal/config"
	"github.com/KennFatt/manly/internal/knowledge"
	"github.com/KennFatt/manly/internal/renderer"
)

func buildBundleTree(workspace *knowledge.Workspace, bundle *knowledge.Bundle, name, prefix string, depth int, display config.Display) renderer.TreeView {
	heading := treeHeading(bundle, prefix)
	if prefix == "" && bundle.Title == "" && name != "" {
		heading = directoryTitle(name)
	}
	return renderer.TreeView{
		Root:         workspace.Root,
		Path:         treeDisplayPath(name, prefix),
		Heading:      heading,
		Description:  bundleDescription(bundle, prefix),
		Depth:        depth,
		ConceptCount: countConceptsUnder(bundle, prefix),
		Nodes:        buildTreeChildren(workspace, bundle, name, prefix, 1, depth),
		HideUsage:    !display.Usage,
	}
}

func buildWorkspaceTree(workspace *knowledge.Workspace, depth int, display config.Display) renderer.TreeView {
	nodes := make([]renderer.TreeNode, 0, len(workspace.Bundles))
	total := 0
	for _, bundle := range workspace.Bundles {
		name := workspaceName(workspace, bundle)
		total += len(bundle.Concepts)
		nodes = append(nodes, buildBundleNode(workspace, bundle, name, depth))
	}
	return renderer.TreeView{
		Root:         workspace.Root,
		Path:         "/",
		Heading:      "Knowledge Workspace",
		Description:  workspace.Description,
		Depth:        depth,
		ConceptCount: total,
		Nodes:        nodes,
		HideUsage:    !display.Usage,
	}
}

func buildBundleNode(workspace *knowledge.Workspace, bundle *knowledge.Bundle, name string, depth int) renderer.TreeNode {
	title := bundle.Title
	if title == "" {
		title = directoryTitle(name)
	}
	node := renderer.TreeNode{
		Kind:         "directory",
		Path:         "/" + name,
		Title:        title,
		Description:  bundle.Description,
		ConceptCount: len(bundle.Concepts),
	}
	if depth == 1 {
		node.Hidden = treeHidden(bundle, "")
		return node
	}
	node.Children = buildTreeChildren(workspace, bundle, name, "", 2, depth)
	return node
}

func buildTreeChildren(workspace *knowledge.Workspace, bundle *knowledge.Bundle, name, prefix string, level, maxDepth int) []renderer.TreeNode {
	nodes := make([]renderer.TreeNode, 0)
	for _, directory := range childDirectories(bundle, prefix) {
		childPrefix := trimDirectoryPrefix(directory)
		nodes = append(nodes, buildTreeDirectoryNode(workspace, bundle, name, childPrefix, level, maxDepth))
	}
	for _, concept := range bundle.ConceptsUnder(prefix, false) {
		nodes = append(nodes, renderer.TreeNode{
			Kind:        "concept",
			Path:        workspace.QualifyID(name, concept.ID),
			Title:       conceptTitle(concept),
			Description: conceptDescription(concept),
			Type:        concept.Type,
			Tags:        append([]string(nil), concept.Tags...),
		})
	}
	return nodes
}

func buildTreeDirectoryNode(workspace *knowledge.Workspace, bundle *knowledge.Bundle, name, prefix string, level, maxDepth int) renderer.TreeNode {
	metadata := bundle.MetadataForDirectory(prefix)
	title := metadata.Title
	if title == "" {
		title = directoryTitle(prefix)
	}
	node := renderer.TreeNode{
		Kind:         "directory",
		Path:         treeDisplayPath(name, prefix),
		Title:        title,
		Description:  metadata.Description,
		ConceptCount: countConceptsUnder(bundle, prefix),
	}
	if level == maxDepth {
		node.Hidden = treeHidden(bundle, prefix)
		return node
	}
	node.Children = buildTreeChildren(workspace, bundle, name, prefix, level+1, maxDepth)
	if len(node.Children) == 0 {
		node.Children = nil
	}
	return node
}

func treeHidden(bundle *knowledge.Bundle, prefix string) *renderer.TreeCut {
	childDirs := childDirectories(bundle, prefix)
	count := countConceptsUnder(bundle, prefix)
	if len(childDirs) == 0 && count == 0 {
		return nil
	}
	return &renderer.TreeCut{Directories: len(childDirs), Concepts: count}
}

func treeHeading(bundle *knowledge.Bundle, prefix string) string {
	if prefix == "" {
		if bundle.Title != "" {
			return bundle.Title
		}
		return "Knowledge Bundle"
	}
	if title := bundle.MetadataForDirectory(prefix).Title; title != "" {
		return title
	}
	return directoryTitle(prefix)
}

func treeDisplayPath(name, prefix string) string {
	if name == "" {
		return directoryDisplay(prefix)
	}
	if prefix == "" {
		return "/" + name
	}
	return "/" + name + "/" + prefix
}
