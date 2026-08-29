package project_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestNewResolverCanonicalRootAndDisplayPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "scenes", "nested", "root.tscn")
	mustWriteFile(t, nested)

	resolver, err := project.NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	if resolver.ProjectRoot() != root {
		t.Fatalf("ProjectRoot() = %q, want %q", resolver.ProjectRoot(), root)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "root", path: root, want: "res://"},
		{name: "nested", path: nested, want: "res://scenes/nested/root.tscn"},
		{name: "relative", path: filepath.Join("scenes", "root.tscn"), want: ""},
		{
			name: "non-clean",
			path: root + string(filepath.Separator) + "scenes" + string(filepath.Separator) + ".." +
				string(filepath.Separator) + "scenes",
			want: "",
		},
		{name: "outside", path: root + "-sibling", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := resolver.DisplayPath(test.path); got != test.want {
				t.Fatalf("DisplayPath(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestNewResolverRejectsInvalidProjectRoots(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	file := filepath.Join(workspace, "project.godot")
	mustWriteFile(t, file)

	tests := []struct {
		name   string
		root   string
		reason project.ResolutionReason
	}{
		{name: "relative", root: "game", reason: project.ResolutionInvalidProjectRoot},
		{name: "missing", root: filepath.Join(workspace, "missing"), reason: project.ResolutionInvalidProjectRoot},
		{name: "regular file", root: file, reason: project.ResolutionInvalidProjectRoot},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := project.NewResolver(test.root)
			_ = assertResolveError(t, err, test.reason)
		})
	}
}

func TestNewResolverPreservesFilesystemCause(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "forbidden")
	_, err := project.NewResolverWithFS(
		root,
		func(string) (fs.FileInfo, error) { return nil, fs.ErrPermission },
		filepath.EvalSymlinks,
	)
	resolveError := assertResolveError(t, err, project.ResolutionFilesystem)
	if resolveError.Candidate != root || !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error = %#v", resolveError)
	}
}

func TestNewResolverCanonicalizesSymlinkedProjectRoot(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	realRoot := filepath.Join(workspace, "real")
	mustMkdirAll(t, realRoot)
	linkedRoot := filepath.Join(workspace, "linked")
	if err := os.Symlink(realRoot, linkedRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	resolver, err := project.NewResolver(linkedRoot)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	if resolver.ProjectRoot() != realRoot {
		t.Fatalf("ProjectRoot() = %q, want %q", resolver.ProjectRoot(), realRoot)
	}
}

func TestResolveSceneInputSuccess(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	scene := filepath.Join(root, "scenes", "root.tscn")
	mustWriteFile(t, scene)
	resolver, err := project.NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}

	tests := []struct {
		name  string
		input string
		cwd   string
	}{
		{name: "absolute", input: scene, cwd: "unused-relative-cwd"},
		{name: "relative", input: filepath.Join("scenes", "root.tscn"), cwd: root},
		{name: "resource", input: "res://scenes/root.tscn", cwd: "unused-relative-cwd"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resolved, resolveErr := resolver.ResolveSceneInput(test.input, test.cwd)
			if resolveErr != nil {
				t.Fatalf("ResolveSceneInput() error = %v", resolveErr)
			}
			want := project.ResolvedPath{
				Canonical: scene,
				Display:   "res://scenes/root.tscn",
				Original:  test.input,
			}
			if resolved != want {
				t.Fatalf("ResolveSceneInput() = %#v, want %#v", resolved, want)
			}
		})
	}
}

