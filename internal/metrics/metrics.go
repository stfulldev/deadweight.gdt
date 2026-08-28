package metrics

// Name is a stable metric identifier used by configuration, budgets, and reports.
type Name string

const (
	Nodes             Name = "nodes"
	TreeDepth         Name = "tree_depth"
	SceneInstances    Name = "scene_instances"
	MeshInstances     Name = "mesh_instances"
	Lights            Name = "lights"
	ShadowLights      Name = "shadow_lights"
	ExternalResources Name = "external_resources"
	SceneDependencies Name = "scene_dependencies"
)

var orderedNames = [...]Name{
	Nodes,
	TreeDepth,
	SceneInstances,
	MeshInstances,
	Lights,
	ShadowLights,
	ExternalResources,
	SceneDependencies,
}

var labels = map[Name]string{
	Nodes:             "Nodes",
	TreeDepth:         "Tree depth",
	SceneInstances:    "Scene instances",
	MeshInstances:     "Mesh instances",
	Lights:            "Lights",
	ShadowLights:      "Shadow lights",
	ExternalResources: "External resources",
	SceneDependencies: "Scene dependencies",
}

// Values contains the eight metrics defined by the MVP 0.1 contract.
type Values struct {
	Nodes             int64 `json:"nodes"`
	TreeDepth         int64 `json:"tree_depth"`
	SceneInstances    int64 `json:"scene_instances"`
	MeshInstances     int64 `json:"mesh_instances"`
	Lights            int64 `json:"lights"`
	ShadowLights      int64 `json:"shadow_lights"`
	ExternalResources int64 `json:"external_resources"`
	SceneDependencies int64 `json:"scene_dependencies"`
}

// OrderedNames returns a defensive copy in the canonical report order.
func OrderedNames() []Name {
	names := make([]Name, len(orderedNames))
	copy(names, orderedNames[:])

	return names
}

// Label returns the human-readable console label for a metric.
func (name Name) Label() string {
	return labels[name]
}

// Valid reports whether name is part of the MVP metric catalog.
func (name Name) Valid() bool {
	_, ok := labels[name]
	return ok
}

// Get returns a metric value by its stable identifier.
func (values Values) Get(name Name) (int64, bool) {
	switch name {
	case Nodes:
		return values.Nodes, true
	case TreeDepth:
		return values.TreeDepth, true
	case SceneInstances:
		return values.SceneInstances, true
	case MeshInstances:
		return values.MeshInstances, true
	case Lights:
		return values.Lights, true
	case ShadowLights:
		return values.ShadowLights, true
	case ExternalResources:
		return values.ExternalResources, true
	case SceneDependencies:
		return values.SceneDependencies, true
	default:
		return 0, false
	}
}
