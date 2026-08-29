package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestJSONReportGoldensValidateAgainstCommittedSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		render func() (string, error)
	}{
		{name: "inspect_complete", render: func() (string, error) {
			return InspectJSON(completeInspect(), Options{Version: "0.2.0-test", Color: true})
		}},
		{name: "inspect_lower_bound", render: func() (string, error) {
			return InspectJSON(lowerBoundInspect(), Options{Version: "0.2.0-test"})
		}},
		{name: "inspect_approximate", render: func() (string, error) {
			return InspectJSON(approximateInspect(), Options{Version: "0.2.0-test"})
		}},
		{name: "check_passed", render: func() (string, error) {
			return CheckJSON(presetCheck(budget.StatusPassed, false), Options{Version: "0.2.0-test"})
		}},
		{name: "check_failed", render: func() (string, error) {
			return CheckJSON(presetCheck(budget.StatusFailed, false), Options{Version: "0.2.0-test"})
		}},
		{name: "check_incomplete", render: func() (string, error) {
			return CheckJSON(presetCheck(budget.StatusIncomplete, true), Options{Version: "0.2.0-test"})
		}},
		{name: "error_coded", render: func() (string, error) {
			return ErrorJSON(&analysis.CycleError{
				Display: []string{"res://A.tscn", "res://B.tscn", "res://A.tscn"},
			}, Options{Version: "0.2.0-test"})
		}},
		{name: "error_uncoded", render: func() (string, error) {
			return ErrorJSON(errors.New("analysis <failed>"), Options{Version: "0.2.0-test"})
		}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := test.render()
			if err != nil {
				t.Fatalf("render error = %v", err)
			}
			assertJSONFraming(t, got)
			validateReportDocument(t, []byte(got))

			goldenPath := filepath.Join("testdata", "golden", "json", test.name+".golden")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("create golden directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0o600); err != nil {
					t.Fatalf("update golden: %v", err)
				}
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if got != string(want) {
				t.Fatalf("golden mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
			}
		})
	}
}

