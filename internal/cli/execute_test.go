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

func TestRootHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Execute([]string{"--help"}, &stdout, &stderr, cli.BuildInfo{Version: "test"})

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, fragment := range []string{"inspect", "check", "presets", "--version"} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Errorf("help does not contain %q:\n%s", fragment, stdout.String())
		}
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
		"Stability:   experimental",
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

func TestPresetsListUsesProductOrderAndLifecycleLabels(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Execute([]string{"presets"}, &stdout, &stderr, cli.BuildInfo{Version: "test"})

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Built-in presets (heuristic, experimental)") {
		t.Fatalf("stdout does not identify lifecycle labels:\n%s", output)
	}

	mobile := strings.Index(output, "\nmobile\n")
	steamDeck := strings.Index(output, "\nsteam-deck\n")
	desktop := strings.Index(output, "\ndesktop\n")
	if mobile < 0 || steamDeck <= mobile || desktop <= steamDeck {
		t.Fatalf("preset IDs are not in product order:\n%s", output)
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
	want := `unknown preset "unknown"; available presets: mobile, steam-deck, desktop`
	if !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
