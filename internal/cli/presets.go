package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stfulldev/deadweight.gdt/internal/report"
)

func newPresetsCommand(service Application, global *globalOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "presets",
		Short: "List built-in heuristic budget presets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := service.ListPresets()
			if err != nil {
				return err
			}

			rendered, err := report.PresetList(result, global.reportOptions(cmd.OutOrStdout()))
			if err != nil {
				return fmt.Errorf("render preset list: %w", err)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), rendered)
			return err
		},
	}

	command.AddCommand(&cobra.Command{
		Use:   "show <id>",
		Short: "Show one built-in preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := service.ShowPreset(args[0])
			if err != nil {
				return err
			}

			rendered, err := report.PresetShow(result, global.reportOptions(cmd.OutOrStdout()))
			if err != nil {
				return fmt.Errorf("render preset: %w", err)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), rendered)
			return err
		},
	})

	return command
}
