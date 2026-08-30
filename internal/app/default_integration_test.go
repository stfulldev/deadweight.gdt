package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestDefaultApplicationInspectsAndChecksTextSceneWithoutGodot(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mustWrite(t, filepath.Join(projectRoot, "project.godot"), "[application]\nconfig/name=\"Fixture\"\n")
	scenePath := filepath.Join(projectRoot, "root.tscn")
	mustWrite(t, scenePath, "[gd_scene format=3]\n\n[node name=\"Root\" type=\"Node\"]\n")
	applicationService := application.NewDefault()

	inspect, err := applicationService.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene:   scenePath,
		Project: projectRoot,
	}})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if inspect.Scene.Display != "res://root.tscn" || inspect.Analysis.Summary.Metrics.Nodes != 1 {
		t.Fatalf("inspect = %#v", inspect)
	}

	check, err := applicationService.Check(application.CheckRequest{
		SceneRequest: application.SceneRequest{
			Scene:   "res://root.tscn",
			Project: projectRoot,
		},
		BudgetOverrides: []string{"nodes=1"},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if check.Evaluation.Status != budget.StatusPassed || len(check.Evaluation.Results) != 1 {
		t.Fatalf("evaluation = %#v", check.Evaluation)
	}
}

func TestDefaultApplicationSupportsFormat4AcrossRecursiveAndInheritedFlows(t *testing.T) {
	t.Parallel()

	root := filepath.Join(fixtureProjects(t), "format4")
	service := application.NewDefault()

	complete, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://root.tscn", Project: root,
	}})
	if err != nil {
		t.Fatalf("inspect format-4 root: %v", err)
	}
	wantComplete := metrics.Values{
		Nodes: 4, TreeDepth: 4, SceneInstances: 2, MeshInstances: 1,
		Lights: 1, ShadowLights: 1, ExternalResources: 2, SceneDependencies: 2,
	}
	if complete.Analysis.Summary.Metrics != wantComplete || complete.Analysis.Status != analysis.AnalysisComplete ||
		complete.Analysis.Reliability != analysis.ReliabilityExact ||
		complete.Analysis.Coverage != (analysis.Coverage{ResolvedSceneInstances: 2, ParsedSceneFiles: 3}) {
		t.Fatalf("format-4 complete analysis = %#v", complete.Analysis)
	}

	inherited, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://derived.tscn", Project: root,
	}})
	if err != nil {
		t.Fatalf("inspect format-4 inherited scene: %v", err)
	}
	wantInherited := metrics.Values{
		Nodes: 3, TreeDepth: 2, MeshInstances: 1, Lights: 1, ShadowLights: 1,
		ExternalResources: 1, SceneDependencies: 1,
	}
	if inherited.Analysis.Summary.Metrics != wantInherited || inherited.Analysis.Status != analysis.AnalysisPartial ||
		inherited.Analysis.Reliability != analysis.ReliabilityApproximate ||
		inherited.Analysis.Coverage != (analysis.Coverage{ParsedSceneFiles: 2, InheritedScenes: 1}) {
		t.Fatalf("format-4 inherited analysis = %#v", inherited.Analysis)
	}

	format3, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://equivalent3.tscn", Project: root,
	}})
	if err != nil {
		t.Fatalf("inspect equivalent format 3: %v", err)
	}
	format4, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://equivalent4.tscn", Project: root,
	}})
	if err != nil {
		t.Fatalf("inspect equivalent format 4: %v", err)
	}
	if format4.Analysis.Summary.Metrics != format3.Analysis.Summary.Metrics ||
		format4.Analysis.Status != format3.Analysis.Status || format4.Analysis.Reliability != format3.Analysis.Reliability ||
		!reflect.DeepEqual(format4.Analysis.MetricConfidence, format3.Analysis.MetricConfidence) {
		t.Fatalf("equivalent analyses differ\nformat 3: %#v\nformat 4: %#v", format3.Analysis, format4.Analysis)
	}

	_, err = service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://future-root.tscn", Project: root,
	}})
	if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeInvalidTSCNRoot ||
		!strings.Contains(err.Error(), "unsupported Godot scene format 5") {
		t.Fatalf("future nested format error/code = %v / %q, %v", err, code, ok)
	}
}

