package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

func TestExecuteRendersCodedErrorOnce(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{
		Use:           "test",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return fmt.Errorf("wrapped: %w", codedTestError{})
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute(root, nil, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2", exitCode)
	}
	if got := stderr.String(); got != "ERROR SB2002: scene dependency cycle\n" {
		t.Fatalf("stderr = %q", got)
	}
	if strings.Count(stderr.String(), string(diagnostic.CodeSceneDependencyCycle)) != 1 {
		t.Fatalf("stderr duplicates diagnostic code: %q", stderr.String())
	}
}

type codedTestError struct{}

func (codedTestError) Error() string {
	return "SB2002: scene dependency cycle"
}

func (codedTestError) DiagnosticCode() diagnostic.Code {
	return diagnostic.CodeSceneDependencyCycle
}

func (codedTestError) DiagnosticMessage() string {
	return "scene dependency cycle"
}
