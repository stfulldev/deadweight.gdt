package report

import (
	"fmt"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
)

// Inspect renders one complete inspect report.
func Inspect(result application.InspectResult, options Options) (string, error) {
	if err := validateInspect(result); err != nil {
		return "", err
	}
	options = normalizedOptions(options)
	style := styler{enabled: options.Color}

	var output strings.Builder
	writeVersion(&output, options)
	fmt.Fprintf(&output, "%-11s%s\n", "Scene:", preferredScenePath(result))
	fmt.Fprintf(&output, "%-11s%s\n", "Project:", result.Project.Directory)
	fmt.Fprintf(&output, "%-11s%s\n", "Analysis:", style.status(strings.ToUpper(string(result.Analysis.Status))))
	fmt.Fprintf(&output, "%-11s%s\n", "Accuracy:", accuracyLabel(result.Analysis.Reliability))

	values := result.Analysis.Summary.Metrics
	confidence := result.Analysis.MetricConfidence
	output.WriteByte('\n')
	writeMetricBlock(&output, "Structure", structureMetrics, values, confidence)
	output.WriteByte('\n')
	writeMetricBlock(&output, "Rendering", renderingMetrics, values, confidence)
	output.WriteByte('\n')
	writeMetricBlock(&output, "Resources", resourceMetrics, values, confidence)
	writeMetricConfidenceQualifications(&output, confidence, result.Analysis.Reliability)

	coverage := result.Analysis.Coverage
	output.WriteString("\nCoverage\n")
	fmt.Fprintf(&output, "  %-26s %10s\n", "Parsed scene files", formatInteger(coverage.ParsedSceneFiles))
	fmt.Fprintf(&output, "  %-26s %10s\n", "Resolved scene instances", formatInteger(coverage.ResolvedSceneInstances))
	fmt.Fprintf(&output, "  %-26s %10s\n", "Unresolved scene instances", formatInteger(coverage.UnresolvedSceneInstances))
	if coverage.InheritedScenes > 0 {
		fmt.Fprintf(&output, "  %-26s %10s\n", "Inherited scenes", formatInteger(coverage.InheritedScenes))
	}

	writeEvidence(&output, result.Analysis, style)
	writeReliabilityWarning(&output, result.Analysis, style)
	if err := writeTopContributors(&output, result, options); err != nil {
		return "", err
	}

	return output.String(), nil
}

func writeReliabilityWarning(output *strings.Builder, result analysis.RecursiveResult, style styler) {
	switch result.Reliability {
	case analysis.ReliabilityExact:
		return
	case analysis.ReliabilityLowerBound:
		count := result.Coverage.UnresolvedSceneInstances
		fmt.Fprintf(
			output,
			"\n%s: Expanded metrics are partial. Values marked with + are known\n",
			style.status("WARNING"),
		)
		if count > 0 {
			fmt.Fprintf(
				output,
				"lower bounds because %s %s could not be analyzed statically.\n",
				formatInteger(count),
				plural(count, "scene instance", "scene instances"),
			)
		} else {
			output.WriteString("lower bounds because some static evidence is unavailable.\n")
		}
	case analysis.ReliabilityApproximate:
		fmt.Fprintf(
			output,
			"\n%s: Expanded metrics are approximate. Values marked with ~ may vary\n",
			style.status("WARNING"),
		)
		output.WriteString("because inherited-scene overrides can change values in either direction.\n")
	}
}