func TestResolveSceneInputFailures(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	root := filepath.Join(workspace, "game")
	outside := filepath.Join(workspace, "outside")
	mustMkdirAll(t, root)
	outsideScene := filepath.Join(outside, "outside.tscn")
	mustWriteFile(t, outsideScene)
	directoryScene := filepath.Join(root, "directory.tscn")
	mustMkdirAll(t, directoryScene)

	resolver, err := project.NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	tests := []struct {
		name   string
		input  string
		cwd    string
		reason project.ResolutionReason
	}{
		{name: "empty", input: "", cwd: root, reason: project.ResolutionInvalidSceneInput},
		{name: "missing", input: "res://missing.tscn", cwd: root, reason: project.ResolutionMissing},
		{name: "non regular", input: directoryScene, cwd: root, reason: project.ResolutionInvalidSceneInput},
		{name: "wrong extension", input: "res://scene.TSCN", cwd: root, reason: project.ResolutionInvalidSceneInput},
		{name: "relative cwd", input: "scene.tscn", cwd: "relative", reason: project.ResolutionInvalidWorkingDir},
		{name: "absolute escape", input: outsideScene, cwd: root, reason: project.ResolutionOutsideProject},
		{
			name:   "resource lexical escape",
			input:  "res://../outside/outside.tscn",
			cwd:    root,
			reason: project.ResolutionOutsideProject,
		},
		{name: "unknown scheme", input: "uid://scene.tscn", cwd: root, reason: project.ResolutionInvalidSceneInput},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, resolveErr := resolver.ResolveSceneInput(test.input, test.cwd)
			_ = assertResolveError(t, resolveErr, test.reason)
		})
	}

	t.Run("symlink escape", func(t *testing.T) {
		escapeLink := filepath.Join(root, "escape")
		if err := os.Symlink(outside, escapeLink); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		_, resolveErr := resolver.ResolveSceneInput("res://escape/outside.tscn", root)
		_ = assertResolveError(t, resolveErr, project.ResolutionOutsideProject)
	})
}

func TestResolveSceneInputPreservesInspectionFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	scene := filepath.Join(root, "secret.tscn")
	resolver, err := project.NewResolverWithFS(
		root,
		func(path string) (fs.FileInfo, error) {
			if path == scene {
				return nil, fs.ErrPermission
			}
			return os.Stat(path)
		},
		filepath.EvalSymlinks,
	)
	if err != nil {
		t.Fatalf("NewResolverWithFS() error = %v", err)
	}

	_, err = resolver.ResolveSceneInput(scene, root)
	resolveError := assertResolveError(t, err, project.ResolutionFilesystem)
	if resolveError.Candidate != scene || !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error = %#v", resolveError)
	}
}

func TestResolveResourceSuccess(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	declaringScene := filepath.Join(root, "scenes", "level", "root.tscn")
	targets := map[string]string{
		"res://assets/material.tres": filepath.Join(root, "assets", "material.tres"),
		"local.png":                  filepath.Join(root, "scenes", "level", "local.png"),
		"../shared.tscn":             filepath.Join(root, "scenes", "shared.tscn"),
	}
	mustWriteFile(t, declaringScene)
	for _, target := range targets {
		mustWriteFile(t, target)
	}
	absoluteRaw := filepath.Join(root, "scripts", "run.gd")
	mustWriteFile(t, absoluteRaw)
	targets[absoluteRaw] = absoluteRaw

	resolver, err := project.NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	for raw, canonical := range targets {
		raw := raw
		canonical := canonical
		t.Run(filepath.Base(canonical), func(t *testing.T) {
			t.Parallel()

			resolution := resolver.ResolveResource(declaringScene, raw)
			if !resolution.Resolved() {
				t.Fatalf("ResolveResource() = %#v", resolution)
			}
			want := project.ResolvedPath{
				Canonical: canonical,
				Display:   resolver.DisplayPath(canonical),
				Original:  raw,
			}
			if resolution.Path != want || resolution.Reason != project.ResolutionResolved {
				t.Fatalf("resolution = %#v, want path %#v", resolution, want)
			}
		})
	}
}

