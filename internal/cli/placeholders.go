package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/report"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

func newInspectCommand(service Application, global *globalOptions) *cobra.Command {
	formatValue := string(presentationText)
	metricValue := ""
	var topLimit int64
	command := &cobra.Command{
		Use:   "inspect <scene>",
		Short: "Inspect effective scene metrics",
		Args:  inspectArguments(&formatValue, &metricValue, &topLimit),
		RunE: func(command *cobra.Command, args []string) error {
			format, _ := parsePresentationFormat(formatValue)
			options := global.reportOptions(command.OutOrStdout())
			options.Contributions, _ = parseContributionSelection(command, metricValue, topLimit)
			result, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
				Scene:   args[0],
				Project: global.project,
				Config:  global.config,
			}})
			if err != nil {
				return wrapPresentationError(err, format, options)
			}

			var rendered string
			if format == presentationJSON {
				rendered, err = report.InspectJSON(result, options)
			} else {
				rendered, err = report.Inspect(result, options)
			}
			if err != nil {
				return wrapPresentationError(fmt.Errorf("render inspect report: %w", err), format, options)
			}

			_, err = fmt.Fprint(command.OutOrStdout(), rendered)
			return wrapPresentationError(err, format, options)
		},
	}
	command.Flags().StringVar(&formatValue, "format", string(presentationText), "output format: text or json")
	command.Flags().StringVar(&metricValue, "metric", "", "metric for the top-contributors view")
	command.Flags().Int64Var(&topLimit, "top", 0, "positive number of contributor rows to show")

	return command
}

func newTreeCommand(service Application, global *globalOptions) *cobra.Command {
	formatValue := string(presentationText)
	command := &cobra.Command{
		Use:   "tree <scene>",
		Short: "Explain the scene dependency tree",
		Args:  sceneArguments(&formatValue),
		RunE: func(command *cobra.Command, args []string) error {
			format, _ := parsePresentationFormat(formatValue)
			options := global.reportOptions(command.OutOrStdout())
			result, err := service.Tree(application.TreeRequest{SceneRequest: application.SceneRequest{
				Scene:   args[0],
				Project: global.project,
				Config:  global.config,
			}})
			if err != nil {
				return wrapPresentationError(err, format, options)
			}

			var rendered string
			if format == presentationJSON {
				rendered, err = report.TreeJSON(result, options)
			} else {
				rendered, err = report.Tree(result, options)
			}
			if err != nil {
				return wrapPresentationError(fmt.Errorf("render dependency tree: %w", err), format, options)
			}

			_, err = fmt.Fprint(command.OutOrStdout(), rendered)
			return wrapPresentationError(err, format, options)
		},
	}
	command.Flags().StringVar(&formatValue, "format", string(presentationText), "output format: text or json")

	return command
}

func newCheckCommand(service Application, global *globalOptions) *cobra.Command {
	var presetID string
	var profileID string
	var budgetOverrides []string
	var failOnPartial bool
	var allowPartial bool
	formatValue := string(presentationText)

	command := &cobra.Command{
		Use:   "check <scene>",
		Short: "Check effective scene metrics against a budget",
		Args:  sceneArguments(&formatValue),
		RunE: func(command *cobra.Command, args []string) error {
			format, _ := parsePresentationFormat(formatValue)
			options := global.reportOptions(command.OutOrStdout())
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
				return wrapPresentationError(err, format, options)
			}

			var rendered string
			if format == presentationJSON {
				rendered, err = report.CheckJSON(result, options)
			} else {
				rendered, err = report.Check(result, options)
			}
			if err != nil {
				return wrapPresentationError(fmt.Errorf("render check report: %w", err), format, options)
			}
			if _, err := fmt.Fprint(command.OutOrStdout(), rendered); err != nil {
				return wrapPresentationError(err, format, options)
			}

			return exitForEvaluation(result.Evaluation.Status)
		},
	}
	command.Flags().StringVar(&presetID, "preset", "", "built-in preset ID")
	command.Flags().StringVar(&profileID, "profile", "", "custom configuration profile ID")
	command.Flags().StringArrayVar(&budgetOverrides, "budget", nil, "final METRIC=LIMIT override (repeatable)")
	command.Flags().BoolVar(&failOnPartial, "fail-on-partial", false, "return exit 3 for partial analysis")
	command.Flags().BoolVar(&allowPartial, "allow-partial", false, "allow partial analysis regardless of config")
	command.Flags().StringVar(&formatValue, "format", string(presentationText), "output format: text or json")
	command.MarkFlagsMutuallyExclusive("preset", "profile")
	command.MarkFlagsMutuallyExclusive("fail-on-partial", "allow-partial")

	return command
}

func newDiffCommand(service Application, global *globalOptions) *cobra.Command {
	formatValue := string(presentationText)
	var metricValues []string
	var failOnReliability bool
	command := &cobra.Command{
		Use:   "diff <before.json> <after.json>",
		Short: "Compare two portable JSON reports",
		Args: func(command *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(command, args); err != nil {
				return err
			}
			if _, err := parsePresentationFormat(formatValue); err != nil {
				return err
			}
			_, err := parseDiffPolicy(metricValues, failOnReliability)
			return err
		},
		RunE: func(command *cobra.Command, args []string) error {
			format, _ := parsePresentationFormat(formatValue)
			options := global.reportOptions(command.OutOrStdout())
			policy, _ := parseDiffPolicy(metricValues, failOnReliability)
			result, err := service.Diff(application.DiffRequest{Before: args[0], After: args[1], Policy: policy})
			if err != nil {
				return wrapPresentationError(err, format, options)
			}
			var rendered string
			if format == presentationJSON {
				rendered, err = report.DiffJSON(result, options)
			} else {
				rendered, err = report.Diff(result, options)
			}
			if err != nil {
				return wrapPresentationError(fmt.Errorf("render diff report: %w", err), format, options)
			}
			if _, err := fmt.Fprint(command.OutOrStdout(), rendered); err != nil {
				return wrapPresentationError(err, format, options)
			}
			return exitForEvaluation(result.Comparison.Enforcement.Status)
		},
	}
	command.Flags().StringVar(&formatValue, "format", string(presentationText), "output format: text or json")
	command.Flags().StringArrayVar(&metricValues, "fail-on-increase", nil, "return a non-zero exit when METRIC increases (repeatable)")
	command.Flags().BoolVar(&failOnReliability, "fail-on-reliability", false, "return exit 3 when report reliability degrades")
	return command
}

func parseDiffPolicy(metricValues []string, failOnReliability bool) (reportdiff.Policy, error) {
	policy := reportdiff.Policy{FailOnReliability: failOnReliability}
	for _, value := range metricValues {
		policy.MetricIncreases = append(policy.MetricIncreases, metrics.Name(value))
	}
	normalized, err := reportdiff.NormalizePolicy(policy)
	if err != nil {
		return reportdiff.Policy{}, fmt.Errorf("invalid --fail-on-increase: %w", err)
	}
	return normalized, nil
}
