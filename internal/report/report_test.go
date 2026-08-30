package report

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestReportGoldens(t *testing.T) {
	t.Parallel()

	catalog, err := preset.Builtins()
	if err != nil {
		t.Fatalf("Builtins() error = %v", err)
	}
	steamDeck, err := catalog.Find("steam-deck")
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}

	tests := []struct {
		name   string
		render func() (string, error)
	}{
		{
			name: "inspect_complete",
			render: func() (string, error) {
				return Inspect(completeInspect(), Options{Version: "0.1.0"})
			},
		},
		{
			name: "inspect_lower_bound",
			render: func() (string, error) {
				return Inspect(lowerBoundInspect(), Options{Version: "0.1.0"})
			},
		},
		{
			name: "inspect_approximate",
			render: func() (string, error) {
				return Inspect(approximateInspect(), Options{Version: "0.1.0"})
			},
		},
		{
			name: "inspect_top_exact",
			render: func() (string, error) {
				return Inspect(completeInspect(), Options{
					Version:       "0.1.0",
					Contributions: ContributionSelection{Metric: metrics.Nodes, Limit: 2},
				})
			},
		},
		{
			name: "inspect_top_lower_bound",
			render: func() (string, error) {
				return Inspect(lowerBoundInspect(), Options{
					Version:       "0.1.0",
					Contributions: ContributionSelection{Metric: metrics.Nodes, Limit: 2},
				})
			},
		},
		{
			name: "inspect_top_approximate",
			render: func() (string, error) {
				return Inspect(approximateInspect(), Options{
					Version:       "0.1.0",
					Contributions: ContributionSelection{Metric: metrics.Nodes, Limit: 2},
				})
			},
		},
		{
			name: "check_passed",
			render: func() (string, error) {
				return Check(presetCheck(budget.StatusPassed, false), Options{Version: "0.1.0"})
			},
		},
		{
			name: "check_failed",
			render: func() (string, error) {
				return Check(presetCheck(budget.StatusFailed, false), Options{Version: "0.1.0"})
			},
		},
		{
			name: "check_incomplete",
			render: func() (string, error) {
				return Check(presetCheck(budget.StatusIncomplete, true), Options{Version: "0.1.0"})
			},
		},
		{
			name: "presets_list",
			render: func() (string, error) {
				return PresetList(application.PresetListResult{Catalog: catalog}, Options{})
			},
		},
		{
			name: "presets_show",
			render: func() (string, error) {
				return PresetShow(application.PresetShowResult{Preset: steamDeck}, Options{})
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, renderErr := test.render()
			if renderErr != nil {
				t.Fatalf("render error = %v", renderErr)
			}
			goldenPath := filepath.Join("testdata", "golden", test.name+".golden")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if writeErr := os.WriteFile(goldenPath, []byte(got), 0o600); writeErr != nil {
					t.Fatalf("update golden: %v", writeErr)
				}
			}
			want, readErr := os.ReadFile(goldenPath)
			if readErr != nil {
				t.Fatalf("read golden: %v", readErr)
			}
			if got != string(want) {
				t.Fatalf("golden mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
			}
			if strings.Contains(got, "\x1b[") {
				t.Fatalf("plain golden contains ANSI: %q", got)
			}
		})
	}
}

