package project_test

import (
	"errors"
	"io/fs"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestResolutionDomainContract(t *testing.T) {
	t.Parallel()

	reasons := []project.ResolutionReason{
		project.ResolutionResolved,
		project.ResolutionInvalidProjectRoot,
		project.ResolutionInvalidWorkingDir,
		project.ResolutionInvalidSceneInput,
		project.ResolutionEmpty,
		project.ResolutionUIDOnly,
		project.ResolutionUserData,
		project.ResolutionUnsupportedTarget,
		project.ResolutionMissing,
		project.ResolutionOutsideProject,
		project.ResolutionFilesystem,
		project.ResolutionInvalidDeclaringScene,
	}
	for _, reason := range reasons {
		if !reason.Valid() {
			t.Errorf("%q.Valid() = false", reason)
		}
	}
	if project.ResolutionReason("unknown").Valid() {
		t.Fatal("unknown resolution reason is valid")
	}

	path := project.ResolvedPath{
		Canonical: "/game/scenes/root.tscn",
		Display:   "res://scenes/root.tscn",
		Original:  "scenes/root.tscn",
	}
	resolved := project.Resolution{Reason: project.ResolutionResolved, Path: path}
	if !resolved.Resolved() || !reflect.DeepEqual(resolved.Path, path) {
		t.Fatalf("resolved = %#v", resolved)
	}

	cause := fs.ErrNotExist
	unresolved := project.Resolution{
		Reason:    project.ResolutionMissing,
		Path:      project.ResolvedPath{Original: "missing.tres"},
		Candidate: "/game/missing.tres",
		Err:       cause,
	}
	if unresolved.Resolved() {
		t.Fatalf("unresolved.Resolved() = true: %#v", unresolved)
	}
	if !errors.Is(unresolved, cause) {
		t.Fatalf("errors.Is(%v, %v) = false", unresolved, cause)
	}
	if unresolved.Path.Original != "missing.tres" || unresolved.Candidate != "/game/missing.tres" {
		t.Fatalf("unresolved evidence = %#v", unresolved)
	}
}

func TestResolveErrorContract(t *testing.T) {
	t.Parallel()

	cause := fs.ErrPermission
	err := &project.ResolveError{
		Reason:    project.ResolutionFilesystem,
		Original:  "res://secret.tres",
		Candidate: "/game/secret.tres",
		Err:       cause,
	}
	var resolveError *project.ResolveError
	if !errors.As(err, &resolveError) {
		t.Fatalf("errors.As(%v) = false", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(%v, %v) = false", err, cause)
	}
	if resolveError.Reason != project.ResolutionFilesystem ||
		resolveError.Original != "res://secret.tres" ||
		resolveError.Candidate != "/game/secret.tres" {
		t.Fatalf("resolve error = %#v", resolveError)
	}

	if !project.ReasonProjectNotFound.Valid() {
		t.Fatal("discovery error reasons changed while adding resolver errors")
	}
}
