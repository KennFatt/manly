package main

import "fmt"

// version is overridden at link time for release builds.
var version = "dev"

type VersionCommand struct{}

func (command VersionCommand) Run(_ *appContext) error {
	fmt.Printf("manly %s\n", version)
	return nil
}
