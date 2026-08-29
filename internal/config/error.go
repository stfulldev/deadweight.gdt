package config

import (
	"fmt"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

// ErrorReason identifies a stable configuration failure boundary.
type ErrorReason string

const (
	ReasonMissingExplicit ErrorReason = "missing_explicit"
	ReasonNotRegular      ErrorReason = "not_regular"
	ReasonFilesystem      ErrorReason = "filesystem"
	ReasonDecode          ErrorReason = "decode"
	ReasonValidation      ErrorReason = "validation"
)

// Valid reports whether reason is part of the version-one error contract.
func (reason ErrorReason) Valid() bool {
	switch reason {
	case ReasonMissingExplicit,
		ReasonNotRegular,
		ReasonFilesystem,
		ReasonDecode,
		ReasonValidation:
		return true
	default:
		return false
	}
}

// Error is a typed configuration failure with source and field evidence.
type Error struct {
	Reason ErrorReason
	Source string
	Field  string
	Detail string
	Err    error
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %s", diagnostic.CodeInvalidConfiguration, err.DiagnosticMessage())
}

// DiagnosticCode exposes the stable configuration diagnostic identifier.
func (err *Error) DiagnosticCode() diagnostic.Code {
	return diagnostic.CodeInvalidConfiguration
}

// DiagnosticMessage returns deterministic, code-free user-facing text.
func (err *Error) DiagnosticMessage() string {
	message := "invalid configuration"
	if err.Source != "" {
		message += fmt.Sprintf(" %q", err.Source)
	}
	if err.Field != "" {
		message += fmt.Sprintf(" field %q", err.Field)
	}

	return message + ": " + err.description()
}

// Unwrap exposes the underlying filesystem or JSON cause when present.
func (err *Error) Unwrap() error {
	return err.Err
}

func (err *Error) description() string {
	if err.Detail != "" {
		return err.Detail
	}
	switch err.Reason {
	case ReasonMissingExplicit:
		return "explicit config does not exist"
	case ReasonNotRegular:
		return "config path is not a regular file"
	case ReasonFilesystem:
		return "cannot access config file"
	case ReasonDecode:
		return "invalid JSON document"
	case ReasonValidation:
		return "invalid value"
	default:
		return "configuration error"
	}
}

func configError(reason ErrorReason, source, field, detail string, cause error) *Error {
	return &Error{
		Reason: reason,
		Source: source,
		Field:  field,
		Detail: detail,
		Err:    cause,
	}
}
