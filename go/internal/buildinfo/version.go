package buildinfo

import (
	"runtime/debug"
	"strings"
)

// Version is set by the build pipeline.
var Version = "dev"

func IsVersionCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	argument := strings.ToLower(strings.TrimSpace(args[0]))
	return argument == "version" || argument == "--version" || argument == "-v"
}

func String() string {
	if Version != "" && Version != "dev" {
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return Version
}
