package main

import (
	"fmt"

	"github.com/KennFatt/manly/internal/knowledge"
)

type CheckCommand struct {
	Strict bool   `help:"Enable advisory checks."`
	Format string `default:"${format}" help:"Output format."`
}

func (command *CheckCommand) Run(app *appContext) error {
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	report, err := knowledge.ValidateWorkspaceRoot(app.root, command.Strict)
	if err != nil {
		return err
	}
	if err := renderCheck(report, format); err != nil {
		return err
	}
	if !report.Valid() {
		return fmt.Errorf("validation failed: %d error(s), %d warning(s)", len(report.Errors), len(report.Warnings))
	}
	return nil
}
