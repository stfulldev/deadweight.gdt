package cli_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

func TestDiffForwardsNormalizedPolicyAndMapsOutcomes(t *testing.T) {
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
			var got application.DiffRequest
			service := &fakeApplication{diff: func(request application.DiffRequest) (application.DiffResult, error) {
				got = request
				return application.DiffResult{Comparison: reportdiff.Result{
					Kind: reportdiff.KindInspect, Scene: "res://root.tscn",
					BeforeReliability: analysis.ReliabilityExact, AfterReliability: analysis.ReliabilityExact,
					Enforcement: reportdiff.Enforcement{Enabled: true, Status: test.status},
				}}, nil
			}}
			var stdout, stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(
				[]string{"--project", "/unused", "diff", "before.json", "after.json", "--fail-on-increase", "lights", "--fail-on-increase", "nodes", "--fail-on-reliability"},
				&stdout, &stderr, cli.BuildInfo{Version: "test"}, service,
			)
			if exitCode != test.wantExit || stderr.Len() != 0 || !strings.Contains(stdout.String(), "Enforcement") {
				t.Fatalf("exit/stdout/stderr = %d/%q/%q", exitCode, stdout.String(), stderr.String())
			}
			want := application.DiffRequest{Before: "before.json", After: "after.json", Policy: reportdiff.Policy{MetricIncreases: []metrics.Name{metrics.Nodes, metrics.Lights}, FailOnReliability: true}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("request = %#v, want %#v", got, want)
			}
		})
	}
}

func TestDiffInvalidSyntaxDoesNotInvokeApplication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing report", args: []string{"diff", "before.json"}, want: "arg(s)"},
		{name: "extra report", args: []string{"diff", "before.json", "after.json", "third.json"}, want: "arg(s)"},
		{name: "invalid format", args: []string{"diff", "before.json", "after.json", "--format", "yaml"}, want: "invalid format"},
		{name: "unknown metric", args: []string{"diff", "before.json", "after.json", "--fail-on-increase", "future"}, want: "unknown metric"},
		{name: "duplicate metric", args: []string{"diff", "before.json", "after.json", "--fail-on-increase", "nodes", "--fail-on-increase", "nodes"}, want: "duplicate metric"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeApplication{}
			var stdout, stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplication(test.args, &stdout, &stderr, cli.BuildInfo{Version: "test"}, service)
			if exitCode != 2 || stdout.Len() != 0 || service.calls != 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("exit/stdout/stderr/calls = %d/%q/%q/%d", exitCode, stdout.String(), stderr.String(), service.calls)
			}
		})
	}
}
