package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

func runList(root string, args []string) error {
	flags := newFlagSet("list")
	recursive := flags.Bool("recursive", false, "include nested concepts")
	formatValue := flags.String("format", string(formatCompact), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--recursive": false, "--format": true})); err != nil {
		return err
	}
	if flags.NArg() > 1 {
		return errors.New("usage: manly list [path] [--recursive] [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	path := "/"
	if flags.NArg() == 1 {
		path = flags.Arg(0)
	}
	prefix, err := directoryPrefix(path)
	if err != nil {
		return err
	}
	concepts := bundle.ConceptsUnder(prefix, *recursive)
	directories := childDirectories(bundle, prefix)
	if *recursive {
		if format == formatJSON {
			return renderJSONRecursiveDirectory(os.Stdout, root, prefix, directories, concepts)
		}
		return renderConceptList(root, bundle, concepts, format, bundleDirectoryTitle(bundle, prefix))
	}
	return renderDirectoryContents(root, bundle, prefix, directories, concepts, format)
}

func runShow(root string, args []string) error {
	flags := newFlagSet("show")
	formatValue := flags.String("format", string(formatCompact), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--format": true})); err != nil {
		return err
	}
	if flags.NArg() < 1 {
		return errors.New("usage: manly show <concept-id-or-directory>... [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	concepts, collection, err := resolveShowConcepts(bundle, flags.Args())
	if err != nil {
		return err
	}
	if collection {
		return renderShowCollection(os.Stdout, bundle, concepts, format)
	}

	concept := concepts[0]
	backlinks, err := bundle.Backlinks(concept.ID)
	if err != nil {
		return err
	}
	return renderShow(os.Stdout, concept, linkViews(concept.Links), backlinks, format)
}

func resolveShowConcepts(bundle *knowledge.Bundle, arguments []string) ([]*knowledge.Concept, bool, error) {
	collection := len(arguments) > 1
	selected := make([]*knowledge.Concept, 0, len(arguments))
	seen := make(map[string]bool)
	for _, argument := range arguments {
		if concept, err := bundle.Get(argument); err == nil {
			if !seen[concept.ID] {
				selected = append(selected, concept)
				seen[concept.ID] = true
			}
			continue
		}

		prefix, err := directoryPrefix(argument)
		if err != nil {
			return nil, false, err
		}
		concepts := bundle.ConceptsUnder(prefix, true)
		if len(concepts) == 0 {
			return nil, false, fmt.Errorf("concept or directory not found: %s", argument)
		}
		collection = true
		for _, concept := range concepts {
			if seen[concept.ID] {
				continue
			}
			selected = append(selected, concept)
			seen[concept.ID] = true
		}
	}
	return selected, collection, nil
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

func directoryTitle(prefix string) string {
	if prefix == "" {
		return "Knowledge Bundle"
	}
	name := filepath.Base(prefix)
	name = strings.ReplaceAll(name, "-", " ")
	return strings.ToUpper(name[:1]) + name[1:]
}
