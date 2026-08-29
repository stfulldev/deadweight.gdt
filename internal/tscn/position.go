package tscn

import (
	"fmt"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

const invalidSceneCode = diagnostic.CodeInvalidTSCNRoot

// Position identifies a one-based source location.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ParseError is a source-aware failure to parse the supported TSCN subset.
type ParseError struct {
	Code     diagnostic.Code
	Source   string
	Position Position
	Message  string
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("%s: %s: %s", err.location(), err.Code, err.Message)
}

// DiagnosticCode exposes the stable code without requiring message parsing.
func (err *ParseError) DiagnosticCode() diagnostic.Code {
	return err.Code
}

// DiagnosticMessage returns source-aware text without duplicating the code prefix.
func (err *ParseError) DiagnosticMessage() string {
	return fmt.Sprintf("%s: %s", err.location(), err.Message)
}

func (err *ParseError) location() string {
	location := err.Source
	if location == "" {
		location = "<input>"
	}

	if err.Position.Line > 0 {
		location = fmt.Sprintf("%s:%d:%d", location, err.Position.Line, err.Position.Column)
	}

	return location
}

func newParseError(source string, position Position, format string, args ...any) *ParseError {
	return &ParseError{
		Code:     invalidSceneCode,
		Source:   source,
		Position: position,
		Message:  fmt.Sprintf(format, args...),
	}
}
