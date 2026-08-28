package budget

import "github.com/stfulldev/deadweight.gdt/internal/metrics"

// Limits contains optional inclusive upper bounds for every MVP metric.
// A nil field means that the corresponding metric is displayed but not checked.
type Limits struct {
	Nodes             *int64 `json:"nodes,omitempty"`
	TreeDepth         *int64 `json:"tree_depth,omitempty"`
	SceneInstances    *int64 `json:"scene_instances,omitempty"`
	MeshInstances     *int64 `json:"mesh_instances,omitempty"`
	Lights            *int64 `json:"lights,omitempty"`
	ShadowLights      *int64 `json:"shadow_lights,omitempty"`
	ExternalResources *int64 `json:"external_resources,omitempty"`
	SceneDependencies *int64 `json:"scene_dependencies,omitempty"`
}

// Get returns a configured limit by its stable metric identifier.
func (limits Limits) Get(name metrics.Name) (int64, bool) {
	var value *int64

	switch name {
	case metrics.Nodes:
		value = limits.Nodes
	case metrics.TreeDepth:
		value = limits.TreeDepth
	case metrics.SceneInstances:
		value = limits.SceneInstances
	case metrics.MeshInstances:
		value = limits.MeshInstances
	case metrics.Lights:
		value = limits.Lights
	case metrics.ShadowLights:
		value = limits.ShadowLights
	case metrics.ExternalResources:
		value = limits.ExternalResources
	case metrics.SceneDependencies:
		value = limits.SceneDependencies
	default:
		return 0, false
	}

	if value == nil {
		return 0, false
	}

	return *value, true
}

// Count returns the number of configured limits.
func (limits Limits) Count() int {
	count := 0
	for _, name := range metrics.OrderedNames() {
		if _, ok := limits.Get(name); ok {
			count++
		}
	}

	return count
}

// Clone returns a deep copy whose optional values do not alias the source.
func (limits Limits) Clone() Limits {
	return Limits{
		Nodes:             cloneInt64(limits.Nodes),
		TreeDepth:         cloneInt64(limits.TreeDepth),
		SceneInstances:    cloneInt64(limits.SceneInstances),
		MeshInstances:     cloneInt64(limits.MeshInstances),
		Lights:            cloneInt64(limits.Lights),
		ShadowLights:      cloneInt64(limits.ShadowLights),
		ExternalResources: cloneInt64(limits.ExternalResources),
		SceneDependencies: cloneInt64(limits.SceneDependencies),
	}
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
