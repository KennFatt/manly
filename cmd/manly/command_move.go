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
	changed, err := knowledge.Move(app.root, command.OldConceptID, command.NewConceptID)
	if err != nil {
		return err
	}
	fmt.Printf("Moved %s to %s; updated %d link(s)\n", command.OldConceptID, command.NewConceptID, changed)
	return nil
}
