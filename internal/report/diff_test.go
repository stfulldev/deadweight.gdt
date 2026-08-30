package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

func TestDiffTextAndJSONGoldens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result application.DiffResult
	}{
		{name: "diff_empty", result: emptyDiffResult()},
		{name: "diff_changed", result: changedDiffResult()},
		{name: "diff_failed", result: failedDiffResult()},
		{name: "diff_incomplete", result: incompleteDiffResult()},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			textReport, err := Diff(test.result, Options{Version: "0.2.0-test", Color: true})
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}
			assertDiffGolden(t, filepath.Join("testdata", "golden", test.name+".golden"), textReport)
			jsonReport, err := DiffJSON(test.result, Options{Version: "0.2.0-test", Color: true})
			if err != nil {
				t.Fatalf("DiffJSON() error = %v", err)
			}
			assertJSONFraming(t, jsonReport)
			validateReportDocument(t, []byte(jsonReport))
			assertDiffGolden(t, filepath.Join("testdata", "golden", "json", test.name+".golden"), jsonReport)
			for _, rendered := range []string{textReport, jsonReport} {
				if strings.Contains(rendered, "before.json") || strings.Contains(rendered, "after.json") || strings.Contains(rendered, "\x1b[") {
					t.Fatalf("portable diff leaked input paths or ANSI: %q", rendered)
				}
			}
		})
	}
}

func TestDiffQualifiedDecreaseUsesUncertainLanguage(t *testing.T) {
	t.Parallel()

	result := emptyDiffResult()
	result.Comparison.Changed = true
	result.Comparison.AfterReliability = analysis.ReliabilityLowerBound
	result.Comparison.ReliabilityChange = &reportdiff.ReliabilityChange{Before: analysis.ReliabilityExact, After: analysis.ReliabilityLowerBound}
	result.Comparison.MetricChanges = []reportdiff.MetricChange{{
		Metric: metrics.Nodes, Before: 100, After: 90, Delta: -10,
		BeforeConfidence: reportdiff.Confidence{Reliability: analysis.ReliabilityExact, Source: reportdiff.ConfidenceMetric},
		AfterConfidence:  reportdiff.Confidence{Reliability: analysis.ReliabilityLowerBound, Reasons: []analysis.ConfidenceReason{analysis.ConfidenceUnavailableScene}, Source: reportdiff.ConfidenceMetric},
		Assessment:       reportdiff.AssessmentUncertain,
	}}
	rendered, err := Diff(result, Options{})
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if !strings.Contains(rendered, "UNCERTAIN (exact -> lower_bound)") || strings.Contains(rendered, "IMPROVEMENT") {
		t.Fatalf("qualified decrease output = %q", rendered)
	}
}

func assertDiffGolden(t *testing.T, path, got string) {
	t.Helper()
	if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Fatalf("diff framing is not exactly one trailing LF: %q", got)
	}
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("golden mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
	}
}

func emptyDiffResult() application.DiffResult {
	return application.DiffResult{Comparison: reportdiff.Result{
		Kind: reportdiff.KindInspect, Scene: "res://levels/city.tscn",
		BeforeReliability: analysis.ReliabilityExact, AfterReliability: analysis.ReliabilityExact,
		Enforcement: reportdiff.Enforcement{Status: budget.StatusPassed},
	}}
}

