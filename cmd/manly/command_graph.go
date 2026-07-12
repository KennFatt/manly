package main

import (
	"errors"
	"fmt"
)

func runLinks(root string, args []string, backlinks bool) error {
	name := "links"
	if backlinks {
		name = "backlinks"
	}
	flags := newFlagSet(name)
	formatValue := flags.String("format", string(formatCompact), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--format": true})); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("usage: manly %s <concept-id> [--format FORMAT]", name)
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	concept, err := bundle.Get(flags.Arg(0))
	if err != nil {
		return err
	}
	if backlinks {
		incoming, err := bundle.Backlinks(concept.ID)
		if err != nil {
			return err
		}
		return renderBacklinks(concept, incoming, format)
	}
	outgoing, err := bundle.Outgoing(concept.ID)
	if err != nil {
		return err
	}
	return renderLinks(concept, outgoing, format)
}

func runGraph(root string, args []string) error {
	flags := newFlagSet("graph")
	depth := flags.Int("depth", 1, "maximum traversal depth")
	formatValue := flags.String("format", string(formatCompact), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--depth": true, "--format": true})); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("usage: manly graph <concept-id> [--depth N] [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	nodes, err := bundle.Graph(flags.Arg(0), *depth)
	if err != nil {
		return err
	}
	return renderGraph(nodes, format)
}
