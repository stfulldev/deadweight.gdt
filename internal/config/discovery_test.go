package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

func TestDiscoverUsesExplicitThenImplicitPriority(t *testing.T) {
	root := t.TempDir()
	implicit := filepath.Join(root, DefaultFilename)
	explicit := filepath.Join(root, "custom.json")
	writeConfigTestFile(t, implicit, `{"version":1,"preset":"mobile"}`)
	writeConfigTestFile(t, explicit, `{"version":1,"preset":"desktop"}`)

	source, present, err := Discover(root, explicit)
	if err != nil {
		t.Fatalf("Discover(explicit) error = %v", err)
	}
	if !present || !source.Explicit || source.Path != explicit {
		t.Fatalf("Discover(explicit) = %#v, %v", source, present)
	}

	source, present, err = Discover(root, "")
	if err != nil {
		t.Fatalf("Discover(implicit) error = %v", err)
	}
	if !present || source.Explicit || source.Path != implicit {
		t.Fatalf("Discover(implicit) = %#v, %v", source, present)
	}
}

func TestDiscoverTreatsMissingImplicitAsAbsent(t *testing.T) {
	source, present, err := Discover(t.TempDir(), "")
	if err != nil || present || source != (Source{}) {
		t.Fatalf("Discover() = %#v, %v, %v; want absent", source, present, err)
	}
}

func TestDiscoverRejectsInvalidSelectedPaths(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name     string
		explicit string
		prepare  func(string) string
		reason   ErrorReason
	}{
		{
			name: "missing explicit", explicit: filepath.Join(root, "missing.json"),
			prepare: func(path string) string { return path }, reason: ReasonMissingExplicit,
		},
		{
			name: "explicit directory", explicit: filepath.Join(root, "directory"),
			prepare: func(path string) string {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
				return path
			}, reason: ReasonNotRegular,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := test.prepare(test.explicit)
			source, present, err := Discover(root, path)
			if err == nil || present || source != (Source{}) {
				t.Fatalf("Discover() = %#v, %v, %v", source, present, err)
			}
			_ = requireConfigError(t, err, test.reason, path, "")
		})
	}

	implicitDirectoryRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(implicitDirectoryRoot, DefaultFilename), 0o700); err != nil {
		t.Fatal(err)
	}
	_, _, err := Discover(implicitDirectoryRoot, "")
	_ = requireConfigError(t, err, ReasonNotRegular, filepath.Join(implicitDirectoryRoot, DefaultFilename), "")
}

func TestReadAndLoadReturnOwnedValuesAndTypedReadFailures(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, DefaultFilename)
	writeConfigTestFile(t, path, `{"version":1,"profile":"shipping","profiles":{"shipping":{"extends":"future-base"}}}`)

	configuration, source, present, err := Load(root, "")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !present || source.Path != path || source.Explicit || configuration.Profile == nil || *configuration.Profile != "shipping" {
		t.Fatalf("Load() = %#v, %#v, %v", configuration, source, present)
	}
	if profile := configuration.Profiles["shipping"]; profile.Extends == nil || *profile.Extends != "future-base" {
		t.Fatalf("profile = %#v", profile)
	}

	discovered, present, err := Discover(root, "")
	if err != nil || !present {
		t.Fatalf("Discover() = %#v, %v, %v", discovered, present, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	got, err := Read(discovered)
	if err == nil || !reflect.DeepEqual(got, Config{}) {
		t.Fatalf("Read(disappeared) = %#v, %v", got, err)
	}
	_ = requireConfigError(t, err, ReasonFilesystem, path, "")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("Read error = %T, want wrapped *os.PathError", err)
	}
}

func TestDiscoverRejectsMissingProjectRootForImplicitConfig(t *testing.T) {
	_, _, err := Discover("", "")
	_ = requireConfigError(t, err, ReasonValidation, "", "project_root")
}

func writeConfigTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func requireConfigError(t *testing.T, err error, reason ErrorReason, source, field string) *Error {
	t.Helper()
	var configErr *Error
	if !errors.As(err, &configErr) {
		t.Fatalf("error = %T %v, want *config.Error", err, err)
	}
	if configErr.Reason != reason || configErr.Source != source || configErr.Field != field {
		t.Fatalf("config error = %#v, want reason/source/field %q/%q/%q", configErr, reason, source, field)
	}
	if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeInvalidConfiguration {
		t.Fatalf("diagnostic code = %q, %v", code, ok)
	}
	if diagnostic.MessageOf(err) != configErr.DiagnosticMessage() {
		t.Fatalf("diagnostic message = %q, want %q", diagnostic.MessageOf(err), configErr.DiagnosticMessage())
	}

	return configErr
}
