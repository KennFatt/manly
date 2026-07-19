package main

import (
	"fmt"

	"github.com/KennFatt/manly/internal/knowledge"
)

type MoveCommand struct {
	OldConceptID string `arg:"" name:"old-concept-id" help:"Current concept ID."`
	NewConceptID string `arg:"" name:"new-concept-id" help:"New concept ID."`
}

func (command *MoveCommand) Run(app *appContext) error {
	workspace, err := loadWorkspace(app.root)
	if err != nil {
		return err
	}
	if workspace.SingleRoot {
		changed, err := knowledge.Move(app.root, command.OldConceptID, command.NewConceptID)
		if err != nil {
			return err
		}
		fmt.Printf("Moved %s to %s; updated %d link(s)\n", command.OldConceptID, command.NewConceptID, changed)
		return nil
	}
	oldRef, err := workspace.ResolveConcept(command.OldConceptID)
	if err != nil {
		return err
	}
	newBundle, newPrefix, newName, err := workspace.ResolveDirectory(command.NewConceptID)
	if err != nil {
		return err
	}
	if newBundle == nil || newPrefix == "" {
		return fmt.Errorf("workspace concept path must include a bundle and concept: %s", command.NewConceptID)
	}
	if oldRef.Bundle != newBundle || oldRef.BundleName != newName {
		return fmt.Errorf("cross-bundle moves are not supported")
	}
	changed, err := knowledge.Move(newBundle.Root, oldRef.Concept.ID, "/"+newPrefix)
	if err != nil {
		return err
	}
	fmt.Printf("Moved %s to %s; updated %d link(s)\n", command.OldConceptID, command.NewConceptID, changed)
	return nil
}
