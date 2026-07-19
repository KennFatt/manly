package main

import (
	"io"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
	"github.com/KennFatt/manly/internal/renderer"
)

type outputFormat = renderer.Format

const (
	formatCompact  = renderer.FormatCompact
	formatFancy    = renderer.FormatFancy
	formatJSON     = renderer.FormatJSON
	formatMarkdown = renderer.FormatMarkdown
)

type conceptView = renderer.Concept
type linkView = renderer.Link

type actionView = renderer.Action

func parseFormat(value string) (outputFormat, error) {
	return renderer.ParseFormat(value)
}

func renderOutput(w io.Writer, format outputFormat, view renderer.View) error {
	outputRenderer, err := renderer.New(format)
	if err != nil {
		return err
	}
	return outputRenderer.Render(w, view)
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

func actionViews(id string, enabled ...bool) []actionView {
	if len(enabled) > 0 && !enabled[0] {
		return nil
	}
	return []actionView{
		{Name: "show", Command: "manly show " + id},
		{Name: "context", Command: "manly context " + id},
		{Name: "edit", Command: "manly edit " + id},
		{Name: "backlinks", Command: "manly backlinks " + id},
	}
}

func backlinkView(backlink knowledge.Backlink) linkView {
	return linkView{
		Label:      backlink.Link.Label,
		Title:      conceptTitle(backlink.Concept),
		Target:     backlink.Concept.ID,
		TargetPath: backlink.Concept.RelPath,
	}
}
