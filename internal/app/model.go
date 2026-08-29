// Package app composes the static-analysis domain services into CLI-ready flows.
package app

import (
	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
	"github.com/stfulldev/deadweight.gdt/internal/project"
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

// CheckRequest requests recursive analysis and effective budget evaluation.
type CheckRequest struct {
	SceneRequest
	Selector        policy.Selector
	BudgetOverrides []string
	PartialOverride budget.PartialOverride
}

// InspectResult is the report-ready evidence produced by an inspect flow.
type InspectResult struct {
	Project       project.Root
	Scene         project.ResolvedPath
	ConfigSource  config.Source
	ConfigPresent bool
	Analysis      analysis.RecursiveResult
}

// CheckResult is the report-ready evidence produced by a check flow.
type CheckResult struct {
	Inspect    InspectResult
	Policy     policy.Effective
	Evaluation budget.Evaluation
}

// PresetListResult contains built-in presets in product order.
type PresetListResult struct {
	Catalog preset.Catalog
}

// PresetShowResult contains one built-in preset selected by stable ID.
type PresetShowResult struct {
	Preset preset.Preset
}
