package report

import (
	"fmt"
	"strings"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

// Diff renders one deterministic human-readable semantic report comparison.
func Diff(result application.DiffResult, options Options) (string, error) {
	comparison := result.Comparison
	if !comparison.Kind.Valid() || comparison.Scene == "" || !comparison.Enforcement.Status.Valid() {
		return "", fmt.Errorf("invalid diff result")
	}
	options = normalizedOptions(options)
	var output strings.Builder
	writeVersion(&output, options)
	fmt.Fprintf(&output, "%-14s%s\n", "Report kind:", comparison.Kind)
	fmt.Fprintf(&output, "%-14s%s\n", "Scene:", comparison.Scene)
	fmt.Fprintf(&output, "%-14s%s -> %s\n", "Reliability:", comparison.BeforeReliability, comparison.AfterReliability)

	if !comparison.Changed {
		output.WriteString("\nNo semantic changes.\n")
	} else {
		writeDiffMetrics(&output, comparison)
		writeDiffCoverage(&output, comparison)
		writeDiffDiagnostics(&output, comparison)
		writeDiffDependencies(&output, comparison)
		writeDiffEvaluation(&output, comparison)
	}
	writeDiffEnforcement(&output, comparison)
	return output.String(), nil
}

func writeDiffMetrics(output *strings.Builder, result reportdiff.Result) {
	if len(result.MetricChanges) == 0 {
		return
	}
	output.WriteString("\nMetrics\n")
	output.WriteString("  Metric                    Before      After      Delta   Assessment\n")
	for _, change := range result.MetricChanges {
		fmt.Fprintf(output, "  %-24s %10s %10s %10s   %s (%s -> %s)", change.Metric, formatInteger(change.Before), formatInteger(change.After), signedInteger(change.Delta), strings.ToUpper(string(change.Assessment)), change.BeforeConfidence.Reliability, change.AfterConfidence.Reliability)
		if change.BeforeConfidence.Source == reportdiff.ConfidenceReportSummary || change.AfterConfidence.Source == reportdiff.ConfidenceReportSummary {
			fmt.Fprintf(output, " [%s -> %s]", change.BeforeConfidence.Source, change.AfterConfidence.Source)
		}
		output.WriteByte('\n')
	}
}

func writeDiffCoverage(output *strings.Builder, result reportdiff.Result) {
	if len(result.CoverageChanges) == 0 {
		return
	}
	output.WriteString("\nCoverage\n")
	for _, change := range result.CoverageChanges {
		fmt.Fprintf(output, "  %-28s %s -> %s (%s)\n", change.Field, formatInteger(change.Before), formatInteger(change.After), signedInteger(change.Delta))
	}
}

func writeDiffDiagnostics(output *strings.Builder, result reportdiff.Result) {
	if len(result.DiagnosticChanges) == 0 {
		return
	}
	output.WriteString("\nDiagnostics\n")
	for _, change := range result.DiagnosticChanges {
		fmt.Fprintf(output, "  %-19s %s %s: %s", strings.ToUpper(string(change.Change)), change.Diagnostic.Severity, change.Diagnostic.Code, change.Diagnostic.Message)
		if change.Change == reportdiff.EvidenceOccurrencesChanged {
			fmt.Fprintf(output, " (%s -> %s, %s)", formatInteger(change.BeforeOccurrences), formatInteger(change.AfterOccurrences), signedInteger(change.Delta))
		}
		output.WriteByte('\n')
	}
}

func writeDiffDependencies(output *strings.Builder, result reportdiff.Result) {
	if len(result.DependencyChanges) == 0 {
		return
	}
	output.WriteString("\nScene dependencies\n")
	for _, change := range result.DependencyChanges {
		fmt.Fprintf(output, "  %-7s %s\n", strings.ToUpper(string(change.Change)), change.Identity)
	}
}

func writeDiffEvaluation(output *strings.Builder, result reportdiff.Result) {
	if result.EvaluationChange == nil {
		return
	}
	output.WriteString("\nBudget evaluation\n")
	fmt.Fprintf(output, "  Verdict:  %s -> %s\n", result.EvaluationChange.Before.Verdict, result.EvaluationChange.After.Verdict)
	fmt.Fprintf(output, "  Exceeded: %s -> %s\n", formatInteger(result.EvaluationChange.Before.Exceeded), formatInteger(result.EvaluationChange.After.Exceeded))
	output.WriteString("  Candidate comparisons:\n")
	for _, item := range result.EvaluationChange.After.Comparisons {
		fmt.Fprintf(output, "    %-22s observed %s, limit %s, passed %t\n", item.Metric, formatInteger(item.Actual), formatInteger(item.Limit), item.Passed)
	}
}

func writeDiffEnforcement(output *strings.Builder, result reportdiff.Result) {
	output.WriteString("\nEnforcement\n")
	if !result.Enforcement.Enabled {
		output.WriteString("  DISABLED — comparison only\n")
		return
	}
	fmt.Fprintf(output, "  %s", result.Enforcement.Status)
	if len(result.Enforcement.Triggers) == 0 {
		output.WriteString(" — no selected regressions\n")
		return
	}
	output.WriteByte('\n')
	for _, trigger := range result.Enforcement.Triggers {
		if trigger.Kind == reportdiff.TriggerMetricIncrease {
			fmt.Fprintf(output, "  - %s increased: %s (%s -> %s)\n", trigger.Metric, strings.ToUpper(string(trigger.Assessment)), trigger.BeforeReliability, trigger.AfterReliability)
		} else {
			fmt.Fprintf(output, "  - reliability degraded: %s -> %s\n", trigger.BeforeReliability, trigger.AfterReliability)
		}
	}
}

func signedInteger(value int64) string {
	if value > 0 {
		return "+" + formatInteger(value)
	}
	return formatInteger(value)
}
