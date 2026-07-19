package main

import (
	"os"

	"github.com/KennFatt/manly/internal/knowledge"
)

type ContextCommand struct {
	Query  string `arg:"" help:"Search query or concept ID."`
	Limit  int    `default:"5" help:"Maximum concepts."`
	Format string `default:"compact" help:"Output format."`
}

func (command *ContextCommand) Run(app *appContext) error {
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(app.root)
	if err != nil {
		return err
	}
	query := command.Query
	var results []knowledge.SearchResult
	if concept, getErr := bundle.Get(query); getErr == nil {
		results = []knowledge.SearchResult{{Concept: concept, Score: 1}}
	} else {
		results = knowledge.Search(bundle, query, knowledge.SearchOptions{Limit: command.Limit})
	}
	if len(results) > command.Limit && command.Limit > 0 {
		results = results[:command.Limit]
	}
	return renderContextResults(os.Stdout, results, query, format)
}
