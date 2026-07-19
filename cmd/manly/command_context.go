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
	workspace, err := loadWorkspace(app.root)
	if err != nil {
		return err
	}
	query := command.Query
	var results []knowledge.SearchResult
	if ref, getErr := workspace.ResolveConcept(query); getErr == nil {
		concept := displayConcept(workspace, *ref)
		results = []knowledge.SearchResult{{Concept: concept, Score: 1}}
	} else if workspace.SingleRoot {
		results = knowledge.Search(workspace.Bundles[0], query, knowledge.SearchOptions{Limit: command.Limit})
	} else {
		refs, searchErr := workspace.Search(query, knowledge.SearchOptions{Limit: command.Limit})
		if searchErr != nil {
			return searchErr
		}
		for _, result := range refs {
			bundle := workspace.ByName[result.BundleName]
			results = append(results, knowledge.SearchResult{
				Concept: displayConcept(workspace, knowledge.ConceptRef{BundleName: result.BundleName, Bundle: bundle, Concept: result.Concept}),
				Score:   result.Score,
			})
		}
	}
	if len(results) > command.Limit && command.Limit > 0 {
		results = results[:command.Limit]
	}
	return renderContextResults(os.Stdout, results, query, format)
}
