package cli_test

import (
	"bytes"
	"strings"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestInspectTopContributorSelectorsAreValidatedBeforeApplication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "metric only", args: []string{"inspect", "res://root.tscn", "--metric", "nodes"}, want: "--metric and --top must be supplied together"},
		{name: "top only", args: []string{"inspect", "res://root.tscn", "--top", "5"}, want: "--metric and --top must be supplied together"},
		{name: "zero", args: []string{"inspect", "res://root.tscn", "--metric", "nodes", "--top", "0"}, want: "--top must be a positive integer"},
		{name: "negative", args: []string{"inspect", "res://root.tscn", "--metric", "nodes", "--top", "-1"}, want: "--top must be a positive integer"},
		{name: "unknown", args: []string{"inspect", "res://root.tscn", "--metric", "triangles", "--top", "5"}, want: "invalid --metric"},
		{name: "resource union", args: []string{"inspect", "res://root.tscn", "--metric", "external_resources", "--top", "5"}, want: "shared unique union"},
		{name: "dependency union", args: []string{"inspect", "res://root.tscn", "--metric", "scene_dependencies", "--top", "5"}, want: "shared unique union"},
		{name: "int64 overflow", args: []string{"inspect", "res://root.tscn", "--metric", "nodes", "--top", "9223372036854775808"}, want: "value out of range"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fake := &fakeApplication{inspect: func(application.InspectRequest) (application.InspectResult, error) {
				return inspectResult(3), nil
			}}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exit := cli.ExecuteWithApplication(test.args, &stdout, &stderr, cli.BuildInfo{Version: "test"}, fake)
			if exit != 2 || fake.calls != 0 || stdout.Len() != 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("exit/calls/stdout/stderr = %d/%d/%q/%q, want usage containing %q", exit, fake.calls, stdout.String(), stderr.String(), test.want)
			}
		})
	}
}

func TestInspectTopContributorsRemainPresentationOnly(t *testing.T) {
	t.Parallel()

	wantRequest := application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene: "res://root.tscn", Project: "/game", Config: "/game/policy.json",
	}}
	for _, format := range []string{"text", "json"} {
		format := format
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			fake := &fakeApplication{inspect: func(request application.InspectRequest) (application.InspectResult, error) {
				if request != wantRequest {
					t.Fatalf("request = %#v, want %#v", request, wantRequest)
				}
				return inspectResult(3), nil
			}}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exit := cli.ExecuteWithApplication([]string{
				"--project", "/game", "--config", "/game/policy.json",
				"inspect", "res://root.tscn", "--metric", "nodes", "--top", "1", "--format", format,
			}, &stdout, &stderr, cli.BuildInfo{Version: "test"}, fake)
			if exit != 0 || stderr.Len() != 0 || fake.calls != 1 {
				t.Fatalf("exit/calls/stderr = %d/%d/%q", exit, fake.calls, stderr.String())
			}
			if format == "text" && !strings.Contains(stdout.String(), "Top contributors — nodes (limit 1)") {
				t.Fatalf("text output lacks top section: %s", stdout.String())
			}
			if format == "json" && !strings.Contains(stdout.String(), `"top_contributors"`) {
				t.Fatalf("JSON output lacks top projection: %s", stdout.String())
			}
		})
	}
}

func TestCheckDoesNotExposeInspectTopContributorFlags(t *testing.T) {
	t.Parallel()

	fake := &fakeApplication{check: func(application.CheckRequest) (application.CheckResult, error) {
		return checkResult(budget.StatusPassed), nil
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exit := cli.ExecuteWithApplication(
		[]string{"check", "res://root.tscn", "--top", "1"},
		&stdout,
		&stderr,
		cli.BuildInfo{Version: "test"},
		fake,
	)
	if exit != 2 || fake.calls != 0 || !strings.Contains(stderr.String(), "unknown flag: --top") {
		t.Fatalf("exit/calls/stderr = %d/%d/%q", exit, fake.calls, stderr.String())
	}
}
