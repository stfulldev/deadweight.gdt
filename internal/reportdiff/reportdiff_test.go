package reportdiff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestDecodeAcceptsCurrentReportsAndLegacyMetricConfidence(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"inspect_complete.golden", "tree_complete.golden", "check_passed.golden"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			contents, err := os.ReadFile(filepath.Join("..", "report", "testdata", "golden", "json", name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			if _, err := Decode(contents); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
		})
	}

	legacy := baselineDocument(KindInspect, analysis.ReliabilityLowerBound, true)
	legacy["future_optional_field"] = map[string]any{"ignored": true}
	snapshot := decodeDocument(t, legacy)
	for _, item := range snapshot.Metrics {
		if item.Confidence.Reliability != analysis.ReliabilityLowerBound || item.Confidence.Source != ConfidenceReportSummary {
			t.Fatalf("legacy confidence = %#v", item.Confidence)
		}
	}
}

func TestDecodeRejectsInvalidAndIncompatibleEvidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		bytes  func(t *testing.T) []byte
		mutate func(map[string]any)
		want   string
	}{
		{name: "malformed", bytes: func(*testing.T) []byte { return []byte("{") }, want: "decode report JSON"},
		{name: "invalid UTF-8", bytes: func(*testing.T) []byte { return []byte{0xff} }, want: "UTF-8"},
		{name: "trailing", bytes: func(t *testing.T) []byte {
			return append(marshalDocument(t, baselineDocument(KindInspect, analysis.ReliabilityExact, false)), []byte("{}")...)
		}, want: "trailing"},
		{name: "unsupported schema", mutate: func(document map[string]any) { document["schema_version"] = 2 }, want: "unsupported"},
		{name: "unsupported kind", mutate: func(document map[string]any) { document["kind"] = "future" }, want: "unsupported report kind"},
		{name: "wrong tool", mutate: func(document map[string]any) { document["tool"].(map[string]any)["name"] = "other" }, want: "tool identity"},
		{name: "non-portable scene", mutate: func(document map[string]any) { document["scene"].(map[string]any)["path"] = `/tmp/root.tscn` }, want: "portable"},
		{name: "missing metric", mutate: func(document map[string]any) { analysisMap(document)["metrics"] = metricMaps(document)[:7] }, want: "exactly 8"},
		{name: "wrong metric order", mutate: func(document map[string]any) { metricMaps(document)[0].(map[string]any)["id"] = "lights" }, want: "position 0"},
		{name: "negative coverage", mutate: func(document map[string]any) {
			analysisMap(document)["coverage"].(map[string]any)["parsed_scene_files"] = -1
		}, want: "non-negative"},
		{name: "mixed payload", mutate: func(document map[string]any) { document["evaluation"] = map[string]any{} }, want: "incompatible payload"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var contents []byte
			if test.bytes != nil {
				contents = test.bytes(t)
			} else {
				document := baselineDocument(KindInspect, analysis.ReliabilityExact, false)
				test.mutate(document)
				contents = marshalDocument(t, document)
			}
			if _, err := Decode(contents); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Decode() error = %v, want %q", err, test.want)
			}
		})
	}

	oversized := make([]byte, MaxInputBytes+1)
	if _, err := Decode(oversized); err == nil || !strings.Contains(err.Error(), "input limit") {
		t.Fatalf("oversized Decode() error = %v", err)
	}
}

