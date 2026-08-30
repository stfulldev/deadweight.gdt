package app

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestInspectOrchestratesSupportedSceneInputsWithoutConfig(t *testing.T) {
	t.Parallel()

	for _, scene := range []string{
		"/game/scenes/root.tscn",
		"scenes/root.tscn",
		"res://scenes/root.tscn",
	} {
		scene := scene
		t.Run(scene, func(t *testing.T) {
			t.Parallel()

			var events []string
			resolver := &resolverStub{resolveSceneInput: func(input, workingDirectory string) (project.ResolvedPath, error) {
				events = append(events, "resolve")
				if input != scene || workingDirectory != "/work" {
					t.Fatalf("ResolveSceneInput(%q, %q)", input, workingDirectory)
				}

				return project.ResolvedPath{
					Canonical: "/game/scenes/root.tscn",
					Display:   "res://scenes/root.tscn",
					Original:  input,
				}, nil
			}}
			application := New(Dependencies{
				WorkingDirectory: func() (string, error) {
					events = append(events, "cwd")
					return "/work", nil
				},
				FindProject: func(request project.Request) (project.Root, error) {
					events = append(events, "project")
					if request.SceneInput != scene || request.WorkingDirectory != "/work" || request.ExplicitProject != "/game" {
						t.Fatalf("project request = %#v", request)
					}

					return project.Root{Directory: "/game", ProjectFile: "/game/project.godot"}, nil
				},
				LoadConfig: func(projectRoot, explicitPath string) (config.Config, config.Source, bool, error) {
					events = append(events, "config")
					if projectRoot != "/game" || explicitPath != "" {
						t.Fatalf("LoadConfig(%q, %q)", projectRoot, explicitPath)
					}

					return config.Config{}, config.Source{}, false, nil
				},
				NewResolver: func(projectRoot string) (SceneResolver, error) {
					events = append(events, "resolver")
					if projectRoot != "/game" {
						t.Fatalf("NewResolver(%q)", projectRoot)
					}

					return resolver, nil
				},
				Analyze: func(gotResolver SceneResolver, root project.ResolvedPath) (analysis.RecursiveResult, error) {
					events = append(events, "analyze")
					if gotResolver != resolver || root.Display != "res://scenes/root.tscn" {
						t.Fatalf("Analyze(%T, %#v)", gotResolver, root)
					}

					return completeAnalysis(metrics.Values{Nodes: 3}), nil
				},
			})

			result, err := application.Inspect(InspectRequest{SceneRequest: SceneRequest{
				Scene:   scene,
				Project: "/game",
			}})
			if err != nil {
				t.Fatalf("Inspect() error = %v", err)
			}
			if result.ConfigPresent || result.ConfigSource != (config.Source{}) {
				t.Fatalf("config evidence = %#v, present %t", result.ConfigSource, result.ConfigPresent)
			}
			if result.Scene.Original != scene || result.Analysis.Summary.Metrics.Nodes != 3 {
				t.Fatalf("result = %#v", result)
			}
			wantEvents := []string{"cwd", "project", "config", "resolver", "resolve", "analyze"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
		})
	}
}

func TestInspectIgnoresConfiguredFailOnPartial(t *testing.T) {
	t.Parallel()

	partialCalls := 0
	application := New(sceneDependencies(
		config.Config{Version: config.CurrentVersion, FailOnPartial: true},
		true,
		partialAnalysis(metrics.Values{Nodes: 2}),
	))
	application.dependencies.ResolvePartial = func(bool, budget.PartialOverride) (bool, error) {
		partialCalls++
		return false, errors.New("inspect must not resolve partial policy")
	}

	result, err := application.Inspect(InspectRequest{SceneRequest: SceneRequest{Scene: "res://root.tscn"}})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Analysis.Status != analysis.AnalysisPartial || partialCalls != 0 {
		t.Fatalf("status = %q, partial calls = %d", result.Analysis.Status, partialCalls)
	}
}

func TestTreeReusesSingleSceneAnalysisWithoutPolicyOrBudgets(t *testing.T) {
	t.Parallel()

	analysisCalls := 0
	policyCalls := 0
	dependencies := sceneDependencies(
		config.Config{Version: config.CurrentVersion, FailOnPartial: true},
		true,
		partialAnalysis(metrics.Values{Nodes: 2}),
	)
	dependencies.Analyze = func(SceneResolver, project.ResolvedPath) (analysis.RecursiveResult, error) {
		analysisCalls++
		return partialAnalysis(metrics.Values{Nodes: 2}), nil
	}
	dependencies.ResolvePolicy = func(string, config.Config, policy.Selector, []string) (policy.Effective, error) {
		policyCalls++
		return policy.Effective{}, errors.New("tree must not resolve policy")
	}
	dependencies.ResolvePartial = func(bool, budget.PartialOverride) (bool, error) {
		policyCalls++
		return false, errors.New("tree must not resolve partial policy")
	}
	dependencies.Evaluate = func(metrics.Values, budget.Limits, analysis.Reliability, bool) (budget.Evaluation, error) {
		policyCalls++
		return budget.Evaluation{}, errors.New("tree must not evaluate budgets")
	}

	application := New(dependencies)
	for _, scene := range []string{"/game/root.tscn", "root.tscn", "res://root.tscn"} {
		result, err := application.Tree(TreeRequest{SceneRequest: SceneRequest{
			Scene:   scene,
			Project: "/game",
			Config:  "/game/policy.json",
		}})
		if err != nil {
			t.Fatalf("Tree(%q) error = %v", scene, err)
		}
		if result.Inspect.Scene.Original != scene ||
			result.Inspect.Analysis.Status != analysis.AnalysisPartial ||
			!result.Inspect.ConfigPresent {
			t.Fatalf("Tree(%q) result = %#v", scene, result)
		}
	}
	if analysisCalls != 3 || policyCalls != 0 {
		t.Fatalf("analysis/policy calls = %d / %d", analysisCalls, policyCalls)
	}
}

func TestTreeReturnsZeroResultForFatalAnalysis(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("cycle")
	dependencies := sceneDependencies(config.Config{}, false, completeAnalysis(metrics.Values{}))
	dependencies.Analyze = func(SceneResolver, project.ResolvedPath) (analysis.RecursiveResult, error) {
		return analysis.RecursiveResult{}, wantErr
	}
	result, err := New(dependencies).Tree(TreeRequest{SceneRequest: SceneRequest{Scene: "res://root.tscn"}})
	if !errors.Is(err, wantErr) || !reflect.DeepEqual(result, TreeResult{}) {
		t.Fatalf("Tree() result/error = %#v / %v", result, err)
	}
}

func TestCheckForwardsOverridesAndReturnsOwnedResult(t *testing.T) {
	t.Parallel()

	configuredLimit := int64(20)
	configuration := config.Config{
		Version:       config.CurrentVersion,
		FailOnPartial: true,
		Budgets:       budget.Limits{Nodes: &configuredLimit},
	}
	dependencies := sceneDependencies(configuration, true, partialAnalysis(metrics.Values{Nodes: 12}))

	returnedLimit := int64(10)
	returnedPolicy := policy.Effective{
		Kind:    policy.KindPreset,
		ID:      "mobile",
		Budgets: budget.Limits{Nodes: &returnedLimit},
	}
	returnedEvaluation := budget.Evaluation{
		Status:        budget.StatusFailed,
		Reliability:   analysis.ReliabilityLowerBound,
		FailOnPartial: false,
		Exceeded:      1,
		Results: []budget.Result{{
			Metric: metrics.Nodes,
			Actual: 12,
			Limit:  10,
			Delta:  2,
		}},
	}

	var capturedBudgets []string
	dependencies.ResolvePolicy = func(
		source string,
		gotConfig config.Config,
		selector policy.Selector,
		budgets []string,
	) (policy.Effective, error) {
		if source != "/game/.deadweight.gdt.json" || gotConfig.FailOnPartial != true {
			t.Fatalf("policy source/config = %q / %#v", source, gotConfig)
		}
		if selector != (policy.Selector{Preset: "mobile"}) {
			t.Fatalf("selector = %#v", selector)
		}
		capturedBudgets = budgets
		return returnedPolicy, nil
	}
	dependencies.ResolvePartial = func(configured bool, override budget.PartialOverride) (bool, error) {
		if !configured || override != budget.PartialAllow {
			t.Fatalf("partial policy = %t / %q", configured, override)
		}

		return false, nil
	}
	dependencies.Evaluate = func(
		values metrics.Values,
		limits budget.Limits,
		reliability analysis.Reliability,
		failOnPartial bool,
	) (budget.Evaluation, error) {
		limit, _ := limits.Get(metrics.Nodes)
		if values.Nodes != 12 || limit != 10 || reliability != analysis.ReliabilityLowerBound || failOnPartial {
			t.Fatalf("Evaluate(%#v, %#v, %q, %t)", values, limits, reliability, failOnPartial)
		}

		return returnedEvaluation, nil
	}
	application := New(dependencies)
	request := CheckRequest{
		SceneRequest:    SceneRequest{Scene: "res://root.tscn"},
		Selector:        policy.Selector{Preset: "mobile"},
		BudgetOverrides: []string{"nodes=11", "nodes=10"},
		PartialOverride: budget.PartialAllow,
	}

	result, err := application.Check(request)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	request.BudgetOverrides[0] = "nodes=999"
	returnedLimit = 999
	returnedEvaluation.Results[0].Actual = 999

	if !reflect.DeepEqual(capturedBudgets, []string{"nodes=11", "nodes=10"}) {
		t.Fatalf("captured budgets = %v", capturedBudgets)
	}
	resultLimit, _ := result.Policy.Budgets.Get(metrics.Nodes)
	if resultLimit != 10 || result.Evaluation.Results[0].Actual != 12 {
		t.Fatalf("result aliases dependency data: policy %d, evaluation %#v", resultLimit, result.Evaluation)
	}
}

func TestCheckRejectsAnEmptyEffectiveBudget(t *testing.T) {
	t.Parallel()

	application := New(sceneDependencies(
		config.Config{Version: config.CurrentVersion},
		false,
		completeAnalysis(metrics.Values{Nodes: 1}),
	))

	result, err := application.Check(CheckRequest{SceneRequest: SceneRequest{Scene: "res://root.tscn"}})
	if err == nil || !strings.Contains(err.Error(), "effective policy has no budget") {
		t.Fatalf("Check() result/error = %#v / %v", result, err)
	}
	if !reflect.DeepEqual(result, CheckResult{}) {
		t.Fatalf("result = %#v, want zero", result)
	}
}

func TestCheckShortCircuitsAfterConfigFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("broken config")
	var events []string
	dependencies := sceneDependencies(config.Config{}, false, completeAnalysis(metrics.Values{}))
	dependencies.LoadConfig = func(string, string) (config.Config, config.Source, bool, error) {
		events = append(events, "config")
		return config.Config{}, config.Source{Path: "/game/config.json"}, true, wantErr
	}
	dependencies.NewResolver = func(string) (SceneResolver, error) {
		events = append(events, "resolver")
		return &resolverStub{}, nil
	}
	dependencies.ResolvePolicy = func(string, config.Config, policy.Selector, []string) (policy.Effective, error) {
		events = append(events, "policy")
		return policy.Effective{}, nil
	}
	application := New(dependencies)

	result, err := application.Check(CheckRequest{SceneRequest: SceneRequest{Scene: "res://root.tscn"}})
	if !errors.Is(err, wantErr) || !reflect.DeepEqual(result, CheckResult{}) {
		t.Fatalf("Check() result/error = %#v / %v", result, err)
	}
	if !reflect.DeepEqual(events, []string{"config"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestInspectReturnsZeroResultForEveryFatalStage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Dependencies, error)
	}{
		{
			name: "working directory",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.WorkingDirectory = func() (string, error) { return "", wantErr }
			},
		},
		{
			name: "project discovery",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.FindProject = func(project.Request) (project.Root, error) {
					return project.Root{}, wantErr
				}
			},
		},
		{
			name: "config load",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.LoadConfig = func(string, string) (config.Config, config.Source, bool, error) {
					return config.Config{}, config.Source{}, false, wantErr
				}
			},
		},
		{
			name: "resolver construction",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.NewResolver = func(string) (SceneResolver, error) { return nil, wantErr }
			},
		},
		{
			name: "scene resolution",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.NewResolver = func(string) (SceneResolver, error) {
					return &resolverStub{resolveSceneInput: func(string, string) (project.ResolvedPath, error) {
						return project.ResolvedPath{}, wantErr
					}}, nil
				}
			},
		},
		{
			name: "analysis",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.Analyze = func(SceneResolver, project.ResolvedPath) (analysis.RecursiveResult, error) {
					return analysis.RecursiveResult{}, wantErr
				}
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			wantErr := errors.New(test.name)
			dependencies := sceneDependencies(
				config.Config{Version: config.CurrentVersion},
				false,
				completeAnalysis(metrics.Values{}),
			)
			test.mutate(&dependencies, wantErr)

			result, err := New(dependencies).Inspect(InspectRequest{SceneRequest: SceneRequest{Scene: "res://root.tscn"}})
			if !errors.Is(err, wantErr) || !reflect.DeepEqual(result, InspectResult{}) {
				t.Fatalf("Inspect() result/error = %#v / %v", result, err)
			}
		})
	}
}

