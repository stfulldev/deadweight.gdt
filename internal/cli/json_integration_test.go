package cli_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestJSONIntegrationSceneInputsArePortableAcrossCheckouts(t *testing.T) {
	fixtures, _ := acceptancePaths(t)
	source := filepath.Join(fixtures, "complete")
	packageDir := cliPackageDirectory(t)
	absoluteScene := filepath.Join(source, "nested.tscn")
	relativeScene, err := filepath.Rel(packageDir, absoluteScene)
	if err != nil {
		t.Fatalf("relative fixture scene: %v", err)
	}

	inputs := []string{"res://nested.tscn", absoluteScene, relativeScene}
	var want string
	for _, input := range inputs {
		stdout := executeJSONSuccess(t, []string{"--project", source, "inspect", input, "--format", "json"})
		if want == "" {
			want = stdout
		} else if stdout != want {
			t.Fatalf("input %q changed portable JSON\n--- want ---\n%s--- got ---\n%s", input, want, stdout)
		}
	}

	checkoutParent := t.TempDir()
	leftRoot := filepath.Join(checkoutParent, "checkout-left")
	rightRoot := filepath.Join(checkoutParent, "checkout-right")
	copyTree(t, source, leftRoot)
	copyTree(t, source, rightRoot)
	left := executeJSONSuccess(t, []string{
		"--project", leftRoot, "inspect", filepath.Join(leftRoot, "nested.tscn"), "--format", "json",
	})
	right := executeJSONSuccess(t, []string{
		"--project", rightRoot, "inspect", filepath.Join(rightRoot, "nested.tscn"), "--format", "json",
	})
	if left != right {
		t.Fatalf("checkout roots changed portable JSON\n--- left ---\n%s--- right ---\n%s", left, right)
	}
	for _, root := range []string{leftRoot, rightRoot, filepath.ToSlash(leftRoot), filepath.ToSlash(rightRoot)} {
		if strings.Contains(left, root) || strings.Contains(right, root) {
			t.Fatalf("successful document leaked checkout root %q", root)
		}
	}
	if strings.Contains(left, `\\`) || strings.Contains(right, `\\`) {
		t.Fatal("successful document contains OS-specific separators")
	}
}

