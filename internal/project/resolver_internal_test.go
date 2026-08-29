package project

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveCandidateContainmentAndSymlinks(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	root := filepath.Join(workspace, "game")
	outside := filepath.Join(workspace, "outside")
	insideFile := filepath.Join(root, "assets", "inside.tres")
	outsideFile := filepath.Join(outside, "outside.tres")
	writeResolverTestFile(t, insideFile)
	writeResolverTestFile(t, outsideFile)

	resolver, err := NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}

	tests := []struct {
		name      string
		candidate string
		reason    ResolutionReason
		canonical string
	}{
		{
			name:      "existing inside",
			candidate: insideFile,
			reason:    ResolutionResolved,
			canonical: insideFile,
		},
		{
			name:      "lexical parent escape",
			candidate: filepath.Join(root, "..", "outside", "outside.tres"),
			reason:    ResolutionOutsideProject,
		},
		{
			name:      "absolute outside",
			candidate: outsideFile,
			reason:    ResolutionOutsideProject,
		},
		{
			name:      "sibling prefix collision",
			candidate: filepath.Join(root+"-old", "resource.tres"),
			reason:    ResolutionOutsideProject,
		},
		{
			name:      "missing inside",
			candidate: filepath.Join(root, "assets", "missing.tres"),
			reason:    ResolutionMissing,
		},
		{
			name:      "non regular",
			candidate: filepath.Join(root, "assets"),
			reason:    ResolutionUnsupportedTarget,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resolution := resolver.resolveCandidate("raw", test.candidate)
			if resolution.Reason != test.reason {
				t.Fatalf("reason = %q, want %q: %v", resolution.Reason, test.reason, resolution)
			}
			if resolution.Path.Canonical != test.canonical {
				t.Fatalf("canonical = %q, want %q", resolution.Path.Canonical, test.canonical)
			}
			if test.reason != ResolutionResolved && resolution.Path.Canonical != "" {
				t.Fatalf("rejected candidate exposed canonical path: %#v", resolution)
			}
		})
	}

	t.Run("symlinks", func(t *testing.T) {
		safeLink := filepath.Join(root, "safe-link")
		escapeLink := filepath.Join(root, "escape-link")
		if err := os.Symlink(filepath.Join(root, "assets"), safeLink); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		if err := os.Symlink(outside, escapeLink); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}

		symlinkTests := []struct {
			name      string
			candidate string
			reason    ResolutionReason
			canonical string
		}{
			{
				name:      "existing safe symlink",
				candidate: filepath.Join(safeLink, "inside.tres"),
				reason:    ResolutionResolved,
				canonical: insideFile,
			},
			{
				name:      "existing escaping symlink",
				candidate: filepath.Join(escapeLink, "outside.tres"),
				reason:    ResolutionOutsideProject,
			},
			{
				name:      "missing below safe symlink",
				candidate: filepath.Join(safeLink, "missing.tres"),
				reason:    ResolutionMissing,
			},
			{
				name:      "missing below escaping symlink",
				candidate: filepath.Join(escapeLink, "missing.tres"),
				reason:    ResolutionOutsideProject,
			},
		}
		for _, test := range symlinkTests {
			t.Run(test.name, func(t *testing.T) {
				resolution := resolver.resolveCandidate("raw", test.candidate)
				if resolution.Reason != test.reason || resolution.Path.Canonical != test.canonical {
					t.Fatalf("resolution = %#v, want reason %q canonical %q", resolution, test.reason, test.canonical)
				}
			})
		}
	})
}

func TestResolveCandidateInjectedFailuresAndRootTermination(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	candidate := filepath.Join(root, "assets", "secret.tres")
	permissionResolver := Resolver{
		projectRoot: root,
		stat: func(path string) (fs.FileInfo, error) {
			if path == candidate {
				return nil, fs.ErrPermission
			}
			return os.Stat(path)
		},
		evalLinks: filepath.EvalSymlinks,
	}
	resolution := permissionResolver.resolveCandidate("secret.tres", candidate)
	if resolution.Reason != ResolutionFilesystem || !errors.Is(resolution, fs.ErrPermission) {
		t.Fatalf("resolution = %#v", resolution)
	}

	var inspected []string
	missingResolver := Resolver{
		projectRoot: root,
		stat: func(path string) (fs.FileInfo, error) {
			inspected = append(inspected, path)
			return nil, fs.ErrNotExist
		},
		evalLinks: filepath.EvalSymlinks,
	}
	resolution = missingResolver.resolveCandidate("missing.tres", candidate)
	if resolution.Reason != ResolutionMissing {
		t.Fatalf("reason = %q, want %q", resolution.Reason, ResolutionMissing)
	}
	volumeRoot := filepath.VolumeName(root) + string(filepath.Separator)
	if len(inspected) == 0 || inspected[len(inspected)-1] != volumeRoot {
		t.Fatalf("inspected = %v, want last %q on %s", inspected, volumeRoot, runtime.GOOS)
	}
}

func TestResolveCandidateDoesNotRepairCase(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	correct := filepath.Join(root, "ExactCase.tres")
	wrong := filepath.Join(root, "exactcase.tres")
	writeResolverTestFile(t, correct)
	if _, err := os.Stat(wrong); err == nil {
		t.Skip("fixture filesystem is case-insensitive")
	}

	resolver, err := NewResolver(root)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	resolution := resolver.resolveCandidate("exactcase.tres", wrong)
	if resolution.Reason != ResolutionMissing {
		t.Fatalf("resolution = %#v, want missing without case repair", resolution)
	}
}

func writeResolverTestFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
