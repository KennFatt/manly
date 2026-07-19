package main

import (
	"os"

	"github.com/KennFatt/manly/internal/knowledge"
)

type SearchCommand struct {
	Query  string `arg:"" help:"Text to search for."`
	Tag    string `help:"Filter by tag."`
	Type   string `help:"Filter by type."`
	Path   string `help:"Restrict results to a path prefix."`
	Limit  int    `default:"10" help:"Maximum results."`
	Format string `default:"compact" help:"Output format."`
}

func (command *SearchCommand) Run(app *appContext) error {
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
	if workspace.SingleRoot {
		results := knowledge.Search(workspace.Bundles[0], command.Query, options)
		return renderSearchResults(os.Stdout, results, command.Query, format)
	}
	results, err := workspace.Search(command.Query, options)
	if err != nil {
		return err
	}
	return renderWorkspaceSearch(workspace, results, command.Query, format)
}