func TestRenderingDoesNotMutateEvidenceOrComparisons(t *testing.T) {
	t.Parallel()

	inspect := lowerBoundInspect()
	wantUnresolved := append([]analysis.UnresolvedInstance(nil), inspect.Analysis.Summary.Unresolved...)
	wantDiagnostics := append([]diagnostic.Diagnostic(nil), inspect.Analysis.Diagnostics...)
	if _, err := Inspect(inspect, Options{}); err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if !reflect.DeepEqual(inspect.Analysis.Summary.Unresolved, wantUnresolved) ||
		!reflect.DeepEqual(inspect.Analysis.Diagnostics, wantDiagnostics) {
		t.Fatal("Inspect() mutated evidence slices")
	}

	check := presetCheck(budget.StatusFailed, false)
	wantComparisons := append([]budget.Result(nil), check.Evaluation.Results...)
	if _, err := Check(check, Options{}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !reflect.DeepEqual(check.Evaluation.Results, wantComparisons) {
		t.Fatal("Check() mutated comparison order")
	}
}

func TestColorChangesOnlyExplicitStatusTokens(t *testing.T) {
	t.Parallel()

	plain, err := Inspect(lowerBoundInspect(), Options{Version: "test"})
	if err != nil {
		t.Fatalf("plain Inspect() error = %v", err)
	}
	colored, err := Inspect(lowerBoundInspect(), Options{Version: "test", Color: true})
	if err != nil {
		t.Fatalf("colored Inspect() error = %v", err)
	}
	if !strings.Contains(colored, "\x1b[33mPARTIAL\x1b[0m") ||
		!strings.Contains(colored, "\x1b[33mWARNING\x1b[0m") {
		t.Fatalf("colored output lacks status ANSI: %q", colored)
	}
	stripped := strings.NewReplacer(ansiYellow, "", ansiGreen, "", ansiRed, "", ansiReset, "").Replace(colored)
	if stripped != plain {
		t.Fatalf("stripped colored output differs from plain\nplain:\n%s\ncolored:\n%s", plain, colored)
	}
}

func TestCustomProfileMetadata(t *testing.T) {
	t.Parallel()

	result := presetCheck(budget.StatusPassed, false)
	result.Policy = policy.Effective{
		Kind: policy.KindProfile,
		ID:   "portable",
		Metadata: policy.Metadata{
			Status:    "custom",
			Stability: "project",
			Renderer:  "compatibility",
			TargetFPS: 30,
			Quality:   "low",
		},
	}

	output, err := Check(result, Options{})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	for _, fragment := range []string{
		"Profile:     portable",
		"Status:      custom (project)",
		"Renderer:    Compatibility",
		"Target FPS:  30",
		"Quality:     Low",
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("output does not contain %q:\n%s", fragment, output)
		}
	}
	if strings.Contains(output, "Built-in presets are") {
		t.Fatalf("custom profile has built-in disclaimer:\n%s", output)
	}
}

func TestFormatIntegerIsLocaleIndependentForSignedValues(t *testing.T) {
	t.Parallel()

	tests := map[int64]string{
		0:                    "0",
		999:                  "999",
		1000:                 "1,000",
		-1234567:             "-1,234,567",
		-9223372036854775808: "-9,223,372,036,854,775,808",
	}
	for value, want := range tests {
		if got := formatInteger(value); got != want {
			t.Errorf("formatInteger(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestEvidenceHelpersCoverFrozenOrderingAndLabels(t *testing.T) {
	t.Parallel()

	items := []diagnostic.Diagnostic{
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning, File: "b", Line: 1, Resource: "z"},
		{Code: diagnostic.CodeInvalidTSCNRoot, Severity: diagnostic.SeverityError, File: "z"},
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning, File: "a", Line: 2, Resource: "b"},
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning, File: "a", Line: 1, Resource: "z"},
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning, File: "a", Line: 1, Resource: "a", Column: 2},
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning, File: "a", Line: 1, Resource: "a", Column: 1, Message: "b"},
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning, File: "a", Line: 1, Resource: "a", Column: 1, Message: "a", Occurrences: 2},
	}
	sorted := sortedDiagnostics(items)
	if sorted[0].Severity != diagnostic.SeverityError {
		t.Fatalf("first severity = %q", sorted[0].Severity)
	}
	for index, want := range []string{
		"a:1:1 [a]",
		"a:1:1 [a]",
		"a:1:2 [a]",
		"a:1 [z]",
		"a:2 [b]",
		"b:1 [z]",
	} {
		if got := diagnosticLocation(sorted[index+1]); got != want {
			t.Errorf("location %d = %q, want %q", index, got, want)
		}
	}

	labels := map[analysis.TargetClassification]string{
		analysis.TargetImportedScene:            "imported PackedScene",
		analysis.TargetInheritedScene:           "inherited scene",
		analysis.TargetMissingExternalResource:  "missing scene resource",
		analysis.TargetUnresolvedPath:           "unresolved path (outside project)",
		analysis.TargetUnsupportedScene:         "unsupported scene",
		analysis.TargetSubResource:              "sub-resource instance",
		analysis.TargetPlaceholder:              "placeholder instance",
		analysis.TargetUnavailableScene:         "unavailable scene",
		analysis.TargetClassification("future"): "future",
		analysis.TargetClassification(""):       "unresolved scene",
	}
	for classification, want := range labels {
		reason := ""
		if classification == analysis.TargetUnresolvedPath {
			reason = "outside_project"
		}
		if got := displayClassification(classification, reason); got != want {
			t.Errorf("displayClassification(%q) = %q, want %q", classification, got, want)
		}
	}
	if got := displayClassification(analysis.TargetUnresolvedPath, ""); got != "unresolved path" {
		t.Errorf("empty unresolved label = %q", got)
	}
}

