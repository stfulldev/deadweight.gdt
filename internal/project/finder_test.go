package project_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestErrorContract(t *testing.T) {
	t.Parallel()

	reasons := []project.ErrorReason{
		project.ReasonInvalidWorkingDirectory,
		project.ReasonInvalidSceneInput,
		project.ReasonInvalidExplicitProject,
		project.ReasonFilesystem,
		project.ReasonProjectNotFound,
	}
	for _, reason := range reasons {
		if !reason.Valid() {
			t.Errorf("%q.Valid() = false", reason)
		}
	}
	if project.ErrorReason("unknown").Valid() {
		t.Fatal("unknown error reason is valid")
	}

	cause := fs.ErrPermission
	err := &project.Error{Reason: project.ReasonFilesystem, Path: "/project", Err: cause}
	var projectError *project.Error
	if !errors.As(err, &projectError) {
		t.Fatalf("errors.As(%v) = false", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(%v, %v) = false", err, cause)
	}
	if projectError.Reason != project.ReasonFilesystem || projectError.Path != "/project" {
		t.Fatalf("project error = %#v", projectError)
	}
}

func TestFindExplicitProject(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	cwd := filepath.Join(workspace, "work")
	projectRoot := filepath.Join(workspace, "game")
	mustMkdirAll(t, cwd)
	mustWriteFile(t, filepath.Join(projectRoot, "project.godot"))

	tests := []struct {
		name     string
		explicit string
	}{
		{name: "relative directory", explicit: filepath.Join("..", "game")},
		{name: "absolute directory", explicit: projectRoot},
		{name: "marker file", explicit: filepath.Join(projectRoot, "project.godot")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root, err := project.NewFinder().Find(project.Request{
				SceneInput:       "res://missing.tscn",
				WorkingDirectory: cwd,
				ExplicitProject:  test.explicit,
			})
			if err != nil {
				t.Fatalf("Find() error = %v", err)
			}
			want := project.Root{
				Directory:   projectRoot,
				ProjectFile: filepath.Join(projectRoot, "project.godot"),
			}
			if !reflect.DeepEqual(root, want) {
				t.Fatalf("Find() = %#v, want %#v", root, want)
			}
		})
	}
}

func TestFindRejectsInvalidExplicitProjectWithoutFallback(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	mustWriteFile(t, filepath.Join(cwd, "project.godot"))
	wrongFile := filepath.Join(cwd, "other.godot")
	mustWriteFile(t, wrongFile)
	missingMarker := filepath.Join(cwd, "missing-marker")
	mustMkdirAll(t, missingMarker)
	nonRegularMarker := filepath.Join(cwd, "non-regular-marker")
	mustMkdirAll(t, filepath.Join(nonRegularMarker, "project.godot"))

	tests := []struct {
		name     string
		explicit string
	}{
		{name: "missing", explicit: filepath.Join(cwd, "absent")},
		{name: "wrong file name", explicit: wrongFile},
		{name: "directory without marker", explicit: missingMarker},
		{name: "non-regular marker", explicit: nonRegularMarker},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := project.NewFinder().Find(project.Request{
				SceneInput:       "res://scene.tscn",
				WorkingDirectory: cwd,
				ExplicitProject:  test.explicit,
			})
			_ = assertProjectError(t, err, project.ReasonInvalidExplicitProject)
		})
	}
}

func TestFindExplicitProjectInspectionError(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	explicit := filepath.Join(cwd, "forbidden")
	var inspected []string
	finder := project.NewFinderWithStat(func(path string) (fs.FileInfo, error) {
		inspected = append(inspected, path)
		if path == explicit {
			return nil, fs.ErrPermission
		}
		return os.Stat(path)
	})

	_, err := finder.Find(project.Request{
		SceneInput:       "res://scene.tscn",
		WorkingDirectory: cwd,
		ExplicitProject:  explicit,
	})
	projectError := assertProjectError(t, err, project.ReasonFilesystem)
	if !errors.Is(err, fs.ErrPermission) || projectError.Path != explicit {
		t.Fatalf("error = %#v", projectError)
	}
	if want := []string{explicit}; !reflect.DeepEqual(inspected, want) {
		t.Fatalf("inspected = %v, want %v", inspected, want)
	}
}

func TestFindFilesystemSceneUsesNearestProject(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	outer := filepath.Join(workspace, "outer")
	inner := filepath.Join(outer, "nested")
	scene := filepath.Join(inner, "scenes", "city.tscn")
	mustWriteFile(t, filepath.Join(outer, "project.godot"))
	mustWriteFile(t, filepath.Join(inner, "project.godot"))
	mustWriteFile(t, scene)

	tests := []struct {
		name  string
		cwd   string
		input string
	}{
		{name: "relative", cwd: inner, input: filepath.Join("scenes", "city.tscn")},
		{name: "absolute", cwd: workspace, input: scene},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root, err := project.NewFinder().Find(project.Request{
				SceneInput:       test.input,
				WorkingDirectory: test.cwd,
			})
			if err != nil {
				t.Fatalf("Find() error = %v", err)
			}
			if root.Directory != inner {
				t.Fatalf("root = %q, want %q", root.Directory, inner)
			}
		})
	}
}

