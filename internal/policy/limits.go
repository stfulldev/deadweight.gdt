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
