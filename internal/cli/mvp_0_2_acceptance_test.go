package cli_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestMVP02CrossFeatureAcceptance(t *testing.T) {
	packageDir := cliPackageDirectory(t)
	fixtureRoot := filepath.Clean(filepath.Join(packageDir, "..", "..", "testdata", "projects", "mvp-0.2"))
	goldenRoot := filepath.Join(packageDir, "testdata", "golden", "mvp_0_2")

	workspace := t.TempDir()
	leftRoot := filepath.Join(workspace, "checkout-left", "project")
	rightRoot := filepath.Join(workspace, "checkout-right-with-a-different-prefix", "project")
	copyTree(t, fixtureRoot, leftRoot)
	copyTree(t, fixtureRoot, rightRoot)

	left := runMVP02Suite(t, leftRoot)
	right := runMVP02Suite(t, rightRoot)
	keys := make([]string, 0, len(left))
	for name := range left {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(goldenRoot, 0o755); err != nil {
			t.Fatalf("create MVP 0.2 golden directory: %v", err)
		}
	}

	for _, name := range keys {
		leftOutput := left[name]
		rightOutput, ok := right[name]
		if !ok {
			t.Fatalf("second checkout did not produce %s output", name)
		}
		if leftOutput != rightOutput {
			t.Fatalf("checkout prefixes changed %s JSON\n--- left ---\n%s--- right ---\n%s", name, leftOutput, rightOutput)
		}
		for _, forbidden := range []string{
			leftRoot, rightRoot, filepath.ToSlash(leftRoot), filepath.ToSlash(rightRoot), `\`,
		} {
			if strings.Contains(leftOutput, forbidden) {
				t.Fatalf("%s JSON leaked checkout-specific value %q", name, forbidden)
			}
		}

		goldenPath := filepath.Join(goldenRoot, name+".golden")
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			if err := os.WriteFile(goldenPath, []byte(leftOutput), 0o600); err != nil {
				t.Fatalf("update golden %s: %v", goldenPath, err)
			}
		}
		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("read golden %s: %v\n--- got ---\n%s", goldenPath, err, leftOutput)
		}
		if leftOutput != string(want) {
			t.Fatalf("%s golden mismatch\n--- want ---\n%s--- got ---\n%s", name, want, leftOutput)
		}
	}
}

func runMVP02Suite(t *testing.T, projectRoot string) map[string]string {
	t.Helper()

	// Verify the reviewable candidate file is independently valid before it is
	// copied over the portable root identity used for the semantic diff.
	executeJSONSuccess(t, []string{
		"--project", projectRoot, "inspect", "res://root.candidate.tscn", "--format", "json",
	})

	outputs := map[string]string{
		"inspect_contributors": executeJSONSuccess(t, []string{
			"--project", projectRoot, "inspect", "res://root.tscn",
			"--metric", "nodes", "--top", "3", "--format", "json",
		}),
		"tree": executeJSONSuccess(t, []string{
			"--project", projectRoot, "tree", "res://root.tscn", "--format", "json",
		}),
		"check_profile": executeJSONSuccess(t, []string{
			"--project", projectRoot, "check", "res://root.tscn",
			"--profile", "shipping", "--format", "json",
		}),
		"profiles": executeJSONSuccess(t, []string{
			"--project", projectRoot, "profiles", "--format", "json",
		}),
		"profile_shipping": executeJSONSuccess(t, []string{
			"--project", projectRoot, "profiles", "show", "shipping", "--format", "json",
		}),
	}

	reportDir := t.TempDir()
	baselinePath := filepath.Join(reportDir, "baseline.json")
	writeMVP02File(t, baselinePath, outputs["inspect_contributors"])
	candidate, err := os.ReadFile(filepath.Join(projectRoot, "root.candidate.tscn"))
	if err != nil {
		t.Fatalf("read candidate scene: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "root.tscn"), candidate, 0o600); err != nil {
		t.Fatalf("activate candidate scene: %v", err)
	}
	candidateReport := executeJSONSuccess(t, []string{
		"--project", projectRoot, "inspect", "res://root.tscn",
		"--metric", "nodes", "--top", "3", "--format", "json",
	})
	candidatePath := filepath.Join(reportDir, "candidate.json")
	writeMVP02File(t, candidatePath, candidateReport)
	outputs["diff"] = executeJSONSuccess(t, []string{
		"diff", baselinePath, candidatePath, "--format", "json",
	})

	return outputs
}

func writeMVP02File(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