func TestFindRejectsInvalidFilesystemScene(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	mustWriteFile(t, filepath.Join(cwd, "project.godot"))
	directoryScene := filepath.Join(cwd, "directory.tscn")
	mustMkdirAll(t, directoryScene)

	for _, input := range []string{"missing.tscn", "scene.TSCN", "scene.txt", directoryScene} {
		input := input
		t.Run(filepath.Base(input), func(t *testing.T) {
			t.Parallel()

			_, err := project.NewFinder().Find(project.Request{
				SceneInput:       input,
				WorkingDirectory: cwd,
			})
			_ = assertProjectError(t, err, project.ReasonInvalidSceneInput)
		})
	}
}

func TestFindResourcePathStartsAtWorkingDirectory(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	cwd := filepath.Join(rootDir, "nested", "deeper")
	mustMkdirAll(t, cwd)
	mustWriteFile(t, filepath.Join(rootDir, "project.godot"))

	root, err := project.NewFinder().Find(project.Request{
		SceneInput:       "res://does/not/exist.tscn",
		WorkingDirectory: cwd,
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if root.Directory != rootDir {
		t.Fatalf("root = %q, want %q", root.Directory, rootDir)
	}
}

func TestFindContinuesPastNonRegularMarker(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	cwd := filepath.Join(rootDir, "nested")
	mustMkdirAll(t, filepath.Join(cwd, "project.godot"))
	mustWriteFile(t, filepath.Join(rootDir, "project.godot"))

	root, err := project.NewFinder().Find(project.Request{
		SceneInput:       "res://scene.tscn",
		WorkingDirectory: cwd,
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if root.Directory != rootDir {
		t.Fatalf("root = %q, want %q", root.Directory, rootDir)
	}
}

func TestFindTerminatesAtFilesystemRoot(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	var inspected []string
	finder := project.NewFinderWithStat(func(path string) (fs.FileInfo, error) {
		if path == cwd {
			return os.Stat(path)
		}
		if filepath.Base(path) == "project.godot" {
			inspected = append(inspected, path)
			return nil, fs.ErrNotExist
		}
		return os.Stat(path)
	})

	_, err := finder.Find(project.Request{
		SceneInput:       "res://scene.tscn",
		WorkingDirectory: cwd,
	})
	projectError := assertProjectError(t, err, project.ReasonProjectNotFound)
	if projectError.Path != cwd || !strings.Contains(err.Error(), "pass --project") {
		t.Fatalf("error = %v", err)
	}

	volumeRoot := filepath.VolumeName(cwd) + string(filepath.Separator)
	wantLast := filepath.Join(volumeRoot, "project.godot")
	if len(inspected) == 0 || inspected[len(inspected)-1] != wantLast {
		t.Fatalf("inspected = %v, want last %q on %s", inspected, wantLast, runtime.GOOS)
	}
}

func TestFindMarkerInspectionError(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	marker := filepath.Join(cwd, "project.godot")
	finder := project.NewFinderWithStat(func(path string) (fs.FileInfo, error) {
		if path == marker {
			return nil, fs.ErrPermission
		}
		return os.Stat(path)
	})

	_, err := finder.Find(project.Request{
		SceneInput:       "res://scene.tscn",
		WorkingDirectory: cwd,
	})
	projectError := assertProjectError(t, err, project.ReasonFilesystem)
	if projectError.Path != marker || !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error = %#v", projectError)
	}
}

func TestFindRejectsInvalidWorkingDirectory(t *testing.T) {
	t.Parallel()

	for _, cwd := range []string{"", "relative/path"} {
		_, err := project.NewFinder().Find(project.Request{
			SceneInput:       "res://scene.tscn",
			WorkingDirectory: cwd,
		})
		_ = assertProjectError(t, err, project.ReasonInvalidWorkingDirectory)
	}
}

func assertProjectError(t *testing.T, err error, reason project.ErrorReason) *project.Error {
	t.Helper()

	var projectError *project.Error
	if !errors.As(err, &projectError) {
		t.Fatalf("error = %v, want *project.Error", err)
	}
	if projectError.Reason != reason {
		t.Fatalf("reason = %q, want %q (error: %v)", projectError.Reason, reason, err)
	}

	return projectError
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	mustMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte("; fixture\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
