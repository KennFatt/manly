package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

type outputFormat string

const (
	formatHuman    outputFormat = "human"
	formatCompact  outputFormat = "compact"
	formatJSON     outputFormat = "json"
	formatMarkdown outputFormat = "markdown"
)

type conceptView struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Type        string   `json:"type,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Content     string   `json:"content,omitempty"`
}

type linkView struct {
	Label      string `json:"label"`
	Target     string `json:"target,omitempty"`
	TargetPath string `json:"target_path,omitempty"`
	URL        string `json:"url,omitempty"`
	Broken     bool   `json:"broken,omitempty"`
	External   bool   `json:"external,omitempty"`
}

type actionView struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func viewConcept(concept *knowledge.Concept, includeContent bool) conceptView {
	view := conceptView{
		ID:          concept.ID,
		Path:        concept.RelPath,
		Type:        concept.Type,
		Title:       conceptTitle(concept),
		Description: conceptDescription(concept),
		Tags:        append([]string(nil), concept.Tags...),
	}
	if includeContent {
		view.Content = strings.TrimSpace(concept.Body)
	}
	return view
}

func conceptTitle(concept *knowledge.Concept) string {
	if concept.Title != "" {
		return concept.Title
	}
	return concept.ID
}

func conceptDescription(concept *knowledge.Concept) string {
	if concept.Description != "" {
		return concept.Description
	}
	for _, line := range strings.Split(concept.Body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

func writeJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func renderConceptList(bundle *knowledge.Bundle, concepts []*knowledge.Concept, format outputFormat, heading string) error {
	switch format {
	case formatJSON:
		entries := make([]conceptView, 0, len(concepts))
		for _, concept := range concepts {
			entries = append(entries, viewConcept(concept, false))
		}
		return writeJSON(map[string]any{"path": heading, "entries": entries})
	case formatMarkdown:
		if heading != "" {
			fmt.Printf("# %s\n\n", heading)
		}
		for _, concept := range concepts {
			fmt.Printf("* [%s](%s) - %s\n", conceptTitle(concept), concept.ID+".md", conceptDescription(concept))
		}
		return nil
	case formatCompact:
		for _, concept := range concepts {
			fmt.Printf("%s\t%s\n", concept.ID, conceptTitle(concept))
		}
		return nil
	default:
		if heading != "" {
			fmt.Printf("%s\n\n", heading)
		}
		for _, concept := range concepts {
			fmt.Printf("  %s\n", concept.ID)
			if description := conceptDescription(concept); description != "" {
				fmt.Printf("      %s\n", description)
			}
			fmt.Printf("      Open: manly show %s\n\n", concept.ID)
		}
		return nil
	}
}

func renderAction(name, command string) {
	fmt.Printf("  %-9s %s\n", name+":", command)
}

func renderLink(link knowledge.Link) linkView {
	return linkView{
		Label:      link.Label,
		Target:     link.TargetID,
		TargetPath: link.TargetPath,
		URL:        link.RawTarget,
		Broken:     link.Broken,
		External:   link.External,
	}
}

func linkNavigationCommand(link linkView) string {
	if link.Target != "" {
		return "manly show " + link.Target
	}
	if strings.HasSuffix(link.TargetPath, "/index.md") {
		return "manly list /" + strings.TrimSuffix(strings.TrimPrefix(link.TargetPath, "/"), "/index.md")
	}
	return ""
}

func actionViews(id string) []actionView {
	return []actionView{
		{Name: "show", Command: "manly show " + id},
		{Name: "context", Command: "manly context " + id},
		{Name: "edit", Command: "manly edit " + id},
		{Name: "backlinks", Command: "manly backlinks " + id},
	}
}

func relativeDisplayPath(path string) string {
	return filepath.ToSlash(path)
}
