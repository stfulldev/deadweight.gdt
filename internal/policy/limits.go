package policy

import (
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func mergeLimits(lower, higher budget.Limits) budget.Limits {
	merged := lower.Clone()
	for _, name := range metrics.OrderedNames() {
		if value, configured := higher.Get(name); configured {
			setLimit(&merged, name, value)
		}
	}

	return merged
}

func mergeResolvedLimits(resolved resolvedProfile, higher budget.Limits, source Layer) resolvedProfile {
	merged := resolved.clone()
	for _, name := range metrics.OrderedNames() {
		if value, configured := higher.Get(name); configured {
			setLimit(&merged.budgets, name, value)
			setLimitSource(&merged.budgetSources, name, source)
		}
	}

	return merged
}

func setLimit(limits *budget.Limits, name metrics.Name, value int64) {
	owned := value
	switch name {
	case metrics.Nodes:
		limits.Nodes = &owned
	case metrics.TreeDepth:
		limits.TreeDepth = &owned
	case metrics.SceneInstances:
		limits.SceneInstances = &owned
	case metrics.MeshInstances:
		limits.MeshInstances = &owned
	case metrics.Lights:
		limits.Lights = &owned
	case metrics.ShadowLights:
		limits.ShadowLights = &owned
	case metrics.ExternalResources:
		limits.ExternalResources = &owned
	case metrics.SceneDependencies:
		limits.SceneDependencies = &owned
	}
}

func setLimitSource(sources *LimitSources, name metrics.Name, source Layer) {
	owned := source
	switch name {
	case metrics.Nodes:
		sources.Nodes = &owned
	case metrics.TreeDepth:
		sources.TreeDepth = &owned
	case metrics.SceneInstances:
		sources.SceneInstances = &owned
	case metrics.MeshInstances:
		sources.MeshInstances = &owned
	case metrics.Lights:
		sources.Lights = &owned
	case metrics.ShadowLights:
		sources.ShadowLights = &owned
	case metrics.ExternalResources:
		sources.ExternalResources = &owned
	case metrics.SceneDependencies:
		sources.SceneDependencies = &owned
	}
}
