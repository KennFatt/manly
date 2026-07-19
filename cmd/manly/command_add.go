package main

import (
	"fmt"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

type AddCommand struct {
	ConceptID   string `arg:"" help:"Concept ID to create."`
	Type        string `required:"" help:"Concept type."`
	Title       string `help:"Concept title."`
	Description string `help:"One-sentence description."`
	Tag         string `help:"Comma-separated tags."`
	Force       bool   `help:"Overwrite an existing concept."`
}

func (command *AddCommand) Run(app *appContext) error {
	id, err := knowledge.Add(app.root, command.ConceptID, knowledge.NewConcept{
		Type:        command.Type,
		Title:       command.Title,
		Description: command.Description,
		Tags:        splitTags(command.Tag),
	}, command.Force)
	if err != nil {
		return err
	}
	fmt.Printf("Created %s\n", id)
	return nil
}

func splitTags(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
