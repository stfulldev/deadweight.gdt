package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// EvalSymlinksFunc returns the path after evaluating symbolic links.
type EvalSymlinksFunc func(path string) (string, error)

// Resolver converts scene and resource references into canonical project paths.
type Resolver struct {
	projectRoot string
	stat        StatFunc
	evalLinks   EvalSymlinksFunc
}

// NewResolver constructs a resolver backed by the host filesystem.
func NewResolver(projectRoot string) (Resolver, error) {
	return NewResolverWithFS(projectRoot, os.Stat, filepath.EvalSymlinks)
}

// NewResolverWithFS constructs a resolver with injectable metadata boundaries.
func NewResolverWithFS(
	projectRoot string,
	stat StatFunc,
	evalLinks EvalSymlinksFunc,
) (Resolver, error) {
	if stat == nil {
		stat = os.Stat
	}
	if evalLinks == nil {
		evalLinks = filepath.EvalSymlinks
	}

	if projectRoot == "" || !filepath.IsAbs(projectRoot) {
		return Resolver{}, resolverFailure(
			ResolutionInvalidProjectRoot,
			projectRoot,
			projectRoot,
			nil,
		)
	}

	candidate := filepath.Clean(projectRoot)
	info, err := stat(candidate)
	if err != nil {
		return Resolver{}, projectRootInspectionFailure(projectRoot, candidate, err)
	}
	if !info.IsDir() {
		return Resolver{}, resolverFailure(
			ResolutionInvalidProjectRoot,
			projectRoot,
			candidate,
			nil,
		)
	}

	canonical, err := evalLinks(candidate)
	if err != nil {
		return Resolver{}, projectRootInspectionFailure(projectRoot, candidate, err)
	}
	canonical = filepath.Clean(canonical)
	if !filepath.IsAbs(canonical) {
		return Resolver{}, resolverFailure(
			ResolutionInvalidProjectRoot,
			projectRoot,
			canonical,
			nil,
		)
	}

	canonicalInfo, err := stat(canonical)
	if err != nil {
		return Resolver{}, projectRootInspectionFailure(projectRoot, canonical, err)
	}
	if !canonicalInfo.IsDir() {
		return Resolver{}, resolverFailure(
			ResolutionInvalidProjectRoot,
			projectRoot,
			canonical,
			nil,
		)
	}

	return Resolver{
		projectRoot: canonical,
		stat:        stat,
		evalLinks:   evalLinks,
	}, nil
}

// ProjectRoot returns the canonical absolute directory used for containment.
func (resolver Resolver) ProjectRoot() string {
	return resolver.projectRoot
}

// DisplayPath converts a canonical in-project path to normalized res:// form.
func (resolver Resolver) DisplayPath(absolute string) string {
	if resolver.projectRoot == "" || !filepath.IsAbs(absolute) || filepath.Clean(absolute) != absolute {
		return ""
	}

	relative, inside := relativeWithin(resolver.projectRoot, absolute)
	if !inside {
		return ""
	}
	if relative == "." {
		return "res://"
	}

	return "res://" + filepath.ToSlash(relative)
}

// ResolveSceneInput resolves one fatal root-scene input inside the project.
func (resolver Resolver) ResolveSceneInput(input, cwd string) (ResolvedPath, error) {
	if input == "" {
		return ResolvedPath{}, resolverFailure(
			ResolutionInvalidSceneInput,
			input,
			"",
			nil,
		)
	}

	var candidate string
	switch {
	case strings.HasPrefix(input, "res://"):
		projectRelative := filepath.FromSlash(strings.TrimPrefix(input, "res://"))
		candidate = filepath.Clean(filepath.Join(resolver.projectRoot, projectRelative))
	case strings.Contains(input, "://"):
		return ResolvedPath{}, resolverFailure(
			ResolutionInvalidSceneInput,
			input,
			input,
			nil,
		)
	case filepath.IsAbs(input):
		candidate = filepath.Clean(input)
	default:
		if cwd == "" || !filepath.IsAbs(cwd) {
			return ResolvedPath{}, resolverFailure(
				ResolutionInvalidWorkingDir,
				cwd,
				cwd,
				nil,
			)
		}
		candidate = filepath.Clean(filepath.Join(cwd, input))
	}

	if filepath.Ext(candidate) != ".tscn" {
		return ResolvedPath{}, resolverFailure(
			ResolutionInvalidSceneInput,
			input,
			candidate,
			nil,
		)
	}

	resolution := resolver.resolveCandidate(input, candidate)
	if resolution.Resolved() {
		return resolution.Path, nil
	}

	reason := resolution.Reason
	if reason == ResolutionUnsupportedTarget {
		reason = ResolutionInvalidSceneInput
	}
	return ResolvedPath{}, resolverFailure(
		reason,
		input,
		resolution.Candidate,
		resolution.Err,
	)
}