func TestRenderValidationFailures(t *testing.T) {
	t.Parallel()

	inspectCases := []struct {
		name   string
		mutate func(*application.InspectResult)
	}{
		{name: "status", mutate: func(result *application.InspectResult) { result.Analysis.Status = "future" }},
		{name: "reliability", mutate: func(result *application.InspectResult) { result.Analysis.Reliability = "future" }},
		{name: "metrics", mutate: func(result *application.InspectResult) { result.Analysis.Summary.Metrics.Nodes = -1 }},
		{name: "coverage", mutate: func(result *application.InspectResult) { result.Analysis.Coverage.ParsedSceneFiles = -1 }},
		{name: "diagnostic", mutate: func(result *application.InspectResult) {
			result.Analysis.Diagnostics = []diagnostic.Diagnostic{{Code: "future", Severity: diagnostic.SeverityWarning}}
		}},
	}
	for _, test := range inspectCases {
		result := completeInspect()
		test.mutate(&result)
		if output, err := Inspect(result, Options{}); err == nil || output != "" {
			t.Errorf("Inspect(%s) output/error = %q / %v", test.name, output, err)
		}
	}

	checkCases := []struct {
		name   string
		mutate func(*application.CheckResult)
	}{
		{name: "policy", mutate: func(result *application.CheckResult) { result.Policy.Kind = "future" }},
		{name: "status", mutate: func(result *application.CheckResult) { result.Evaluation.Status = "future" }},
		{name: "metric", mutate: func(result *application.CheckResult) { result.Evaluation.Results[0].Metric = "future" }},
		{name: "negative", mutate: func(result *application.CheckResult) { result.Evaluation.Results[0].Actual = -1 }},
	}
	for _, test := range checkCases {
		result := presetCheck(budget.StatusPassed, false)
		test.mutate(&result)
		if output, err := Check(result, Options{}); err == nil || output != "" {
			t.Errorf("Check(%s) output/error = %q / %v", test.name, output, err)
		}
	}
}

func TestFallbackPresentationValues(t *testing.T) {
	t.Parallel()

	result := completeInspect()
	result.Scene.Display = ""
	if got := preferredScenePath(result); got != result.Scene.Original {
		t.Fatalf("original fallback = %q", got)
	}
	result.Scene.Original = ""
	if got := preferredScenePath(result); got != result.Scene.Canonical {
		t.Fatalf("canonical fallback = %q", got)
	}
	result.Scene.Canonical = ""
	if got := preferredScenePath(result); got != "<unknown>" {
		t.Fatalf("unknown fallback = %q", got)
	}
	if got := displayRenderer("custom_renderer"); got != "Custom renderer" {
		t.Fatalf("custom renderer = %q", got)
	}
	if got := displayTitle("unspecified"); got != "Unspecified" {
		t.Fatalf("unspecified title = %q", got)
	}
}

