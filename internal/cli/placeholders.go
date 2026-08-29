package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/report"
)

func newInspectCommand(service Application, global *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <scene>",
		Short: "Inspect effective scene metrics",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			result, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
				Scene:   args[0],
				Project: global.project,
				Config:  global.config,
			}})
			if err != nil {
				return err
			}

			rendered, err := report.Inspect(result, global.reportOptions(command.OutOrStdout()))
			if err != nil {
				return fmt.Errorf("render inspect report: %w", err)
			}

			_, err = fmt.Fprint(command.OutOrStdout(), rendered)
			return err
		},
	}
}

func newCheckCommand(service Application, global *globalOptions) *cobra.Command {
	var presetID string
	var profileID string
	var budgetOverrides []string
	var failOnPartial bool
	var allowPartial bool

	command := &cobra.Command{
		Use:   "check <scene>",
		Short: "Check effective scene metrics against a budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			partialOverride := budget.PartialInherit
			if failOnPartial {
				partialOverride = budget.PartialFail
			}
			if allowPartial {
				partialOverride = budget.PartialAllow
			}

			result, err := service.Check(application.CheckRequest{
				SceneRequest: application.SceneRequest{
					Scene:   args[0],
					Project: global.project,
					Config:  global.config,
				},
				Selector: policy.Selector{
					Preset:  presetID,
					Profile: profileID,
				},
				BudgetOverrides: append([]string(nil), budgetOverrides...),
				PartialOverride: partialOverride,
			})
			if err != nil {
				return err
			}

			rendered, err := report.Check(result, global.reportOptions(command.OutOrStdout()))
			if err != nil {
				return fmt.Errorf("render check report: %w", err)
			}
			if _, err := fmt.Fprint(command.OutOrStdout(), rendered); err != nil {
				return err
			}

			return exitForEvaluation(result.Evaluation.Status)
		},
	}
	command.Flags().StringVar(&presetID, "preset", "", "built-in preset ID")
	command.Flags().StringVar(&profileID, "profile", "", "custom configuration profile ID")
	command.Flags().StringArrayVar(&budgetOverrides, "budget", nil, "final METRIC=LIMIT override (repeatable)")
	command.Flags().BoolVar(&failOnPartial, "fail-on-partial", false, "return exit 3 for partial analysis")
	command.Flags().BoolVar(&allowPartial, "allow-partial", false, "allow partial analysis regardless of config")
	command.MarkFlagsMutuallyExclusive("preset", "profile")
	command.MarkFlagsMutuallyExclusive("fail-on-partial", "allow-partial")

	return command
}
