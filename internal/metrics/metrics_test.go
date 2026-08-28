package metrics_test

import (
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestOrderedNames(t *testing.T) {
	t.Parallel()

	want := []metrics.Name{
		metrics.Nodes,
		metrics.TreeDepth,
		metrics.SceneInstances,
		metrics.MeshInstances,
		metrics.Lights,
		metrics.ShadowLights,
		metrics.ExternalResources,
		metrics.SceneDependencies,
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
