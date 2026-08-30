package cli_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFormat4CommandsReuseExistingAnalysisAndPresentationFlows(t *testing.T) {
	t.Parallel()

	fixtures, _ := acceptancePaths(t)
	projectRoot := filepath.Join(fixtures, "format4")

	textOutput, textError, exit := executeDefaultCLI(t, []string{
		"--project", projectRoot, "inspect", "res://root.tscn",
	})
	if exit != 0 || textError != "" {
		t.Fatalf("text inspect exit/stderr = %d/%q", exit, textError)
	}
	for _, required := range []string{
		"Analysis:  COMPLETE", "Nodes                               4",
		"Scene instances                     2", "Scene dependencies                  2",
	} {
		if !strings.Contains(textOutput, required) {
			t.Errorf("text inspect lacks %q: %s", required, textOutput)
		}
	}

	inspectJSON := executeJSONSuccess(t, []string{
		"--project", projectRoot, "inspect", "res://root.tscn",
		"--metric", "nodes", "--top", "3", "--format", "json",
	})
	for _, required := range []string{
		`"kind": "inspect"`, `"status": "complete"`, `"reliability": "exact"`,
		`"id": "nodes"`, `"value": 4`, `"top_contributors"`, `"scene": "res://leaf.tscn"`,
	} {
		if !strings.Contains(inspectJSON, required) {
			t.Errorf("format-4 inspect JSON lacks %q: %s", required, inspectJSON)
		}
	}

	treeJSON := executeJSONSuccess(t, []string{
		"--project", projectRoot, "tree", "res://root.tscn", "--format", "json",
	})
	for _, required := range []string{
		`"kind": "tree"`, `"target": "res://child.tscn"`, `"target": "res://leaf.tscn"`,
	} {
		if !strings.Contains(treeJSON, required) {
			t.Errorf("format-4 tree JSON lacks %q: %s", required, treeJSON)
		}
	}

	checkJSON := executeJSONSuccess(t, []string{
		"--project", projectRoot, "check", "res://root.tscn",
		"--budget", "nodes=4", "--format", "json",
	})
	for _, required := range []string{`"kind": "check"`, `"verdict": "PASSED"`, `"observed": 4`} {
		if !strings.Contains(checkJSON, required) {
			t.Errorf("format-4 check JSON lacks %q: %s", required, checkJSON)
		}
	}

	inheritedOutput, inheritedError, exit := executeDefaultCLI(t, []string{
		"--project", projectRoot, "inspect", "res://derived.tscn",
	})
	if exit != 0 || inheritedError != "" || !strings.Contains(inheritedOutput, "Analysis:  PARTIAL") ||
		!strings.Contains(inheritedOutput, "Accuracy:  approximate") {
		t.Fatalf("inherited inspect exit/stdout/stderr = %d/%q/%q", exit, inheritedOutput, inheritedError)
	}

	_, futureError, exit := executeDefaultCLI(t, []string{
		"--project", projectRoot, "inspect", "res://future-root.tscn", "--format", "json",
	})
	if exit != 2 || !strings.Contains(futureError, `"kind": "error"`) ||
		!strings.Contains(futureError, "unsupported Godot scene format 5") {
		t.Fatalf("future format exit/stderr = %d/%q", exit, futureError)
	}
}
