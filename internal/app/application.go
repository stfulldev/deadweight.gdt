package app

import (
	"fmt"
	"io"
	"os"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

// SceneResolver is the secure root/resource path boundary required by the app.
type SceneResolver interface {
	analysis.ResourceResolver
	ResolveSceneInput(input, workingDirectory string) (project.ResolvedPath, error)
}

// Dependencies contains replaceable application effects. New supplies a
// production default for every nil field.
type Dependencies struct {
	WorkingDirectory   func() (string, error)
	FindProject        func(project.Request) (project.Root, error)
	NewResolver        func(projectRoot string) (SceneResolver, error)
	LoadConfig         func(projectRoot, explicitPath string) (config.Config, config.Source, bool, error)
	Analyze            func(SceneResolver, project.ResolvedPath) (analysis.RecursiveResult, error)
	ResolvePolicy      func(string, config.Config, policy.Selector, []string) (policy.Effective, error)
	ResolvePartial     func(bool, budget.PartialOverride) (bool, error)
	Evaluate           func(metrics.Values, budget.Limits, analysis.Reliability, bool) (budget.Evaluation, error)
	LoadBuiltInPresets func() (preset.Catalog, error)
	ReadFile           func(string) ([]byte, error)
}

// Application executes the standalone CLI command flows.
type Application struct {
	dependencies Dependencies
}

// New constructs an application, filling omitted effects with production
// implementations.
func New(dependencies Dependencies) *Application {
	if dependencies.WorkingDirectory == nil {
		dependencies.WorkingDirectory = os.Getwd
	}
	if dependencies.FindProject == nil {
		finder := project.NewFinder()
		dependencies.FindProject = finder.Find
	}
	if dependencies.NewResolver == nil {
		dependencies.NewResolver = func(projectRoot string) (SceneResolver, error) {
			resolver, err := project.NewResolver(projectRoot)
			if err != nil {
				return nil, err
			}

			return resolver, nil
		}
	}
	if dependencies.LoadConfig == nil {
		dependencies.LoadConfig = config.Load
	}
	if dependencies.Analyze == nil {
		dependencies.Analyze = analyzeScene
	}
	if dependencies.ResolvePolicy == nil {
		dependencies.ResolvePolicy = policy.Resolve
	}
	if dependencies.ResolvePartial == nil {
		dependencies.ResolvePartial = budget.ResolveFailOnPartial
	}
	if dependencies.Evaluate == nil {
		dependencies.Evaluate = budget.Evaluate
	}
	if dependencies.LoadBuiltInPresets == nil {
		dependencies.LoadBuiltInPresets = preset.Builtins
	}
	if dependencies.ReadFile == nil {
		dependencies.ReadFile = os.ReadFile
	}

	return &Application{dependencies: dependencies}
}

// NewDefault constructs the production application.
func NewDefault() *Application {
	return New(Dependencies{})
}

// Inspect executes project/config discovery, secure scene resolution, and
// recursive static analysis without applying budget policy.
func (application *Application) Inspect(request InspectRequest) (InspectResult, error) {
	result, _, err := application.inspect(request.SceneRequest)
	return result, err
}

// Tree executes the same single-scene analysis as Inspect for dependency-tree
// presentation, without applying policy or budget evaluation.
func (application *Application) Tree(request TreeRequest) (TreeResult, error) {
	result, _, err := application.inspect(request.SceneRequest)
	if err != nil {
		return TreeResult{}, err
	}

	return TreeResult{Inspect: result}, nil
}

// Check executes analysis followed by effective policy and budget evaluation.
func (application *Application) Check(request CheckRequest) (CheckResult, error) {
	inspect, configuration, err := application.inspect(request.SceneRequest)
	if err != nil {
		return CheckResult{}, err
	}

	source := inspect.ConfigSource.Path
	effective, err := application.dependencies.ResolvePolicy(
		source,
		configuration,
		request.Selector,
		append([]string(nil), request.BudgetOverrides...),
	)
	if err != nil {
		return CheckResult{}, fmt.Errorf("resolve effective policy: %w", err)
	}

	failOnPartial, err := application.dependencies.ResolvePartial(
		configuration.FailOnPartial,
		request.PartialOverride,
	)
	if err != nil {
		return CheckResult{}, fmt.Errorf("resolve partial policy: %w", err)
	}

	evaluation, err := application.dependencies.Evaluate(
		inspect.Analysis.Summary.Metrics,
		effective.Budgets,
		inspect.Analysis.Reliability,
		failOnPartial,
	)
	if err != nil {
		return CheckResult{}, fmt.Errorf("evaluate budgets: %w", err)
	}

	return CheckResult{
		Inspect:    inspect,
		Policy:     effective.Clone(),
		Evaluation: evaluation.Clone(),
	}, nil
}

// Diff reads and compares two portable reports without consulting project state.
func (application *Application) Diff(request DiffRequest) (DiffResult, error) {
	if err := application.validate(); err != nil {
		return DiffResult{}, err
	}
	if request.Before == "" || request.After == "" {
		return DiffResult{}, fmt.Errorf("both baseline report paths are required")
	}
	policy, err := reportdiff.NormalizePolicy(request.Policy)
	if err != nil {
		return DiffResult{}, fmt.Errorf("validate diff policy: %w", err)
	}
	before, err := application.readReport(request.Before)
	if err != nil {
		return DiffResult{}, fmt.Errorf("read baseline %q: %w", request.Before, err)
	}
	after, err := application.readReport(request.After)
	if err != nil {
		return DiffResult{}, fmt.Errorf("read candidate %q: %w", request.After, err)
	}
	comparison, err := reportdiff.Compare(before, after, policy)
	if err != nil {
		return DiffResult{}, fmt.Errorf("compare reports: %w", err)
	}
	return DiffResult{Comparison: comparison}, nil
}

func (application *Application) readReport(filename string) (reportdiff.Snapshot, error) {
	contents, err := application.dependencies.ReadFile(filename)
	if err != nil {
		return reportdiff.Snapshot{}, err
	}
	if len(contents) > reportdiff.MaxInputBytes {
		return reportdiff.Snapshot{}, fmt.Errorf("report exceeds %d-byte input limit", reportdiff.MaxInputBytes)
	}
	return reportdiff.Decode(contents)
}

// ListPresets returns built-in presets without consulting project state.
func (application *Application) ListPresets() (PresetListResult, error) {
	if err := application.validate(); err != nil {
		return PresetListResult{}, err
	}

	catalog, err := application.dependencies.LoadBuiltInPresets()
	if err != nil {
		return PresetListResult{}, fmt.Errorf("load built-in presets: %w", err)
	}

	return PresetListResult{Catalog: cloneCatalog(catalog)}, nil
}

// ShowPreset returns one built-in preset without consulting project state.
func (application *Application) ShowPreset(id string) (PresetShowResult, error) {
	result, err := application.ListPresets()
	if err != nil {
		return PresetShowResult{}, err
	}

	item, err := result.Catalog.Find(id)
	if err != nil {
		return PresetShowResult{}, err
	}

	return PresetShowResult{Preset: item}, nil
}

func (application *Application) inspect(request SceneRequest) (InspectResult, config.Config, error) {
	if err := application.validate(); err != nil {
		return InspectResult{}, config.Config{}, err
	}

	workingDirectory, err := application.dependencies.WorkingDirectory()
	if err != nil {
		return InspectResult{}, config.Config{}, fmt.Errorf("resolve working directory: %w", err)
	}

	root, err := application.dependencies.FindProject(project.Request{
		SceneInput:       request.Scene,
		WorkingDirectory: workingDirectory,
		ExplicitProject:  request.Project,
	})
	if err != nil {
		return InspectResult{}, config.Config{}, fmt.Errorf("discover project: %w", err)
	}

	configuration, source, present, err := application.dependencies.LoadConfig(root.Directory, request.Config)
	if err != nil {
		return InspectResult{}, config.Config{}, fmt.Errorf("load configuration: %w", err)
	}
	if !present {
		configuration = config.Config{Version: config.CurrentVersion}
	}

	resolver, err := application.dependencies.NewResolver(root.Directory)
	if err != nil {
		return InspectResult{}, config.Config{}, fmt.Errorf("create scene resolver: %w", err)
	}

	scene, err := resolver.ResolveSceneInput(request.Scene, workingDirectory)
	if err != nil {
		return InspectResult{}, config.Config{}, fmt.Errorf("resolve root scene: %w", err)
	}

	result, err := application.dependencies.Analyze(resolver, scene)
	if err != nil {
		return InspectResult{}, config.Config{}, fmt.Errorf("analyze root scene: %w", err)
	}

	return InspectResult{
		Project:       root,
		Scene:         scene,
		ConfigSource:  source,
		ConfigPresent: present,
		Analysis:      result,
	}, configuration.Clone(), nil
}

func (application *Application) validate() error {
	if application == nil {
		return fmt.Errorf("application is not initialized")
	}

	return nil
}

func analyzeScene(resolver SceneResolver, scene project.ResolvedPath) (analysis.RecursiveResult, error) {
	analyzer, err := analysis.NewRecursiveAnalyzer(
		resolver,
		func(path project.ResolvedPath) (io.ReadCloser, error) {
			return os.Open(path.Canonical)
		},
		tscn.Parse,
	)
	if err != nil {
		return analysis.RecursiveResult{}, err
	}

	return analyzer.Analyze(scene)
}

func cloneCatalog(catalog preset.Catalog) preset.Catalog {
	cloned := make(preset.Catalog, len(catalog))
	for index, item := range catalog {
		item.Budgets = item.Budgets.Clone()
		cloned[index] = item
	}

	return cloned
}