// ResolveResource resolves a declared resource into a typed nonfatal result.
func (resolver Resolver) ResolveResource(fromScene, raw string) Resolution {
	switch {
	case raw == "":
		return unresolvedResolution(ResolutionEmpty, raw, "", nil)
	case strings.HasPrefix(raw, "uid://"):
		return unresolvedResolution(ResolutionUIDOnly, raw, "", nil)
	case strings.HasPrefix(raw, "user://"):
		return unresolvedResolution(ResolutionUserData, raw, "", nil)
	}

	var candidate string
	switch {
	case strings.HasPrefix(raw, "res://"):
		projectRelative := filepath.FromSlash(strings.TrimPrefix(raw, "res://"))
		candidate = filepath.Clean(filepath.Join(resolver.projectRoot, projectRelative))
	case strings.Contains(raw, "://"):
		return unresolvedResolution(ResolutionUnsupportedTarget, raw, raw, nil)
	case filepath.IsAbs(raw):
		candidate = filepath.Clean(raw)
	default:
		declaring := resolver.validateDeclaringScene(raw, fromScene)
		if !declaring.Resolved() {
			return declaring
		}
		candidate = filepath.Clean(filepath.Join(filepath.Dir(declaring.Path.Canonical), raw))
	}

	return resolver.resolveCandidate(raw, candidate)
}

func (resolver Resolver) validateDeclaringScene(raw, fromScene string) Resolution {
	if fromScene == "" || !filepath.IsAbs(fromScene) || filepath.Clean(fromScene) != fromScene {
		return unresolvedResolution(
			ResolutionInvalidDeclaringScene,
			raw,
			fromScene,
			nil,
		)
	}
	if _, inside := relativeWithin(resolver.projectRoot, fromScene); !inside {
		return unresolvedResolution(
			ResolutionInvalidDeclaringScene,
			raw,
			fromScene,
			nil,
		)
	}

	info, err := resolver.stat(fromScene)
	if err != nil {
		reason := ResolutionFilesystem
		if errors.Is(err, fs.ErrNotExist) {
			reason = ResolutionInvalidDeclaringScene
		}

		return unresolvedResolution(reason, raw, fromScene, err)
	}
	if !info.Mode().IsRegular() {
		return unresolvedResolution(
			ResolutionInvalidDeclaringScene,
			raw,
			fromScene,
			nil,
		)
	}

	canonical, err := resolver.evalLinks(fromScene)
	if err != nil {
		reason := ResolutionFilesystem
		if errors.Is(err, fs.ErrNotExist) {
			reason = ResolutionInvalidDeclaringScene
		}

		return unresolvedResolution(reason, raw, fromScene, err)
	}
	canonical = filepath.Clean(canonical)
	if canonical != fromScene {
		return unresolvedResolution(
			ResolutionInvalidDeclaringScene,
			raw,
			fromScene,
			nil,
		)
	}
	if _, inside := relativeWithin(resolver.projectRoot, canonical); !inside {
		return unresolvedResolution(
			ResolutionInvalidDeclaringScene,
			raw,
			fromScene,
			nil,
		)
	}

	return Resolution{
		Reason: ResolutionResolved,
		Path: ResolvedPath{
			Canonical: canonical,
			Display:   resolver.DisplayPath(canonical),
			Original:  fromScene,
		},
	}
}

func relativeWithin(root, target string) (string, bool) {
	relative, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(relative) {
		return "", false
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}

	return relative, true
}

