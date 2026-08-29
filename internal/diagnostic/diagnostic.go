package diagnostic

import "fmt"

// Severity describes whether a diagnostic is informational or fatal to analysis.
type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Valid reports whether severity is part of the MVP diagnostic taxonomy.
func (severity Severity) Valid() bool {
	return severity == SeverityWarning || severity == SeverityError
}

// Diagnostic is a stable, source-aware analyzer message.
type Diagnostic struct {
	Code        Code     `json:"code"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	File        string   `json:"file,omitempty"`
	Line        int      `json:"line,omitempty"`
	Column      int      `json:"column,omitempty"`
	Resource    string   `json:"resource,omitempty"`
	Occurrences int64    `json:"occurrences,omitempty"`
}

// ValidationError reports an invalid diagnostic field or field combination.
type ValidationError struct {
	Field   string
	Message string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("diagnostic %s: %s", err.Field, err.Message)
}

// Validate checks that a diagnostic satisfies the MVP domain contract.
func (item Diagnostic) Validate() error {
	if !item.Code.Valid() {
		return &ValidationError{Field: "code", Message: fmt.Sprintf("unknown code %q", item.Code)}
	}
	if !item.Severity.Valid() {
		return &ValidationError{Field: "severity", Message: fmt.Sprintf("unknown severity %q", item.Severity)}
	}

	wantSeverity, _ := item.Code.Severity()
	if item.Severity != wantSeverity {
		return &ValidationError{
			Field:   "severity",
			Message: fmt.Sprintf("code %q requires severity %q, got %q", item.Code, wantSeverity, item.Severity),
		}
	}
	if item.Line < 0 {
		return &ValidationError{Field: "line", Message: "must be non-negative"}
	}
	if item.Column < 0 {
		return &ValidationError{Field: "column", Message: "must be non-negative"}
	}
	if item.Occurrences < 0 {
		return &ValidationError{Field: "occurrences", Message: "must be non-negative"}
	}

	return nil
}