func TestDefaultApplicationFixtureMatrix(t *testing.T) {
	t.Parallel()

	fixtures := fixtureProjects(t)
	tests := []struct {
		name          string
		group         string
		scene         string
		explicit      bool
		workingAtRoot bool
		wantDisplay   string
		wantMetrics   metrics.Values
		wantStatus    analysis.AnalysisStatus
		wantAccuracy  analysis.Reliability
		wantCoverage  analysis.Coverage
		wantResources int
		wantDeps      []string
		wantCodes     []diagnostic.Code
	}{
		{
			name: "absolute simple scene with nearest project discovery", group: "complete",
			scene: "simple.tscn", wantDisplay: "res://simple.tscn",
			wantMetrics: metrics.Values{Nodes: 3, TreeDepth: 2, MeshInstances: 1, Lights: 1, ShadowLights: 1},
			wantStatus:  analysis.AnalysisComplete, wantAccuracy: analysis.ReliabilityExact,
			wantCoverage: analysis.Coverage{ParsedSceneFiles: 1},
		},
		{
			name: "res path nested project", group: "complete", scene: "res://nested.tscn", explicit: true,
			wantDisplay: "res://nested.tscn",
			wantMetrics: metrics.Values{
				Nodes: 4, TreeDepth: 3, SceneInstances: 1, MeshInstances: 1,
				Lights: 1, ShadowLights: 1, ExternalResources: 2, SceneDependencies: 1,
			},
			wantStatus: analysis.AnalysisComplete, wantAccuracy: analysis.ReliabilityExact,
			wantCoverage:  analysis.Coverage{ResolvedSceneInstances: 1, ParsedSceneFiles: 2},
			wantResources: 2, wantDeps: []string{"deps/child.tscn"},
		},
		{
			name: "repeated scene uses one canonical parse", group: "repeated", scene: "res://city.tscn", explicit: true,
			wantDisplay: "res://city.tscn",
			wantMetrics: metrics.Values{
				Nodes: 7, TreeDepth: 3, SceneInstances: 3, MeshInstances: 3,
				Lights: 3, ShadowLights: 3, ExternalResources: 1, SceneDependencies: 1,
			},
			wantStatus: analysis.AnalysisComplete, wantAccuracy: analysis.ReliabilityExact,
			wantCoverage:  analysis.Coverage{ResolvedSceneInstances: 3, ParsedSceneFiles: 2},
			wantResources: 1, wantDeps: []string{"lamp.tscn"},
		},
		{
			name: "filesystem path relative to working directory", group: "relative-paths",
			scene: "levels/city.tscn", explicit: true, workingAtRoot: true,
			wantDisplay: "res://levels/city.tscn",
			wantMetrics: metrics.Values{
				Nodes: 2, TreeDepth: 2, SceneInstances: 1, Lights: 1,
				ExternalResources: 1, SceneDependencies: 1,
			},
			wantStatus: analysis.AnalysisComplete, wantAccuracy: analysis.ReliabilityExact,
			wantCoverage:  analysis.Coverage{ResolvedSceneInstances: 1, ParsedSceneFiles: 2},
			wantResources: 1, wantDeps: []string{"props/lamp.tscn"},
		},
		{
			name: "missing text scene is a lower bound", group: "unresolved",
			scene: "res://missing-tscn.tscn", explicit: true, wantDisplay: "res://missing-tscn.tscn",
			wantMetrics: metrics.Values{Nodes: 2, TreeDepth: 2, SceneInstances: 1, ExternalResources: 1},
			wantStatus:  analysis.AnalysisPartial, wantAccuracy: analysis.ReliabilityLowerBound,
			wantCoverage:  analysis.Coverage{UnresolvedSceneInstances: 1, ParsedSceneFiles: 1},
			wantResources: 1, wantCodes: []diagnostic.Code{diagnostic.CodeUnavailableResource},
		},
		{
			name: "imported scene is a lower bound", group: "unresolved",
			scene: "res://imported-glb.tscn", explicit: true, wantDisplay: "res://imported-glb.tscn",
			wantMetrics: metrics.Values{Nodes: 2, TreeDepth: 2, SceneInstances: 1, ExternalResources: 1},
			wantStatus:  analysis.AnalysisPartial, wantAccuracy: analysis.ReliabilityLowerBound,
			wantCoverage:  analysis.Coverage{UnresolvedSceneInstances: 1, ParsedSceneFiles: 1},
			wantResources: 1, wantCodes: []diagnostic.Code{diagnostic.CodeImportedScene},
		},
		{
			name: "placeholder is a lower bound", group: "unresolved",
			scene: "res://placeholder.tscn", explicit: true, wantDisplay: "res://placeholder.tscn",
			wantMetrics: metrics.Values{Nodes: 2, TreeDepth: 2, SceneInstances: 1},
			wantStatus:  analysis.AnalysisPartial, wantAccuracy: analysis.ReliabilityLowerBound,
			wantCoverage: analysis.Coverage{UnresolvedSceneInstances: 1, ParsedSceneFiles: 1},
			wantCodes:    []diagnostic.Code{diagnostic.CodeInstancePlaceholder},
		},
		{
			name: "inherited scene is approximate", group: "inherited",
			scene: "res://zombie.tscn", explicit: true, wantDisplay: "res://zombie.tscn",
			wantMetrics: metrics.Values{Nodes: 3, TreeDepth: 3, MeshInstances: 1, ExternalResources: 1, SceneDependencies: 1},
			wantStatus:  analysis.AnalysisPartial, wantAccuracy: analysis.ReliabilityApproximate,
			wantCoverage:  analysis.Coverage{ParsedSceneFiles: 2, InheritedScenes: 1},
			wantResources: 1, wantDeps: []string{"enemy.tscn"},
			wantCodes: []diagnostic.Code{diagnostic.CodeInheritedScene},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root := filepath.Join(fixtures, test.group)
			scene := test.scene
			if !strings.HasPrefix(scene, "res://") && !test.workingAtRoot {
				scene = filepath.Join(root, scene)
			}
			service := application.NewDefault()
			if test.workingAtRoot {
				service = application.New(application.Dependencies{WorkingDirectory: func() (string, error) {
					return root, nil
				}})
			}
			projectRoot := ""
			if test.explicit {
				projectRoot = root
			}

			result, err := service.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
				Scene: scene, Project: projectRoot,
			}})
			if err != nil {
				t.Fatalf("Inspect() error = %v", err)
			}
			if result.Scene.Display != test.wantDisplay || result.Analysis.Summary.Metrics != test.wantMetrics ||
				result.Analysis.Status != test.wantStatus || result.Analysis.Reliability != test.wantAccuracy ||
				result.Analysis.Coverage != test.wantCoverage || result.Analysis.ParsedSceneFiles != test.wantCoverage.ParsedSceneFiles {
				t.Fatalf("fixture result = %#v", result)
			}
			if len(result.Analysis.Summary.ExternalResources) != test.wantResources {
				t.Fatalf("resources = %#v, want %d", result.Analysis.Summary.ExternalResources, test.wantResources)
			}
			var wantDeps []string
			if len(test.wantDeps) > 0 {
				wantDeps = make([]string, len(test.wantDeps))
			}
			for index, path := range test.wantDeps {
				wantDeps[index] = filepath.Join(root, filepath.FromSlash(path))
			}
			if !reflect.DeepEqual(result.Analysis.Summary.Dependencies, wantDeps) {
				t.Fatalf("dependencies = %#v, want %#v", result.Analysis.Summary.Dependencies, wantDeps)
			}
			if codes := fixtureDiagnosticCodes(result.Analysis.Diagnostics); !reflect.DeepEqual(codes, test.wantCodes) {
				t.Fatalf("diagnostic codes = %#v, want %#v", codes, test.wantCodes)
			}
		})
	}
}