func (resolver Resolver) resolveCandidate(original, rawCandidate string) Resolution {
	candidate := filepath.Clean(rawCandidate)
	if !filepath.IsAbs(candidate) {
		return unresolvedResolution(
			ResolutionOutsideProject,
			original,
			candidate,
			nil,
		)
	}
	if _, inside := relativeWithin(resolver.projectRoot, candidate); !inside {
		return unresolvedResolution(
			ResolutionOutsideProject,
			original,
			candidate,
			nil,
		)
	}

	info, err := resolver.stat(candidate)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return resolver.resolveMissingCandidate(original, candidate, err)
		}

		return unresolvedResolution(
			ResolutionFilesystem,
			original,
			candidate,
			err,
		)
	}

	canonical, err := resolver.evalLinks(candidate)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return resolver.resolveMissingCandidate(original, candidate, err)
		}

		return unresolvedResolution(
			ResolutionFilesystem,
			original,
			candidate,
			err,
		)
	}
	canonical = filepath.Clean(canonical)
	if !filepath.IsAbs(canonical) {
		return unresolvedResolution(
			ResolutionFilesystem,
			original,
			candidate,
			fmt.Errorf("evaluated path %q is not absolute", canonical),
		)
	}
	if _, inside := relativeWithin(resolver.projectRoot, canonical); !inside {
		return unresolvedResolution(
			ResolutionOutsideProject,
			original,
			candidate,
			nil,
		)
	}
	if !info.Mode().IsRegular() {
		return unresolvedResolution(
			ResolutionUnsupportedTarget,
			original,
			candidate,
			nil,
		)
	}

	return Resolution{
		Reason: ResolutionResolved,
		Path: ResolvedPath{
			Canonical: canonical,
			Display:   resolver.DisplayPath(canonical),
			Original:  original,
		},
	}
}

func (resolver Resolver) resolveMissingCandidate(
	original string,
	candidate string,
	missingErr error,
) Resolution {
	for current := filepath.Dir(candidate); ; current = filepath.Dir(current) {
		_, err := resolver.stat(current)
		switch {
		case err == nil:
			canonicalAncestor, evalErr := resolver.evalLinks(current)
			if evalErr != nil {
				if errors.Is(evalErr, fs.ErrNotExist) {
					break
				}

				return unresolvedResolution(
					ResolutionFilesystem,
					original,
					candidate,
					evalErr,
				)
			}
			canonicalAncestor = filepath.Clean(canonicalAncestor)
			if !filepath.IsAbs(canonicalAncestor) {
				return unresolvedResolution(
					ResolutionFilesystem,
					original,
					candidate,
					fmt.Errorf("evaluated ancestor %q is not absolute", canonicalAncestor),
				)
			}

			suffix, relErr := filepath.Rel(current, candidate)
			if relErr != nil {
				return unresolvedResolution(
					ResolutionFilesystem,
					original,
					candidate,
					relErr,
				)
			}
			canonicalCandidate := filepath.Clean(filepath.Join(canonicalAncestor, suffix))
			if _, inside := relativeWithin(resolver.projectRoot, canonicalCandidate); !inside {
				return unresolvedResolution(
					ResolutionOutsideProject,
					original,
					candidate,
					nil,
				)
			}

			return unresolvedResolution(
				ResolutionMissing,
				original,
				candidate,
				missingErr,
			)
		case errors.Is(err, fs.ErrNotExist):
			// Keep walking until an existing ancestor or the filesystem root.
		default:
			return unresolvedResolution(
				ResolutionFilesystem,
				original,
				candidate,
				err,
			)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return unresolvedResolution(
				ResolutionMissing,
				original,
				candidate,
				missingErr,
			)
		}
	}
}

func unresolvedResolution(
	reason ResolutionReason,
	original string,
	candidate string,
	err error,
) Resolution {
	return Resolution{
		Reason: reason,
		Path: ResolvedPath{
			Original: original,
		},
		Candidate: candidate,
		Err:       err,
	}
}

func projectRootInspectionFailure(original, candidate string, err error) *ResolveError {
	reason := ResolutionFilesystem
	if errors.Is(err, fs.ErrNotExist) {
		reason = ResolutionInvalidProjectRoot
	}

	return resolverFailure(reason, original, candidate, err)
}

func resolverFailure(
	reason ResolutionReason,
	original string,
	candidate string,
	err error,
) *ResolveError {
	return &ResolveError{
		Reason:    reason,
		Original:  original,
		Candidate: candidate,
		Err:       err,
	}
}
