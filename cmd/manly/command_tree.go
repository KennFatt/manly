package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

type TreeCommand struct {
	Path   string `arg:"" optional:"" default:"/" help:"Directory path to show as a tree."`
	Depth  int    `default:"2" help:"Maximum tree depth below the requested path."`
	Format string `default:"${format}" help:"Output format."`
}

func (command *TreeCommand) Run(app *appContext) error {
	if command.Depth < 1 {
		return fmt.Errorf("--depth must be at least 1")
	}
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	workspace, err := loadWorkspace(app.root)
	if err != nil {
		return err
	}
	if conceptID, ok := treeConceptPath(workspace, command.Path); ok {
		return fmt.Errorf("tree expects a directory path: %s (use manly show %s)", conceptID, conceptID)
	}
	if !workspace.SingleRoot {
		bundle, prefix, name, err := workspace.ResolveDirectory(command.Path)
		if err != nil {
			return err
		}
		if bundle == nil {
			if strings.Trim(command.Path, " /") != "" {
				return fmt.Errorf("directory not found: %s", command.Path)
			}
			return renderOutput(os.Stdout, format, buildWorkspaceTree(workspace, command.Depth, app.display))
		}
		if !treeDirectoryExists(bundle, prefix) {
			return fmt.Errorf("directory not found: %s", command.Path)
		}
		return renderOutput(os.Stdout, format, buildBundleTree(workspace, bundle, name, prefix, command.Depth, app.display))
	}
	bundle := workspace.Bundles[0]
	prefix, err := directoryPrefix(command.Path)
	if err != nil {
		return err
	}
	if !treeDirectoryExists(bundle, prefix) {
		return fmt.Errorf("directory not found: %s", command.Path)
	}
	return renderOutput(os.Stdout, format, buildBundleTree(workspace, bundle, "", prefix, command.Depth, app.display))
}

func treeConceptPath(workspace *knowledge.Workspace, path string) (string, bool) {
	if strings.Trim(path, " /") == "" {
		return "", false
	}
	ref, err := workspace.ResolveConcept(path)
	if err != nil || ref == nil || ref.Concept == nil {
		return "", false
	}
	return workspace.QualifyID(ref.BundleName, ref.Concept.ID), true
}

func treeDirectoryExists(bundle *knowledge.Bundle, prefix string) bool {
	if prefix == "" {
		return true
	}
	if countConceptsUnder(bundle, prefix) > 0 {
		return true
	}
	_, ok := bundle.Directories[prefix]
	return ok
}
