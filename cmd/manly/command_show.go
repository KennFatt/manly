package main

import (
	"fmt"
	"os"

	"github.com/KennFatt/manly/internal/knowledge"
)

type ShowCommand struct {
	Concepts []string `arg:"" help:"Concept IDs or directories to show."`
	Format   string   `default:"compact" help:"Output format."`
}

func (command *ShowCommand) Run(app *appContext) error {
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(app.root)
	if err != nil {
		return err
	}
	concepts, collection, err := resolveShowConcepts(bundle, command.Concepts)
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
