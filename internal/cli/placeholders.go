package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
)

type renderOptions struct {
	noColor bool
}

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

			_, err = fmt.Fprint(command.OutOrStdout(), renderInspect(result, renderOptions{noColor: global.noColor}))
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

			if _, err := fmt.Fprint(command.OutOrStdout(), renderCheck(result, renderOptions{noColor: global.noColor})); err != nil {
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

func renderInspect(result application.InspectResult, options renderOptions) string {
	var output strings.Builder
	renderAnalysisHeader(&output, result, options)
	output.WriteString("\nMetrics\n")
	for _, name := range metrics.OrderedNames() {
		value, _ := result.Analysis.Summary.Metrics.Get(name)
		fmt.Fprintf(&output, "  %-26s %8s\n", name.Label(), formatInteger(value))
	}
	fmt.Fprintf(&output, "\nDiagnostics: %d\n", len(result.Analysis.Diagnostics))

	return output.String()
}

func renderCheck(result application.CheckResult, options renderOptions) string {
	var output strings.Builder
	renderAnalysisHeader(&output, result.Inspect, options)
	fmt.Fprintf(&output, "Policy: %s\n", displayPolicy(result.Policy))
	output.WriteString("\nBudgets\n")
	for _, comparison := range result.Evaluation.Results {
		verdict := "PASS"
		if !comparison.Passed {
			verdict = "FAIL"
		}
		fmt.Fprintf(
			&output,
			"  %-26s %8s / %-8s %s\n",
			comparison.Metric.Label(),
			formatInteger(comparison.Actual),
			formatInteger(comparison.Limit),
			verdict,
		)
	}
	fmt.Fprintf(&output, "\nResult: %s\n", result.Evaluation.Status)

	return output.String()
}

func renderAnalysisHeader(output *strings.Builder, result application.InspectResult, options renderOptions) {
	_ = options.noColor
	scene := result.Scene.Display
	if scene == "" {
		scene = result.Scene.Original
	}
	if scene == "" {
		scene = result.Scene.Canonical
	}
	configuration := "none"
	if result.ConfigPresent {
		configuration = result.ConfigSource.Path
	}

	fmt.Fprintf(output, "Scene: %s\n", scene)
	fmt.Fprintf(output, "Project: %s\n", result.Project.Directory)
	fmt.Fprintf(output, "Config: %s\n", configuration)
	fmt.Fprintf(
		output,
		"Analysis: %s (%s)\n",
		strings.ToUpper(string(result.Analysis.Status)),
		result.Analysis.Reliability,
	)
}

func displayPolicy(effective policy.Effective) string {
	if effective.ID == "" {
		return "project/CLI overrides"
	}

	return string(effective.Kind) + " " + effective.ID
}