func TestMixedMetricConfidenceUsesScopedMarkersAndQualifications(t *testing.T) {
	t.Parallel()

	result := ordinaryResourceInspect()
	rendered, err := Inspect(result, Options{Version: "test"})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	for _, exact := range []string{
		"Nodes",
		"Tree depth",
		"Scene instances",
		"Mesh instances",
		"Lights",
		"Shadow lights",
		"Scene dependencies",
	} {
		if !strings.Contains(rendered, exact) {
			t.Errorf("mixed report lacks %q", exact)
		}
	}
	if !strings.Contains(rendered, "External resources") ||
		!strings.Contains(rendered, "218+") ||
		!strings.Contains(rendered, "Metric confidence") ||
		!strings.Contains(rendered, "Nodes                      exact") ||
		!strings.Contains(rendered, "some static evidence is unavailable") ||
		strings.Contains(rendered, "0 scene instances") {
		t.Fatalf("mixed confidence report =\n%s", rendered)
	}

	check := presetCheck(budget.StatusPassed, false)
	check.Inspect = result
	check.Evaluation.Reliability = result.Analysis.Reliability
	checkText, err := Check(check, Options{Version: "test"})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !strings.Contains(checkText, "External resources") ||
		!strings.Contains(checkText, "218+") ||
		strings.Contains(checkText, "2,841+") {
		t.Fatalf("mixed check markers =\n%s", checkText)
	}
}

func completeInspect() application.InspectResult {
	result := application.InspectResult{
		Project: project.Root{Directory: "<PROJECT>", ProjectFile: "<PROJECT>/project.godot"},
		Scene: project.ResolvedPath{
			Canonical: "<PROJECT>/levels/city.tscn",
			Display:   "res://levels/city.tscn",
			Original:  "res://levels/city.tscn",
		},
		Analysis: analysis.RecursiveResult{
			Summary: analysis.ExpandedSummary{
				Metrics: cityMetrics(),
				Contributions: []analysis.SceneContribution{{
					Kind:             analysis.ContributionRoot,
					SceneCanonical:   "<PROJECT>/levels/city.tscn",
					SceneDisplay:     "res://levels/city.tscn",
					Occurrences:      1,
					Values:           fixtureContributionValues(cityMetrics()),
					DepthCandidate:   analysis.OptionalDepth{Value: cityMetrics().TreeDepth, Known: true},
					Reliability:      analysis.ReliabilityExact,
					MetricConfidence: fixtureMetricConfidence(analysis.ReliabilityExact),
				}},
			},
			Status:           analysis.AnalysisComplete,
			Reliability:      analysis.ReliabilityExact,
			MetricConfidence: fixtureMetricConfidence(analysis.ReliabilityExact),
			Coverage: analysis.Coverage{
				ParsedSceneFiles:         13,
				ResolvedSceneInstances:   184,
				UnresolvedSceneInstances: 0,
			},
		},
	}

	return result
}

func ordinaryResourceInspect() application.InspectResult {
	result := completeInspect()
	confidence, err := result.Analysis.MetricConfidence.With(
		metrics.ExternalResources,
		analysis.ReliabilityLowerBound,
		analysis.ConfidenceUnavailableResource,
	)
	if err != nil {
		panic(err)
	}
	result.Analysis.Status = analysis.AnalysisPartial
	result.Analysis.Reliability = analysis.ReliabilityLowerBound
	result.Analysis.MetricConfidence = confidence
	row := &result.Analysis.Summary.Contributions[0]
	row.MetricConfidence = confidence
	row.Reliability = confidence.Reliability()

	return result
}

func fixtureContributionValues(values metrics.Values) analysis.ContributionValues {
	return analysis.ContributionValues{
		Nodes:          values.Nodes,
		SceneInstances: values.SceneInstances,
		MeshInstances:  values.MeshInstances,
		Lights:         values.Lights,
		ShadowLights:   values.ShadowLights,
	}
}

