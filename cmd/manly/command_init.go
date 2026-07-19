package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type InitCommand struct {
	Force bool `help:"Allow creation of an existing root index."`
}

func (command *InitCommand) Run(app *appContext) error {
	if err := os.MkdirAll(app.root, 0o755); err != nil {
		return fmt.Errorf("create root: %w", err)
	}
	indexPath := filepath.Join(app.root, "index.md")
	if _, err := os.Stat(indexPath); err == nil && !command.Force {
		return fmt.Errorf("root index already exists: %s", indexPath)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check root index: %w", err)
	}
	content := "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Knowledge Bundle\n\n"
	if err := os.WriteFile(indexPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write root index: %w", err)
	}
	fmt.Printf("Initialized OKF bundle at %s\n", app.root)
	return nil
}
