package main

import (
	"os"

	"github.com/alecthomas/kong"
)

// appContext contains runtime values shared by command handlers.
type appContext struct {
	root string
}

type cli struct {
	Root string `help:"Knowledge bundle or workspace root (overrides MANLY_ROOT)." placeholder:"PATH"`

	Init      InitCommand      `cmd help:"Initialize an OKF bundle."`
	List      ListCommand      `cmd help:"List directories or concepts."`
	Show      ShowCommand      `cmd help:"Show one or more concepts."`
	Search    SearchCommand    `cmd help:"Search concepts."`
	Context   ContextCommand   `cmd help:"Retrieve bounded agent context."`
	Links     LinksCommand     `cmd help:"Show outgoing links."`
	Backlinks BacklinksCommand `cmd help:"Show incoming links."`
	Graph     GraphCommand     `cmd help:"Traverse linked concepts."`
	Add       AddCommand       `cmd help:"Create a concept."`
	Edit      EditCommand      `cmd help:"Open a concept in $EDITOR."`
	Move      MoveCommand      `cmd help:"Move a concept and update links."`
	Index     IndexCommand     `cmd help:"Update marked generated index sections."`
	Check     CheckCommand     `cmd help:"Validate the bundle."`
	Version   VersionCommand   `cmd help:"Print the manly executable version."`
}

func newParser(commandLine *cli, exit func(int)) (*kong.Kong, error) {
	return kong.New(
		commandLine,
		kong.Name("manly"),
		kong.Description("Navigate local OKF knowledge bundles and workspaces."),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(exit),
	)
}
