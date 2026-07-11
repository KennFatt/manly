package main

import (
	"errors"
	"fmt"

	"github.com/KennFatt/manly/internal/knowledge"
)

func runLinks(root string, args []string, backlinks bool) error {
	name := "links"
	if backlinks {
		name = "backlinks"
	}
	flags := newFlagSet(name)
	formatValue := flags.String("format", string(formatHuman), "output format")
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
	formatValue := flags.String("format", string(formatHuman), "output format")
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

func renderLinks(concept *knowledge.Concept, links []knowledge.Link, format outputFormat) error {
	views := make([]linkView, 0, len(links))
	for _, link := range links {
		if link.External || link.TargetID != "" || link.TargetPath != "" || link.Broken {
			views = append(views, renderLink(link))
		}
	}
	switch format {
	case formatJSON:
		return writeJSON(map[string]any{"source": concept.ID, "links": views})
	case formatMarkdown:
		fmt.Printf("# Links from %s\n\n", concept.ID)
		for _, link := range views {
			fmt.Printf("* [%s](%s)\n", link.Label, link.URL)
		}
	default:
		fmt.Printf("Links from %s\n\n", concept.ID)
		for index, link := range views {
			if link.Target != "" {
				fmt.Printf("[%d] %s\n    %s\n    manly show %s\n\n", index+1, link.Label, link.Target, link.Target)
			} else if command := linkNavigationCommand(link); command != "" {
				fmt.Printf("[%d] %s\n    %s\n    %s\n\n", index+1, link.Label, link.TargetPath, command)
			} else {
				fmt.Printf("[%d] %s\n    %s\n\n", index+1, link.Label, link.URL)
			}
		}
		if len(views) == 0 {
			fmt.Println("No links.")
		}
	}
	return nil
}

func renderBacklinks(concept *knowledge.Concept, backlinks []knowledge.Backlink, format outputFormat) error {
	views := make([]linkView, 0, len(backlinks))
	for _, backlink := range backlinks {
		views = append(views, linkView{Label: backlink.Link.Label, Target: backlink.Concept.ID, TargetPath: backlink.Concept.RelPath})
	}
	switch format {
	case formatJSON:
		return writeJSON(map[string]any{"target": concept.ID, "backlinks": views})
	case formatMarkdown:
		fmt.Printf("# Backlinks to %s\n\n", concept.ID)
		for _, link := range views {
			fmt.Printf("* [%s](%s.md)\n", link.Label, link.Target)
		}
	default:
		fmt.Printf("Backlinks to %s\n\n", concept.ID)
		for index, backlink := range backlinks {
			fmt.Printf("[%d] %s\n    %s (%s)\n    manly show %s\n\n", index+1, conceptTitle(backlink.Concept), backlink.Concept.ID, backlink.Link.Label, backlink.Concept.ID)
		}
		if len(backlinks) == 0 {
			fmt.Println("No backlinks.")
		}
	}
	return nil
}

func renderGraph(nodes []knowledge.GraphNode, format outputFormat) error {
	type graphView struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Depth int    `json:"depth"`
	}
	views := make([]graphView, 0, len(nodes))
	for _, node := range nodes {
		views = append(views, graphView{ID: node.Concept.ID, Title: conceptTitle(node.Concept), Depth: node.Depth})
	}
	switch format {
	case formatJSON:
		return writeJSON(map[string]any{"nodes": views})
	case formatMarkdown:
		fmt.Println("# Concept graph")
		for _, node := range views {
			fmt.Printf("* [%s](%s.md) - depth %d\n", node.Title, node.ID, node.Depth)
		}
	default:
		fmt.Println("Concept graph")
		for _, node := range views {
			fmt.Printf("%*s%s  %s\n", node.Depth*2, "", node.ID, node.Title)
		}
	}
	return nil
}