func TestCompareCoversSemanticEvidenceAndIgnoresEnvelopeMetadata(t *testing.T) {
	t.Parallel()

	beforeDocument := baselineDocument(KindInspect, analysis.ReliabilityExact, false)
	beforeDocument["tool"].(map[string]any)["version"] = "before"
	beforeAnalysis := analysisMap(beforeDocument)
	beforeAnalysis["diagnostics"] = []any{diagnosticDocument(1, "old")}
	beforeAnalysis["unique_evidence"] = []any{dependencyDocument("res://old.tscn")}

	afterDocument := baselineDocument(KindInspect, analysis.ReliabilityExact, false)
	afterDocument["tool"].(map[string]any)["version"] = "after"
	setMetricValue(afterDocument, metrics.Nodes, 12)
	afterAnalysis := analysisMap(afterDocument)
	afterAnalysis["coverage"].(map[string]any)["parsed_scene_files"] = int64(2)
	afterAnalysis["diagnostics"] = []any{diagnosticDocument(3, "old"), diagnosticDocument(1, "new")}
	afterAnalysis["unique_evidence"] = []any{dependencyDocument("res://new.tscn")}

	before := decodeDocument(t, beforeDocument)
	after := decodeDocument(t, afterDocument)
	result, err := Compare(before, after, Policy{})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if !result.Changed || len(result.MetricChanges) != 1 || result.MetricChanges[0].Delta != 2 || result.MetricChanges[0].Assessment != AssessmentRegression {
		t.Fatalf("metric changes = %#v", result.MetricChanges)
	}
	if len(result.CoverageChanges) != 1 || result.CoverageChanges[0].Delta != 1 {
		t.Fatalf("coverage changes = %#v", result.CoverageChanges)
	}
	if len(result.DiagnosticChanges) != 2 || len(result.DependencyChanges) != 2 {
		t.Fatalf("diagnostic/dependency changes = %#v / %#v", result.DiagnosticChanges, result.DependencyChanges)
	}
	if result.Enforcement.Enabled || result.Enforcement.Status != budget.StatusPassed {
		t.Fatalf("default enforcement = %#v", result.Enforcement)
	}

	equal, err := Compare(before, before, Policy{})
	if err != nil || equal.Changed {
		t.Fatalf("equal Compare() = %#v / %v", equal, err)
	}
	result.DiagnosticChanges[0].Diagnostic.Message = "mutated"
	result.DependencyChanges[0].Identity = "res://mutated.tscn"
	repeated, err := Compare(before, after, Policy{})
	if err != nil || repeated.DiagnosticChanges[0].Diagnostic.Message == "mutated" || repeated.DependencyChanges[0].Identity == "res://mutated.tscn" {
		t.Fatalf("comparison result aliases inputs: %#v / %v", repeated, err)
	}
}

func TestCompareConfidenceAssessmentAndEnforcement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		beforeValue       int64
		afterValue        int64
		beforeReliability analysis.Reliability
		afterReliability  analysis.Reliability
		wantAssessment    Assessment
		wantStatus        budget.Status
	}{
		{name: "exact increase", beforeValue: 100, afterValue: 110, beforeReliability: analysis.ReliabilityExact, afterReliability: analysis.ReliabilityExact, wantAssessment: AssessmentRegression, wantStatus: budget.StatusFailed},
		{name: "larger lower bound", beforeValue: 100, afterValue: 110, beforeReliability: analysis.ReliabilityExact, afterReliability: analysis.ReliabilityLowerBound, wantAssessment: AssessmentRegression, wantStatus: budget.StatusFailed},
		{name: "smaller partial candidate", beforeValue: 100, afterValue: 90, beforeReliability: analysis.ReliabilityExact, afterReliability: analysis.ReliabilityLowerBound, wantAssessment: AssessmentUncertain, wantStatus: budget.StatusPassed},
		{name: "uncertain increase", beforeValue: 100, afterValue: 110, beforeReliability: analysis.ReliabilityLowerBound, afterReliability: analysis.ReliabilityLowerBound, wantAssessment: AssessmentUncertain, wantStatus: budget.StatusIncomplete},
		{name: "proven improvement", beforeValue: 100, afterValue: 90, beforeReliability: analysis.ReliabilityLowerBound, afterReliability: analysis.ReliabilityExact, wantAssessment: AssessmentImprovement, wantStatus: budget.StatusPassed},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			beforeDocument := baselineDocument(KindInspect, test.beforeReliability, false)
			afterDocument := baselineDocument(KindInspect, test.afterReliability, false)
			setMetricValue(beforeDocument, metrics.Nodes, test.beforeValue)
			setMetricValue(afterDocument, metrics.Nodes, test.afterValue)
			result, err := Compare(decodeDocument(t, beforeDocument), decodeDocument(t, afterDocument), Policy{MetricIncreases: []metrics.Name{metrics.Nodes}})
			if err != nil {
				t.Fatalf("Compare() error = %v", err)
			}
			if result.MetricChanges[0].Assessment != test.wantAssessment || result.Enforcement.Status != test.wantStatus {
				t.Fatalf("assessment/status = %s/%s, want %s/%s", result.MetricChanges[0].Assessment, result.Enforcement.Status, test.wantAssessment, test.wantStatus)
			}
		})
	}

	before := baselineDocument(KindInspect, analysis.ReliabilityExact, false)
	after := baselineDocument(KindInspect, analysis.ReliabilityLowerBound, false)
	setMetricExact(after, metrics.Nodes)
	setMetricValue(after, metrics.Nodes, 11)
	result, err := Compare(decodeDocument(t, before), decodeDocument(t, after), Policy{MetricIncreases: []metrics.Name{metrics.Nodes}, FailOnReliability: true})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if result.Enforcement.Status != budget.StatusIncomplete || len(result.Enforcement.Triggers) != 2 {
		t.Fatalf("reliability priority = %#v", result.Enforcement)
	}
}