func TestCheckReturnsZeroResultForEveryPolicyStageFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Dependencies, error)
	}{
		{
			name: "policy resolution",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.ResolvePolicy = func(string, config.Config, policy.Selector, []string) (policy.Effective, error) {
					return policy.Effective{}, wantErr
				}
			},
		},
		{
			name: "partial policy",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.ResolvePartial = func(bool, budget.PartialOverride) (bool, error) {
					return false, wantErr
				}
			},
		},
		{
			name: "budget evaluation",
			mutate: func(dependencies *Dependencies, wantErr error) {
				dependencies.Evaluate = func(metrics.Values, budget.Limits, analysis.Reliability, bool) (budget.Evaluation, error) {
					return budget.Evaluation{}, wantErr
				}
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			wantErr := errors.New(test.name)
			dependencies := sceneDependencies(
				config.Config{Version: config.CurrentVersion},
				false,
				completeAnalysis(metrics.Values{Nodes: 1}),
			)
			limit := int64(1)
			dependencies.ResolvePolicy = func(string, config.Config, policy.Selector, []string) (policy.Effective, error) {
				return policy.Effective{Budgets: budget.Limits{Nodes: &limit}}, nil
			}
			test.mutate(&dependencies, wantErr)

			result, err := New(dependencies).Check(CheckRequest{SceneRequest: SceneRequest{Scene: "res://root.tscn"}})
			if !errors.Is(err, wantErr) || !reflect.DeepEqual(result, CheckResult{}) {
				t.Fatalf("Check() result/error = %#v / %v", result, err)
			}
		})
	}
}

