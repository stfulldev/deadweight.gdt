package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestJSONInspectPreservesApplicationRequest(t *testing.T) {
	t.Parallel()

	var requests []application.InspectRequest
	service := &fakeApplication{inspect: func(request application.InspectRequest) (application.InspectResult, error) {
		requests = append(requests, request)
		return inspectResult(3), nil
	}}
	want := application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://root.tscn", Project: "/game", Config: "/game/policy.json",
	}}

	var textOut, textErr bytes.Buffer
	textExit := cli.ExecuteWithApplication(
		[]string{"--project", "/game", "--config", "/game/policy.json", "inspect", "res://root.tscn"},
		&textOut, &textErr, cli.BuildInfo{Version: "test"}, service,
	)
	var jsonOut, jsonErr bytes.Buffer
	jsonExit := cli.ExecuteWithApplication(
		[]string{"--project", "/game", "--config", "/game/policy.json", "inspect", "res://root.tscn", "--format", "json"},
		&jsonOut, &jsonErr, cli.BuildInfo{Version: "test"}, service,
	)

	if textExit != 0 || jsonExit != 0 || textErr.Len() != 0 || jsonErr.Len() != 0 {
		t.Fatalf("text/json exits and stderr = %d/%d %q/%q", textExit, jsonExit, textErr.String(), jsonErr.String())
	}
	if len(requests) != 2 || !reflect.DeepEqual(requests[0], want) || !reflect.DeepEqual(requests[1], want) {
		t.Fatalf("requests = %#v, want two %#v", requests, want)
	}
	if !strings.Contains(textOut.String(), "Analysis:  COMPLETE") {
		t.Fatalf("default output is not text: %q", textOut.String())
	}
	assertCommandJSONKind(t, jsonOut.Bytes(), "inspect")
}

func TestJSONCheckPreservesReportFirstExitOutcomes(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		status   budget.Status
		wantExit int
	}{
		{name: "passed", status: budget.StatusPassed, wantExit: 0},
		{name: "failed", status: budget.StatusFailed, wantExit: 1},
		{name: "incomplete", status: budget.StatusIncomplete, wantExit: 3},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var got application.CheckRequest
			service := &fakeApplication{check: func(request application.CheckRequest) (application.CheckResult, error) {
				got = request
				return checkResult(test.status), nil
			}}
			var stdout, stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				[]string{"check", "res://root.tscn", "--preset", "mobile", "--format", "json"},
				&stdout, &stderr, cli.BuildInfo{Version: "test"}, service,
			)
			if exitCode != test.wantExit || stderr.Len() != 0 {
				t.Fatalf("exit/stderr = %d/%q, want %d/empty", exitCode, stderr.String(), test.wantExit)
			}
			if got.Scene != "res://root.tscn" || got.Selector.Preset != "mobile" {
				t.Fatalf("request = %#v", got)
			}
			assertCommandJSONKind(t, stdout.Bytes(), "check")
			if !strings.Contains(stdout.String(), `"verdict": "`+string(test.status)+`"`) {
				t.Fatalf("stdout lacks verdict %q: %s", test.status, stdout.String())
			}
		})
	}
}

func TestInvalidFormatAndPresetFormatDoNotInvokeApplication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		fragment string
	}{
		{name: "inspect", args: []string{"inspect", "scene.tscn", "--format", "yaml"}, fragment: `invalid format "yaml"`},
		{name: "check", args: []string{"check", "scene.tscn", "--format", "future"}, fragment: `invalid format "future"`},
		{name: "preset list", args: []string{"presets", "--format", "json"}, fragment: "unknown flag"},
		{name: "preset show", args: []string{"presets", "show", "mobile", "--format", "json"}, fragment: "unknown flag"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeApplication{}
			var stdout, stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(test.args, &stdout, &stderr, cli.BuildInfo{Version: "test"}, service)
			if exitCode != 2 || stdout.Len() != 0 || service.calls != 0 {
				t.Fatalf("exit/stdout/calls = %d/%q/%d", exitCode, stdout.String(), service.calls)
			}
			if !strings.Contains(stderr.String(), test.fragment) {
				t.Fatalf("stderr lacks %q: %q", test.fragment, stderr.String())
			}
		})
	}
}

