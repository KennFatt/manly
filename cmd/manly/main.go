package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "manly: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	root, commandArgs, err := parseGlobalArgs(args)
	if err != nil {
		return err
	}
	if len(commandArgs) == 0 || commandArgs[0] == "help" || commandArgs[0] == "--help" || commandArgs[0] == "-h" {
		printUsage()
		return nil
	}

	switch commandArgs[0] {
	case "init":
		return runInit(root, commandArgs[1:])
	case "list":
		return runList(root, commandArgs[1:])
	case "show":
		return runShow(root, commandArgs[1:])
	case "search":
		return runSearch(root, commandArgs[1:])
	case "context":
		return runContext(root, commandArgs[1:])
	case "links":
		return runLinks(root, commandArgs[1:], false)
	case "backlinks":
		return runLinks(root, commandArgs[1:], true)
	case "graph":
		return runGraph(root, commandArgs[1:])
	case "add":
		return runAdd(root, commandArgs[1:])
	case "edit":
		return runEdit(root, commandArgs[1:])
	case "move":
		return runMove(root, commandArgs[1:])
	case "index":
		return runIndex(root, commandArgs[1:])
	case "check":
		return runCheck(root, commandArgs[1:])
	default:
		return fmt.Errorf("unknown command %q; use 'manly help' for usage", commandArgs[0])
	}
}

func parseGlobalArgs(args []string) (string, []string, error) {
	root := os.Getenv("MANLY_ROOT")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", nil, fmt.Errorf("resolve home directory: %w", err)
		}
		root = filepath.Join(home, ".okf")
	}
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "--root" || argument == "-root":
			if index+1 >= len(args) {
				return "", nil, errors.New("--root requires a path")
			}
			root = args[index+1]
			index++
		case strings.HasPrefix(argument, "--root="):
			root = strings.TrimPrefix(argument, "--root=")
		case argument == "--help" || argument == "-h":
			return root, []string{"help"}, nil
		default:
			return root, args[index:], nil
		}
	}
	return root, nil, nil
}

func loadBundle(root string) (*knowledge.Bundle, error) {
	return knowledge.Load(root)
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	return flags
}

func normalizeFlagArgs(args []string, valueFlags map[string]bool) []string {
	var flags []string
	var positional []string
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if !strings.HasPrefix(argument, "-") || argument == "-" {
			positional = append(positional, argument)
			continue
		}
		flags = append(flags, argument)
		name := argument
		if equals := strings.IndexByte(name, '='); equals >= 0 {
			name = name[:equals]
		}
		if valueFlags[name] && !strings.Contains(argument, "=") && index+1 < len(args) {
			index++
			flags = append(flags, args[index])
		}
	}
	return append(flags, positional...)
}

func runInit(root string, args []string) error {
	flags := newFlagSet("init")
	force := flags.Bool("force", false, "allow creation of an existing root index")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: manly init [--force]")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create root: %w", err)
	}
	indexPath := filepath.Join(root, "index.md")
	if _, err := os.Stat(indexPath); err == nil && !*force {
		return fmt.Errorf("root index already exists: %s", indexPath)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check root index: %w", err)
	}
	content := "---\nokf_version: \"0.1\"\n---\n\n# Knowledge Bundle\n\n"
	if err := os.WriteFile(indexPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write root index: %w", err)
	}
	fmt.Printf("Initialized OKF bundle at %s\n", root)
	return nil
}

func runEdit(root string, args []string) error {
	flags := newFlagSet("edit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("usage: manly edit <concept-id>")
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	concept, err := bundle.Get(flags.Arg(0))
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
	command := exec.Command(parts[0], append(parts[1:], concept.AbsPath)...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func printUsage() {
	fmt.Print(`manly: navigate a global OKF knowledge bundle

Usage:
  manly [--root PATH] <command> [options]

Commands:
  init       Initialize an OKF bundle
  list       List directories or concepts
  show       Show one concept
  search     Search concepts
  context    Retrieve bounded agent context
  links      Show outgoing links
  backlinks  Show incoming links
  graph      Traverse linked concepts
  add        Create a concept
  edit       Open a concept in $EDITOR
  move       Move a concept and update links
  index      Update marked generated index sections
  check      Validate the bundle

Output formats:
  compact, fancy, json, markdown

The root defaults to $MANLY_ROOT or ~/.okf.
`)
}
