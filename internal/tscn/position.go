package tscn

import "fmt"

const invalidSceneCode = "SB2001"

// Position identifies a one-based source location.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ParseError is a source-aware failure to parse the supported TSCN subset.
type ParseError struct {
	Code     string
	Source   string
	Position Position
	Message  string
}

func (err *ParseError) Error() string {
	location := err.Source
	if location == "" {
		location = "<input>"
	}

	if err.Position.Line > 0 {
		location = fmt.Sprintf("%s:%d:%d", location, err.Position.Line, err.Position.Column)
	}

	return fmt.Sprintf("%s: %s: %s", location, err.Code, err.Message)
}

func newParseError(source string, position Position, format string, args ...any) *ParseError {
	return &ParseError{
		Code:     invalidSceneCode,
		Source:   source,
		Position: position,
		Message:  fmt.Sprintf(format, args...),
	}
}