func changedDiffResult() application.DiffResult {
	result := application.DiffResult{Comparison: reportdiff.Result{
		Kind: reportdiff.KindCheck, Scene: "res://levels/city.tscn",
		BeforeReliability: analysis.ReliabilityExact, AfterReliability: analysis.ReliabilityLowerBound,
		Changed: true,
		MetricChanges: []reportdiff.MetricChange{
			{Metric: metrics.Nodes, Before: 100, After: 112, Delta: 12, BeforeConfidence: exactDiffConfidence(), AfterConfidence: lowerDiffConfidence(), Assessment: reportdiff.AssessmentRegression},
			{Metric: metrics.MeshInstances, Before: 50, After: 48, Delta: -2, BeforeConfidence: exactDiffConfidence(), AfterConfidence: lowerDiffConfidence(), Assessment: reportdiff.AssessmentUncertain},
		},
		ReliabilityChange: &reportdiff.ReliabilityChange{Before: analysis.ReliabilityExact, After: analysis.ReliabilityLowerBound},
		CoverageChanges:   []reportdiff.CoverageChange{{Field: "parsed_scene_files", Before: 2, After: 3, Delta: 1}},
		DiagnosticChanges: []reportdiff.DiagnosticChange{
			{Change: reportdiff.EvidenceAdded, Diagnostic: reportdiff.Diagnostic{Code: diagnostic.CodeUnavailableResource, Severity: diagnostic.SeverityWarning, Message: "resource unavailable", Occurrences: 1, SourcePath: "res://levels/city.tscn", Line: 4}, AfterOccurrences: 1, Delta: 1},
			{Change: reportdiff.EvidenceOccurrencesChanged, Diagnostic: reportdiff.Diagnostic{Code: diagnostic.CodeUnresolvedSceneInstance, Severity: diagnostic.SeverityWarning, Message: "scene unavailable", Occurrences: 3}, BeforeOccurrences: 1, AfterOccurrences: 3, Delta: 2},
		},
		DependencyChanges: []reportdiff.DependencyChange{{Change: reportdiff.EvidenceRemoved, Identity: "res://old.tscn"}, {Change: reportdiff.EvidenceAdded, Identity: "res://new.tscn"}},
		EvaluationChange: &reportdiff.EvaluationChange{
			Before: reportdiff.Evaluation{Verdict: budget.StatusPassed, Comparisons: []budget.Result{{Metric: metrics.Nodes, Actual: 100, Limit: 100, Delta: 0, Passed: true}}},
			After:  reportdiff.Evaluation{Verdict: budget.StatusFailed, Exceeded: 1, Comparisons: []budget.Result{{Metric: metrics.Nodes, Actual: 112, Limit: 100, Delta: 12, Passed: false}}},
		},
		Enforcement: reportdiff.Enforcement{Status: budget.StatusPassed},
	}}
	return result
}

func failedDiffResult() application.DiffResult {
	result := emptyDiffResult()
	result.Comparison.Changed = true
	result.Comparison.MetricChanges = []reportdiff.MetricChange{{Metric: metrics.Nodes, Before: 10, After: 12, Delta: 2, BeforeConfidence: exactDiffConfidence(), AfterConfidence: exactDiffConfidence(), Assessment: reportdiff.AssessmentRegression}}
	result.Comparison.Enforcement = reportdiff.Enforcement{Enabled: true, Status: budget.StatusFailed, Triggers: []reportdiff.Trigger{{Kind: reportdiff.TriggerMetricIncrease, Metric: metrics.Nodes, Assessment: reportdiff.AssessmentRegression, BeforeReliability: analysis.ReliabilityExact, AfterReliability: analysis.ReliabilityExact}}}
	return result
}

func incompleteDiffResult() application.DiffResult {
	result := emptyDiffResult()
	result.Comparison.Changed = true
	result.Comparison.AfterReliability = analysis.ReliabilityLowerBound
	result.Comparison.ReliabilityChange = &reportdiff.ReliabilityChange{Before: analysis.ReliabilityExact, After: analysis.ReliabilityLowerBound}
	result.Comparison.Enforcement = reportdiff.Enforcement{Enabled: true, Status: budget.StatusIncomplete, Triggers: []reportdiff.Trigger{{Kind: reportdiff.TriggerReliabilityDegradation, BeforeReliability: analysis.ReliabilityExact, AfterReliability: analysis.ReliabilityLowerBound}}}
	return result
}

func exactDiffConfidence() reportdiff.Confidence {
	return reportdiff.Confidence{Reliability: analysis.ReliabilityExact, Source: reportdiff.ConfidenceMetric}
}

func lowerDiffConfidence() reportdiff.Confidence {
	return reportdiff.Confidence{Reliability: analysis.ReliabilityLowerBound, Reasons: []analysis.ConfidenceReason{analysis.ConfidenceUnavailableScene}, Source: reportdiff.ConfidenceMetric}
}
