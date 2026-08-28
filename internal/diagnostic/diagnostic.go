package diagnostic

// Severity describes whether a diagnostic is informational or fatal to analysis.
type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Diagnostic is a stable, source-aware analyzer message.
type Diagnostic struct {
	Code        string   `json:"code"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	File        string   `json:"file,omitempty"`
	Line        int      `json:"line,omitempty"`
	Column      int      `json:"column,omitempty"`
	Resource    string   `json:"resource,omitempty"`
	Occurrences int64    `json:"occurrences,omitempty"`
}