func TestDefaultApplicationFixtureFailuresAreCoded(t *testing.T) {
	t.Parallel()

	fixtures := fixtureProjects(t)
	t.Run("complete cycle chain", func(t *testing.T) {
		t.Parallel()

		root := filepath.Join(fixtures, "cyclic")
		_, err := application.NewDefault().Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
			Scene: "res://A.tscn", Project: root,
		}})
		if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeSceneDependencyCycle {
			t.Fatalf("error/code = %v / %q, %v", err, code, ok)
		}
		var cycle *analysis.CycleError
		if !errors.As(err, &cycle) {
			t.Fatalf("error type = %T, want wrapped *analysis.CycleError", err)
		}
		want := []string{"res://A.tscn", "res://B.tscn", "res://C.tscn", "res://A.tscn"}
		if !reflect.DeepEqual(cycle.Display, want) {
			t.Fatalf("cycle = %#v, want %#v", cycle.Display, want)
		}
	})

	for _, test := range []struct {
		name, scene, message string
	}{
		{name: "format two", scene: "format2.tscn", message: "unsupported Godot scene format 2"},
		{name: "unclosed string", scene: "unclosed-string.tscn", message: "unterminated string"},
		{name: "invalid external resource id", scene: "bad-ext-id.tscn", message: "must be a scalar"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root := filepath.Join(fixtures, "malformed")
			_, err := application.NewDefault().Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
				Scene: "res://" + test.scene, Project: root,
			}})
			if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeInvalidTSCNRoot ||
				!strings.Contains(err.Error(), test.message) {
				t.Fatalf("error/code = %v / %q, %v", err, code, ok)
			}
		})
	}
}

func fixtureProjects(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve fixture source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", "testdata", "projects"))
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Fatalf("fixture root %q: %v", root, err)
	}
	return root
}

func fixtureDiagnosticCodes(items []diagnostic.Diagnostic) []diagnostic.Code {
	var codes []diagnostic.Code
	if len(items) > 0 {
		codes = make([]diagnostic.Code, len(items))
	}
	for index, item := range items {
		codes[index] = item.Code
	}
	return codes
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
