// Package report renders deterministic human-readable console output.
package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

var (
	structureMetrics = []metrics.Name{
		metrics.Nodes,
		metrics.TreeDepth,
		metrics.SceneInstances,
		metrics.SceneDependencies,
	}
	renderingMetrics = []metrics.Name{
		metrics.MeshInstances,
		metrics.Lights,
		metrics.ShadowLights,
	}
	resourceMetrics = []metrics.Name{
		metrics.ExternalResources,
	}
)

// Options controls presentation without changing report meaning.
type Options struct {
	Version       string
	Color         bool
	Contributions ContributionSelection
}

// ContributionSelection requests an opt-in inspect top-contributors projection.
type ContributionSelection struct {
	Metric metrics.Name
	Limit  int64
}

// Present reports whether a contribution projection was requested.
func (selection ContributionSelection) Present() bool {
	return selection.Metric != "" || selection.Limit != 0
}

type styler struct {
	enabled bool
}

func (style styler) status(value string) string {
	switch value {
	case "PASS", "PASSED", "COMPLETE":
		return style.wrap(ansiGreen, value)
	case "FAIL", "FAILED", "ERROR":
		return style.wrap(ansiRed, value)
	case "WARNING", "PARTIAL", "INCOMPLETE":
		return style.wrap(ansiYellow, value)
	default:
		return value
	}
}

func (style styler) wrap(code, value string) string {
	if !style.enabled {
		return value
	}

	return code + value + ansiReset
}

func normalizedOptions(options Options) Options {
	if options.Version == "" {
		options.Version = "dev"
	}

	return options
}

func formatInteger(value int64) string {
	raw := strconv.FormatInt(value, 10)
	sign := ""
	if strings.HasPrefix(raw, "-") {
		sign = "-"
		raw = strings.TrimPrefix(raw, "-")
	}
	if len(raw) <= 3 {
		return sign + raw
	}

	first := len(raw) % 3
	if first == 0 {
		first = 3
	}

	var output strings.Builder
	output.WriteString(sign)
	output.WriteString(raw[:first])
	for index := first; index < len(raw); index += 3 {
		output.WriteByte(',')
		output.WriteString(raw[index : index+3])
	}

	return output.String()
}

func reliabilityMarker(reliability analysis.Reliability) string {
	switch reliability {
	case analysis.ReliabilityLowerBound:
		return "+"
	case analysis.ReliabilityApproximate:
		return "~"
	default:
		return ""
	}
}

func accuracyLabel(reliability analysis.Reliability) string {
	switch reliability {
	case analysis.ReliabilityExact:
		return "exact"
	case analysis.ReliabilityLowerBound:
		return "lower bound"
	case analysis.ReliabilityApproximate:
		return "approximate"
	default:
		return string(reliability)
	}
}

func formatMetric(value int64, reliability analysis.Reliability) string {
	return formatInteger(value) + reliabilityMarker(reliability)
}

func preferredScenePath(result application.InspectResult) string {
	for _, candidate := range []string{result.Scene.Display, result.Scene.Original, result.Scene.Canonical} {
		if candidate != "" {
			return candidate
		}
	}

	return "<unknown>"
}

func validateInspect(result application.InspectResult) error {
	if !result.Analysis.Status.Valid() {
		return fmt.Errorf("invalid analysis status %q", result.Analysis.Status)
	}
	if !result.Analysis.Reliability.Valid() {
		return fmt.Errorf("invalid analysis reliability %q", result.Analysis.Reliability)
	}
	if err := result.Analysis.Summary.Metrics.Validate(); err != nil {
		return err
	}
	if err := result.Analysis.Coverage.Validate(); err != nil {
		return err
	}
	for _, item := range result.Analysis.Diagnostics {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func writeVersion(output *strings.Builder, options Options) {
	fmt.Fprintf(output, "deadweight.gdt %s\n\n", options.Version)
}

func writeMetricBlock(
	output *strings.Builder,
	title string,
	names []metrics.Name,
	values metrics.Values,
	reliability analysis.Reliability,
) {
	output.WriteString(title)
	output.WriteByte('\n')
	for _, name := range names {
		value, _ := values.Get(name)
		fmt.Fprintf(output, "  %-26s %10s\n", name.Label(), formatMetric(value, reliability))
	}
}

func displayRenderer(value string) string {
	switch value {
	case "forward_plus":
		return "Forward+"
	case "mobile":
		return "Mobile"
	case "compatibility":
		return "Compatibility"
	default:
		return displayTitle(value)
	}
}

func displayTitle(value string) string {
	if value == "" || value == "unspecified" {
		return "Unspecified"
	}

	return strings.ToUpper(value[:1]) + strings.ReplaceAll(value[1:], "_", " ")
}

func plural(count int64, singular, pluralValue string) string {
	if count == 1 {
		return singular
	}

	return pluralValue
}
