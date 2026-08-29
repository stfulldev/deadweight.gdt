package diagnostic_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

func TestCatalog(t *testing.T) {
	t.Parallel()

	want := []diagnostic.Definition{
		{Code: diagnostic.CodeUnresolvedSceneInstance, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeInheritedScene, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeUnavailableResource, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeInstancePlaceholder, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeUnsupportedResourcePath, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeUnclassifiedCustomType, Severity: diagnostic.SeverityWarning},
		{Code: diagnostic.CodeInvalidTSCNRoot, Severity: diagnostic.SeverityError},
		{Code: diagnostic.CodeSceneDependencyCycle, Severity: diagnostic.SeverityError},
		{Code: diagnostic.CodeInvalidConfiguration, Severity: diagnostic.SeverityError},
		{Code: diagnostic.CodeArithmeticOverflow, Severity: diagnostic.SeverityError},
	}

	got := diagnostic.Catalog()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Catalog() = %#v, want %#v", got, want)
	}

	got[0] = diagnostic.Definition{Code: "UNKNOWN", Severity: "unknown"}
	if diagnostic.Catalog()[0] != want[0] {
		t.Fatal("Catalog returned mutable package state")
	}

	for _, definition := range want {
		if !definition.Code.Valid() {
			t.Errorf("%q.Valid() = false", definition.Code)
		}
		severity, ok := definition.Code.Severity()
		if !ok || severity != definition.Severity {
			t.Errorf("%q.Severity() = %q, %v", definition.Code, severity, ok)
		}
	}

	if diagnostic.Code("UNKNOWN").Valid() {
		t.Fatal("unknown code is valid")
	}
	if severity, ok := diagnostic.Code("UNKNOWN").Severity(); ok || severity != "" {
		t.Fatalf("unknown severity lookup = %q, %v", severity, ok)
	}
}

func TestSeverityValid(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		severity diagnostic.Severity
		want     bool
	}{
		{severity: diagnostic.SeverityWarning, want: true},
		{severity: diagnostic.SeverityError, want: true},
		{severity: "unknown", want: false},
	} {
		if got := test.severity.Valid(); got != test.want {
			t.Errorf("%q.Valid() = %v, want %v", test.severity, got, test.want)
		}
	}
}

func TestDiagnosticValidate(t *testing.T) {
	t.Parallel()

	valid := diagnostic.Diagnostic{
		Code:        diagnostic.CodeUnresolvedSceneInstance,
		Severity:    diagnostic.SeverityWarning,
		Message:     "scene could not be resolved",
		Line:        1,
		Column:      2,
		Occurrences: 0,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name  string
		item  diagnostic.Diagnostic
		field string
	}{
		{name: "unknown code", item: diagnostic.Diagnostic{Code: "UNKNOWN", Severity: diagnostic.SeverityWarning}, field: "code"},
		{name: "unknown severity", item: diagnostic.Diagnostic{Code: diagnostic.CodeUnresolvedSceneInstance, Severity: "unknown"}, field: "severity"},
		{name: "inconsistent warning", item: diagnostic.Diagnostic{Code: diagnostic.CodeUnresolvedSceneInstance, Severity: diagnostic.SeverityError}, field: "severity"},
		{name: "inconsistent error", item: diagnostic.Diagnostic{Code: diagnostic.CodeInvalidTSCNRoot, Severity: diagnostic.SeverityWarning}, field: "severity"},
		{name: "negative line", item: diagnostic.Diagnostic{Code: diagnostic.CodeInvalidTSCNRoot, Severity: diagnostic.SeverityError, Line: -1}, field: "line"},
		{name: "negative column", item: diagnostic.Diagnostic{Code: diagnostic.CodeInvalidTSCNRoot, Severity: diagnostic.SeverityError, Column: -1}, field: "column"},
		{name: "negative occurrences", item: diagnostic.Diagnostic{Code: diagnostic.CodeInvalidTSCNRoot, Severity: diagnostic.SeverityError, Occurrences: -1}, field: "occurrences"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.item.Validate()
			var validationError *diagnostic.ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Validate() error = %v, want *diagnostic.ValidationError", err)
			}
			if validationError.Field != test.field {
				t.Fatalf("ValidationError.Field = %q, want %q", validationError.Field, test.field)
			}
		})
	}
}

func TestCodedErrorDiscovery(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("outer context: %w", testCodedError{
		code:    diagnostic.CodeSceneDependencyCycle,
		message: "cycle detected",
	})

	code, ok := diagnostic.CodeOf(err)
	if !ok || code != diagnostic.CodeSceneDependencyCycle {
		t.Fatalf("CodeOf() = %q, %v", code, ok)
	}
	if got := diagnostic.MessageOf(err); got != "cycle detected" {
		t.Fatalf("MessageOf() = %q", got)
	}

	unknown := testCodedError{code: "UNKNOWN", message: "unknown"}
	if code, ok := diagnostic.CodeOf(unknown); ok || code != "UNKNOWN" {
		t.Fatalf("CodeOf(unknown) = %q, %v", code, ok)
	}
	if code, ok := diagnostic.CodeOf(errors.New("plain")); ok || code != "" {
		t.Fatalf("CodeOf(plain) = %q, %v", code, ok)
	}
}

type testCodedError struct {
	code    diagnostic.Code
	message string
}

func (err testCodedError) Error() string {
	return fmt.Sprintf("%s: %s", err.code, err.message)
}

func (err testCodedError) DiagnosticCode() diagnostic.Code {
	return err.code
}

func (err testCodedError) DiagnosticMessage() string {
	return err.message
}
