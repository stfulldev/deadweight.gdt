package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Execute([]string{"--version"}, &stdout, &stderr, cli.BuildInfo{Version: "0.1.0-test"})

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if stdout.String() != "deadweight.gdt 0.1.0-test\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestPresetsShowSteamDeck(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Execute([]string{"presets", "show", "steam-deck"}, &stdout, &stderr, cli.BuildInfo{Version: "test"})

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}

	wantFragments := []string{
		"Preset:      Steam Deck",
		"Status:      heuristic",
		"Renderer:    Forward+",
		"Nodes                         3,000",
		"Shadow lights                     8",
		"not a performance guarantee",
	}
	for _, fragment := range wantFragments {
		if !strings.Contains(stdout.String(), fragment) {
			t.Errorf("stdout does not contain %q:\n%s", fragment, stdout.String())
		}
	}
}

func TestUnknownPresetUsesFatalUsageExitCode(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Execute([]string{"presets", "show", "unknown"}, &stdout, &stderr, cli.BuildInfo{Version: "test"})

	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2", exitCode)
	}
	if !strings.Contains(stderr.String(), "available presets: mobile, steam-deck, desktop") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestAnalyzerCommandsFailClearlyUntilImplemented(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Execute([]string{"inspect", "scene.tscn"}, &stdout, &stderr, cli.BuildInfo{Version: "test"})

	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2", exitCode)
	}
	if !strings.Contains(stderr.String(), "scene analyzer is not implemented yet") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
