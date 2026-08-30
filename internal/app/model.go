// Package app composes the static-analysis domain services into CLI-ready flows.
package app

import (
	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

// SceneRequest contains the common inputs for one scene application flow.
type SceneRequest struct {
	Scene   string
	Project string
	Config  string
}

// InspectRequest requests recursive analysis without budget evaluation.
type InspectRequest struct {
	SceneRequest
}

// TreeRequest requests recursive analysis for dependency-tree presentation.
type TreeRequest struct {
	SceneRequest
}

// CheckRequest requests recursive analysis and effective budget evaluation.
type CheckRequest struct {
	SceneRequest
	Selector        policy.Selector
	BudgetOverrides []string
	PartialOverride budget.PartialOverride
}

// DiffRequest requests an offline comparison of two portable JSON reports.
type DiffRequest struct {
	Before string
	After  string
	Policy reportdiff.Policy
}

// ProfileRequest contains project/config inputs for custom-profile commands.
type ProfileRequest struct {
	Project string
	Config  string
}

// ProfileShowRequest selects one custom profile in a project configuration.
type ProfileShowRequest struct {
	ProfileRequest
	ID string
}

// InspectResult is the report-ready evidence produced by an inspect flow.
type InspectResult struct {
	Project       project.Root
	Scene         project.ResolvedPath
	ConfigSource  config.Source
	ConfigPresent bool
	Analysis      analysis.RecursiveResult
}

// TreeResult is the report-ready evidence produced by a dependency-tree flow.
// Inspect holds the shared single-scene analysis result without applying policy.
type TreeResult struct {
	Inspect InspectResult
}

// CheckResult is the report-ready evidence produced by a check flow.
type CheckResult struct {
	Inspect    InspectResult
	Policy     policy.Effective
	Evaluation budget.Evaluation
}

// DiffResult is the report-ready result of one offline comparison.
type DiffResult struct {
	Comparison reportdiff.Result
}

// PresetListResult contains built-in presets in product order.
type PresetListResult struct {
	Catalog preset.Catalog
}

// PresetShowResult contains one built-in preset selected by stable ID.
type PresetShowResult struct {
	Preset preset.Preset
}

// ProfileListResult contains custom profiles from one selected configuration.
type ProfileListResult struct {
	Project      project.Root
	ConfigSource config.Source
	Profiles     []policy.ProfileSummary
}

// ProfileShowResult contains one explained effective custom profile.
type ProfileShowResult struct {
	Project      project.Root
	ConfigSource config.Source
	Explanation  policy.Explanation
}