func lowerBoundInspect() application.InspectResult {
	result := completeInspect()
	result.Analysis.Status = analysis.AnalysisPartial
	result.Analysis.Reliability = analysis.ReliabilityLowerBound
	result.Analysis.MetricConfidence = fixtureMetricConfidence(analysis.ReliabilityLowerBound)
	result.Analysis.Coverage.ResolvedSceneInstances = 179
	result.Analysis.Coverage.UnresolvedSceneInstances = 5
	result.Analysis.Summary.Unresolved = []analysis.UnresolvedInstance{
		{
			Classification:   analysis.TargetImportedScene,
			ResolutionReason: project.ResolutionUnsupportedTarget,
			RawTarget:        "res://models/tree.glb",
			Occurrences:      2,
		},
		{
			Classification:   analysis.TargetImportedScene,
			ResolutionReason: project.ResolutionUnsupportedTarget,
			RawTarget:        "res://models/car.glb",
			Occurrences:      1,
		},
		{
			Classification:   analysis.TargetImportedScene,
			ResolutionReason: project.ResolutionUnsupportedTarget,
			RawTarget:        "res://models/car.glb",
			Occurrences:      2,
		},
	}
	result.Analysis.Diagnostics = []diagnostic.Diagnostic{
		{
			Code:        diagnostic.CodeImportedScene,
			Severity:    diagnostic.SeverityWarning,
			Message:     "imported PackedScene cannot be expanded statically",
			File:        "res://models/tree.glb",
			Occurrences: 2,
		},
		{
			Code:        diagnostic.CodeImportedScene,
			Severity:    diagnostic.SeverityWarning,
			Message:     "imported PackedScene cannot be expanded statically",
			File:        "res://models/car.glb",
			Occurrences: 3,
		},
	}
	root := &result.Analysis.Summary.Contributions[0]
	root.Values.Nodes -= 5
	root.Values.SceneInstances -= 5
	result.Analysis.Summary.Contributions = append(result.Analysis.Summary.Contributions,
		analysis.SceneContribution{
			Kind:             analysis.ContributionUnresolved,
			SceneDisplay:     "res://models/tree.glb",
			DeclaringScene:   "<PROJECT>/levels/city.tscn",
			DeclaringDisplay: "res://levels/city.tscn",
			MountPath:        "Trees",
			RawTarget:        "res://models/tree.glb",
			Classification:   analysis.TargetImportedScene,
			Occurrences:      2,
			Values:           analysis.ContributionValues{Nodes: 2, SceneInstances: 2},
			DepthCandidate:   analysis.OptionalDepth{Value: 4, Known: true},
			Reliability:      analysis.ReliabilityLowerBound,
			MetricConfidence: fixtureMetricConfidence(analysis.ReliabilityLowerBound),
		},
		analysis.SceneContribution{
			Kind:             analysis.ContributionUnresolved,
			SceneDisplay:     "res://models/car.glb",
			DeclaringScene:   "<PROJECT>/levels/city.tscn",
			DeclaringDisplay: "res://levels/city.tscn",
			MountPath:        "Cars",
			RawTarget:        "res://models/car.glb",
			Classification:   analysis.TargetImportedScene,
			Occurrences:      3,
			Values:           analysis.ContributionValues{Nodes: 3, SceneInstances: 3},
			DepthCandidate:   analysis.OptionalDepth{Value: 3, Known: true},
			Reliability:      analysis.ReliabilityLowerBound,
			MetricConfidence: fixtureMetricConfidence(analysis.ReliabilityLowerBound),
		},
	)

	return result
}