func TestTreeJSONIntegrationSceneInputsAndCheckoutsArePortable(t *testing.T) {
	fixtures, _ := acceptancePaths(t)
	source := filepath.Join(fixtures, "complete")
	packageDir := cliPackageDirectory(t)
	absoluteScene := filepath.Join(source, "nested.tscn")
	relativeScene, err := filepath.Rel(packageDir, absoluteScene)
	if err != nil {
		t.Fatalf("relative fixture scene: %v", err)
	}

	inputs := []string{"res://nested.tscn", absoluteScene, relativeScene}
	var want string
	for _, input := range inputs {
		stdout := executeJSONSuccess(t, []string{"--project", source, "tree", input, "--format", "json"})
		if want == "" {
			want = stdout
		} else if stdout != want {
			t.Fatalf("tree input %q changed portable JSON\n--- want ---\n%s--- got ---\n%s", input, want, stdout)
		}
	}
	for _, required := range []string{
		`"kind": "tree"`, `"root": "res://nested.tscn"`,
		`"source": "res://nested.tscn"`, `"target": "res://deps/child.tscn"`,
	} {
		if !strings.Contains(want, required) {
			t.Errorf("tree JSON lacks %q: %s", required, want)
		}
	}

	checkoutParent := t.TempDir()
	leftRoot := filepath.Join(checkoutParent, "tree-left")
	rightRoot := filepath.Join(checkoutParent, "tree-right")
	copyTree(t, source, leftRoot)
	copyTree(t, source, rightRoot)
	left := executeJSONSuccess(t, []string{
		"--project", leftRoot, "tree", filepath.Join(leftRoot, "nested.tscn"), "--format", "json",
	})
	right := executeJSONSuccess(t, []string{
		"--project", rightRoot, "tree", filepath.Join(rightRoot, "nested.tscn"), "--format", "json",
	})
	if left != right {
		t.Fatalf("checkout roots changed tree JSON\n--- left ---\n%s--- right ---\n%s", left, right)
	}
	for _, forbidden := range []string{leftRoot, rightRoot, filepath.ToSlash(leftRoot), filepath.ToSlash(rightRoot), `\`} {
		if strings.Contains(left, forbidden) || strings.Contains(right, forbidden) {
			t.Fatalf("tree JSON leaked checkout-specific value %q", forbidden)
		}
	}
}

func TestJSONIntegrationConfigurationProvenanceAndPresetMetadata(t *testing.T) {
	fixtures, _ := acceptancePaths(t)
	source := filepath.Join(fixtures, "complete")
	workspace := t.TempDir()
	projectRoot := filepath.Join(workspace, "project")
	copyTree(t, source, projectRoot)

	absent := executeJSONSuccess(t, []string{
		"--project", projectRoot, "inspect", "res://simple.tscn", "--format", "json",
	})
	if !strings.Contains(absent, `"present": false`) || !strings.Contains(absent, `"selection": "absent"`) {
		t.Fatalf("absent configuration provenance missing: %s", absent)
	}

	implicitPath := filepath.Join(projectRoot, ".deadweight.gdt.json")
	if err := os.WriteFile(implicitPath, []byte("{\"version\":1}\n"), 0o600); err != nil {
		t.Fatalf("write implicit config: %v", err)
	}
	implicit := executeJSONSuccess(t, []string{
		"--project", projectRoot, "inspect", "res://simple.tscn", "--format", "json",
	})
	if !strings.Contains(implicit, `"selection": "implicit"`) ||
		!strings.Contains(implicit, `"path": "res://.deadweight.gdt.json"`) {
		t.Fatalf("implicit configuration provenance missing: %s", implicit)
	}

	explicitPath := filepath.Join(workspace, "outside-policy.json")
	if err := os.WriteFile(explicitPath, []byte("{\"version\":1}\n"), 0o600); err != nil {
		t.Fatalf("write explicit config: %v", err)
	}
	explicit := executeJSONSuccess(t, []string{
		"--project", projectRoot, "--config", explicitPath,
		"inspect", "res://simple.tscn", "--format", "json",
	})
	if !strings.Contains(explicit, `"selection": "explicit"`) || strings.Contains(explicit, workspace) {
		t.Fatalf("explicit configuration provenance is not portable: %s", explicit)
	}

	preset := executeJSONSuccess(t, []string{
		"--project", projectRoot, "--config", explicitPath,
		"check", "res://simple.tscn", "--preset", "steam-deck", "--format", "json",
	})
	for _, required := range []string{
		`"kind": "preset"`, `"id": "steam-deck"`, `"status": "heuristic"`,
		`"stability": "experimental"`, `"fail_on_partial": false`,
	} {
		if !strings.Contains(preset, required) {
			t.Errorf("preset JSON lacks %q: %s", required, preset)
		}
	}
}

func TestJSONIntegrationGroupsDiagnosticsWithoutCanonicalPaths(t *testing.T) {
	fixtures, _ := acceptancePaths(t)
	projectRoot := filepath.Join(fixtures, "unresolved")
	stdout := executeJSONSuccess(t, []string{
		"--project", projectRoot, "inspect", "res://missing-tscn.tscn", "--format", "json",
	})
	for _, required := range []string{
		`"status": "partial"`, `"reliability": "lower_bound"`,
		`"code": "SB1004"`, `"occurrences": 1`,
	} {
		if !strings.Contains(stdout, required) {
			t.Errorf("partial JSON lacks %q: %s", required, stdout)
		}
	}
	if strings.Contains(stdout, projectRoot) || strings.Contains(stdout, filepath.ToSlash(projectRoot)) {
		t.Fatalf("partial JSON leaked canonical root: %s", stdout)
	}
}

func executeJSONSuccess(t *testing.T, args []string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	exitCode := cli.ExecuteWithApplication(
		args, &stdout, &stderr, cli.BuildInfo{Version: "0.2.0-test"}, application.NewDefault(),
	)
	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("args %v: exit/stderr = %d/%q", args, exitCode, stderr.String())
	}
	if !strings.HasSuffix(stdout.String(), "\n") || strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("args %v: invalid JSON framing: %q", args, stdout.String())
	}
	return stdout.String()
}

func cliPackageDirectory(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve CLI test source path")
	}
	return filepath.Dir(source)
}

func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, contents, 0o600)
	})
	if err != nil {
		t.Fatalf("copy fixture tree: %v", err)
	}
}
