package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestProfilesEndToEndAndCheckPolicyParity(t *testing.T) {
	projectRoot := t.TempDir()
	writeProfileFixture(t, filepath.Join(projectRoot, "project.godot"), "[application]\nconfig/name=\"Profile fixture\"\n")
	writeProfileFixture(t, filepath.Join(projectRoot, "root.tscn"), "[gd_scene format=3]\n\n[node name=\"Root\" type=\"Node\"]\n")
	configPath := filepath.Join(projectRoot, ".deadweight.gdt.json")
	writeProfileFixture(t, configPath, `{
  "version": 1,
  "fail_on_partial": false,
  "budgets": {"nodes": 90, "lights": 7},
  "profiles": {
    "shipping": {"extends": "base", "name": "Shipping", "quality": "low"},
    "ci": {"budgets": {"nodes": 1}},
    "base": {"extends": "steam-deck", "platform": "handheld", "budgets": {"nodes": 100}}
  }
}`)

	listText, listErr, exit := executeDefaultCLI(t, []string{"--project", projectRoot, "profiles"})
	if exit != 0 || listErr != "" {
		t.Fatalf("profiles exit/stdout/stderr = %d / %q / %q", exit, listText, listErr)
	}
	ciIndex := strings.Index(listText, "ci\n")
	baseIndex := strings.Index(listText, "base\n")
	shippingIndex := strings.Index(listText, "shipping\n")
	if baseIndex < 0 || ciIndex <= baseIndex || shippingIndex <= ciIndex {
		t.Fatalf("non-canonical profile list: %s", listText)
	}

	listJSON, listJSONErr, exit := executeDefaultCLI(t, []string{"--project", projectRoot, "profiles", "--format", "json"})
	if exit != 0 || listJSONErr != "" || !strings.Contains(listJSON, `"kind": "profiles"`) {
		t.Fatalf("profiles JSON exit/stdout/stderr = %d / %q / %q", exit, listJSON, listJSONErr)
	}

	showJSON, showErr, exit := executeDefaultCLI(t, []string{
		"--project", projectRoot, "profiles", "show", "shipping", "--format", "json",
	})
	if exit != 0 || showErr != "" {
		t.Fatalf("profiles show exit/stdout/stderr = %d / %q / %q", exit, showJSON, showErr)
	}
	checkJSON, checkErr, exit := executeDefaultCLI(t, []string{
		"--project", projectRoot, "check", "res://root.tscn", "--profile", "shipping", "--format", "json",
	})
	if exit != 0 || checkErr != "" {
		t.Fatalf("check exit/stdout/stderr = %d / %q / %q", exit, checkJSON, checkErr)
	}

	var profileDocument struct {
		Profile struct {
			Chain []struct {
				Kind string `json:"kind"`
				ID   string `json:"id"`
			} `json:"chain"`
			Metadata map[string]struct {
				Value any `json:"value"`
			} `json:"metadata"`
			Budgets []struct {
				Metric string `json:"metric"`
				Limit  int64  `json:"limit"`
			} `json:"budgets"`
			FailOnPartial struct {
				Value bool `json:"value"`
			} `json:"fail_on_partial"`
		} `json:"profile"`
	}
	var checkDocument struct {
		Policy struct {
			Metadata      map[string]any `json:"metadata"`
			FailOnPartial bool           `json:"fail_on_partial"`
		} `json:"policy"`
		Evaluation struct {
			Comparisons []struct {
				Metric string `json:"metric"`
				Limit  int64  `json:"limit"`
			} `json:"comparisons"`
		} `json:"evaluation"`
	}
	if err := json.Unmarshal([]byte(showJSON), &profileDocument); err != nil {
		t.Fatalf("decode profile JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(checkJSON), &checkDocument); err != nil {
		t.Fatalf("decode check JSON: %v", err)
	}
	wantChain := []struct{ Kind, ID string }{
		{Kind: "preset", ID: "steam-deck"},
		{Kind: "profile", ID: "base"},
		{Kind: "profile", ID: "shipping"},
	}
	gotChain := make([]struct{ Kind, ID string }, 0, len(profileDocument.Profile.Chain))
	for _, layer := range profileDocument.Profile.Chain {
		gotChain = append(gotChain, struct{ Kind, ID string }{Kind: layer.Kind, ID: layer.ID})
	}
	if !reflect.DeepEqual(gotChain, wantChain) {
		t.Fatalf("profile chain = %#v, want %#v", gotChain, wantChain)
	}
	for name, checkValue := range checkDocument.Policy.Metadata {
		if profileDocument.Profile.Metadata[name].Value != checkValue {
			t.Fatalf("metadata %s = %#v, check %#v", name, profileDocument.Profile.Metadata[name].Value, checkValue)
		}
	}
	profileBudgets := make(map[string]int64)
	for _, entry := range profileDocument.Profile.Budgets {
		profileBudgets[entry.Metric] = entry.Limit
	}
	checkBudgets := make(map[string]int64)
	for _, entry := range checkDocument.Evaluation.Comparisons {
		checkBudgets[entry.Metric] = entry.Limit
	}
	if !reflect.DeepEqual(profileBudgets, checkBudgets) || profileBudgets["nodes"] != 90 || profileBudgets["lights"] != 7 {
		t.Fatalf("profile/check budgets = %#v / %#v", profileBudgets, checkBudgets)
	}
	if profileDocument.Profile.FailOnPartial.Value != checkDocument.Policy.FailOnPartial {
		t.Fatalf("partial policy = %t / %t", profileDocument.Profile.FailOnPartial.Value, checkDocument.Policy.FailOnPartial)
	}
}

func TestProfilesEndToEndRejectsAbsentInvalidAndCyclicConfiguration(t *testing.T) {
	projectRoot := t.TempDir()
	writeProfileFixture(t, filepath.Join(projectRoot, "project.godot"), "[application]\n")

	_, absentErr, exit := executeDefaultCLI(t, []string{"--project", projectRoot, "profiles"})
	if exit != 2 || !strings.Contains(absentErr, ".deadweight.gdt.json") || !strings.Contains(absentErr, "--config") {
		t.Fatalf("absent config exit/error = %d / %q", exit, absentErr)
	}

	invalidPath := filepath.Join(projectRoot, "invalid.json")
	writeProfileFixture(t, invalidPath, `{"version":1,"unknown":true}`)
	_, invalidErr, exit := executeDefaultCLI(t, []string{"--project", projectRoot, "--config", invalidPath, "profiles"})
	if exit != 2 || !strings.Contains(invalidErr, "unknown") {
		t.Fatalf("invalid config exit/error = %d / %q", exit, invalidErr)
	}

	cyclePath := filepath.Join(projectRoot, "cycle.json")
	writeProfileFixture(t, cyclePath, `{
  "version": 1,
  "profiles": {"a": {"extends": "b"}, "b": {"extends": "a"}}
}`)
	_, cycleErr, exit := executeDefaultCLI(t, []string{"--project", projectRoot, "--config", cyclePath, "profiles"})
	if exit != 2 || !strings.Contains(cycleErr, "a -> b -> a") {
		t.Fatalf("cycle config exit/error = %d / %q", exit, cycleErr)
	}
}

func executeDefaultCLI(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	exit := cli.Execute(args, &stdout, &stderr, cli.BuildInfo{Version: "test"})
	return stdout.String(), stderr.String(), exit
}

func writeProfileFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
