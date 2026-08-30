package cli_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestTreeForwardsOneEquivalentRequestForTextAndJSON(t *testing.T) {
	t.Parallel()

	want := application.TreeRequest{SceneRequest: application.SceneRequest{
		Scene:   "res://root.tscn",
		Project: "/game",
		Config:  "/game/policy.json",
	}}
	for _, format := range []string{"text", "json"} {
		format := format
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			var got application.TreeRequest
			service := &fakeApplication{tree: func(request application.TreeRequest) (application.TreeResult, error) {
				got = request
				return treeResult(false), nil
			}}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				[]string{
					"--project", "/game", "--config", "/game/policy.json",
					"tree", "res://root.tscn", "--format", format,
				},
				&stdout,
				&stderr,
				cli.BuildInfo{Version: "test"},
				service,
			)
			if exitCode != 0 || stderr.Len() != 0 || service.calls != 1 {
				t.Fatalf("exit/stderr/calls = %d / %q / %d", exitCode, stderr.String(), service.calls)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("request = %#v, want %#v", got, want)
			}
			if format == "text" {
				if !strings.Contains(stdout.String(), "Dependencies\nres://root.tscn") {
					t.Fatalf("text tree missing: %s", stdout.String())
				}
			} else {
				var document struct {
					Kind string `json:"kind"`
				}
				if err := json.Unmarshal(stdout.Bytes(), &document); err != nil || document.Kind != "tree" {
					t.Fatalf("JSON kind/error = %q / %v: %s", document.Kind, err, stdout.String())
				}
			}
		})
	}
}

func TestTreeSuccessfulPartialEvidenceExitsZero(t *testing.T) {
	t.Parallel()

	service := &fakeApplication{tree: func(application.TreeRequest) (application.TreeResult, error) {
		return treeResult(true), nil
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.ExecuteWithApplication(
		[]string{"tree", "res://root.tscn"},
		&stdout,
		&stderr,
		cli.BuildInfo{Version: "test"},
		service,
	)
	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit/stdout/stderr = %d / %q / %q", exitCode, stdout.String(), stderr.String())
	}
	for _, fragment := range []string{"Analysis:  PARTIAL", "lower_bound", "imported_scene", "WARNING"} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Errorf("partial tree lacks %q: %s", fragment, stdout.String())
		}
	}
}

func TestTreeFatalCycleUsesSelectedPresentationAndExitTwo(t *testing.T) {
	t.Parallel()

	for _, format := range []string{"text", "json"} {
		format := format
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			service := &fakeApplication{tree: func(application.TreeRequest) (application.TreeResult, error) {
				return application.TreeResult{}, &analysis.CycleError{
					Display: []string{"res://A.tscn", "res://B.tscn", "res://A.tscn"},
				}
			}}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				[]string{"tree", "res://A.tscn", "--format", format},
				&stdout,
				&stderr,
				cli.BuildInfo{Version: "test"},
				service,
			)
			if exitCode != 2 || stdout.Len() != 0 || service.calls != 1 {
				t.Fatalf("exit/stdout/calls = %d / %q / %d", exitCode, stdout.String(), service.calls)
			}
			if !strings.Contains(stderr.String(), string(diagnostic.CodeSceneDependencyCycle)) {
				t.Fatalf("cycle stderr = %q", stderr.String())
			}
			if gotJSON := strings.Contains(stderr.String(), `"kind": "error"`); gotJSON != (format == "json") {
				t.Fatalf("JSON error present = %t for %s: %s", gotJSON, format, stderr.String())
			}
		})
	}
}

func treeResult(partial bool) application.TreeResult {
	inspect := inspectResult(1)
	inspect.Analysis.Coverage.ParsedSceneFiles = 1
	inspect.Analysis.Graph = analysis.DependencyGraph{
		RootCanonical: "/game/root.tscn",
		RootDisplay:   "res://root.tscn",
		Nodes: []analysis.GraphNode{{
			Canonical: "/game/root.tscn",
			Display:   "res://root.tscn",
		}},
	}
	if partial {
		inspect.Analysis.Status = analysis.AnalysisPartial
		inspect.Analysis.Reliability = analysis.ReliabilityLowerBound
		confidence, err := analysis.UniformMetricConfidence(
			analysis.ReliabilityLowerBound,
			analysis.ConfidenceImportedScene,
		)
		if err != nil {
			panic(err)
		}
		inspect.Analysis.MetricConfidence = confidence
		inspect.Analysis.Coverage.UnresolvedSceneInstances = 1
		inspect.Analysis.Summary.Contributions[0].Reliability = analysis.ReliabilityLowerBound
		inspect.Analysis.Summary.Contributions[0].MetricConfidence = confidence
		inspect.Analysis.Graph.Edges = []analysis.GraphEdge{{
			FromCanonical:    "/game/root.tscn",
			FromDisplay:      "res://root.tscn",
			RawTarget:        "res://model.glb",
			ResourceID:       "1_model",
			Kind:             analysis.EdgeInstance,
			Classification:   analysis.TargetImportedScene,
			ResolutionReason: project.ResolutionResolved,
			Occurrences:      1,
		}}
	}

	return application.TreeResult{Inspect: inspect}
}
