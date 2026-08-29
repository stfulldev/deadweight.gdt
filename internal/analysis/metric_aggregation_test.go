package analysis

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestFinalizeMetricValues(t *testing.T) {
	base := metrics.Values{
		Nodes: 1, TreeDepth: 2, SceneInstances: 3, MeshInstances: 4,
		Lights: 5, ShadowLights: 6, ExternalResources: 99, SceneDependencies: 99,
	}
	tests := []struct {
		name                  string
		values                metrics.Values
		externalResources     uint64
		sceneDependencies     int64
		want                  metrics.Values
		wantOverflow          bool
		wantInvalidMetricName metrics.Name
	}{
		{
			name:   "all eight values",
			values: base, externalResources: 7, sceneDependencies: 8,
			want: metrics.Values{
				Nodes: 1, TreeDepth: 2, SceneInstances: 3, MeshInstances: 4,
				Lights: 5, ShadowLights: 6, ExternalResources: 7, SceneDependencies: 8,
			},
		},
		{
			name: "cardinality overflow", values: base,
			externalResources: uint64(math.MaxInt64) + 1, wantOverflow: true,
		},
		{
			name: "negative occurrence value", values: func() metrics.Values {
				values := base
				values.MeshInstances = -1
				return values
			}(),
			wantInvalidMetricName: metrics.MeshInstances,
		},
		{
			name: "negative graph value", values: base, sceneDependencies: -1,
			wantInvalidMetricName: metrics.SceneDependencies,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := finalizeMetricValues(test.values, test.externalResources, test.sceneDependencies)
			switch {
			case test.wantOverflow:
				assertOverflow(t, err)
				if got != (metrics.Values{}) {
					t.Fatalf("values = %#v, want zero values", got)
				}
			case test.wantInvalidMetricName != "":
				var valueErr *metrics.ValueError
				if !errors.As(err, &valueErr) || valueErr.Name != test.wantInvalidMetricName {
					t.Fatalf("error = %T %v, want ValueError for %q", err, err, test.wantInvalidMetricName)
				}
				if got != (metrics.Values{}) {
					t.Fatalf("values = %#v, want zero values", got)
				}
			default:
				if err != nil || got != test.want {
					t.Fatalf("finalizeMetricValues() = %#v, %v; want %#v, nil", got, err, test.want)
				}
				if test.values.ExternalResources != 99 || test.values.SceneDependencies != 99 {
					t.Fatalf("input mutated = %#v", test.values)
				}
			}
		})
	}
}

func TestFinalizeMetricsUsesRetainedRootEvidenceInCanonicalOrder(t *testing.T) {
	summary := ExpandedSummary{
		Metrics: metrics.Values{
			Nodes: 11, TreeDepth: 12, SceneInstances: 13, MeshInstances: 14,
			Lights: 15, ShadowLights: 16,
		},
		ExternalResources: []ResourceIdentity{
			{Resolved: true, Canonical: "/project/a.res"},
			{Resolved: true, Canonical: "/project/b.res"},
			{DeclaringScene: "/project/root.tscn", ResourceID: "3", RawPath: "missing.res"},
		},
	}
	got, err := finalizeMetrics(summary, DependencyGraph{SceneDependencies: 2})
	if err != nil {
		t.Fatalf("finalizeMetrics() error = %v", err)
	}
	want := metrics.Values{
		Nodes: 11, TreeDepth: 12, SceneInstances: 13, MeshInstances: 14,
		Lights: 15, ShadowLights: 16, ExternalResources: 3, SceneDependencies: 2,
	}
	if got != want {
		t.Fatalf("values = %#v, want %#v", got, want)
	}
	wantNames := []metrics.Name{
		metrics.Nodes, metrics.TreeDepth, metrics.SceneInstances, metrics.MeshInstances,
		metrics.Lights, metrics.ShadowLights, metrics.ExternalResources, metrics.SceneDependencies,
	}
	if names := metrics.OrderedNames(); !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("ordered names = %#v, want %#v", names, wantNames)
	}
	wantValues := []int64{11, 12, 13, 14, 15, 16, 3, 2}
	for index, name := range wantNames {
		value, ok := got.Get(name)
		wantValue := wantValues[index]
		if !ok || value != wantValue {
			t.Fatalf("Get(%q) = %d, %v; want %d, true", name, value, ok, wantValue)
		}
	}
}

