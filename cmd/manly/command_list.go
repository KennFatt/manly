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
	Level     *int   `help:"Maximum recursive listing level."`
	Format    string `default:"${format}" help:"Output format."`
}

func (command *ListCommand) Run(app *appContext) error {
	if err := command.validate(); err != nil {
		return err
	}
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
				return renderWorkspaceRecursiveList(workspace, command.Level, format, app.display)
			}
			return renderWorkspaceRootList(workspace, format, app.display)
		}
		return renderWorkspaceDirectory(workspace, bundle, name, prefix, command.Recursive, command.Level, format, app.display)
	}
	bundle := workspace.Bundles[0]
	prefix, err := directoryPrefix(command.Path)
	if err != nil {
		return err
	}
	concepts := listConcepts(bundle, prefix, command.Recursive, command.Level)
	directories := listDirectories(bundle, prefix, command.Recursive, command.Level)
	if command.Recursive {
		visibleDirectories := []string(nil)
		if command.Level != nil {
			visibleDirectories = directories
		}
		if format == formatJSON {
			return renderJSONRecursiveDirectory(os.Stdout, app.root, prefix, directories, concepts, app.display)
		}
		if format == formatAgent {
			return renderAgentConceptList(app.root, prefix, concepts, visibleDirectories)
		}
		return renderConceptList(app.root, bundle, prefix, concepts, visibleDirectories, format, bundleDirectoryTitle(bundle, prefix), app.display)
	}
	return renderDirectoryContents(app.root, bundle, prefix, directories, concepts, format, app.display)
}

func (command *ListCommand) validate() error {
	if command.Level == nil {
		return nil
	}
	if *command.Level < 1 {
		return fmt.Errorf("--level must be at least 1")
	}
	if !command.Recursive {
		return fmt.Errorf("--level requires --recursive")
	}
	return nil
}

func listConcepts(bundle *knowledge.Bundle, prefix string, recursive bool, level *int) []*knowledge.Concept {
	if !recursive {
		return bundle.ConceptsUnder(prefix, false)
	}
	if level == nil {
		return bundle.ConceptsUnder(prefix, true)
	}
	return bundle.ConceptsUnderLevel(prefix, *level)
}

func listDirectories(bundle *knowledge.Bundle, prefix string, recursive bool, level *int) []string {
	if !recursive || level == nil {
		return childDirectories(bundle, prefix)
	}
	return directoriesAtLevel(bundle, prefix, *level)
}

func directoriesAtLevel(bundle *knowledge.Bundle, prefix string, level int) []string {
	seen := make(map[string]bool)
	if level < 1 {
		return []string{}
	}
	for _, concept := range bundle.Concepts {
		directory := filepath.ToSlash(filepath.Dir(concept.RelPath))
		if directory == "." {
			directory = ""
		}
		if prefix != "" && directory != prefix && !strings.HasPrefix(directory, prefix+"/") {
			continue
		}
		relative := strings.Trim(strings.TrimPrefix(directory, prefix), "/")
		if relative == "" {
			continue
		}
		segments := strings.Split(relative, "/")
		if len(segments) != level {
			continue
		}
		path := strings.Join(segments, "/")
		if prefix != "" {
			path = prefix + "/" + path
		}
		seen["/"+path] = true
	}
	result := make([]string, 0, len(seen))
	for directory := range seen {
		result = append(result, directory)
	}
	sort.Strings(result)
	return result
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