func TestCompareRejectsIncompatibleReportsAndInvalidPolicy(t *testing.T) {
	t.Parallel()

	before := decodeDocument(t, baselineDocument(KindInspect, analysis.ReliabilityExact, false))
	afterDocument := baselineDocument(KindInspect, analysis.ReliabilityExact, false)
	afterDocument["scene"].(map[string]any)["path"] = "res://other.tscn"
	after := decodeDocument(t, afterDocument)
	if _, err := Compare(before, after, Policy{}); err == nil || !strings.Contains(err.Error(), "root scenes") {
		t.Fatalf("scene mismatch error = %v", err)
	}
	if _, err := NormalizePolicy(Policy{MetricIncreases: []metrics.Name{metrics.Nodes, metrics.Nodes}}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate policy error = %v", err)
	}
	if _, err := NormalizePolicy(Policy{MetricIncreases: []metrics.Name{"future"}}); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown policy error = %v", err)
	}
}

func TestCompareRetainsChangedCheckEvaluation(t *testing.T) {
	t.Parallel()

	beforeDocument := baselineDocument(KindCheck, analysis.ReliabilityExact, false)
	afterDocument := baselineDocument(KindCheck, analysis.ReliabilityExact, false)
	setMetricValue(afterDocument, metrics.Nodes, 12)
	afterDocument["evaluation"] = map[string]any{
		"comparisons": []any{map[string]any{"metric": "nodes", "observed": int64(12), "limit": int64(10), "delta": int64(2), "passed": false}},
		"exceeded":    int64(1), "verdict": string(budget.StatusFailed),
	}
	result, err := Compare(decodeDocument(t, beforeDocument), decodeDocument(t, afterDocument), Policy{})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if result.EvaluationChange == nil || result.EvaluationChange.Before.Verdict != budget.StatusPassed || result.EvaluationChange.After.Verdict != budget.StatusFailed || result.EvaluationChange.After.Comparisons[0].Limit != 10 {
		t.Fatalf("evaluation change = %#v", result.EvaluationChange)
	}
}

