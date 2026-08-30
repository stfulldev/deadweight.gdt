package reportdiff

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

// Compare validates compatibility and returns an owned semantic diff.
func Compare(before, after Snapshot, policy Policy) (Result, error) {
	normalizedPolicy, err := NormalizePolicy(policy)
	if err != nil {
		return Result{}, err
	}
	if before.Kind != after.Kind {
		return Result{}, fmt.Errorf("incompatible report kinds %q and %q", before.Kind, after.Kind)
	}
	if before.Scene != after.Scene {
		return Result{}, fmt.Errorf("incompatible root scenes %q and %q", before.Scene, after.Scene)
	}
	result := Result{
		Kind: before.Kind, Scene: before.Scene,
		BeforeReliability: before.Reliability, AfterReliability: after.Reliability,
	}
	result.MetricChanges = compareMetrics(before.Metrics, after.Metrics)
	if before.Reliability != after.Reliability {
		result.ReliabilityChange = &ReliabilityChange{Before: before.Reliability, After: after.Reliability}
	}
	result.CoverageChanges = compareCoverage(before.Coverage, after.Coverage)
	result.DiagnosticChanges = compareDiagnostics(before.Diagnostics, after.Diagnostics)
	result.DependencyChanges = compareDependencies(before.Dependencies, after.Dependencies)
	if !reflect.DeepEqual(before.Evaluation, after.Evaluation) {
		if before.Evaluation == nil || after.Evaluation == nil {
			return Result{}, fmt.Errorf("compatible check evaluation is missing")
		}
		result.EvaluationChange = &EvaluationChange{Before: cloneEvaluation(*before.Evaluation), After: cloneEvaluation(*after.Evaluation)}
	}
	result.Changed = len(result.MetricChanges) > 0 || result.ReliabilityChange != nil || len(result.CoverageChanges) > 0 || len(result.DiagnosticChanges) > 0 || len(result.DependencyChanges) > 0 || result.EvaluationChange != nil
	result.Enforcement = evaluatePolicy(result, normalizedPolicy)
	return cloneResult(result), nil
}

func compareMetrics(before, after []MetricSnapshot) []MetricChange {
	result := make([]MetricChange, 0)
	for index := range before {
		left, right := before[index], after[index]
		if left.Value == right.Value && reflect.DeepEqual(left.Confidence, right.Confidence) {
			continue
		}
		result = append(result, MetricChange{
			Metric: left.Metric, Before: left.Value, After: right.Value, Delta: right.Value - left.Value,
			BeforeConfidence: cloneConfidence(left.Confidence), AfterConfidence: cloneConfidence(right.Confidence),
			Assessment: assessMetric(left.Value, right.Value, left.Confidence.Reliability, right.Confidence.Reliability),
		})
	}
	return result
}

func assessMetric(before, after int64, beforeReliability, afterReliability analysis.Reliability) Assessment {
	if after > before && beforeReliability == analysis.ReliabilityExact && (afterReliability == analysis.ReliabilityExact || afterReliability == analysis.ReliabilityLowerBound) {
		return AssessmentRegression
	}
	if after < before && afterReliability == analysis.ReliabilityExact && (beforeReliability == analysis.ReliabilityExact || beforeReliability == analysis.ReliabilityLowerBound) {
		return AssessmentImprovement
	}
	return AssessmentUncertain
}

func compareCoverage(before, after analysis.Coverage) []CoverageChange {
	fields := []struct {
		name          string
		before, after int64
	}{
		{"parsed_scene_files", before.ParsedSceneFiles, after.ParsedSceneFiles},
		{"resolved_scene_instances", before.ResolvedSceneInstances, after.ResolvedSceneInstances},
		{"unresolved_scene_instances", before.UnresolvedSceneInstances, after.UnresolvedSceneInstances},
		{"inherited_scenes", before.InheritedScenes, after.InheritedScenes},
	}
	result := make([]CoverageChange, 0)
	for _, field := range fields {
		if field.before != field.after {
			result = append(result, CoverageChange{Field: field.name, Before: field.before, After: field.after, Delta: field.after - field.before})
		}
	}
	return result
}