func TestRecursiveAnalyzerFinalizesSupportedLiteralMetrics(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	loader := &memorySceneEffects{sources: map[string]string{root.Canonical: `[gd_scene format=3]
[node name="Root" type="Node3D"]
[node name="Mesh" type="MeshInstance3D" parent="."]
[node name="Sun" type="DirectionalLight3D" parent="."]
shadow_enabled = true
[node name="Bulb" type="OmniLight3D" parent="."]
[node name="Cone" type="SpotLight3D" parent="."]
shadow_enabled = false
[node name="CustomMesh" type="CustomMeshInstance3D" parent="."]
[node name="MultiMesh" type="MultiMeshInstance3D" parent="."]
[node name="Light2D" type="PointLight2D" parent="."]
[node name="CustomLight" type="CustomOmniLight3D" parent="."]
`}, errors: map[string]error{}}
	analyzer := newTestRecursiveAnalyzer(t, &memoryResolver{results: map[string]project.Resolution{}}, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	want := metrics.Values{Nodes: 9, TreeDepth: 2, MeshInstances: 1, Lights: 3, ShadowLights: 1}
	if result.Summary.Metrics != want {
		t.Fatalf("metrics = %#v, want %#v", result.Summary.Metrics, want)
	}
}

func TestRecursiveAnalyzerKeepsUnresolvedResourceIdentitiesDocumentLocal(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	left := testScenePath(rootDir, "left.tscn")
	right := testScenePath(rootDir, "right.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "left.tscn"):  resolvedResource("left.tscn", left),
		resolverKey(root.Canonical, "right.tscn"): resolvedResource("right.tscn", right),
	}}
	loader := &memorySceneEffects{sources: map[string]string{
		root.Canonical: sceneWithMounts("Root", []resourceMount{{"1_left", "left.tscn"}, {"2_right", "right.tscn"}}),
		left.Canonical: `[gd_scene format=3]
[ext_resource type="Resource" path="shared.res" id="left_asset"]
[node name="Left" type="Node3D"]
`,
		right.Canonical: `[gd_scene format=3]
[ext_resource type="Resource" path="shared.res" id="right_asset"]
[node name="Right" type="Node3D"]
`,
	}, errors: map[string]error{}}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	want := metrics.Values{
		Nodes: 3, TreeDepth: 2, SceneInstances: 2,
		ExternalResources: 4, SceneDependencies: 2,
	}
	if result.Summary.Metrics != want {
		t.Fatalf("metrics = %#v, want %#v", result.Summary.Metrics, want)
	}
	unresolved := make(map[string]ResourceIdentity)
	for _, resource := range result.Summary.ExternalResources {
		if !resource.Resolved {
			unresolved[resource.ResourceID] = resource
		}
	}
	if len(unresolved) != 2 || unresolved["left_asset"].DeclaringScene != left.Canonical ||
		unresolved["right_asset"].DeclaringScene != right.Canonical {
		t.Fatalf("unresolved identities = %#v", unresolved)
	}
}

func TestRecursiveAnalyzerFinalizesCityBuildingLampExample(t *testing.T) {
	rootDir := t.TempDir()
	city := testScenePath(rootDir, "city.tscn")
	building := testScenePath(rootDir, "building.tscn")
	lamp := testScenePath(rootDir, "lamp.tscn")
	texture := testResourcePath(rootDir, "shared.png")

	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(city.Canonical, "building.tscn"): resolvedResource("building.tscn", building),
		resolverKey(city.Canonical, "lamp.tscn"):     resolvedResource("lamp.tscn", lamp),
		resolverKey(building.Canonical, "lamp.tscn"): resolvedResource("lamp.tscn", lamp),
		resolverKey(lamp.Canonical, "shared.png"):    resolvedResource("shared.png", texture),
	}}

	var citySource strings.Builder
	citySource.WriteString("[gd_scene format=3]\n")
	citySource.WriteString("[ext_resource type=\"PackedScene\" path=\"building.tscn\" id=\"1_building\"]\n")
	citySource.WriteString("[ext_resource type=\"PackedScene\" path=\"lamp.tscn\" id=\"2_lamp\"]\n")
	writeOrdinaryNodes(&citySource, "City", 10)
	for index := 0; index < 2; index++ {
		fmt.Fprintf(&citySource, "[node name=\"Building%d\" parent=\".\" instance=ExtResource(\"1_building\")]\n", index)
	}
	for index := 0; index < 3; index++ {
		fmt.Fprintf(&citySource, "[node name=\"Lamp%d\" parent=\".\" instance=ExtResource(\"2_lamp\")]\n", index)
	}

	var buildingSource strings.Builder
	buildingSource.WriteString("[gd_scene format=3]\n")
	buildingSource.WriteString("[ext_resource type=\"PackedScene\" path=\"lamp.tscn\" id=\"1_lamp\"]\n")
	writeOrdinaryNodes(&buildingSource, "Building", 16)
	buildingSource.WriteString("[node name=\"NestedLamp\" parent=\".\" instance=ExtResource(\"1_lamp\")]\n")

	var lampSource strings.Builder
	lampSource.WriteString("[gd_scene format=3]\n")
	lampSource.WriteString("[ext_resource type=\"Texture2D\" path=\"shared.png\" id=\"1_texture\"]\n")
	writeOrdinaryNodes(&lampSource, "Lamp", 4)

	loader := &memorySceneEffects{sources: map[string]string{
		city.Canonical:     citySource.String(),
		building.Canonical: buildingSource.String(),
		lamp.Canonical:     lampSource.String(),
	}, errors: map[string]error{}}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(city)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Summary.Metrics.Nodes != 62 || result.Summary.Metrics.SceneInstances != 7 ||
		result.Summary.Metrics.ExternalResources != 3 || result.Summary.Metrics.SceneDependencies != 2 {
		t.Fatalf("metrics = %#v", result.Summary.Metrics)
	}
	if occurrences(result.Summary.ExternalResources, texture.Canonical) != 1 {
		t.Fatalf("shared texture identities = %#v", result.Summary.ExternalResources)
	}
	if !reflect.DeepEqual(result.Summary.Dependencies, []string{building.Canonical, lamp.Canonical}) {
		t.Fatalf("dependencies = %#v", result.Summary.Dependencies)
	}
	for _, path := range []project.ResolvedPath{city, building, lamp} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func writeOrdinaryNodes(source *strings.Builder, prefix string, count int) {
	fmt.Fprintf(source, "[node name=\"%s\" type=\"Node3D\"]\n", prefix)
	for index := 1; index < count; index++ {
		fmt.Fprintf(source, "[node name=\"%s%d\" type=\"Node3D\" parent=\".\"]\n", prefix, index)
	}
}
