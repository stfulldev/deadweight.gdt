package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

// BuildInfo contains values injected at build time.
type BuildInfo struct {
	Version string
}

// Application is the command-facing boundary for the four MVP flows.
type Application interface {
	Inspect(application.InspectRequest) (application.InspectResult, error)
	Check(application.CheckRequest) (application.CheckResult, error)
	ListPresets() (application.PresetListResult, error)
	ShowPreset(string) (application.PresetShowResult, error)
}

type globalOptions struct {
	project string
	config  string
	noColor bool
}

type exitSignal struct {
	code int
}

func (signal *exitSignal) Error() string {
	return fmt.Sprintf("command outcome requires exit code %d", signal.code)
}

// Execute runs the CLI and maps command failures to the MVP usage/fatal exit code.
func Execute(args []string, stdout, stderr io.Writer, info BuildInfo) int {
	return ExecuteWithApplication(args, stdout, stderr, info, application.NewDefault())
}

// ExecuteWithApplication runs the CLI with an injected application service.
func ExecuteWithApplication(
	args []string,
	stdout, stderr io.Writer,
	info BuildInfo,
	service Application,
) int {
	return execute(NewRootWithApplication(info, service), args, stdout, stderr)
}

func execute(root *cobra.Command, args []string, stdout, stderr io.Writer) int {
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.Execute(); err != nil {
		var signal *exitSignal
		if errors.As(err, &signal) {
			return signal.code
		}
		if code, ok := diagnostic.CodeOf(err); ok {
			_, _ = fmt.Fprintf(stderr, "ERROR %s: %s\n", code, diagnostic.MessageOf(err))
		} else {
			_, _ = fmt.Fprintf(stderr, "ERROR: %v\n", err)
		}
		return 2
	}

	return 0
}

// NewRoot constructs a fresh command tree suitable for execution and tests.
func NewRoot(info BuildInfo) *cobra.Command {
	return NewRootWithApplication(info, application.NewDefault())
}

// NewRootWithApplication constructs a fresh command tree with injected flows.
func NewRootWithApplication(info BuildInfo, service Application) *cobra.Command {
	if service == nil {
		service = application.NewDefault()
	}
	options := &globalOptions{}
	root := &cobra.Command{
		Use:           "deadweight.gdt",
		Short:         "Static scene-complexity budgets for Godot 4 text scenes",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       info.Version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetVersionTemplate("deadweight.gdt {{.Version}}\n")
	root.PersistentFlags().StringVar(&options.project, "project", "", "explicit Godot project root or project.godot")
	root.PersistentFlags().StringVar(&options.config, "config", "", "explicit configuration file")
	root.PersistentFlags().BoolVar(&options.noColor, "no-color", false, "disable ANSI color")

	root.AddCommand(
		newInspectCommand(service, options),
		newCheckCommand(service, options),
		newPresetsCommand(service),
	)

	return root
}

func exitForEvaluation(status budget.Status) error {
	switch status {
	case budget.StatusPassed:
		return nil
	case budget.StatusFailed:
		return &exitSignal{code: 1}
	case budget.StatusIncomplete:
		return &exitSignal{code: 3}
	default:
		return fmt.Errorf("invalid check outcome %q", status)
	}
}
