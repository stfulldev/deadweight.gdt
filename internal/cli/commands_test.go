package cli_test

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestInspectForwardsGlobalFlagsToInjectedApplication(t *testing.T) {
	t.Parallel()

	var got application.InspectRequest
	service := &fakeApplication{inspect: func(request application.InspectRequest) (application.InspectResult, error) {
		got = request
		return inspectResult(3), nil
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.ExecuteWithApplication(
		[]string{
			"--project", "/game",
			"--config", "/game/policy.json",
			"--no-color",
			"inspect", "res://root.tscn",
		},
		&stdout,
		&stderr,
		cli.BuildInfo{Version: "test"},
		service,
	)

	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit/stderr = %d / %q", exitCode, stderr.String())
	}
	want := application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene:   "res://root.tscn",
		Project: "/game",
		Config:  "/game/policy.json",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("request = %#v, want %#v", got, want)
	}
	for _, fragment := range []string{
		"Scene:     res://root.tscn",
		"Analysis:  COMPLETE",
		"Accuracy:  exact",
		"Nodes                               3",
	} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Errorf("stdout does not contain %q:\n%s", fragment, stdout.String())
		}
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("stdout contains ANSI: %q", stdout.String())
	}
}

func TestCheckForwardsFlagsAndMapsNonFatalOutcomes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   budget.Status
		wantExit int
	}{
		{name: "passed", status: budget.StatusPassed, wantExit: 0},
		{name: "failed", status: budget.StatusFailed, wantExit: 1},
		{name: "incomplete", status: budget.StatusIncomplete, wantExit: 3},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var got application.CheckRequest
			service := &fakeApplication{check: func(request application.CheckRequest) (application.CheckResult, error) {
				got = request
				return checkResult(test.status), nil
			}}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				[]string{
					"--project", "/game",
					"--config", "/game/policy.json",
					"check", "res://root.tscn",
					"--preset", "mobile",
					"--budget", "nodes=4",
					"--budget", "nodes=3",
					"--allow-partial",
				},
				&stdout,
				&stderr,
				cli.BuildInfo{Version: "test"},
				service,
			)

			if exitCode != test.wantExit || stderr.Len() != 0 {
				t.Fatalf("exit/stderr = %d / %q, want %d / empty", exitCode, stderr.String(), test.wantExit)
			}
			want := application.CheckRequest{
				SceneRequest: application.SceneRequest{
					Scene:   "res://root.tscn",
					Project: "/game",
					Config:  "/game/policy.json",
				},
				Selector:        policy.Selector{Preset: "mobile"},
				BudgetOverrides: []string{"nodes=4", "nodes=3"},
				PartialOverride: budget.PartialAllow,
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("request = %#v, want %#v", got, want)
			}
			if !strings.Contains(stdout.String(), verdictHeading(test.status)) {
				t.Fatalf("stdout = %q", stdout.String())
			}
		})
	}
}