func TestPresetFlowsDoNotConsultProjectState(t *testing.T) {
	t.Parallel()

	limit := int64(5)
	source := preset.Catalog{{
		ID:      "test",
		Name:    "Test",
		Budgets: budget.Limits{Nodes: &limit},
	}}
	loadCalls := 0
	application := New(Dependencies{
		WorkingDirectory: func() (string, error) {
			t.Fatal("preset flow consulted working directory")
			return "", nil
		},
		FindProject: func(project.Request) (project.Root, error) {
			t.Fatal("preset flow consulted project finder")
			return project.Root{}, nil
		},
		LoadConfig: func(string, string) (config.Config, config.Source, bool, error) {
			t.Fatal("preset flow consulted config")
			return config.Config{}, config.Source{}, false, nil
		},
		LoadBuiltInPresets: func() (preset.Catalog, error) {
			loadCalls++
			return source, nil
		},
	})

	listed, err := application.ListPresets()
	if err != nil {
		t.Fatalf("ListPresets() error = %v", err)
	}
	shown, err := application.ShowPreset("test")
	if err != nil {
		t.Fatalf("ShowPreset() error = %v", err)
	}
	limit = 99
	listedLimit, _ := listed.Catalog[0].Budgets.Get(metrics.Nodes)
	shownLimit, _ := shown.Preset.Budgets.Get(metrics.Nodes)
	if listedLimit != 5 || shownLimit != 5 || loadCalls != 2 {
		t.Fatalf("limits/calls = %d / %d / %d", listedLimit, shownLimit, loadCalls)
	}

	_, err = application.ShowPreset("missing")
	if err == nil || !strings.Contains(err.Error(), "available presets: test") {
		t.Fatalf("ShowPreset(missing) error = %v", err)
	}
}

