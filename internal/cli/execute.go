package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// BuildInfo contains values injected at build time.
type BuildInfo struct {
	Version string
}

// Execute runs the CLI and maps command failures to the MVP usage/fatal exit code.
func Execute(args []string, stdout, stderr io.Writer, info BuildInfo) int {
	root := NewRoot(info)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 2
	}

	return 0
}

// NewRoot constructs a fresh command tree suitable for execution and tests.
func NewRoot(info BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:           "deadweight.gdt",
		Short:         "Static scene-complexity budgets for Godot 4 text scenes",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       info.Version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetVersionTemplate("deadweight.gdt {{.Version}}\n")

	root.AddCommand(
		newInspectCommand(),
		newCheckCommand(),
		newPresetsCommand(),
	)

	return root
}
