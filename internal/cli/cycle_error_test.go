package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stfulldev/deadweight.gdt/internal/analysis"
)

func TestExecuteMapsAnalysisCycleToFatalExitCodeAndCompleteChain(t *testing.T) {
	root := &cobra.Command{
		Use:           "test",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return &analysis.CycleError{
				Canonical: []string{"/project/a.tscn", "/project/b.tscn", "/project/a.tscn"},
				Display:   []string{"res://a.tscn", "res://b.tscn", "res://a.tscn"},
			}
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute(root, nil, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2", exitCode)
	}
	want := "ERROR SB2002: scene dependency cycle\n\n  res://a.tscn\n  → res://b.tscn\n  → res://a.tscn\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}