func TestJSONProjectionIsPortableOrderedAndImmutable(t *testing.T) {
	t.Parallel()

	left := portableInspectFixture(
		`/tmp/checkout-one`,
		`/tmp/checkout-one/scenes/root.tscn`,
		`/tmp/checkout-one/deadweight.gdt.json`,
		`/tmp/checkout-one/scenes/child.tscn`,
	)
	right := portableInspectFixture(
		`D:\\work\\checkout-two`,
		`D:\\work\\checkout-two\\scenes\\root.tscn`,
		`D:\\work\\checkout-two\\deadweight.gdt.json`,
		`D:\\work\\checkout-two\\scenes\\child.tscn`,
	)

	leftBefore := snapshotJSON(t, left)
	gotLeft, err := InspectJSON(left, Options{Version: "test", Color: true})
	if err != nil {
		t.Fatalf("left InspectJSON() error = %v", err)
	}
	gotRight, err := InspectJSON(right, Options{Version: "test"})
	if err != nil {
		t.Fatalf("right InspectJSON() error = %v", err)
	}
	if gotLeft != gotRight {
		t.Fatalf("portable documents differ\n--- left ---\n%s--- right ---\n%s", gotLeft, gotRight)
	}
	if after := snapshotJSON(t, left); !bytes.Equal(after, leftBefore) {
		t.Fatal("InspectJSON mutated caller-owned result")
	}
	for _, forbidden := range []string{"checkout-one", "checkout-two", `\\`, "\x1b["} {
		if strings.Contains(gotLeft, forbidden) {
			t.Fatalf("portable document contains %q: %s", forbidden, gotLeft)
		}
	}
	for _, required := range []string{
		`"path": "res://scenes/root.tscn"`,
		`"selection": "explicit"`,
		`"path": "res://deadweight.gdt.json"`,
		`"path": "res://scenes/child.tscn"`,
	} {
		if !strings.Contains(gotLeft, required) {
			t.Errorf("portable document lacks %q: %s", required, gotLeft)
		}
	}

	var document struct {
		Analysis struct {
			Metrics []struct {
				ID metrics.Name `json:"id"`
			} `json:"metrics"`
		} `json:"analysis"`
	}
	if err := json.Unmarshal([]byte(gotLeft), &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	gotOrder := make([]metrics.Name, 0, len(document.Analysis.Metrics))
	for _, metric := range document.Analysis.Metrics {
		gotOrder = append(gotOrder, metric.ID)
	}
	if want := metrics.OrderedNames(); !reflect.DeepEqual(gotOrder, want) {
		t.Fatalf("metric order = %v, want %v", gotOrder, want)
	}
}

func TestJSONCheckCanonicalizesComparisonsAndChecksIntegerInvariants(t *testing.T) {
	t.Parallel()

	result := presetCheck(budget.StatusFailed, false)
	before := snapshotJSON(t, result)
	rendered, err := CheckJSON(result, Options{Version: "test"})
	if err != nil {
		t.Fatalf("CheckJSON() error = %v", err)
	}
	if after := snapshotJSON(t, result); !bytes.Equal(after, before) {
		t.Fatal("CheckJSON mutated caller-owned comparison order")
	}
	last := -1
	for _, name := range metrics.OrderedNames() {
		position := strings.Index(rendered, `"metric": "`+string(name)+`"`)
		if position < 0 {
			t.Fatalf("missing comparison %q: %s", name, rendered)
		}
		if position <= last {
			t.Fatalf("comparison %q is outside canonical order: %s", name, rendered)
		}
		last = position
	}

	invalidDelta := result
	invalidDelta.Evaluation.Results = append([]budget.Result(nil), result.Evaluation.Results...)
	invalidDelta.Evaluation.Results[0].Delta++
	if output, err := CheckJSON(invalidDelta, Options{}); err == nil || output != "" {
		t.Fatalf("invalid delta output/error = %q / %v", output, err)
	}

	invalidCount := result
	invalidCount.Evaluation.Exceeded++
	if output, err := CheckJSON(invalidCount, Options{}); err == nil || output != "" {
		t.Fatalf("invalid exceeded output/error = %q / %v", output, err)
	}

	maximum := completeInspect()
	maximum.Analysis.Summary.Metrics = metrics.Values{
		Nodes: math.MaxInt64, TreeDepth: math.MaxInt64, SceneInstances: math.MaxInt64,
		MeshInstances: math.MaxInt64, Lights: math.MaxInt64, ShadowLights: math.MaxInt64,
		ExternalResources: math.MaxInt64, SceneDependencies: math.MaxInt64,
	}
	maximum.Analysis.Coverage = analysis.Coverage{
		ParsedSceneFiles: math.MaxInt64, ResolvedSceneInstances: math.MaxInt64,
		UnresolvedSceneInstances: math.MaxInt64, InheritedScenes: math.MaxInt64,
	}
	maximumJSON, err := InspectJSON(maximum, Options{Version: "test"})
	if err != nil {
		t.Fatalf("maximum InspectJSON() error = %v", err)
	}
	validateReportDocument(t, []byte(maximumJSON))
	if !strings.Contains(maximumJSON, "9223372036854775807") {
		t.Fatal("maximum signed 64-bit integer was not encoded as an integer")
	}
}

func TestJSONSchemaRejectsInvalidKindsEnumsAndVersions(t *testing.T) {
	t.Parallel()

	valid, err := InspectJSON(completeInspect(), Options{Version: "test"})
	if err != nil {
		t.Fatalf("InspectJSON() error = %v", err)
	}
	var base map[string]any
	decoder := json.NewDecoder(strings.NewReader(valid))
	decoder.UseNumber()
	if err := decoder.Decode(&base); err != nil {
		t.Fatalf("decode valid document: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing discriminator", mutate: func(document map[string]any) { delete(document, "kind") }},
		{name: "mixed payload", mutate: func(document map[string]any) {
			document["evaluation"] = map[string]any{"comparisons": []any{}, "exceeded": 0, "verdict": "PASSED"}
		}},
		{name: "invalid enum", mutate: func(document map[string]any) {
			document["analysis"].(map[string]any)["status"] = "future"
		}},
		{name: "unsupported schema", mutate: func(document map[string]any) { document["schema_version"] = 2 }},
		{name: "unknown kind", mutate: func(document map[string]any) { document["kind"] = "future" }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			document := cloneDocumentMap(t, base)
			test.mutate(document)
			if err := reportSchema(t).Validate(document); err == nil {
				t.Fatalf("schema accepted invalid document: %#v", document)
			}
		})
	}

	compatible := cloneDocumentMap(t, base)
	compatible["future_optional_field"] = map[string]any{"ignored": true}
	if err := reportSchema(t).Validate(compatible); err != nil {
		t.Fatalf("schema rejected compatible optional root field: %v", err)
	}
}

