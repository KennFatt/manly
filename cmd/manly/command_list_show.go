package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

func runList(root string, args []string) error {
	flags := newFlagSet("list")
	recursive := flags.Bool("recursive", false, "include nested concepts")
	formatValue := flags.String("format", string(formatHuman), "output format")
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
			return renderDirectoryJSON(prefix, directories, concepts, true)
		}
		return renderConceptList(bundle, concepts, format, bundleDirectoryTitle(bundle, prefix))
	}
	return renderDirectoryContents(bundle, prefix, directories, concepts, format)
}

func runShow(root string, args []string) error {
	flags := newFlagSet("show")
	formatValue := flags.String("format", string(formatHuman), "output format")
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
	links := concept.Links
	outgoing := make([]linkView, 0, len(links))
	for _, link := range links {
		if link.External || link.TargetID != "" || link.TargetPath != "" || link.Broken {
			outgoing = append(outgoing, renderLink(link))
		}
	}
	backlinks, err := bundle.Backlinks(concept.ID)
	if err != nil {
		return err
	}
	switch format {
	case formatJSON:
		backlinkViews := make([]linkView, 0, len(backlinks))
		for _, backlink := range backlinks {
			backlinkViews = append(backlinkViews, linkView{Label: backlink.Link.Label, Target: backlink.Concept.ID, TargetPath: backlink.Concept.RelPath})
		}
		return writeJSON(map[string]any{
			"concept":   viewConcept(concept, true),
			"links":     outgoing,
			"backlinks": backlinkViews,
			"actions":   actionViews(concept.ID),
		})
	case formatCompact:
		fmt.Printf("%s\n%s\n", concept.ID, strings.TrimSpace(concept.Body))
		return nil
	case formatMarkdown:
		fmt.Printf("# %s\n\n%s\n", conceptTitle(concept), strings.TrimSpace(concept.Body))
		return nil
	default:
		fmt.Printf("%s\n\n", strings.TrimSpace(concept.Body))
		if len(outgoing) > 0 {
			fmt.Println("Links:")
			for index, link := range outgoing {
				if link.Target != "" {
					fmt.Printf("[%d] %s\n    %s\n    manly show %s\n\n", index+1, link.Label, link.Target, link.Target)
				} else if command := linkNavigationCommand(link); command != "" {
					fmt.Printf("[%d] %s\n    %s\n    %s\n\n", index+1, link.Label, link.TargetPath, command)
				} else if link.URL != "" {
					fmt.Printf("[%d] %s\n    %s\n\n", index+1, link.Label, link.URL)
				} else {
					fmt.Printf("[%d] %s\n    broken: %s\n\n", index+1, link.Label, link.URL)
				}
			}
		}
		if len(backlinks) > 0 {
			fmt.Println("Backlinks:")
			for _, backlink := range backlinks {
				fmt.Printf("  %s (%s)\n", backlink.Concept.ID, backlink.Link.Label)
			}
			fmt.Println()
		}
		fmt.Println("Actions:")
		renderAction("Open", "manly show "+concept.ID)
		renderAction("Context", "manly context "+concept.ID)
		renderAction("Edit", "manly edit "+concept.ID)
		renderAction("Backlinks", "manly backlinks "+concept.ID)
		return nil
	}
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

func renderDirectoryContents(bundle *knowledge.Bundle, prefix string, directories []string, concepts []*knowledge.Concept, format outputFormat) error {
	switch format {
	case formatJSON:
		return renderDirectoryJSON(prefix, directories, concepts, false)
	case formatMarkdown:
		fmt.Printf("# %s\n\n", bundleDirectoryTitle(bundle, prefix))
		for _, directory := range directories {
			fmt.Printf("* [%s](%s/)\n", filepath.Base(directory), directory)
		}
		for _, concept := range concepts {
			fmt.Printf("* [%s](%s.md) - %s\n", conceptTitle(concept), concept.ID, conceptDescription(concept))
		}
	default:
		fmt.Printf("%s\n\n", bundleDirectoryTitle(bundle, prefix))
		for _, directory := range directories {
			fmt.Printf("  %-24s %d concepts\n", directory, countConceptsUnder(bundle, strings.TrimPrefix(directory, "/")))
		}
		for _, concept := range concepts {
			fmt.Printf("  %s\n", concept.ID)
			if description := conceptDescription(concept); description != "" {
				fmt.Printf("      %s\n", description)
			}
			fmt.Printf("      Open: manly show %s\n", concept.ID)
		}
		fmt.Printf("\n%d concept(s)\n", countConceptsUnder(bundle, prefix))
	}
	return nil
}

func renderDirectoryJSON(prefix string, directories []string, concepts []*knowledge.Concept, recursive bool) error {
	type entryView struct {
		Concept conceptView  `json:"concept"`
		Actions []actionView `json:"actions"`
	}
	entries := make([]entryView, 0, len(concepts))
	for _, concept := range concepts {
		entries = append(entries, entryView{Concept: viewConcept(concept, false), Actions: actionViews(concept.ID)})
	}
	return writeJSON(map[string]any{
		"path":        directoryDisplay(prefix),
		"recursive":   recursive,
		"directories": directories,
		"entries":     entries,
	})
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
