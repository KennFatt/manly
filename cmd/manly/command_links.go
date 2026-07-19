package main

type LinksCommand struct {
	ConceptID string `arg:"" help:"Concept ID."`
	Format    string `default:"compact" help:"Output format."`
}

func (command *LinksCommand) Run(app *appContext) error {
	return runLinkCommand(app.root, command.ConceptID, command.Format, false)
}

func runLinkCommand(root, conceptID, formatValue string, backlinks bool) error {
	format, err := parseFormat(formatValue)
	if err != nil {
		return err
	}
	bundle, err := loadBundle(root)
	if err != nil {
		return err
	}
	concept, err := bundle.Get(conceptID)
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
