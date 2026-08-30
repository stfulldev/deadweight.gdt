// Package policy resolves presets, custom profiles, and overrides into one
// effective check policy without performing scene I/O or budget evaluation.
package policy

import (
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

// Kind identifies the selected base policy domain. KindNone is the zero value.
type Kind string

const (
	KindNone    Kind = ""
	KindPreset  Kind = "preset"
	KindProfile Kind = "profile"
)

// Valid reports whether kind is part of the MVP policy contract.
func (kind Kind) Valid() bool {
	return kind == KindNone || kind == KindPreset || kind == KindProfile
}

// Selector contains the mutually exclusive CLI base selectors.
// Empty fields mean that configuration selection should be used.
type Selector struct {
	Preset  string
	Profile string
}

// Metadata is the complete metadata inherited by an effective base policy.
type Metadata struct {
	Name        string
	Description string
	Platform    string
	Renderer    string
	TargetFPS   int64
	Quality     string
	Status      string
	Stability   string
}

// Effective is one fully resolved check policy. KindNone and an empty ID mean
// the policy was formed exclusively from project or CLI budget overrides.
type Effective struct {
	Kind     Kind
	ID       string
	Metadata Metadata
	Budgets  budget.Limits
}

// Clone returns a deep copy whose optional budget values do not alias input.
func (effective Effective) Clone() Effective {
	effective.Budgets = effective.Budgets.Clone()
	return effective
}

// LayerKind identifies one stable policy value source domain.
type LayerKind string

const (
	LayerDefault LayerKind = "default"
	LayerPreset  LayerKind = "preset"
	LayerProfile LayerKind = "profile"
	LayerProject LayerKind = "project"
)

// Valid reports whether kind is a supported explanation layer.
func (kind LayerKind) Valid() bool {
	return kind == LayerDefault || kind == LayerPreset || kind == LayerProfile || kind == LayerProject
}

// Layer identifies the source of one effective policy value.
type Layer struct {
	Kind LayerKind
	ID   string
}

// MetadataSources identifies the source layer of every effective metadata field.
type MetadataSources struct {
	Name        Layer
	Description Layer
	Platform    Layer
	Renderer    Layer
	TargetFPS   Layer
	Quality     Layer
	Status      Layer
	Stability   Layer
}

// LimitSources mirrors optional effective budgets with one source per value.
type LimitSources struct {
	Nodes             *Layer
	TreeDepth         *Layer
	SceneInstances    *Layer
	MeshInstances     *Layer
	Lights            *Layer
	ShadowLights      *Layer
	ExternalResources *Layer
	SceneDependencies *Layer
}

// Get returns the source for one metric when that effective budget is present.
func (sources LimitSources) Get(name metrics.Name) (Layer, bool) {
	var source *Layer
	switch name {
	case metrics.Nodes:
		source = sources.Nodes
	case metrics.TreeDepth:
		source = sources.TreeDepth
	case metrics.SceneInstances:
		source = sources.SceneInstances
	case metrics.MeshInstances:
		source = sources.MeshInstances
	case metrics.Lights:
		source = sources.Lights
	case metrics.ShadowLights:
		source = sources.ShadowLights
	case metrics.ExternalResources:
		source = sources.ExternalResources
	case metrics.SceneDependencies:
		source = sources.SceneDependencies
	}
	if source == nil {
		return Layer{}, false
	}

	return *source, true
}

// Clone returns owned source pointers.
func (sources LimitSources) Clone() LimitSources {
	return LimitSources{
		Nodes:             cloneLayer(sources.Nodes),
		TreeDepth:         cloneLayer(sources.TreeDepth),
		SceneInstances:    cloneLayer(sources.SceneInstances),
		MeshInstances:     cloneLayer(sources.MeshInstances),
		Lights:            cloneLayer(sources.Lights),
		ShadowLights:      cloneLayer(sources.ShadowLights),
		ExternalResources: cloneLayer(sources.ExternalResources),
		SceneDependencies: cloneLayer(sources.SceneDependencies),
	}
}

// ProfileSummary is one custom profile in canonical discovery order.
type ProfileSummary struct {
	ID          string
	Extends     string
	Name        string
	Description string
}

// Explanation is an effective custom policy with its merge evidence.
type Explanation struct {
	Effective           Effective
	FailOnPartial       bool
	FailOnPartialSource Layer
	Chain               []Layer
	MetadataSources     MetadataSources
	BudgetSources       LimitSources
}

// Clone returns an explanation with no caller-owned slices or pointers shared.
func (explanation Explanation) Clone() Explanation {
	explanation.Effective = explanation.Effective.Clone()
	explanation.Chain = append([]Layer(nil), explanation.Chain...)
	explanation.BudgetSources = explanation.BudgetSources.Clone()
	return explanation
}

func cloneLayer(source *Layer) *Layer {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}