func baselineDocument(kind Kind, reliability analysis.Reliability, legacy bool) map[string]any {
	status := string(analysis.AnalysisComplete)
	if reliability != analysis.ReliabilityExact {
		status = string(analysis.AnalysisPartial)
	}
	metricDocuments := make([]any, 0, len(metrics.OrderedNames()))
	for index, name := range metrics.OrderedNames() {
		metric := map[string]any{"id": string(name), "value": int64(index + 10)}
		if !legacy {
			reasons := []any{}
			if reliability != analysis.ReliabilityExact {
				reasons = []any{string(analysis.ConfidenceUnavailableScene)}
			}
			metric["confidence"] = map[string]any{"reliability": string(reliability), "reasons": reasons}
		}
		metricDocuments = append(metricDocuments, metric)
	}
	document := map[string]any{
		"schema_version": 1,
		"kind":           string(kind),
		"tool":           map[string]any{"name": "deadweight.gdt", "version": "test"},
		"scene":          map[string]any{"path": "res://root.tscn"},
		"configuration":  map[string]any{"present": false, "selection": "absent"},
		"analysis": map[string]any{
			"status": status, "reliability": string(reliability), "metrics": metricDocuments,
			"coverage":    map[string]any{"parsed_scene_files": int64(1), "resolved_scene_instances": int64(0), "unresolved_scene_instances": int64(0), "inherited_scenes": int64(0)},
			"diagnostics": []any{}, "contributions": []any{map[string]any{}}, "unique_evidence": []any{},
		},
	}
	if kind == KindTree {
		document["dependency_tree"] = map[string]any{"root": "res://root.tscn", "entries": []any{}}
	}
	if kind == KindCheck {
		document["policy"] = map[string]any{"kind": "overrides", "metadata": map[string]any{}, "fail_on_partial": false}
		document["evaluation"] = map[string]any{
			"comparisons": []any{map[string]any{"metric": "nodes", "observed": int64(10), "limit": int64(10), "delta": int64(0), "passed": true}},
			"exceeded":    int64(0), "verdict": string(budget.StatusPassed),
		}
	}
	return document
}

func analysisMap(document map[string]any) map[string]any {
	return document["analysis"].(map[string]any)
}
func metricMaps(document map[string]any) []any { return analysisMap(document)["metrics"].([]any) }

func setMetricValue(document map[string]any, name metrics.Name, value int64) {
	for _, raw := range metricMaps(document) {
		metric := raw.(map[string]any)
		if metric["id"] == string(name) {
			metric["value"] = value
			return
		}
	}
}

func setMetricExact(document map[string]any, name metrics.Name) {
	for _, raw := range metricMaps(document) {
		metric := raw.(map[string]any)
		if metric["id"] == string(name) {
			metric["confidence"] = map[string]any{"reliability": "exact", "reasons": []any{}}
			return
		}
	}
}

func diagnosticDocument(occurrences int64, message string) map[string]any {
	return map[string]any{"code": "SB1004", "severity": "warning", "message": message, "occurrences": occurrences, "source": map[string]any{"path": "res://root.tscn", "line": int64(1)}}
}

func dependencyDocument(identity string) map[string]any {
	return map[string]any{"metric": "scene_dependencies", "identity": identity, "referrers": []any{map[string]any{"scene": "res://root.tscn", "edge_kind": "instance", "occurrences": int64(1)}}}
}

func decodeDocument(t *testing.T, document map[string]any) Snapshot {
	t.Helper()
	result, err := Decode(marshalDocument(t, document))
	if err != nil {
		t.Fatalf("Decode() error = %v\ndocument: %#v", err, document)
	}
	return result
}

func marshalDocument(t *testing.T, document map[string]any) []byte {
	t.Helper()
	contents, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal document: %v", err)
	}
	return contents
}

func TestCompareDoesNotMutateSnapshots(t *testing.T) {
	t.Parallel()
	before := decodeDocument(t, baselineDocument(KindInspect, analysis.ReliabilityExact, false))
	after := decodeDocument(t, baselineDocument(KindInspect, analysis.ReliabilityExact, false))
	wantBefore, wantAfter := before, after
	if _, err := Compare(before, after, Policy{}); err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if !reflect.DeepEqual(before, wantBefore) || !reflect.DeepEqual(after, wantAfter) {
		t.Fatal("Compare() mutated an input snapshot")
	}
}