func TestResolveResourceUnresolvedClassifications(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	root := filepath.Join(workspace, "game")
	declaringScene := filepath.Join(root, "scenes", "root.tscn")
	outside := filepath.Join(workspace, "outside.tres")
	nonRegular := filepath.Join(root, "directory")
	mustWriteFile(t, declaringScene)
	mustWriteFile(t, outside)
	mustMkdirAll(t, nonRegular)

	resolver, err := project.NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	tests := []struct {
		name      string
		fromScene string
		raw       string
		reason    project.ResolutionReason
	}{
		{name: "empty", fromScene: declaringScene, raw: "", reason: project.ResolutionEmpty},
		{name: "uid", fromScene: declaringScene, raw: "uid://abc", reason: project.ResolutionUIDOnly},
		{name: "user data", fromScene: declaringScene, raw: "user://save.dat", reason: project.ResolutionUserData},
		{name: "unknown scheme", fromScene: declaringScene, raw: "https://example.test/file", reason: project.ResolutionUnsupportedTarget},
		{name: "missing", fromScene: declaringScene, raw: "missing.tres", reason: project.ResolutionMissing},
		{name: "non regular", fromScene: declaringScene, raw: nonRegular, reason: project.ResolutionUnsupportedTarget},
		{name: "outside", fromScene: declaringScene, raw: outside, reason: project.ResolutionOutsideProject},
		{name: "relative declaring scene", fromScene: "scenes/root.tscn", raw: "local.tres", reason: project.ResolutionInvalidDeclaringScene},
		{
			name:      "non clean declaring scene",
			fromScene: filepath.Dir(declaringScene) + string(filepath.Separator) + ".." + string(filepath.Separator) + "scenes" + string(filepath.Separator) + "root.tscn",
			raw:       "local.tres",
			reason:    project.ResolutionInvalidDeclaringScene,
		},
		{name: "outside declaring scene", fromScene: outside, raw: "local.tres", reason: project.ResolutionInvalidDeclaringScene},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resolution := resolver.ResolveResource(test.fromScene, test.raw)
			if resolution.Resolved() || resolution.Reason != test.reason {
				t.Fatalf("resolution = %#v, want unresolved reason %q", resolution, test.reason)
			}
			if resolution.Path.Original != test.raw {
				t.Fatalf("original = %q, want %q", resolution.Path.Original, test.raw)
			}
		})
	}
}

func TestResolveResourceSchemesAvoidFilesystem(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	statCalls := 0
	evalCalls := 0
	resolver, err := project.NewResolverWithFS(
		root,
		func(path string) (fs.FileInfo, error) {
			statCalls++
			return os.Stat(path)
		},
		func(path string) (string, error) {
			evalCalls++
			return filepath.EvalSymlinks(path)
		},
	)
	if err != nil {
		t.Fatalf("NewResolverWithFS() error = %v", err)
	}
	statCalls = 0
	evalCalls = 0

	for _, raw := range []string{"", "uid://abc", "user://save", "https://example.test/file"} {
		_ = resolver.ResolveResource("", raw)
	}
	if statCalls != 0 || evalCalls != 0 {
		t.Fatalf("scheme classification used filesystem: stat=%d eval=%d", statCalls, evalCalls)
	}
}

func TestResolveResourcePreservesFilesystemCause(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	declaringScene := filepath.Join(root, "root.tscn")
	target := filepath.Join(root, "secret.tres")
	mustWriteFile(t, declaringScene)
	resolver, err := project.NewResolverWithFS(
		root,
		func(path string) (fs.FileInfo, error) {
			if path == target {
				return nil, fs.ErrPermission
			}
			return os.Stat(path)
		},
		filepath.EvalSymlinks,
	)
	if err != nil {
		t.Fatalf("NewResolverWithFS() error = %v", err)
	}

	resolution := resolver.ResolveResource(declaringScene, target)
	if resolution.Reason != project.ResolutionFilesystem ||
		resolution.Candidate != target ||
		!errors.Is(resolution, fs.ErrPermission) {
		t.Fatalf("resolution = %#v", resolution)
	}
}

func TestResolveResourceRejectsNoncanonicalDeclaringScene(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	realScene := filepath.Join(root, "real.tscn")
	linkedScene := filepath.Join(root, "linked.tscn")
	mustWriteFile(t, realScene)
	if err := os.Symlink(realScene, linkedScene); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	resolver, err := project.NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}

	resolution := resolver.ResolveResource(linkedScene, "local.tres")
	if resolution.Reason != project.ResolutionInvalidDeclaringScene {
		t.Fatalf("resolution = %#v", resolution)
	}
}

func assertResolveError(
	t *testing.T,
	err error,
	reason project.ResolutionReason,
) *project.ResolveError {
	t.Helper()

	if err == nil {
		t.Fatalf("error = nil, want reason %q", reason)
	}
	var resolveError *project.ResolveError
	if !errors.As(err, &resolveError) {
		t.Fatalf("error type = %T, want *project.ResolveError: %v", err, err)
	}
	if resolveError.Reason != reason {
		t.Fatalf("reason = %q, want %q: %v", resolveError.Reason, reason, err)
	}

	return resolveError
}
