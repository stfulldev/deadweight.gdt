package metrics_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestCatalog(t *testing.T) {
	t.Parallel()

	want := []metrics.Definition{
		{Name: metrics.Nodes, Label: "Nodes"},
		{Name: metrics.TreeDepth, Label: "Tree depth"},
		{Name: metrics.SceneInstances, Label: "Scene instances"},
		{Name: metrics.MeshInstances, Label: "Mesh instances"},
		{Name: metrics.Lights, Label: "Lights"},
		{Name: metrics.ShadowLights, Label: "Shadow lights"},
		{Name: metrics.ExternalResources, Label: "External resources"},
		{Name: metrics.SceneDependencies, Label: "Scene dependencies"},
	}

	got := metrics.Catalog()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Catalog() = %v, want %v", got, want)
	}

	got[0] = metrics.Definition{Name: "mutated", Label: "Mutated"}
	if metrics.Catalog()[0] != want[0] {
		t.Fatal("Catalog returned mutable package state")
	}

	for _, definition := range want {
		if !definition.Name.Valid() {
			t.Errorf("%q.Valid() = false", definition.Name)
		}
		if got := definition.Name.Label(); got != definition.Label {
			t.Errorf("%q.Label() = %q, want %q", definition.Name, got, definition.Label)
		}
	}

	unknown := metrics.Name("unknown")
	if unknown.Valid() {
		t.Fatal("unknown.Valid() = true")
	}
	if got := unknown.Label(); got != "" {
		t.Fatalf("unknown.Label() = %q, want empty", got)
	}
}

func TestOrderedNames(t *testing.T) {
	t.Parallel()

	want := make([]metrics.Name, 0, len(metrics.Catalog()))
	for _, definition := range metrics.Catalog() {
		want = append(want, definition.Name)
	}

	got := metrics.OrderedNames()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OrderedNames() = %v, want %v", got, want)
	}

	got[0] = "mutated"
	if metrics.OrderedNames()[0] != metrics.Nodes {
		t.Fatal("OrderedNames returned mutable package state")
	}
}

func TestValuesGet(t *testing.T) {
	t.Parallel()

	values := metrics.Values{
		Nodes:             1,
		TreeDepth:         2,
		SceneInstances:    3,
		MeshInstances:     4,
		Lights:            5,
		ShadowLights:      6,
		ExternalResources: 7,
		SceneDependencies: 8,
	}

	for index, name := range metrics.OrderedNames() {
		got, ok := values.Get(name)
		if !ok {
			t.Fatalf("Get(%q) unexpectedly returned ok=false", name)
		}

		want := int64(index + 1)
		if got != want {
			t.Errorf("Get(%q) = %d, want %d", name, got, want)
		}
	}

	if _, ok := values.Get("unknown"); ok {
		t.Fatal("Get(unknown) unexpectedly returned ok=true")
	}
}

func TestValuesValidate(t *testing.T) {
	t.Parallel()

	valid := metrics.Values{
		Nodes:             0,
		TreeDepth:         1,
		SceneInstances:    2,
		MeshInstances:     3,
		Lights:            4,
		ShadowLights:      5,
		ExternalResources: 6,
		SceneDependencies: 7,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	for _, definition := range metrics.Catalog() {
		definition := definition
		t.Run(string(definition.Name), func(t *testing.T) {
			t.Parallel()

			values := valid
			setMetricValue(&values, definition.Name, -1)

			err := values.Validate()
			var valueError *metrics.ValueError
			if !errors.As(err, &valueError) {
				t.Fatalf("Validate() error = %v, want *metrics.ValueError", err)
			}
			if valueError.Name != definition.Name || valueError.Value != -1 {
				t.Fatalf("ValueError = %#v", valueError)
			}
		})
	}
}

func setMetricValue(values *metrics.Values, name metrics.Name, value int64) {
	switch name {
	case metrics.Nodes:
		values.Nodes = value
	case metrics.TreeDepth:
		values.TreeDepth = value
	case metrics.SceneInstances:
		values.SceneInstances = value
	case metrics.MeshInstances:
		values.MeshInstances = value
	case metrics.Lights:
		values.Lights = value
	case metrics.ShadowLights:
		values.ShadowLights = value
	case metrics.ExternalResources:
		values.ExternalResources = value
	case metrics.SceneDependencies:
		values.SceneDependencies = value
	}
}
