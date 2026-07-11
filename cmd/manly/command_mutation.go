package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

func runAdd(root string, args []string) error {
	flags := newFlagSet("add")
	typeValue := flags.String("type", "", "concept type")
	title := flags.String("title", "", "concept title")
	description := flags.String("description", "", "one-sentence description")
	tagsValue := flags.String("tag", "", "comma-separated tags")
	force := flags.Bool("force", false, "overwrite an existing concept")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--type": true, "--title": true, "--description": true, "--tag": true, "--force": false})); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("usage: manly add <concept-id> --type TYPE [--title TITLE] [--description TEXT] [--tag TAG,...]")
	}
	tags := splitTags(*tagsValue)
	id, err := knowledge.Add(root, flags.Arg(0), knowledge.NewConcept{
		Type:        *typeValue,
		Title:       *title,
		Description: *description,
		Tags:        tags,
	}, *force)
	if err != nil {
		return err
	}
	fmt.Printf("Created %s\n", id)
	return nil
}

func runMove(root string, args []string) error {
	flags := newFlagSet("move")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 2 {
		return errors.New("usage: manly move <old-concept-id> <new-concept-id>")
	}
	changed, err := knowledge.Move(root, flags.Arg(0), flags.Arg(1))
	if err != nil {
		return err
	}
	fmt.Printf("Moved %s to %s; updated %d link(s)\n", flags.Arg(0), flags.Arg(1), changed)
	return nil
}

func runIndex(root string, args []string) error {
	flags := newFlagSet("index")
	checkOnly := flags.Bool("check", false, "report stale generated sections without writing")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--check": false})); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: manly index [--check]")
	}
	changed, err := knowledge.UpdateIndexes(root, *checkOnly)
	if err != nil {
		return err
	}
	if *checkOnly {
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

func runCheck(root string, args []string) error {
	flags := newFlagSet("check")
	strict := flags.Bool("strict", false, "enable advisory checks")
	formatValue := flags.String("format", string(formatHuman), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--strict": false, "--format": true})); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: manly check [--strict] [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	report, err := knowledge.Validate(root, *strict)
	if err != nil {
		return err
	}
	if format == formatJSON {
		if err := writeJSON(report); err != nil {
			return err
		}
		if !report.Valid() {
			return errors.New("validation failed")
		}
		return nil
	}
	for _, issue := range report.Errors {
		fmt.Printf("ERROR: %s: %s\n", issue.Path, issue.Message)
	}
	for _, issue := range report.Warnings {
		fmt.Printf("WARNING: %s: %s\n", issue.Path, issue.Message)
	}
	if !report.Valid() {
		return fmt.Errorf("validation failed: %d error(s), %d warning(s)", len(report.Errors), len(report.Warnings))
	}
	fmt.Printf("OKF validation passed: %d warning(s)\n", len(report.Warnings))
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
