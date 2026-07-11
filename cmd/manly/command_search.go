package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/KennFatt/manly/internal/knowledge"
)

func runSearch(root string, args []string) error {
	flags := newFlagSet("search")
	tag := flags.String("tag", "", "filter by tag")
	typeFilter := flags.String("type", "", "filter by type")
	pathFilter := flags.String("path", "", "filter by path")
	limit := flags.Int("limit", 10, "maximum results")
	formatValue := flags.String("format", string(formatHuman), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--tag": true, "--type": true, "--path": true, "--limit": true, "--format": true})); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("usage: manly search <query> [--tag TAG] [--type TYPE] [--path PATH] [--limit N] [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	results := knowledge.Search(bundle, flags.Arg(0), knowledge.SearchOptions{
		Tag:   *tag,
		Type:  *typeFilter,
		Path:  *pathFilter,
		Limit: *limit,
	})
	return renderSearchResults(results, flags.Arg(0), format)
}

func runContext(root string, args []string) error {
	flags := newFlagSet("context")
	limit := flags.Int("limit", 5, "maximum concepts")
	formatValue := flags.String("format", string(formatHuman), "output format")
	if err := flags.Parse(normalizeFlagArgs(args, map[string]bool{"--limit": true, "--format": true})); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("usage: manly context <query-or-concept-id> [--limit N] [--format FORMAT]")
	}
	format, err := parseFormat(*formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	query := flags.Arg(0)
	var results []knowledge.SearchResult
	if concept, getErr := bundle.Get(query); getErr == nil {
		results = []knowledge.SearchResult{{Concept: concept, Score: 1}}
	} else {
		results = knowledge.Search(bundle, query, knowledge.SearchOptions{Limit: *limit})
	}
	if len(results) > *limit && *limit > 0 {
		results = results[:*limit]
	}
	return renderContextResults(results, query, format)
}

func renderSearchResults(results []knowledge.SearchResult, query string, format outputFormat) error {
	type searchView struct {
		Concept conceptView  `json:"concept"`
		Score   float64      `json:"score"`
		Actions []actionView `json:"actions"`
	}
	views := make([]searchView, 0, len(results))
	for _, result := range results {
		views = append(views, searchView{
			Concept: viewConcept(result.Concept, false),
			Score:   result.Score,
			Actions: actionViews(result.Concept.ID),
		})
	}
	switch format {
	case formatJSON:
		return writeJSON(map[string]any{"query": query, "results": views})
	case formatMarkdown:
		fmt.Printf("# Search results for %q\n\n", query)
		for _, result := range results {
			fmt.Printf("* [%s](%s.md) - %s\n", conceptTitle(result.Concept), result.Concept.ID, conceptDescription(result.Concept))
		}
		return nil
	case formatCompact:
		for _, result := range results {
			fmt.Printf("%.2f\t%s\t%s\n", result.Score, result.Concept.ID, conceptTitle(result.Concept))
		}
		return nil
	default:
		fmt.Printf("Search results for %q\n\n", query)
		for index, result := range results {
			fmt.Printf("[%d] %s\n", index+1, conceptTitle(result.Concept))
			fmt.Printf("    %s\n", result.Concept.ID)
			if description := conceptDescription(result.Concept); description != "" {
				fmt.Printf("    %s\n", description)
			}
			fmt.Printf("    Open:    manly show %s\n", result.Concept.ID)
			fmt.Printf("    Context: manly context %s\n\n", result.Concept.ID)
		}
		if len(results) == 0 {
			fmt.Println("No matching concepts.")
		}
		return nil
	}
}

func renderContextResults(results []knowledge.SearchResult, query string, format outputFormat) error {
	type contextView struct {
		Concept conceptView  `json:"concept"`
		Score   float64      `json:"score"`
		Links   []linkView   `json:"links"`
		Actions []actionView `json:"actions"`
	}
	views := make([]contextView, 0, len(results))
	for _, result := range results {
		links := make([]linkView, 0, len(result.Concept.Links))
		for _, link := range result.Concept.Links {
			if link.External || link.TargetID != "" || link.TargetPath != "" || link.Broken {
				links = append(links, renderLink(link))
			}
		}
		views = append(views, contextView{
			Concept: viewConcept(result.Concept, true),
			Score:   result.Score,
			Links:   links,
			Actions: actionViews(result.Concept.ID),
		})
	}
	switch format {
	case formatJSON:
		return writeJSON(map[string]any{"query": query, "results": views})
	case formatMarkdown:
		for _, result := range results {
			fmt.Printf("## %s\n\n%s\n\n", conceptTitle(result.Concept), strings.TrimSpace(result.Concept.Body))
		}
		return nil
	case formatCompact:
		for _, result := range results {
			fmt.Printf("%s\n%s\n\n", result.Concept.ID, strings.TrimSpace(result.Concept.Body))
		}
		return nil
	default:
		fmt.Printf("Context for %q\n\n", query)
		for _, result := range results {
			fmt.Printf("## %s\n", conceptTitle(result.Concept))
			fmt.Printf("ID: %s\n\n%s\n\n", result.Concept.ID, strings.TrimSpace(result.Concept.Body))
			fmt.Printf("Open: manly show %s\n\n", result.Concept.ID)
		}
		if len(results) == 0 {
			fmt.Println("No matching concepts.")
		}
		return nil
	}
}
