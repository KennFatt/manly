package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

type EditCommand struct {
	ConceptID string `arg:"" help:"Concept ID to edit."`
}

func (command *EditCommand) Run(app *appContext) error {
	bundle, err := loadBundle(app.root)
	if err != nil {
		return err
	}
	concept, err := bundle.Get(command.ConceptID)
	if err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("EDITOR is not set")
	}
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return errors.New("EDITOR is empty")
	}
	commandLine := exec.Command(parts[0], append(parts[1:], concept.AbsPath)...)
	commandLine.Stdin = os.Stdin
	commandLine.Stdout = os.Stdout
	commandLine.Stderr = os.Stderr
	return commandLine.Run()
}
