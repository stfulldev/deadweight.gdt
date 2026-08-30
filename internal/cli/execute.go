package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/report"
)

// BuildInfo contains values injected at build time.
type BuildInfo struct {
	Version string
}

// Application is the command-facing boundary for the standalone CLI flows.
type Application interface {
	Inspect(application.InspectRequest) (application.InspectResult, error)
	Tree(application.TreeRequest) (application.TreeResult, error)
	Check(application.CheckRequest) (application.CheckResult, error)
	Diff(application.DiffRequest) (application.DiffResult, error)
	ListPresets() (application.PresetListResult, error)
	ShowPreset(string) (application.PresetShowResult, error)
	ListProfiles(application.ProfileRequest) (application.ProfileListResult, error)
	ShowProfile(application.ProfileShowRequest) (application.ProfileShowResult, error)
}

type globalOptions struct {
	project string
	config  string
	noColor bool
	version string
	runtime PresentationRuntime
}

// PresentationRuntime contains environment effects used only for color policy.
type PresentationRuntime struct {
	LookupEnv  func(string) (string, bool)
	IsTerminal func(io.Writer) bool
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
	return ExecuteWithApplicationAndRuntime(
		args,
		stdout,
		stderr,
		info,
		service,
		PresentationRuntime{},
	)
}

// ExecuteWithApplicationAndRuntime runs the CLI with injected application and
// presentation environment effects.
func ExecuteWithApplicationAndRuntime(
	args []string,
	stdout, stderr io.Writer,
	info BuildInfo,
	service Application,
	runtime PresentationRuntime,
) int {
	return execute(NewRootWithApplicationAndRuntime(info, service, runtime), args, stdout, stderr)
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
		var failure *presentationError
		if errors.As(err, &failure) && failure.format == presentationJSON {
			rendered, renderErr := report.ErrorJSON(failure.err, failure.options)
			if renderErr == nil {
				_, _ = fmt.Fprint(stderr, rendered)
				return 2
			}
		}
		_, _ = fmt.Fprint(stderr, report.Error(err))
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
	return NewRootWithApplicationAndRuntime(info, service, PresentationRuntime{})
}

// NewRootWithApplicationAndRuntime constructs a fresh command tree with all
// application and presentation effects injected.
func NewRootWithApplicationAndRuntime(
	info BuildInfo,
	service Application,
	runtime PresentationRuntime,
) *cobra.Command {
	if service == nil {
		service = application.NewDefault()
	}
	options := &globalOptions{version: info.Version, runtime: normalizePresentationRuntime(runtime)}
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
		newTreeCommand(service, options),
		newCheckCommand(service, options),
		newDiffCommand(service, options),
		newPresetsCommand(service, options),
		newProfilesCommand(service, options),
	)

	return root
}

func normalizePresentationRuntime(runtime PresentationRuntime) PresentationRuntime {
	if runtime.LookupEnv == nil {
		runtime.LookupEnv = os.LookupEnv
	}
	if runtime.IsTerminal == nil {
		runtime.IsTerminal = isTerminalWriter
	}

	return runtime
}

func isTerminalWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

func (options *globalOptions) reportOptions(writer io.Writer) report.Options {
	color := !options.noColor && options.runtime.IsTerminal(writer)
	if _, present := options.runtime.LookupEnv("NO_COLOR"); present {
		color = false
	}

	return report.Options{Version: options.version, Color: color}
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
