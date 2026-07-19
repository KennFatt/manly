package main

import "github.com/KennFatt/manly/internal/knowledge"

type LinksCommand struct {
	ConceptID string `arg:"" help:"Concept ID."`
	Format    string `default:"${format}" help:"Output format."`
}

func (command *LinksCommand) Run(app *appContext) error {
	return runLinkCommand(app.root, command.ConceptID, command.Format, false)
}

func runLinkCommand(root, conceptID, formatValue string, backlinks bool) error {
	format, err := parseFormat(formatValue)
	if err != nil {
		return err
	}
	workspace, err := loadWorkspace(root)
	if err != nil {
		return err
	}
	ref, err := workspace.ResolveConcept(conceptID)
	if err != nil {
		return err
	}
	if workspace.SingleRoot {
		if backlinks {
			incoming, err := ref.Bundle.Backlinks(ref.Concept.ID)
			if err != nil {
				return err
			}
			return renderBacklinks(ref.Concept, incoming, format)
		}
		outgoing, err := ref.Bundle.Outgoing(ref.Concept.ID)
		if err != nil {
			return err
		}
		return renderLinks(ref.Concept, outgoing, format)
	}
	displayed := displayConcept(workspace, *ref)
	if backlinks {
		incoming, err := ref.Bundle.Backlinks(ref.Concept.ID)
		if err != nil {
			return err
		}
		for index := range incoming {
			incoming[index].Concept = displayConcept(workspace, knowledge.ConceptRef{BundleName: ref.BundleName, Bundle: ref.Bundle, Concept: incoming[index].Concept})
			incoming[index].Link = displayLinks(workspace, ref.BundleName, []knowledge.Link{incoming[index].Link})[0]
		}
		return renderBacklinks(displayed, incoming, format)
	}
	outgoing, err := ref.Bundle.Outgoing(ref.Concept.ID)
	if err != nil {
		return err
	}
	return renderLinks(displayed, displayLinks(workspace, ref.BundleName, outgoing), format)
}
