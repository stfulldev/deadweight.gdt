package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestDiffIntegrationEqualAndDeterministic(t *testing.T) {
	t.Parallel()

	baseline := filepath.Join("..", "report", "testdata", "golden", "json", "inspect_complete.golden")
	args := []string{"--project", "/unused", "--config", "/also-unused", "diff", baseline, baseline}
	firstOut, firstErr, firstExit := executeDiff(t, args)
	secondOut, secondErr, secondExit := executeDiff(t, args)
	if firstExit != 0 || firstErr != "" || firstOut != secondOut || firstErr != secondErr || firstExit != secondExit {
		t.Fatalf("deterministic executions = (%d, %q, %q) / (%d, %q, %q)", firstExit, firstOut, firstErr, secondExit, secondOut, secondErr)
	}
	if !strings.Contains(firstOut, "No semantic changes.") || strings.Contains(firstOut, baseline) || strings.Contains(firstOut, "\x1b[") {
		t.Fatalf("equal text output = %q", firstOut)
	}

	jsonOut, jsonErr, jsonExit := executeDiff(t, []string{"diff", baseline, baseline, "--format", "json"})
	if jsonExit != 0 || jsonErr != "" || !strings.Contains(jsonOut, `"kind": "diff"`) || !strings.Contains(jsonOut, `"changed": false`) || strings.Contains(jsonOut, baseline) {
		t.Fatalf("equal JSON = %d / %q / %q", jsonExit, jsonOut, jsonErr)
	}
}

func TestDiffIntegrationOptInOutcomesRenderBeforeExit(t *testing.T) {
	t.Parallel()

	baseline := filepath.Join("..", "report", "testdata", "golden", "json", "inspect_complete.golden")
	candidate := mutateReport(t, baseline, func(document map[string]any) {
		metrics := document["analysis"].(map[string]any)["metrics"].([]any)
		metrics[0].(map[string]any)["value"] = float64(3000)
	})
	stdout, stderr, exitCode := executeDiff(t, []string{"diff", baseline, candidate, "--fail-on-increase", "nodes"})
	if exitCode != 1 || stderr != "" || !strings.Contains(stdout, "REGRESSION") || !strings.Contains(stdout, "Enforcement\n  FAILED") {
		t.Fatalf("failed diff = %d / %q / %q", exitCode, stdout, stderr)
	}

	partial := filepath.Join("..", "report", "testdata", "golden", "json", "inspect_lower_bound.golden")
	stdout, stderr, exitCode = executeDiff(t, []string{"diff", baseline, partial, "--fail-on-reliability"})
	if exitCode != 3 || stderr != "" || !strings.Contains(stdout, "INCOMPLETE") || !strings.Contains(stdout, "reliability degraded") {
		t.Fatalf("incomplete diff = %d / %q / %q", exitCode, stdout, stderr)
	}
}

func TestDiffIntegrationFatalInputsHaveNoPartialStdout(t *testing.T) {
	t.Parallel()

	baseline := filepath.Join("..", "report", "testdata", "golden", "json", "inspect_complete.golden")
	malformed := filepath.Join(t.TempDir(), "malformed.json")
	if err := os.WriteFile(malformed, []byte("{\n"), 0o600); err != nil {
		t.Fatalf("write malformed fixture: %v", err)
	}
	stdout, stderr, exitCode := executeDiff(t, []string{"diff", baseline, malformed, "--format", "json"})
	if exitCode != 2 || stdout != "" || !strings.Contains(stderr, `"kind": "error"`) || !strings.Contains(stderr, "decode report JSON") {
		t.Fatalf("malformed JSON outcome = %d / %q / %q", exitCode, stdout, stderr)
	}

	stdout, stderr, exitCode = executeDiff(t, []string{"diff", "missing-before.json", "missing-after.json", "--fail-on-increase", "future"})
	if exitCode != 2 || stdout != "" || !strings.Contains(stderr, "unknown metric") || strings.Contains(stderr, "no such file") {
		t.Fatalf("pre-read validation outcome = %d / %q / %q", exitCode, stdout, stderr)
	}

	otherScene := mutateReport(t, baseline, func(document map[string]any) {
		document["scene"].(map[string]any)["path"] = "res://levels/other.tscn"
	})
	stdout, stderr, exitCode = executeDiff(t, []string{"diff", baseline, otherScene})
	if exitCode != 2 || stdout != "" || !strings.Contains(stderr, "incompatible root scenes") {
		t.Fatalf("scene mismatch outcome = %d / %q / %q", exitCode, stdout, stderr)
	}
}

func mutateReport(t *testing.T, source string, mutate func(map[string]any)) string {
	t.Helper()
	contents, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read report fixture: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("decode report fixture: %v", err)
	}
	mutate(document)
	contents, err = json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatalf("encode candidate fixture: %v", err)
	}
	path := filepath.Join(t.TempDir(), "candidate.json")
	if err := os.WriteFile(path, append(contents, '\n'), 0o600); err != nil {
		t.Fatalf("write candidate fixture: %v", err)
	}
	return path
}

func executeDiff(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	exitCode := cli.ExecuteWithApplication(args, &stdout, &stderr, cli.BuildInfo{Version: "0.2.0-test"}, application.NewDefault())
	return stdout.String(), stderr.String(), exitCode
}
