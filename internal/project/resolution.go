package project

import "fmt"

// ResolutionReason identifies a stable path-resolution outcome.
type ResolutionReason string

const (
	ResolutionResolved              ResolutionReason = "resolved"
	ResolutionInvalidProjectRoot    ResolutionReason = "invalid_project_root"
	ResolutionInvalidWorkingDir     ResolutionReason = "invalid_working_directory"
	ResolutionInvalidSceneInput     ResolutionReason = "invalid_scene_input"
	ResolutionEmpty                 ResolutionReason = "empty"
	ResolutionUIDOnly               ResolutionReason = "uid_only"
	ResolutionUserData              ResolutionReason = "user_data"
	ResolutionUnsupportedTarget     ResolutionReason = "unsupported_target"
	ResolutionMissing               ResolutionReason = "missing"
	ResolutionOutsideProject        ResolutionReason = "outside_project"
	ResolutionFilesystem            ResolutionReason = "filesystem"
	ResolutionInvalidDeclaringScene ResolutionReason = "invalid_declaring_scene"
)

// Valid reports whether reason is part of the secure path-resolution contract.
func (reason ResolutionReason) Valid() bool {
	switch reason {
	case ResolutionResolved,
		ResolutionInvalidProjectRoot,
		ResolutionInvalidWorkingDir,
		ResolutionInvalidSceneInput,
		ResolutionEmpty,
		ResolutionUIDOnly,
		ResolutionUserData,
		ResolutionUnsupportedTarget,
		ResolutionMissing,
		ResolutionOutsideProject,
		ResolutionFilesystem,
		ResolutionInvalidDeclaringScene:
		return true
	default:
		return false
	}
}

// ResolvedPath preserves filesystem, display, and diagnostic path identities.
type ResolvedPath struct {
	Canonical string
	Display   string
	Original  string
}

// Resolution represents either a resolved resource or typed unresolved evidence.
type Resolution struct {
	Path      ResolvedPath
	Reason    ResolutionReason
	Candidate string
	Err       error
}

// Resolved reports whether the resource has a usable canonical path.
func (resolution Resolution) Resolved() bool {
	return resolution.Reason == ResolutionResolved && resolution.Path.Canonical != ""
}

// Error provides human-readable unresolved evidence without defining control flow.
func (resolution Resolution) Error() string {
	if resolution.Resolved() {
		return fmt.Sprintf("resolved resource %q", resolution.Path.Display)
	}

	return formatResolutionFailure(
		resolution.Reason,
		resolution.Path.Original,
		resolution.Candidate,
		resolution.Err,
	)
}

// Unwrap exposes an underlying filesystem cause when one exists.
func (resolution Resolution) Unwrap() error {
	return resolution.Err
}

// ResolveError is a fatal resolver-construction or root-scene failure.
type ResolveError struct {
	Reason    ResolutionReason
	Original  string
	Candidate string
	Err       error
}

func (err *ResolveError) Error() string {
	return formatResolutionFailure(err.Reason, err.Original, err.Candidate, err.Err)
}

// Unwrap exposes an underlying filesystem cause when one exists.
func (err *ResolveError) Unwrap() error {
	return err.Err
}

func formatResolutionFailure(reason ResolutionReason, original, candidate string, cause error) string {
	value := original
	if value == "" {
		value = candidate
	}
	if value == "" {
		value = "<empty>"
	}

	detail := resolutionReasonDescription(reason)
	if cause != nil {
		detail = cause.Error()
	}

	return fmt.Sprintf("path resolution %s for %q: %s", reason, value, detail)
}

func resolutionReasonDescription(reason ResolutionReason) string {
	switch reason {
	case ResolutionInvalidProjectRoot:
		return "project root must be an existing absolute directory"
	case ResolutionInvalidWorkingDir:
		return "working directory must be absolute"
	case ResolutionInvalidSceneInput:
		return "root scene must be a regular .tscn file"
	case ResolutionEmpty:
		return "path is empty"
	case ResolutionUIDOnly:
		return "uid-only references are unsupported without Godot metadata"
	case ResolutionUserData:
		return "user:// references are outside the project resource tree"
	case ResolutionUnsupportedTarget:
		return "target scheme or file type is unsupported"
	case ResolutionMissing:
		return "target does not exist"
	case ResolutionOutsideProject:
		return "target is outside the project root"
	case ResolutionFilesystem:
		return "filesystem metadata is unavailable"
	case ResolutionInvalidDeclaringScene:
		return "declaring scene must be a canonical in-project file"
	case ResolutionResolved:
		return "resolved"
	default:
		return "unknown resolution reason"
	}
}
