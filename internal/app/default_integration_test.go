package app_test

import (
	"os"
	"path/filepath"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
)

func TestDefaultApplicationInspectsAndChecksTextSceneWithoutGodot(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mustWrite(t, filepath.Join(projectRoot, "project.godot"), "[application]\nconfig/name=\"Fixture\"\n")
	scenePath := filepath.Join(projectRoot, "root.tscn")
	mustWrite(t, scenePath, "[gd_scene format=3]\n\n[node name=\"Root\" type=\"Node\"]\n")
	applicationService := application.NewDefault()

	inspect, err := applicationService.Inspect(application.InspectRequest{SceneRequest: application.SceneRequest{
		Scene:   scenePath,
		Project: projectRoot,
	}})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if inspect.Scene.Display != "res://root.tscn" || inspect.Analysis.Summary.Metrics.Nodes != 1 {
		t.Fatalf("inspect = %#v", inspect)
	}

	check, err := applicationService.Check(application.CheckRequest{
		SceneRequest: application.SceneRequest{
			Scene:   "res://root.tscn",
			Project: projectRoot,
		},
		BudgetOverrides: []string{"nodes=1"},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if check.Evaluation.Status != budget.StatusPassed || len(check.Evaluation.Results) != 1 {
		t.Fatalf("evaluation = %#v", check.Evaluation)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
