package app

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

func TestDiffReadsOnlyRequestedReportsAndAvoidsSceneEffects(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile(filepath.Join("..", "report", "testdata", "golden", "json", "inspect_complete.golden"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var reads []string
	application := New(Dependencies{
		ReadFile: func(filename string) ([]byte, error) {
			reads = append(reads, filename)
			return append([]byte(nil), contents...), nil
		},
		WorkingDirectory: func() (string, error) {
			t.Fatal("Diff consulted the working directory")
			return "", nil
		},
		FindProject: func(_ project.Request) (project.Root, error) {
			t.Fatal("Diff attempted project discovery")
			return project.Root{}, nil
		},
	})
	result, err := application.Diff(DiffRequest{
		Before: "before.json", After: "after.json",
		Policy: reportdiff.Policy{MetricIncreases: []metrics.Name{metrics.Nodes}},
	})
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if !reflect.DeepEqual(reads, []string{"before.json", "after.json"}) {
		t.Fatalf("reads = %#v", reads)
	}
	if result.Comparison.Changed || !result.Comparison.Enforcement.Enabled {
		t.Fatalf("result = %#v", result)
	}
}

func TestDiffValidatesPolicyBeforeReadingAndBoundsInput(t *testing.T) {
	t.Parallel()

	reads := 0
	application := New(Dependencies{ReadFile: func(string) ([]byte, error) {
		reads++
		return make([]byte, reportdiff.MaxInputBytes+1), nil
	}})
	_, err := application.Diff(DiffRequest{
		Before: "before.json", After: "after.json",
		Policy: reportdiff.Policy{MetricIncreases: []metrics.Name{metrics.Nodes, metrics.Nodes}},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate") || reads != 0 {
		t.Fatalf("invalid policy error/reads = %v/%d", err, reads)
	}
	_, err = application.Diff(DiffRequest{Before: "before.json", After: "after.json"})
	if err == nil || !strings.Contains(err.Error(), "input limit") || reads != 1 {
		t.Fatalf("oversized error/reads = %v/%d", err, reads)
	}
}

func TestDiffReportsReadAndDecodeContext(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("disk unavailable")
	application := New(Dependencies{ReadFile: func(filename string) ([]byte, error) {
		if filename == "before.json" {
			return []byte(`{"broken":`), nil
		}
		return nil, wantErr
	}})
	if _, err := application.Diff(DiffRequest{Before: "before.json", After: "after.json"}); err == nil || !strings.Contains(err.Error(), `read baseline "before.json"`) || !strings.Contains(err.Error(), "decode report JSON") {
		t.Fatalf("baseline decode error = %v", err)
	}
	application = New(Dependencies{ReadFile: func(filename string) ([]byte, error) {
		if filename == "before.json" {
			contents, readErr := os.ReadFile(filepath.Join("..", "report", "testdata", "golden", "json", "inspect_complete.golden"))
			return contents, readErr
		}
		return nil, wantErr
	}})
	if _, err := application.Diff(DiffRequest{Before: "before.json", After: "after.json"}); !errors.Is(err, wantErr) || !strings.Contains(err.Error(), `read candidate "after.json"`) {
		t.Fatalf("candidate read error = %v", err)
	}
}
