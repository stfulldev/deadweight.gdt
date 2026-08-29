package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/report"
)

type presentationFormat string

const (
	presentationText presentationFormat = "text"
	presentationJSON presentationFormat = "json"
)

func parsePresentationFormat(value string) (presentationFormat, error) {
	format := presentationFormat(value)
	switch format {
	case presentationText, presentationJSON:
		return format, nil
	default:
		return "", fmt.Errorf("invalid format %q; want text or json", value)
	}
}

func inspectArguments(format, metricValue *string, topLimit *int64) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := sceneArguments(format)(command, args); err != nil {
			return err
		}
		_, err := parseContributionSelection(command, *metricValue, *topLimit)
		return err
	}
}

func parseContributionSelection(
	command *cobra.Command,
	metricValue string,
	topLimit int64,
) (report.ContributionSelection, error) {
	metricSet := command.Flags().Changed("metric")
	topSet := command.Flags().Changed("top")
	if metricSet != topSet {
		return report.ContributionSelection{}, fmt.Errorf("--metric and --top must be supplied together")
	}
	if !metricSet {
		return report.ContributionSelection{}, nil
	}
	if topLimit <= 0 {
		return report.ContributionSelection{}, fmt.Errorf("--top must be a positive integer, got %d", topLimit)
	}
	name := metrics.Name(metricValue)
	switch name {
	case metrics.Nodes,
		metrics.TreeDepth,
		metrics.SceneInstances,
		metrics.MeshInstances,
		metrics.Lights,
		metrics.ShadowLights:
		return report.ContributionSelection{Metric: name, Limit: topLimit}, nil
	case metrics.ExternalResources, metrics.SceneDependencies:
		return report.ContributionSelection{}, fmt.Errorf(
			"--metric %s is a shared unique union and has no additive top-owner ranking",
			metricValue,
		)
	default:
		return report.ContributionSelection{}, fmt.Errorf(
			"invalid --metric %q; want nodes, tree_depth, scene_instances, mesh_instances, lights, or shadow_lights",
			metricValue,
		)
	}
}

func sceneArguments(format *string) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(command, args); err != nil {
			return err
		}
		_, err := parsePresentationFormat(*format)
		return err
	}
}

type presentationError struct {
	err     error
	format  presentationFormat
	options report.Options
}

func (failure *presentationError) Error() string {
	return failure.err.Error()
}

func (failure *presentationError) Unwrap() error {
	return failure.err
}

func wrapPresentationError(err error, format presentationFormat, options report.Options) error {
	if err == nil || format != presentationJSON {
		return err
	}
	return &presentationError{err: err, format: format, options: options}
}
