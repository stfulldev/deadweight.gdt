package main

import (
	"os"

	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr, cli.BuildInfo{
		Version: version,
	}))
}
