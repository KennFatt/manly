package main

import (
	"os"

	"github.com/KennFatt/manly/internal/knowledge"
)

type ContextCommand struct {
	Query  string `arg:"" help:"Search query or concept ID."`
	Tag    string `help:"Filter by tag."`
	Type   string `help:"Filter by type."`
	Path   string `help:"Restrict results to a path prefix."`
	Limit  int    `default:"5" help:"Maximum concepts."`
	Format string `default:"${format}" help:"Output format."`
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
	options := knowledge.SearchOptions{
		Tag:   command.Tag,
		Type:  command.Type,
		Path:  command.Path,
		Limit: command.Limit,
	}
	query := command.Query
	var results []knowledge.SearchResult
	if ref, getErr := workspace.ResolveConcept(query); getErr == nil {
		concept := displayConcept(workspace, *ref)
		results = []knowledge.SearchResult{{
			Concept: concept,
			Score:   knowledge.RankExactID.Weight(),
			Match: knowledge.Match{
				MatchedFields: []string{"id"},
				MatchedTerms:  []string{concept.ID},
				Rank:          knowledge.RankExactID,
			},
		}}
	} else if workspace.SingleRoot {
		results = knowledge.Search(workspace.Bundles[0], query, options)
	} else {
		refs, searchErr := workspace.Search(query, options)
		if searchErr != nil {
			return searchErr
		}
		for _, result := range refs {
			bundle := workspace.ByName[result.BundleName]
			results = append(results, knowledge.SearchResult{
				Concept:    displayConcept(workspace, knowledge.ConceptRef{BundleName: result.BundleName, Bundle: bundle, Concept: result.Concept}),
				Score:      result.Score,
				Match:      result.Match,
				BundleName: result.BundleName,
			})
		}
	}
	if len(results) > command.Limit && command.Limit > 0 {
		results = results[:command.Limit]
	}
	return renderContextResults(os.Stdout, workspace, results, query, format)
}