func TestCheckForwardsProfileAndFailOnPartial(t *testing.T) {
	t.Parallel()

	var got application.CheckRequest
	service := &fakeApplication{check: func(request application.CheckRequest) (application.CheckResult, error) {
		got = request
		return checkResult(budget.StatusPassed), nil
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.ExecuteWithApplication(
		[]string{"check", "scene.tscn", "--profile", "portable", "--fail-on-partial"},
		&stdout,
		&stderr,
		cli.BuildInfo{Version: "test"},
		service,
	)

	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit/stderr = %d / %q", exitCode, stderr.String())
	}
	if got.Selector != (policy.Selector{Profile: "portable"}) || got.PartialOverride != budget.PartialFail {
		t.Fatalf("request = %#v", got)
	}
}

func TestInvalidCommandSyntaxDoesNotInvokeApplication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		fragments []string
	}{
		{
			name:      "inspect missing scene",
			args:      []string{"inspect"},
			fragments: []string{"arg(s)"},
		},
		{
			name:      "inspect extra scene",
			args:      []string{"inspect", "one.tscn", "two.tscn"},
			fragments: []string{"arg(s)"},
		},
		{
			name:      "selector conflict",
			args:      []string{"check", "scene.tscn", "--preset", "mobile", "--profile", "custom"},
			fragments: []string{"preset", "profile"},
		},
		{
			name:      "partial conflict",
			args:      []string{"check", "scene.tscn", "--fail-on-partial", "--allow-partial"},
			fragments: []string{"fail-on-partial", "allow-partial"},
		},
		{
			name:      "preset list extra argument",
			args:      []string{"presets", "extra"},
			fragments: []string{"unknown command"},
		},
		{
			name:      "preset show missing id",
			args:      []string{"presets", "show"},
			fragments: []string{"arg(s)"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &fakeApplication{}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				test.args,
				&stdout,
				&stderr,
				cli.BuildInfo{Version: "test"},
				service,
			)

			if exitCode != 2 {
				t.Fatalf("exit = %d, stderr = %q", exitCode, stderr.String())
			}
			if service.calls != 0 {
				t.Fatalf("application calls = %d", service.calls)
			}
			for _, fragment := range test.fragments {
				if !strings.Contains(stderr.String(), fragment) {
					t.Errorf("stderr does not contain %q: %q", fragment, stderr.String())
				}
			}
		})
	}
}

func TestPresetCommandsUseInjectedApplicationOnly(t *testing.T) {
	t.Parallel()

	listCalls := 0
	showCalls := 0
	service := &fakeApplication{
		listPresets: func() (application.PresetListResult, error) {
			listCalls++
			return application.PresetListResult{Catalog: preset.Catalog{{ID: "test", Description: "Test preset"}}}, nil
		},
		showPreset: func(id string) (application.PresetShowResult, error) {
			showCalls++
			if id != "test" {
				t.Fatalf("preset id = %q", id)
			}

			return application.PresetShowResult{Preset: preset.Preset{ID: "test", Name: "Test"}}, nil
		},
	}

	for _, args := range [][]string{
		{"--project", "/does/not/exist", "--config", "/also/missing", "presets"},
		{"presets", "show", "test"},
	} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		exitCode := cli.ExecuteWithApplication(args, &stdout, &stderr, cli.BuildInfo{Version: "test"}, service)
		if exitCode != 0 || stderr.Len() != 0 {
			t.Fatalf("args %v: exit/stderr = %d / %q", args, exitCode, stderr.String())
		}
	}
	if listCalls != 1 || showCalls != 1 || service.calls != 2 {
		t.Fatalf("list/show/total calls = %d / %d / %d", listCalls, showCalls, service.calls)
	}
}

func TestApplicationFailureRemainsFatal(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("analysis failed")
	service := &fakeApplication{inspect: func(application.InspectRequest) (application.InspectResult, error) {
		return application.InspectResult{}, wantErr
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.ExecuteWithApplication(
		[]string{"inspect", "scene.tscn"},
		&stdout,
		&stderr,
		cli.BuildInfo{Version: "test"},
		service,
	)

	if exitCode != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), wantErr.Error()) {
		t.Fatalf("exit/stdout/stderr = %d / %q / %q", exitCode, stdout.String(), stderr.String())
	}
}

func TestColorPolicyUsesTerminalAndBothSuppressionInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		terminal   bool
		noColor    bool
		noColorEnv bool
		wantANSI   bool
	}{
		{name: "terminal color", terminal: true, wantANSI: true},
		{name: "explicit no color", terminal: true, noColor: true},
		{name: "environment no color", terminal: true, noColorEnv: true},
		{name: "non terminal", terminal: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &fakeApplication{inspect: func(application.InspectRequest) (application.InspectResult, error) {
				return inspectResult(3), nil
			}}
			args := []string{"inspect", "res://root.tscn"}
			if test.noColor {
				args = append([]string{"--no-color"}, args...)
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplicationAndRuntime(
				args,
				&stdout,
				&stderr,
				cli.BuildInfo{Version: "test"},
				service,
				cli.PresentationRuntime{
					LookupEnv: func(name string) (string, bool) {
						if name == "NO_COLOR" && test.noColorEnv {
							return "", true
						}

						return "", false
					},
					IsTerminal: func(io.Writer) bool { return test.terminal },
				},
			)
			if exitCode != 0 || stderr.Len() != 0 {
				t.Fatalf("exit/stderr = %d / %q", exitCode, stderr.String())
			}
			if got := strings.Contains(stdout.String(), "\x1b["); got != test.wantANSI {
				t.Fatalf("ANSI present = %t, want %t:\n%q", got, test.wantANSI, stdout.String())
			}
			if !strings.Contains(stdout.String(), "COMPLETE") {
				t.Fatalf("textual status missing: %q", stdout.String())
			}
		})
	}
}

