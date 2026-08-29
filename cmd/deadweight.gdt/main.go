package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr, cli.BuildInfo{
		Version: buildVersion(),
	}))
}

func buildVersion() string {
	moduleVersion := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		moduleVersion = info.Main.Version
	}

	return resolveVersion(version, moduleVersion)
}

func resolveVersion(injected, moduleVersion string) string {
	if !isDevelopmentVersion(injected) {
		return injected
	}
	if !isDevelopmentVersion(moduleVersion) {
		return strings.TrimPrefix(moduleVersion, "v")
	}

	return "dev"
}

func isDevelopmentVersion(value string) bool {
	return value == "" || value == "dev" || value == "(devel)"
}
