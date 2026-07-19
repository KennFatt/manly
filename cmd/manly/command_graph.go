package main

import "github.com/KennFatt/manly/internal/knowledge"

type GraphCommand struct {
	ConceptID string `arg:"" help:"Starting concept ID."`
	Depth     int    `default:"1" help:"Maximum traversal depth."`
	Format    string `default:"compact" help:"Output format."`
}

func (command *GraphCommand) Run(app *appContext) error {
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	workspace, err := loadWorkspace(app.root)
	if err != nil {
		return err
	}
	ref, err := workspace.ResolveConcept(command.ConceptID)
	if err != nil {
		return err
	}
	nodes, err := ref.Bundle.Graph(ref.Concept.ID, command.Depth)
	if err != nil {
		return err
	}
	if !workspace.SingleRoot {
		for index := range nodes {
			nodes[index].Concept = displayConcept(workspace, knowledge.ConceptRef{BundleName: ref.BundleName, Bundle: ref.Bundle, Concept: nodes[index].Concept})
		}
	}
	return renderGraph(nodes, format)
}
