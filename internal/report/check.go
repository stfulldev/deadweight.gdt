package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
)

// Check renders one complete policy comparison report.
func Check(result application.CheckResult, options Options) (string, error) {
	if err := validateInspect(result.Inspect); err != nil {
		return "", err
	}
	if !result.Policy.Kind.Valid() {
		return "", fmt.Errorf("invalid policy kind %q", result.Policy.Kind)
	}
	if !result.Evaluation.Status.Valid() {
		return "", fmt.Errorf("invalid check status %q", result.Evaluation.Status)
	}
	comparisons, err := sortedComparisons(result.Evaluation.Results)
	if err != nil {
		return "", err
	}
	options = normalizedOptions(options)
	style := styler{enabled: options.Color}

	var output strings.Builder
	writeVersion(&output, options)
	fmt.Fprintf(&output, "%-13s%s\n", "Scene:", preferredScenePath(result.Inspect))
	fmt.Fprintf(
		&output,
		"%-13s%s\n",
		"Analysis:",
		style.status(strings.ToUpper(string(result.Inspect.Analysis.Status))),
	)
	if result.Inspect.Analysis.Reliability != analysis.ReliabilityExact {
		fmt.Fprintf(&output, "%-13s%s\n", "Accuracy:", accuracyLabel(result.Inspect.Analysis.Reliability))
	}
	writePolicyMetadata(&output, result.Policy)

	output.WriteString("\nMetric                        Actual     Budget   Result\n")
	output.WriteString("--------------------------------------------------------\n")
	for _, comparison := range comparisons {
		verdict := "PASS"
		if !comparison.Passed {
			verdict = "FAIL"
		}
		actual := formatMetric(comparison.Actual, result.Inspect.Analysis.Reliability)
		fmt.Fprintf(
			&output,
			"%-26s %10s %10s   %s",
			comparison.Metric.Label(),
			actual,
			formatInteger(comparison.Limit),
			style.status(verdict),
		)
		if !comparison.Passed {
			fmt.Fprintf(&output, "  +%s", formatInteger(comparison.Actual-comparison.Limit))
		}
		output.WriteByte('\n')
	}

	writeCheckSummary(&output, result, style)
	if result.Evaluation.Status != budget.StatusIncomplete {
		writeReliabilityWarning(&output, result.Inspect.Analysis, style)
	}
	if result.Policy.Kind == policy.KindPreset {
		output.WriteString("\nBuilt-in presets are heuristic guardrails, not performance guarantees.\n")
		output.WriteString("Profile the game on target hardware.\n")
	}

	return output.String(), nil
}

func writePolicyMetadata(output *strings.Builder, effective policy.Effective) {
	switch effective.Kind {
	case policy.KindPreset:
		fmt.Fprintf(output, "%-13s%s\n", "Preset:", effective.ID)
	case policy.KindProfile:
		fmt.Fprintf(output, "%-13s%s\n", "Profile:", effective.ID)
	default:
		fmt.Fprintf(output, "%-13s%s\n", "Policy:", "project/CLI overrides")
	}

	metadata := effective.Metadata
	if metadata.Status != "" {
		status := metadata.Status
		if metadata.Stability != "" {
			status += " (" + metadata.Stability + ")"
		}
		fmt.Fprintf(output, "%-13s%s\n", "Status:", status)
	}
	if metadata.Renderer != "" && metadata.Renderer != "unspecified" {
		fmt.Fprintf(output, "%-13s%s\n", "Renderer:", displayRenderer(metadata.Renderer))
	}
	if metadata.TargetFPS > 0 {
		fmt.Fprintf(output, "%-13s%s\n", "Target FPS:", formatInteger(metadata.TargetFPS))
	}
	if metadata.Quality != "" && metadata.Quality != "custom" {
		fmt.Fprintf(output, "%-13s%s\n", "Quality:", displayTitle(metadata.Quality))
	}
}

func writeCheckSummary(output *strings.Builder, result application.CheckResult, style styler) {
	output.WriteByte('\n')
	switch result.Evaluation.Status {
	case budget.StatusPassed:
		fmt.Fprintf(
			output,
			"%s — all %s %s within limits\n",
			style.status("PASSED"),
			formatInteger(int64(len(result.Evaluation.Results))),
			plural(int64(len(result.Evaluation.Results)), "budget", "budgets"),
		)
	case budget.StatusFailed:
		count := int64(result.Evaluation.Exceeded)
		fmt.Fprintf(
			output,
			"%s — %s %s exceeded\n",
			style.status("FAILED"),
			formatInteger(count),
			plural(count, "budget", "budgets"),
		)
	case budget.StatusIncomplete:
		count := result.Inspect.Analysis.Coverage.UnresolvedSceneInstances
		reason := "static analysis is partial"
		if count > 0 {
			reason = fmt.Sprintf(
				"%s %s could not be analyzed statically",
				formatInteger(count),
				plural(count, "scene instance", "scene instances"),
			)
		} else if result.Inspect.Analysis.Coverage.InheritedScenes > 0 {
			reason = "inherited-scene analysis is approximate"
		}
		fmt.Fprintf(output, "%s — %s\n", style.status("INCOMPLETE"), reason)
		output.WriteString("Configured policy fail_on_partial=true rejects partial analysis.\n")
	}
}

func sortedComparisons(items []budget.Result) ([]budget.Result, error) {
	order := make(map[metrics.Name]int, len(metrics.OrderedNames()))
	for index, name := range metrics.OrderedNames() {
		order[name] = index
	}

	result := append([]budget.Result(nil), items...)
	for _, item := range result {
		if !item.Metric.Valid() {
			return nil, fmt.Errorf("invalid comparison metric %q", item.Metric)
		}
		if item.Actual < 0 || item.Limit < 0 {
			return nil, fmt.Errorf("negative comparison for %q", item.Metric)
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		return order[result[left].Metric] < order[result[right].Metric]
	})

	return result, nil
}
