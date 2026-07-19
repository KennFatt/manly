package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KennFatt/manly/internal/config"
	"github.com/KennFatt/manly/internal/knowledge"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "manly: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}
	resolvedConfig, err := config.Load(home)
	if err != nil {
		return err
	}

	commandLine := &cli{}
	exitCode := -1
	parser, err := newParser(commandLine, func(code int) {
		exitCode = code
	}, resolvedConfig)
	if err != nil {
		return err
	}

	parseArgs := args
	if wantsTopLevelHelp(args) {
		parseArgs = []string{"--help"}
	}
	context, err := parser.Parse(parseArgs)
	if exitCode >= 0 {
		if exitCode == 0 {
			return nil
		}
		if err == nil {
			return fmt.Errorf("parser exited with status %d", exitCode)
		}
	}
	if err != nil {
		return err
	}

	root, err := resolveRoot(commandLine.Root, resolvedConfig.Root)
	if err != nil {
		return err
	}
	return context.Run(&appContext{root: root, display: resolvedConfig.Display})
}

func wantsTopLevelHelp(args []string) bool {
	if len(args) == 0 {
		return true
	}
	for index := 0; index < len(args); index++ {
		switch {
		case args[index] == "--root":
			index++
		case strings.HasPrefix(args[index], "--root="):
			// The value is part of this argument.
		case args[index] == "help":
			return true
		default:
			return false
		}
	}
	return false
}

func resolveRoot(explicit string, configured ...string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if root := os.Getenv("MANLY_ROOT"); root != "" {
		return root, nil
	}
	if len(configured) > 0 && configured[0] != "" {
		return configured[0], nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".okf"), nil
}

func loadBundle(root string) (*knowledge.Bundle, error) {
	return knowledge.Load(root)
}
