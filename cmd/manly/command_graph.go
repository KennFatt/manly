package main

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
	bundle, err := loadBundle(app.root)
	if err != nil {
		return err
	}
	nodes, err := bundle.Graph(command.ConceptID, command.Depth)
	if err != nil {
		return err
	}
	return renderGraph(nodes, format)
}
