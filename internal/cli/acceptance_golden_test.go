package cli_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
)

func TestAcceptanceGoldens(t *testing.T) {
	fixtures, goldenDir := acceptancePaths(t)
	tests := []struct {
		name       string
		group      string
		args       func(string) []string
		terminal   bool
		noColorEnv bool
		orphan     bool
	}{
		{
			name: "inspect_complete", group: "complete",
			args: func(root string) []string { return []string{"--project", root, "inspect", "res://nested.tscn"} },
		},
		{
			name: "inspect_partial", group: "unresolved",
			args: func(root string) []string { return []string{"--project", root, "inspect", "res://missing-tscn.tscn"} },
		},
		{
			name: "inspect_approximate", group: "inherited",
			args: func(root string) []string { return []string{"--project", root, "inspect", "res://zombie.tscn"} },
		},
		{
			name: "check_pass", group: "complete",
			args: func(root string) []string {
				return []string{"--project", root, "check", "res://simple.tscn", "--budget", "nodes=3"}
			},
		},
		{
			name: "check_fail", group: "complete",
			args: func(root string) []string {
				return []string{"--project", root, "check", "res://simple.tscn", "--budget", "nodes=2"}
			},
		},
		{
			name: "check_partial_rejected", group: "unresolved",
			args: func(root string) []string {
				return []string{"--project", root, "check", "res://missing-tscn.tscn", "--budget", "nodes=10", "--fail-on-partial"}
			},
		},
		{
			name: "cycle_error", group: "cyclic",
			args: func(root string) []string { return []string{"--project", root, "inspect", "res://A.tscn"} },
		},
		{
			name: "config_error", group: "malformed",
			args: func(root string) []string {
				return []string{"--project", root, "--config", filepath.Join(root, "invalid-config.json"), "inspect", "res://format2.tscn"}
			},
		},
		{
			name: "missing_project", orphan: true,
			args: func(root string) []string { return []string{"inspect", filepath.Join(root, "orphan.tscn")} },
		},
		{
			name: "no_color", group: "complete", terminal: true, noColorEnv: true,
			args: func(root string) []string { return []string{"--project", root, "inspect", "res://simple.tscn"} },
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			root := filepath.Join(fixtures, test.group)
			if test.orphan {
				root = t.TempDir()
				if err := os.WriteFile(
					filepath.Join(root, "orphan.tscn"),
					[]byte("[gd_scene format=3]\n[node name=\"Root\" type=\"Node3D\"]\n"),
					0o600,
				); err != nil {
					t.Fatalf("write orphan scene: %v", err)
				}
			}

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := cli.ExecuteWithApplicationAndRuntime(
				test.args(root),
				&stdout,
				&stderr,
				cli.BuildInfo{Version: "0.1.0-test"},
				application.NewDefault(),
				cli.PresentationRuntime{
					LookupEnv: func(key string) (string, bool) {
						if key == "NO_COLOR" && test.noColorEnv {
							return "1", true
						}
						return "", false
					},
					IsTerminal: func(io.Writer) bool { return test.terminal },
				},
			)

			got := acceptanceSnapshot(t, root, exitCode, stdout.String(), stderr.String())
			goldenPath := filepath.Join(goldenDir, test.name+".golden")
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v\n--- got ---\n%s", goldenPath, err, got)
			}
			if got != string(want) {
				t.Fatalf("golden mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
			}
			if strings.Contains(got, "\x1b[") {
				t.Fatalf("acceptance golden contains ANSI: %q", got)
			}
		})
	}
}

func acceptancePaths(t *testing.T) (string, string) {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve acceptance source path")
	}
	packageDir := filepath.Dir(source)
	return filepath.Clean(filepath.Join(packageDir, "..", "..", "testdata", "projects")),
		filepath.Join(packageDir, "testdata", "golden", "acceptance")
}

func acceptanceSnapshot(t *testing.T, root string, exitCode int, stdout, stderr string) string {
	t.Helper()

	return fmt.Sprintf(
		"exit: %d\n\n[stdout]\n%s[stderr]\n%s",
		exitCode,
		normalizeAcceptanceOutput(t, root, stdout),
		normalizeAcceptanceOutput(t, root, stderr),
	)
}

func normalizeAcceptanceOutput(t *testing.T, root, output string) string {
	t.Helper()
	paths := []string{filepath.Clean(root), filepath.ToSlash(filepath.Clean(root))}
	if canonical, err := filepath.EvalSymlinks(root); err == nil {
		paths = append(paths, filepath.Clean(canonical), filepath.ToSlash(filepath.Clean(canonical)))
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		escaped := strings.ReplaceAll(path, `\`, `\\`)
		output = strings.ReplaceAll(output, escaped, "<PROJECT>")
		output = strings.ReplaceAll(output, path, "<PROJECT>")
	}
	output = strings.ReplaceAll(output, `\\`, "/")
	output = strings.ReplaceAll(output, `\`, "/")
	for _, path := range paths {
		if strings.Contains(output, path) || strings.Contains(output, strings.ReplaceAll(path, `\`, `\\`)) {
			t.Fatalf("snapshot leaked fixture root: %q", output)
		}
	}
	return output
}

func TestNormalizeAcceptanceOutputWindowsQuotedPath(t *testing.T) {
	root := `D:\a\deadweight.gdt\deadweight.gdt\testdata\projects\malformed`
	output := `ERROR: invalid configuration "D:\\a\\deadweight.gdt\\deadweight.gdt\\testdata\\projects\\malformed\\invalid-config.json"`
	want := `ERROR: invalid configuration "<PROJECT>/invalid-config.json"`

	if got := normalizeAcceptanceOutput(t, root, output); got != want {
		t.Fatalf("normalizeAcceptanceOutput() = %q, want %q", got, want)
	}
}
