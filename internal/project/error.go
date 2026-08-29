package project

import "fmt"

// ErrorReason identifies a stable project-discovery failure class.
type ErrorReason string

const (
	ReasonInvalidWorkingDirectory ErrorReason = "invalid_working_directory"
	ReasonInvalidSceneInput       ErrorReason = "invalid_scene_input"
	ReasonInvalidExplicitProject  ErrorReason = "invalid_explicit_project"
	ReasonFilesystem              ErrorReason = "filesystem"
	ReasonProjectNotFound         ErrorReason = "project_not_found"
)

// Valid reports whether reason is part of the project-discovery contract.
func (reason ErrorReason) Valid() bool {
	switch reason {
	case ReasonInvalidWorkingDirectory,
		ReasonInvalidSceneInput,
		ReasonInvalidExplicitProject,
		ReasonFilesystem,
		ReasonProjectNotFound:
		return true
	default:
		return false
	}
}

// Error is a typed, path-aware project-discovery failure.
type Error struct {
	Reason ErrorReason
	Path   string
	Detail string
	Err    error
}

func (err *Error) Error() string {
	switch err.Reason {
	case ReasonInvalidWorkingDirectory:
		return fmt.Sprintf("invalid working directory %q: %s", err.Path, err.description())
	case ReasonInvalidSceneInput:
		return fmt.Sprintf("invalid scene input %q: %s", err.Path, err.description())
	case ReasonInvalidExplicitProject:
		return fmt.Sprintf("invalid --project %q: %s", err.Path, err.description())
	case ReasonFilesystem:
		return fmt.Sprintf("cannot inspect %q: %s", err.Path, err.description())
	case ReasonProjectNotFound:
		return fmt.Sprintf(
			"no regular project.godot found from %q; run from inside a Godot project or pass --project",
			err.Path,
		)
	default:
		return fmt.Sprintf("project discovery failed for %q: %s", err.Path, err.description())
	}
}

// Unwrap exposes an underlying filesystem cause when one exists.
func (err *Error) Unwrap() error {
	return err.Err
}

func (err *Error) description() string {
	if err.Detail != "" {
		return err.Detail
	}
	if err.Err != nil {
		return err.Err.Error()
	}

	return "invalid value"
}
