package main

type BacklinksCommand struct {
	ConceptID string `arg:"" help:"Concept ID."`
	Format    string `default:"${format}" help:"Output format."`
}

func (command *BacklinksCommand) Run(app *appContext) error {
	return runLinkCommand(app.root, command.ConceptID, command.Format, true)
}
