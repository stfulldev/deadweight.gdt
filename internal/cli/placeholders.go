package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

var errAnalyzerNotImplemented = errors.New("scene analyzer is not implemented yet; see docs/MVP_0.1_SPEC.md")

func newInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <scene>",
		Short: "Inspect effective scene metrics",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errAnalyzerNotImplemented
		},
	}
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check <scene>",
		Short: "Check effective scene metrics against a budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errAnalyzerNotImplemented
		},
	}
}
