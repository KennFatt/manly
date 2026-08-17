package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

type ListCommand struct {
	Path      string `arg:"" optional:"" default:"/" help:"Directory path to list."`
	Recursive bool   `default:"${recursive}" negatable help:"Include nested concepts."`
	Format    string `default:"${format}" help:"Output format."`
}

func (command *ListCommand) Run(app *appContext) error {
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	workspace, err := loadWorkspace(app.root)
	if err != nil {
		return err
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
			if command.Recursive {
				return renderWorkspaceRecursiveList(workspace, format, app.display)
			}
			return renderWorkspaceRootList(workspace, format, app.display)
		}
		return renderWorkspaceDirectory(workspace, bundle, name, prefix, command.Recursive, format, app.display)
	}
	bundle := workspace.Bundles[0]
	prefix, err := directoryPrefix(command.Path)
	if err != nil {
		return err
	}
	concepts := bundle.ConceptsUnder(prefix, command.Recursive)
	directories := childDirectories(bundle, prefix)
	if command.Recursive {
		if format == formatJSON {
			return renderJSONRecursiveDirectory(os.Stdout, app.root, prefix, directories, concepts, app.display)
		}
		if format == formatAgent {
			return renderAgentConceptList(app.root, prefix, concepts)
		}
		return renderConceptList(app.root, bundle, prefix, concepts, format, bundleDirectoryTitle(bundle, prefix), app.display)
	}
	return renderDirectoryContents(app.root, bundle, prefix, directories, concepts, format, app.display)
}

func directoryPrefix(value string) (string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Trim(value, "/")
	if value == "" {
		return "", nil
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("invalid directory path: %q", value)
	}
	return clean, nil
}

func childDirectories(bundle *knowledge.Bundle, prefix string) []string {
	seen := make(map[string]bool)
	for _, concept := range bundle.Concepts {
		directory := filepath.ToSlash(filepath.Dir(concept.RelPath))
		if directory == "." {
			directory = ""
		}
		if prefix != "" {
			if directory != prefix && !strings.HasPrefix(directory, prefix+"/") {
				continue
			}
			directory = strings.TrimPrefix(strings.TrimPrefix(directory, prefix), "/")
		} else {
			directory = strings.TrimPrefix(directory, "/")
		}
		if directory == "" {
			continue
		}
		if index := strings.Index(directory, "/"); index >= 0 {
			directory = directory[:index]
		}
		if prefix != "" {
			directory = prefix + "/" + directory
		}
		seen[directory] = true
	}
	result := make([]string, 0, len(seen))
	for directory := range seen {
		result = append(result, "/"+directory)
	}
	sort.Strings(result)
	return result
}

func countConceptsUnder(bundle *knowledge.Bundle, prefix string) int {
	count := 0
	for _, concept := range bundle.Concepts {
		directory := filepath.ToSlash(filepath.Dir(concept.RelPath))
		if directory == "." {
			directory = ""
		}
		if prefix == "" || directory == prefix || strings.HasPrefix(directory, prefix+"/") {
			count++
		}
	}
	return count
}

func directoryDisplay(prefix string) string {
	if prefix == "" {
		return "/"
	}
	return "/" + prefix
}

func bundleDirectoryTitle(bundle *knowledge.Bundle, prefix string) string {
	if prefix == "" && bundle.Title != "" {
		return bundle.Title
	}
	return directoryTitle(prefix)
}

func bundleDescription(bundle *knowledge.Bundle, prefix string) string {
	if prefix == "" {
		return bundle.Description
	}
	return bundle.MetadataForDirectory(prefix).Description
}

func directoryTitle(prefix string) string {
	if prefix == "" {
		return "Knowledge Bundle"
	}
	name := filepath.Base(prefix)
	name = strings.ReplaceAll(name, "-", " ")
	return strings.ToUpper(name[:1]) + name[1:]
}
