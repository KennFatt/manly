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
	concepts := conceptsInDirectory(bundle, prefix, *recursive)
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
	if flags.NArg() != 1 {
		return errors.New("usage: manly show <concept-id> [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	concept, err := bundle.Get(flags.Arg(0))
	if err != nil {
		return err
	}
	outgoing := make([]linkView, 0, len(concept.Links))
	for _, link := range concept.Links {
		if link.External || link.TargetID != "" || link.TargetPath != "" || link.Broken {
			outgoing = append(outgoing, renderLink(link))
		}
	}
	backlinks, err := bundle.Backlinks(concept.ID)
	if err != nil {
		return err
	}
	return renderShow(os.Stdout, concept, outgoing, backlinks, format)
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

func conceptsInDirectory(bundle *knowledge.Bundle, prefix string, recursive bool) []*knowledge.Concept {
	var concepts []*knowledge.Concept
	for _, concept := range bundle.Concepts {
		directory := filepath.ToSlash(filepath.Dir(concept.RelPath))
		if directory == "." {
			directory = ""
		}
		if recursive {
			if prefix == "" || directory == prefix || strings.HasPrefix(directory, prefix+"/") {
				concepts = append(concepts, concept)
			}
		} else if directory == prefix {
			concepts = append(concepts, concept)
		}
	}
	sort.Slice(concepts, func(i, j int) bool { return concepts[i].ID < concepts[j].ID })
	return concepts
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
