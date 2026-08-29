package metrics

import "fmt"

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

// Definition describes one metric in canonical report order.
type Definition struct {
	Name  Name
	Label string
}

var definitions = [...]Definition{
	{Name: Nodes, Label: "Nodes"},
	{Name: TreeDepth, Label: "Tree depth"},
	{Name: SceneInstances, Label: "Scene instances"},
	{Name: MeshInstances, Label: "Mesh instances"},
	{Name: Lights, Label: "Lights"},
	{Name: ShadowLights, Label: "Shadow lights"},
	{Name: ExternalResources, Label: "External resources"},
	{Name: SceneDependencies, Label: "Scene dependencies"},
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

// ValueError reports a metric value that violates the domain contract.
type ValueError struct {
	Name  Name
	Value int64
}

func (err *ValueError) Error() string {
	return fmt.Sprintf("metric %q must be non-negative, got %d", err.Name, err.Value)
}

// Catalog returns a defensive copy of the metric definitions in canonical order.
func Catalog() []Definition {
	catalog := make([]Definition, len(definitions))
	copy(catalog, definitions[:])

	return catalog
}

// OrderedNames returns a defensive copy in the canonical report order.
func OrderedNames() []Name {
	names := make([]Name, 0, len(definitions))
	for _, definition := range definitions {
		names = append(names, definition.Name)
	}

	return names
}

// Label returns the human-readable console label for a metric.
func (name Name) Label() string {
	for _, definition := range definitions {
		if definition.Name == name {
			return definition.Label
		}
	}

	return ""
}

// Valid reports whether name is part of the MVP metric catalog.
func (name Name) Valid() bool {
	return name.Label() != ""
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

// Validate checks that every metric value satisfies the MVP domain contract.
func (values Values) Validate() error {
	for _, definition := range definitions {
		value, _ := values.Get(definition.Name)
		if value < 0 {
			return &ValueError{Name: definition.Name, Value: value}
		}
	}

	return nil
}
