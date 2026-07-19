package main

import (
	"fmt"
	"runtime/debug"
)

// version is overridden at link time for release builds.
var version = "dev"

type VersionCommand struct{}

func (command VersionCommand) Run(_ *appContext) error {
	fmt.Printf("manly %s\n", executableVersion())
	return nil
}

func executableVersion() string {
	if version != "dev" {
		return version
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if ok && buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		return buildInfo.Main.Version
	}

	return version
}