func TestJSONFatalRoutingForCodedAndUncodedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{
			name: "coded cycle",
			err: &analysis.CycleError{
				Display: []string{"res://a.tscn", "res://b.tscn", "res://a.tscn"},
			},
			wantCode: "SB2002",
		},
		{name: "uncoded", err: errors.New("analysis failed")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeApplication{inspect: func(application.InspectRequest) (application.InspectResult, error) {
				return application.InspectResult{}, test.err
			}}
			var stdout, stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				[]string{"inspect", "res://root.tscn", "--format", "json"},
				&stdout, &stderr, cli.BuildInfo{Version: "test"}, service,
			)
			if exitCode != 2 || stdout.Len() != 0 {
				t.Fatalf("exit/stdout = %d/%q", exitCode, stdout.String())
			}
			assertCommandJSONKind(t, stderr.Bytes(), "error")
			if test.wantCode != "" && !strings.Contains(stderr.String(), `"code": "`+test.wantCode+`"`) {
				t.Fatalf("stderr lacks code %q: %s", test.wantCode, stderr.String())
			}
			if test.wantCode == "" && strings.Contains(stderr.String(), `"code"`) {
				t.Fatalf("uncoded fatal contains code: %s", stderr.String())
			}
			for _, forbidden := range []string{"ERROR:", "goroutine ", "\x1b["} {
				if strings.Contains(stderr.String(), forbidden) {
					t.Fatalf("stderr contains %q: %s", forbidden, stderr.String())
				}
			}
		})
	}
}

func TestJSONColorInputsProduceIdenticalBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		terminal   bool
		noColor    bool
		noColorEnv bool
	}{
		{name: "redirected"},
		{name: "terminal", terminal: true},
		{name: "no color flag", terminal: true, noColor: true},
		{name: "no color environment", terminal: true, noColorEnv: true},
	}
	var want string
	for _, test := range tests {
		service := &fakeApplication{inspect: func(application.InspectRequest) (application.InspectResult, error) {
			return inspectResult(3), nil
		}}
		args := []string{"inspect", "res://root.tscn", "--format", "json"}
		if test.noColor {
			args = append([]string{"--no-color"}, args...)
		}
		var stdout, stderr bytes.Buffer
		exitCode := cli.ExecuteWithApplicationAndRuntime(
			args, &stdout, &stderr, cli.BuildInfo{Version: "test"}, service,
			cli.PresentationRuntime{
				LookupEnv: func(name string) (string, bool) {
					return "1", name == "NO_COLOR" && test.noColorEnv
				},
				IsTerminal: func(io.Writer) bool { return test.terminal },
			},
		)
		if exitCode != 0 || stderr.Len() != 0 {
			t.Fatalf("%s: exit/stderr = %d/%q", test.name, exitCode, stderr.String())
		}
		if want == "" {
			want = stdout.String()
		} else if stdout.String() != want {
			t.Fatalf("%s: JSON differs by color runtime\n--- want ---\n%s--- got ---\n%s", test.name, want, stdout.String())
		}
	}
}

func TestJSONSelectionRemainsTextForPreSelectionParseFailure(t *testing.T) {
	t.Parallel()

	service := &fakeApplication{}
	var stdout, stderr bytes.Buffer
	exitCode := cli.ExecuteWithApplication(
		[]string{"inspect", "--format", "json"},
		&stdout, &stderr, cli.BuildInfo{Version: "test"}, service,
	)
	if exitCode != 2 || stdout.Len() != 0 || service.calls != 0 {
		t.Fatalf("exit/stdout/calls = %d/%q/%d", exitCode, stdout.String(), service.calls)
	}
	if strings.HasPrefix(strings.TrimSpace(stderr.String()), "{") || !strings.Contains(stderr.String(), "arg(s)") {
		t.Fatalf("pre-selection error is not text usage output: %q", stderr.String())
	}
}

func assertCommandJSONKind(t *testing.T, encoded []byte, wantKind string) {
	t.Helper()
	if !json.Valid(encoded) {
		t.Fatalf("invalid JSON: %s", encoded)
	}
	var envelope struct {
		SchemaVersion int    `json:"schema_version"`
		Kind          string `json:"kind"`
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if envelope.SchemaVersion != 1 || envelope.Kind != wantKind {
		t.Fatalf("envelope = %#v, want version 1 kind %q", envelope, wantKind)
	}
	if strings.Contains(string(encoded), "\r") || !bytes.HasSuffix(encoded, []byte("\n")) || bytes.HasSuffix(encoded, []byte("\n\n")) {
		t.Fatalf("invalid JSON framing: %q", encoded)
	}
}