func TestJSONFatalDocumentIsMachineOnly(t *testing.T) {
	t.Parallel()

	rendered, err := ErrorJSON(&analysis.CycleError{
		Display: []string{"res://a.tscn", "res://b.tscn", "res://a.tscn"},
	}, Options{Version: "test", Color: true})
	if err != nil {
		t.Fatalf("ErrorJSON() error = %v", err)
	}
	assertJSONFraming(t, rendered)
	for _, required := range []string{`"kind": "error"`, `"code": "SB2002"`, `"severity": "error"`, `res://b.tscn`} {
		if !strings.Contains(rendered, required) {
			t.Errorf("fatal document lacks %q: %s", required, rendered)
		}
	}
	for _, forbidden := range []string{"ERROR SB2002:", "goroutine ", "\x1b["} {
		if strings.Contains(rendered, forbidden) {
			t.Errorf("fatal document contains %q: %s", forbidden, rendered)
		}
	}
	if output, err := ErrorJSON(nil, Options{}); err == nil || output != "" {
		t.Fatalf("ErrorJSON(nil) output/error = %q / %v", output, err)
	}
}

func portableInspectFixture(root, scene, configPath, diagnosticPath string) application.InspectResult {
	result := completeInspect()
	result.Project = project.Root{Directory: root, ProjectFile: root + `/project.godot`}
	result.Scene = project.ResolvedPath{Canonical: scene}
	result.ConfigPresent = true
	result.ConfigSource = config.Source{Path: configPath, Explicit: true}
	result.Analysis.Status = analysis.AnalysisPartial
	result.Analysis.Reliability = analysis.ReliabilityLowerBound
	result.Analysis.Diagnostics = []diagnostic.Diagnostic{{
		Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning,
		Message: "imported PackedScene cannot be expanded statically",
		File:    diagnosticPath, Line: 4, Column: 2, Occurrences: 2,
	}}
	return result
}

func assertJSONFraming(t *testing.T, rendered string) {
	t.Helper()
	if !json.Valid([]byte(rendered)) {
		t.Fatalf("invalid JSON: %q", rendered)
	}
	if strings.Contains(rendered, "\r") || !strings.HasSuffix(rendered, "\n") || strings.HasSuffix(rendered, "\n\n") {
		t.Fatalf("invalid JSON framing: %q", rendered)
	}
	if strings.Contains(rendered, "\x1b[") {
		t.Fatalf("JSON contains ANSI: %q", rendered)
	}
}

func validateReportDocument(t *testing.T, encoded []byte) {
	t.Helper()
	var document any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		t.Fatalf("decode report document: %v", err)
	}
	if err := reportSchema(t).Validate(document); err != nil {
		t.Fatalf("schema validation: %v\ndocument:\n%s", err, encoded)
	}
}

func reportSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "schema", "deadweight.gdt.report-v1.schema.json"))
	if err != nil {
		t.Fatalf("resolve report schema: %v", err)
	}
	compiled, err := jsonschema.NewCompiler().Compile(path)
	if err != nil {
		t.Fatalf("compile report schema: %v", err)
	}
	return compiled
}

func cloneDocumentMap(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("encode document clone: %v", err)
	}
	var result map[string]any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode document clone: %v", err)
	}
	return result
}

func snapshotJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("snapshot JSON: %v", err)
	}
	return encoded
}