func TestNilApplicationAndPresetLoadFailureReturnErrors(t *testing.T) {
	t.Parallel()

	var nilApplication *Application
	if _, err := nilApplication.Inspect(InspectRequest{}); err == nil {
		t.Fatal("nil Inspect() error = nil")
	}
	if _, err := nilApplication.Tree(TreeRequest{}); err == nil {
		t.Fatal("nil Tree() error = nil")
	}
	if _, err := nilApplication.ListPresets(); err == nil {
		t.Fatal("nil ListPresets() error = nil")
	}

	wantErr := errors.New("embedded preset failure")
	application := New(Dependencies{LoadBuiltInPresets: func() (preset.Catalog, error) {
		return nil, wantErr
	}})
	result, err := application.ListPresets()
	if !errors.Is(err, wantErr) || !reflect.DeepEqual(result, PresetListResult{}) {
		t.Fatalf("ListPresets() result/error = %#v / %v", result, err)
	}
}

type resolverStub struct {
	resolveSceneInput func(string, string) (project.ResolvedPath, error)
}

func (resolver *resolverStub) ResolveSceneInput(input, workingDirectory string) (project.ResolvedPath, error) {
	if resolver.resolveSceneInput == nil {
		return project.ResolvedPath{
			Canonical: "/game/root.tscn",
			Display:   "res://root.tscn",
			Original:  input,
		}, nil
	}

	return resolver.resolveSceneInput(input, workingDirectory)
}

