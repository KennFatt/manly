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
	workspace, err := loadWorkspace(app.root)
	if err != nil {
		return err
	}
	input := knowledge.NewConcept{
		Type:        command.Type,
		Title:       command.Title,
		Description: command.Description,
		Tags:        splitTags(command.Tag),
	}
	if workspace.SingleRoot {
		id, err := knowledge.Add(app.root, command.ConceptID, input, command.Force)
		if err != nil {
			return err
		}
		fmt.Printf("Created %s\n", id)
		return nil
	}
	bundle, prefix, name, err := workspace.ResolveDirectory(command.ConceptID)
	if err != nil {
		return err
	}
	if bundle == nil || prefix == "" {
		return fmt.Errorf("workspace concept path must include a bundle and concept: %s", command.ConceptID)
	}
	id, err := knowledge.Add(bundle.Root, "/"+prefix, input, command.Force)
	if err != nil {
		return err
	}
	fmt.Printf("Created %s\n", workspace.QualifyID(name, id))
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
