package report

import (
	"errors"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
)

func TestErrorRendering(t *testing.T) {
	t.Parallel()

	cycle := &analysis.CycleError{Display: []string{
		"res://scenes/A.tscn",
		"res://scenes/B.tscn",
		"res://scenes/C.tscn",
		"res://scenes/A.tscn",
	}}
	wantCycle := "ERROR SB2002: scene dependency cycle\n\n" +
		"  res://scenes/A.tscn\n" +
		"  → res://scenes/B.tscn\n" +
		"  → res://scenes/C.tscn\n" +
		"  → res://scenes/A.tscn\n"
	if got := Error(cycle); got != wantCycle {
		t.Fatalf("cycle error = %q, want %q", got, wantCycle)
	}

	if got, want := Error(errors.New("fatal failure")), "ERROR: fatal failure\n"; got != want {
		t.Fatalf("plain error = %q, want %q", got, want)
	}
	if got := Error(nil); got != "" {
		t.Fatalf("nil error = %q", got)
	}
	for _, output := range []string{Error(cycle), Error(errors.New("fatal failure"))} {
		for _, forbidden := range []string{"\x1b[", "goroutine ", "runtime/debug.Stack", ".go:"} {
			if strings.Contains(output, forbidden) {
				t.Fatalf("error output contains %q: %q", forbidden, output)
			}
		}
	}
}