func approximateInspect() application.InspectResult {
	result := completeInspect()
	result.Analysis.Status = analysis.AnalysisPartial
	result.Analysis.Reliability = analysis.ReliabilityApproximate
	result.Analysis.MetricConfidence = fixtureMetricConfidence(analysis.ReliabilityApproximate)
	result.Analysis.Coverage.InheritedScenes = 1
	result.Analysis.Summary.InheritedTargets = []analysis.InheritedTarget{{
		Classification: analysis.TargetInheritedScene,
		TargetDisplay:  "res://actors/zombie.tscn",
		Occurrences:    1,
	}}
	result.Analysis.Diagnostics = []diagnostic.Diagnostic{{
		Code:        diagnostic.CodeInheritedScene,
		Severity:    diagnostic.SeverityWarning,
		Message:     "inherited-scene overrides make expanded metrics approximate",
		File:        "res://actors/zombie.tscn",
		Occurrences: 1,
	}}
	result.Analysis.Summary.Contributions[0].Reliability = analysis.ReliabilityApproximate
	result.Analysis.Summary.Contributions[0].MetricConfidence = fixtureMetricConfidence(analysis.ReliabilityApproximate)

	return result
}

func fixtureMetricConfidence(reliability analysis.Reliability) analysis.MetricConfidence {
	reasons := []analysis.ConfidenceReason(nil)
	switch reliability {
	case analysis.ReliabilityLowerBound:
		reasons = []analysis.ConfidenceReason{analysis.ConfidenceImportedScene}
	case analysis.ReliabilityApproximate:
		reasons = []analysis.ConfidenceReason{analysis.ConfidenceInheritedScene}
	}
	confidence, err := analysis.UniformMetricConfidence(reliability, reasons...)
	if err != nil {
		panic(err)
	}

	return confidence
}

func presetCheck(status budget.Status, partial bool) application.CheckResult {
	inspect := completeInspect()
	if partial {
		inspect = lowerBoundInspect()
	}
	comparisons := failedComparisons()
	exceeded := 3
	if status == budget.StatusPassed {
		comparisons = passedComparisons()
		exceeded = 0
	}

	return application.CheckResult{
		Inspect: inspect,
		Policy: policy.Effective{
			Kind: policy.KindPreset,
			ID:   "steam-deck",
			Metadata: policy.Metadata{
				Status:    "heuristic",
				Stability: "experimental",
				Renderer:  "forward_plus",
				TargetFPS: 60,
				Quality:   "balanced",
			},
		},
		Evaluation: budget.Evaluation{
			Status:        status,
			Reliability:   inspect.Analysis.Reliability,
			FailOnPartial: status == budget.StatusIncomplete,
			Exceeded:      exceeded,
			Results:       comparisons,
		},
	}
}

func failedComparisons() []budget.Result {
	return []budget.Result{
		comparison(metrics.SceneDependencies, 12, 80),
		comparison(metrics.ExternalResources, 218, 300),
		comparison(metrics.ShadowLights, 9, 8),
		comparison(metrics.Lights, 43, 32),
		comparison(metrics.MeshInstances, 1024, 1000),
		comparison(metrics.SceneInstances, 184, 250),
		comparison(metrics.TreeDepth, 17, 20),
		comparison(metrics.Nodes, 2841, 3000),
	}
}

func passedComparisons() []budget.Result {
	return []budget.Result{
		comparison(metrics.SceneDependencies, 12, 80),
		comparison(metrics.ExternalResources, 218, 300),
		comparison(metrics.ShadowLights, 9, 10),
		comparison(metrics.Lights, 43, 50),
		comparison(metrics.MeshInstances, 1024, 1100),
		comparison(metrics.SceneInstances, 184, 250),
		comparison(metrics.TreeDepth, 17, 20),
		comparison(metrics.Nodes, 2841, 3000),
	}
}

func comparison(name metrics.Name, actual, limit int64) budget.Result {
	return budget.Result{
		Metric: name,
		Actual: actual,
		Limit:  limit,
		Delta:  actual - limit,
		Passed: actual <= limit,
	}
}

func cityMetrics() metrics.Values {
	return metrics.Values{
		Nodes:             2841,
		TreeDepth:         17,
		SceneInstances:    184,
		MeshInstances:     1024,
		Lights:            43,
		ShadowLights:      9,
		ExternalResources: 218,
		SceneDependencies: 12,
	}
}
