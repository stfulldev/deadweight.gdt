package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

type projectedContribution struct {
	SourceIndex int
	Kind        analysis.ContributionKind
	Scene       string
	Declaring   string
	MountPath   string
	ResourceID  string
	RawTarget   string
	Reason      analysis.TargetClassification
	Occurrences int64
	Reliability analysis.Reliability
	Value       int64
	ValueKnown  bool
}

func validateContributionSelection(selection ContributionSelection) error {
	if !selection.Present() {
		return nil
	}
	if selection.Limit <= 0 {
		return fmt.Errorf("top contributor limit must be positive, got %d", selection.Limit)
	}
	switch selection.Metric {
	case metrics.Nodes,
		metrics.TreeDepth,
		metrics.SceneInstances,
		metrics.MeshInstances,
		metrics.Lights,
		metrics.ShadowLights:
		return nil
	case metrics.ExternalResources, metrics.SceneDependencies:
		return fmt.Errorf("metric %q is a shared unique union and has no additive top-owner ranking", selection.Metric)
	default:
		return fmt.Errorf("invalid top contributor metric %q", selection.Metric)
	}
}

func topContributions(
	result application.InspectResult,
	selection ContributionSelection,
) ([]projectedContribution, error) {
	if err := validateContributionSelection(selection); err != nil {
		return nil, err
	}
	if !selection.Present() {
		return nil, nil
	}

	projected := make([]projectedContribution, 0, len(result.Analysis.Summary.Contributions))
	for index, item := range result.Analysis.Summary.Contributions {
		if err := item.Validate(); err != nil {
			return nil, err
		}
		value, known := item.Values.Get(selection.Metric)
		if selection.Metric == metrics.TreeDepth {
			value = item.DepthCandidate.Value
			known = item.DepthCandidate.Known
		}
		projected = append(projected, projectedContribution{
			SourceIndex: index,
			Kind:        item.Kind,
			Scene:       portableContributionIdentity(result.Project.Directory, item),
			Declaring:   portableContributionDeclaring(result.Project.Directory, item),
			MountPath:   portableMountPath(item.MountPath),
			ResourceID:  item.ResourceID,
			RawTarget:   portableRawTarget(result.Project.Directory, item.RawTarget),
			Reason:      item.Classification,
			Occurrences: item.Occurrences,
			Reliability: item.Reliability,
			Value:       value,
			ValueKnown:  known,
		})
	}

	sort.Slice(projected, func(left, right int) bool {
		first := projected[left]
		second := projected[right]
		if first.ValueKnown != second.ValueKnown {
			return first.ValueKnown
		}
		if first.Value != second.Value {
			return first.Value > second.Value
		}
		if first.Scene != second.Scene {
			return first.Scene < second.Scene
		}
		if first.Declaring != second.Declaring {
			return first.Declaring < second.Declaring
		}
		if first.MountPath != second.MountPath {
			return first.MountPath < second.MountPath
		}
		if first.Kind != second.Kind {
			return first.Kind < second.Kind
		}
		if first.ResourceID != second.ResourceID {
			return first.ResourceID < second.ResourceID
		}
		if first.RawTarget != second.RawTarget {
			return first.RawTarget < second.RawTarget
		}
		if first.Reason != second.Reason {
			return first.Reason < second.Reason
		}

		return first.SourceIndex < second.SourceIndex
	})
	if selection.Limit < int64(len(projected)) {
		projected = projected[:int(selection.Limit)]
	}

	return projected, nil
}

func portableContributionIdentity(projectRoot string, item analysis.SceneContribution) string {
	for _, candidate := range []string{item.SceneDisplay, item.SceneOriginal, item.SceneCanonical} {
		if portable, ok := portableOptionalPath(projectRoot, candidate); ok {
			return portable
		}
	}
	if raw := portableRawTarget(projectRoot, item.RawTarget); raw != "" {
		return raw
	}
	if item.ResourceID != "" {
		return "ExtResource(" + item.ResourceID + ")"
	}
	if item.Classification != "" {
		return "<" + string(item.Classification) + ">"
	}

	return "<unknown>"
}

func portableContributionDeclaring(projectRoot string, item analysis.SceneContribution) string {
	for _, candidate := range []string{item.DeclaringDisplay, item.DeclaringScene} {
		if portable, ok := portableOptionalPath(projectRoot, candidate); ok {
			return portable
		}
	}

	return ""
}

func portableRawTarget(projectRoot, value string) string {
	if portable, ok := portableOptionalResource(projectRoot, value); ok {
		return portable
	}

	return ""
}

func portableMountPath(value string) string {
	return strings.ReplaceAll(value, `\`, "/")
}

func writeTopContributors(
	output *strings.Builder,
	result application.InspectResult,
	options Options,
) error {
	items, err := topContributions(result, options.Contributions)
	if err != nil {
		return err
	}
	if !options.Contributions.Present() {
		return nil
	}

	fmt.Fprintf(
		output,
		"\nTop contributors — %s (limit %s)\n",
		options.Contributions.Metric,
		formatInteger(options.Contributions.Limit),
	)
	for _, item := range items {
		value := "unknown"
		if item.ValueKnown {
			value = formatMetric(item.Value, item.Reliability)
		}
		context := "root"
		if item.Declaring != "" {
			context = "via " + item.Declaring
			if item.MountPath != "" {
				context += ":" + item.MountPath
			}
		}
		fmt.Fprintf(
			output,
			"  %10s  ×%-6s %-11s %-34s %s\n",
			value,
			formatInteger(item.Occurrences),
			item.Reliability,
			item.Scene,
			context,
		)
	}

	return nil
}