type fakeApplication struct {
	inspect     func(application.InspectRequest) (application.InspectResult, error)
	check       func(application.CheckRequest) (application.CheckResult, error)
	listPresets func() (application.PresetListResult, error)
	showPreset  func(string) (application.PresetShowResult, error)
	calls       int
}

func (fake *fakeApplication) Inspect(request application.InspectRequest) (application.InspectResult, error) {
	fake.calls++
	if fake.inspect == nil {
		return application.InspectResult{}, errors.New("unexpected Inspect call")
	}

	return fake.inspect(request)
}

func (fake *fakeApplication) Check(request application.CheckRequest) (application.CheckResult, error) {
	fake.calls++
	if fake.check == nil {
		return application.CheckResult{}, errors.New("unexpected Check call")
	}

	return fake.check(request)
}

func (fake *fakeApplication) ListPresets() (application.PresetListResult, error) {
	fake.calls++
	if fake.listPresets == nil {
		return application.PresetListResult{}, errors.New("unexpected ListPresets call")
	}

	return fake.listPresets()
}

func (fake *fakeApplication) ShowPreset(id string) (application.PresetShowResult, error) {
	fake.calls++
	if fake.showPreset == nil {
		return application.PresetShowResult{}, errors.New("unexpected ShowPreset call")
	}

	return fake.showPreset(id)
}

func inspectResult(nodes int64) application.InspectResult {
	return application.InspectResult{
		Project: project.Root{Directory: "/game", ProjectFile: "/game/project.godot"},
		Scene: project.ResolvedPath{
			Canonical: "/game/root.tscn",
			Display:   "res://root.tscn",
			Original:  "res://root.tscn",
		},
		ConfigSource:  config.Source{Path: "/game/policy.json", Explicit: true},
		ConfigPresent: true,
		Analysis: analysis.RecursiveResult{
			Summary:     analysis.ExpandedSummary{Metrics: metrics.Values{Nodes: nodes}},
			Status:      analysis.AnalysisComplete,
			Reliability: analysis.ReliabilityExact,
		},
	}
}

func checkResult(status budget.Status) application.CheckResult {
	inspect := inspectResult(3)
	actual := int64(3)
	limit := int64(2)
	passed := false
	exceeded := 1
	if status == budget.StatusPassed {
		actual = 1
		passed = true
		exceeded = 0
	}
	if status == budget.StatusIncomplete {
		inspect.Analysis.Status = analysis.AnalysisPartial
		inspect.Analysis.Reliability = analysis.ReliabilityLowerBound
		inspect.Analysis.Coverage.UnresolvedSceneInstances = 1
	}

	return application.CheckResult{
		Inspect: inspect,
		Policy: policy.Effective{
			Kind: policy.KindPreset,
			ID:   "mobile",
		},
		Evaluation: budget.Evaluation{
			Status:        status,
			Reliability:   inspect.Analysis.Reliability,
			FailOnPartial: status == budget.StatusIncomplete,
			Exceeded:      exceeded,
			Results: []budget.Result{{
				Metric: metrics.Nodes,
				Actual: actual,
				Limit:  limit,
				Delta:  actual - limit,
				Passed: passed,
			}},
		},
	}
}

func verdictHeading(status budget.Status) string {
	switch status {
	case budget.StatusPassed:
		return "PASSED —"
	case budget.StatusFailed:
		return "FAILED —"
	case budget.StatusIncomplete:
		return "INCOMPLETE —"
	default:
		return string(status)
	}
}
