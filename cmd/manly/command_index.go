package main

import (
	"fmt"

	"github.com/KennFatt/manly/internal/knowledge"
)

type IndexCommand struct {
	Check bool `help:"Report stale generated sections without writing."`
}

func (command *IndexCommand) Run(app *appContext) error {
	changed, err := knowledge.UpdateIndexes(app.root, command.Check)
	if err != nil {
		return err
	}
	if command.Check {
		if len(changed) == 0 {
			fmt.Println("Generated index sections are up to date.")
			return nil
		}
		for _, path := range changed {
			fmt.Printf("stale: %s\n", path)
		}
		return fmt.Errorf("%d generated index section(s) are stale", len(changed))
	}
	if len(changed) == 0 {
		fmt.Println("No marked generated index sections found.")
		return nil
	}
	for _, path := range changed {
		fmt.Printf("Updated %s\n", path)
	}
	return nil
}
