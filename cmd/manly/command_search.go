package main

import (
	"errors"
	"os"

	"github.com/KennFatt/manly/internal/knowledge"
)

func runSearch(root string, args []string) error {
	flags := newFlagSet("search")
	tag := flags.String("tag", "", "filter by tag")
	typeFilter := flags.String("type", "", "filter by type")
	pathFilter := flags.String("path", "", "filter by path")
	limit := flags.Int("limit", 10, "maximum results")
	formatValue := flags.String("format", string(formatCompact), "output format")
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
	return renderSearchResults(os.Stdout, results, flags.Arg(0), format)
}

func runContext(root string, args []string) error {
	flags := newFlagSet("context")
	limit := flags.Int("limit", 5, "maximum concepts")
	formatValue := flags.String("format", string(formatCompact), "output format")
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
	return renderContextResults(os.Stdout, results, query, format)
}