func compareDiagnostics(before, after []Diagnostic) []DiagnosticChange {
	left := make(map[string]Diagnostic, len(before))
	right := make(map[string]Diagnostic, len(after))
	keys := make(map[string]struct{})
	for _, item := range before {
		key := diagnosticKey(item)
		left[key] = item
		keys[key] = struct{}{}
	}
	for _, item := range after {
		key := diagnosticKey(item)
		right[key] = item
		keys[key] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	result := make([]DiagnosticChange, 0)
	for _, key := range ordered {
		beforeItem, beforeOK := left[key]
		afterItem, afterOK := right[key]
		switch {
		case !beforeOK:
			result = append(result, DiagnosticChange{Change: EvidenceAdded, Diagnostic: afterItem, AfterOccurrences: afterItem.Occurrences, Delta: afterItem.Occurrences})
		case !afterOK:
			result = append(result, DiagnosticChange{Change: EvidenceRemoved, Diagnostic: beforeItem, BeforeOccurrences: beforeItem.Occurrences, Delta: -beforeItem.Occurrences})
		case beforeItem.Occurrences != afterItem.Occurrences:
			item := afterItem
			item.Occurrences = afterItem.Occurrences
			result = append(result, DiagnosticChange{Change: EvidenceOccurrencesChanged, Diagnostic: item, BeforeOccurrences: beforeItem.Occurrences, AfterOccurrences: afterItem.Occurrences, Delta: afterItem.Occurrences - beforeItem.Occurrences})
		}
	}
	return result
}

func compareDependencies(before, after []string) []DependencyChange {
	left := make(map[string]struct{}, len(before))
	right := make(map[string]struct{}, len(after))
	keys := make(map[string]struct{})
	for _, item := range before {
		left[item] = struct{}{}
		keys[item] = struct{}{}
	}
	for _, item := range after {
		right[item] = struct{}{}
		keys[item] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	result := make([]DependencyChange, 0)
	for _, key := range ordered {
		_, inBefore := left[key]
		_, inAfter := right[key]
		if inBefore && !inAfter {
			result = append(result, DependencyChange{Change: EvidenceRemoved, Identity: key})
		}
		if !inBefore && inAfter {
			result = append(result, DependencyChange{Change: EvidenceAdded, Identity: key})
		}
	}
	return result
}

func evaluatePolicy(result Result, policy Policy) Enforcement {
	enforcement := Enforcement{Enabled: len(policy.MetricIncreases) > 0 || policy.FailOnReliability, Status: budget.StatusPassed}
	selected := make(map[metrics.Name]struct{}, len(policy.MetricIncreases))
	for _, name := range policy.MetricIncreases {
		selected[name] = struct{}{}
	}
	for _, change := range result.MetricChanges {
		if change.Delta <= 0 {
			continue
		}
		if _, ok := selected[change.Metric]; !ok {
			continue
		}
		trigger := Trigger{Kind: TriggerMetricIncrease, Metric: change.Metric, Assessment: change.Assessment, BeforeReliability: change.BeforeConfidence.Reliability, AfterReliability: change.AfterConfidence.Reliability}
		enforcement.Triggers = append(enforcement.Triggers, trigger)
		if change.Assessment == AssessmentRegression && enforcement.Status != budget.StatusIncomplete {
			enforcement.Status = budget.StatusFailed
		} else if change.Assessment != AssessmentRegression {
			enforcement.Status = budget.StatusIncomplete
		}
	}
	if policy.FailOnReliability && reliabilityRank(result.AfterReliability) > reliabilityRank(result.BeforeReliability) {
		enforcement.Triggers = append(enforcement.Triggers, Trigger{Kind: TriggerReliabilityDegradation, BeforeReliability: result.BeforeReliability, AfterReliability: result.AfterReliability})
		enforcement.Status = budget.StatusIncomplete
	}
	return enforcement
}

func cloneConfidence(value Confidence) Confidence {
	value.Reasons = append([]analysis.ConfidenceReason(nil), value.Reasons...)
	return value
}
func cloneEvaluation(value Evaluation) Evaluation {
	value.Comparisons = append([]budget.Result(nil), value.Comparisons...)
	return value
}
func cloneResult(value Result) Result {
	value.MetricChanges = append([]MetricChange(nil), value.MetricChanges...)
	for index := range value.MetricChanges {
		value.MetricChanges[index].BeforeConfidence = cloneConfidence(value.MetricChanges[index].BeforeConfidence)
		value.MetricChanges[index].AfterConfidence = cloneConfidence(value.MetricChanges[index].AfterConfidence)
	}
	value.CoverageChanges = append([]CoverageChange(nil), value.CoverageChanges...)
	value.DiagnosticChanges = append([]DiagnosticChange(nil), value.DiagnosticChanges...)
	value.DependencyChanges = append([]DependencyChange(nil), value.DependencyChanges...)
	value.Enforcement.Triggers = append([]Trigger(nil), value.Enforcement.Triggers...)
	if value.ReliabilityChange != nil {
		cloned := *value.ReliabilityChange
		value.ReliabilityChange = &cloned
	}
	if value.EvaluationChange != nil {
		cloned := &EvaluationChange{Before: cloneEvaluation(value.EvaluationChange.Before), After: cloneEvaluation(value.EvaluationChange.After)}
		value.EvaluationChange = cloned
	}
	return value
}