func (*resolverStub) ResolveResource(string, string) project.Resolution {
	return project.Resolution{}
}

func sceneDependencies(
	configuration config.Config,
	configPresent bool,
	result analysis.RecursiveResult,
) Dependencies {
	return Dependencies{
		WorkingDirectory: func() (string, error) {
			return "/work", nil
		},
		FindProject: func(project.Request) (project.Root, error) {
			return project.Root{Directory: "/game", ProjectFile: "/game/project.godot"}, nil
		},
		LoadConfig: func(string, string) (config.Config, config.Source, bool, error) {
			source := config.Source{}
			if configPresent {
				source = config.Source{Path: "/game/.deadweight.gdt.json"}
			}

			return configuration, source, configPresent, nil
		},
		NewResolver: func(string) (SceneResolver, error) {
			return &resolverStub{}, nil
		},
		Analyze: func(SceneResolver, project.ResolvedPath) (analysis.RecursiveResult, error) {
			return result, nil
		},
	}
}

func completeAnalysis(values metrics.Values) analysis.RecursiveResult {
	return analysis.RecursiveResult{
		Summary:     analysis.ExpandedSummary{Metrics: values},
		Status:      analysis.AnalysisComplete,
		Reliability: analysis.ReliabilityExact,
	}
}

func partialAnalysis(values metrics.Values) analysis.RecursiveResult {
	return analysis.RecursiveResult{
		Summary:     analysis.ExpandedSummary{Metrics: values},
		Status:      analysis.AnalysisPartial,
		Reliability: analysis.ReliabilityLowerBound,
	}
}
