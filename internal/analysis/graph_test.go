package analysis

import (
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func TestRecursiveAnalyzerBuildsRootOnlyGraph(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{}}
	loader := &memorySceneEffects{
		sources: map[string]string{root.Canonical: `[gd_scene format=3]
[node name="Root" type="Node3D"]
`},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)
	builds := 0
	analyzer.summarize = func(document *tscn.Document) (LocalSummary, error) {
		builds++
		return BuildLocalSummary(document)
	}

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	wantGraph := DependencyGraph{
		RootCanonical: root.Canonical,
		RootDisplay:   root.Display,
		Nodes:         []GraphNode{{Canonical: root.Canonical, Display: root.Display}},
	}
	if !reflect.DeepEqual(result.Graph, wantGraph) {
		t.Fatalf("Graph = %#v, want %#v", result.Graph, wantGraph)
	}
	if result.Summary.Metrics != (metrics.Values{Nodes: 1, TreeDepth: 1}) || len(result.Summary.Dependencies) != 0 {
		t.Fatalf("Summary = %#v", result.Summary)
	}
	if result.ParsedSceneFiles != 1 {
		t.Fatalf("ParsedSceneFiles = %d, want 1", result.ParsedSceneFiles)
	}
	if builds != 1 {
		t.Fatalf("local summary builds = %d, want 1", builds)
	}
	requireMemorySceneEffects(t, loader, root, 1)
}

func TestRecursiveAnalyzerBuildsExactChainGraph(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	leaf := testScenePath(rootDir, "leaf.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
		resolverKey(child.Canonical, "leaf.tscn"): resolvedResource("leaf.tscn", leaf),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:  sceneWithMounts("Root", []resourceMount{{"1_child", "child.tscn"}}),
			child.Canonical: sceneWithMounts("Child", []resourceMount{{"1_leaf", "leaf.tscn"}}),
			leaf.Canonical: `[gd_scene format=3]
[node name="Leaf" type="Node3D"]
`,
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Graph.SceneDependencies != 2 || len(result.Graph.Nodes) != 3 || len(result.Graph.Edges) != 2 {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	wantDependencies := []string{child.Canonical, leaf.Canonical}
	sort.Strings(wantDependencies)
	if !reflect.DeepEqual(result.Summary.Dependencies, wantDependencies) {
		t.Fatalf("Dependencies = %#v, want %#v", result.Summary.Dependencies, wantDependencies)
	}
	if result.Summary.Metrics != (metrics.Values{Nodes: 3, TreeDepth: 3, SceneInstances: 2}) {
		t.Fatalf("Metrics = %#v", result.Summary.Metrics)
	}
	if result.ParsedSceneFiles != 3 {
		t.Fatalf("ParsedSceneFiles = %d, want 3", result.ParsedSceneFiles)
	}
	for _, path := range []project.ResolvedPath{root, child, leaf} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerCompactsOneHundredRepeatedGraphEdges(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
	}}
	var mounts strings.Builder
	for index := 0; index < 100; index++ {
		mounts.WriteString("[node name=\"Child")
		mounts.WriteString(testDecimal(index))
		mounts.WriteString("\" parent=\".\" instance=ExtResource(\"1_child\")]\n")
	}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="child.tscn" id="1_child"]
[node name="Root" type="Node3D"]
` + mounts.String(),
			child.Canonical: `[gd_scene format=3]
[node name="Child" type="MeshInstance3D"]
`,
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)
	builds := make(map[*tscn.Document]int)
	analyzer.summarize = func(document *tscn.Document) (LocalSummary, error) {
		builds[document]++
		return BuildLocalSummary(document)
	}

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Graph.SceneDependencies != 1 || len(result.Graph.Nodes) != 2 || len(result.Graph.Edges) != 1 {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	edge := result.Graph.Edges[0]
	if !edge.Resolved || edge.Kind != EdgeInstance || edge.ToCanonical != child.Canonical || edge.Occurrences != 100 {
		t.Fatalf("Edge = %#v", edge)
	}
	if result.Summary.Metrics != (metrics.Values{Nodes: 101, TreeDepth: 2, SceneInstances: 100, MeshInstances: 100}) {
		t.Fatalf("Metrics = %#v", result.Summary.Metrics)
	}
	if result.ParsedSceneFiles != 2 || len(builds) != 2 || len(resolver.calls) != 1 {
		t.Fatalf("loader/build/resolver counts = %#v/%#v/%#v", loader.calls, builds, resolver.calls)
	}
	for _, path := range []project.ResolvedPath{root, child} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerBuildsDiamondOnce(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	left := testScenePath(rootDir, "left.tscn")
	right := testScenePath(rootDir, "right.tscn")
	shared := testScenePath(rootDir, "shared.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "left.tscn"):    resolvedResource("left.tscn", left),
		resolverKey(root.Canonical, "right.tscn"):   resolvedResource("right.tscn", right),
		resolverKey(left.Canonical, "shared.tscn"):  resolvedResource("shared.tscn", shared),
		resolverKey(right.Canonical, "shared.tscn"): resolvedResource("shared.tscn", shared),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:   sceneWithMounts("Root", []resourceMount{{"1_left", "left.tscn"}, {"2_right", "right.tscn"}}),
			left.Canonical:   sceneWithMounts("Left", []resourceMount{{"1_shared", "shared.tscn"}}),
			right.Canonical:  sceneWithMounts("Right", []resourceMount{{"1_shared", "shared.tscn"}}),
			shared.Canonical: "[gd_scene format=3]\n[node name=\"Shared\" type=\"Node3D\"]\n",
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Graph.SceneDependencies != 3 || len(result.Graph.Nodes) != 4 || len(result.Graph.Edges) != 4 {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	if result.Summary.Metrics.Nodes != 5 || result.Summary.Metrics.SceneInstances != 4 {
		t.Fatalf("Metrics = %#v", result.Summary.Metrics)
	}
	if result.ParsedSceneFiles != 4 {
		t.Fatalf("ParsedSceneFiles = %d, want 4", result.ParsedSceneFiles)
	}
	for _, path := range []project.ResolvedPath{root, left, right, shared} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerKeepsUnresolvedTargetsOutOfGraphNodes(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	imported := testResourcePath(rootDir, "model.glb")
	unsupported := testResourcePath(rootDir, "logic.gd")
	unavailable := testScenePath(rootDir, "unavailable.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "missing.tscn"): {
			Reason: project.ResolutionMissing,
			Path:   project.ResolvedPath{Original: "missing.tscn"},
		},
		resolverKey(root.Canonical, "model.glb"):        resolvedResource("model.glb", imported),
		resolverKey(root.Canonical, "logic.gd"):         resolvedResource("logic.gd", unsupported),
		resolverKey(root.Canonical, "unavailable.tscn"): resolvedResource("unavailable.tscn", unavailable),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="missing.tscn" id="1_missing"]
[ext_resource type="PackedScene" path="model.glb" id="2_imported"]
[ext_resource type="PackedScene" path="logic.gd" id="3_unsupported"]
[ext_resource type="PackedScene" path="unavailable.tscn" id="4_unavailable"]
[node name="Root" type="Node3D"]
[node name="WrongID" parent="." instance=ExtResource("0_unknown")]
[node name="Missing" parent="." instance=ExtResource("1_missing")]
[node name="Embedded" parent="." instance=SubResource("Scene_1")]
[node name="Placeholder" parent="." instance_placeholder="res://later.tscn"]
[node name="Imported" parent="." instance=ExtResource("2_imported")]
[node name="Unsupported" parent="." instance=ExtResource("3_unsupported")]
[node name="Unavailable" parent="." instance=ExtResource("4_unavailable")]
`},
		errors: map[string]error{unavailable.Canonical: errors.New("unavailable")},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Graph.SceneDependencies != 0 || len(result.Graph.Nodes) != 1 || len(result.Graph.Edges) != 7 {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	want := map[TargetClassification]int{
		TargetMissingExternalResource: 1,
		TargetUnresolvedPath:          1,
		TargetSubResource:             1,
		TargetPlaceholder:             1,
		TargetImportedScene:           1,
		TargetUnsupportedScene:        1,
		TargetUnavailableScene:        1,
	}
	got := make(map[TargetClassification]int)
	for _, edge := range result.Graph.Edges {
		got[edge.Classification]++
		if edge.Resolved || edge.ToCanonical != "" || edge.ToDisplay != "" || edge.Occurrences != 1 || edge.Kind != EdgeInstance {
			t.Errorf("unresolved edge = %#v", edge)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("classifications = %#v, want %#v", got, want)
	}
	if result.Summary.Metrics.Nodes != 8 || result.Summary.Coverage.Unresolved != 7 {
		t.Fatalf("Summary = %#v", result.Summary)
	}
	if result.ParsedSceneFiles != 1 {
		t.Fatalf("ParsedSceneFiles = %d, want only root", result.ParsedSceneFiles)
	}
	if loader.calls[unavailable.Canonical] != 1 {
		t.Fatalf("unavailable loader calls = %d", loader.calls[unavailable.Canonical])
	}
	requireMemorySceneEffects(t, loader, root, 1)
	if loader.parses[unavailable.Canonical] != 0 || loader.closes[unavailable.Canonical] != 0 {
		t.Fatalf("unavailable effects = parses %d, closes %d", loader.parses[unavailable.Canonical], loader.closes[unavailable.Canonical])
	}
}

func TestRecursiveAnalyzerRetainsUnresolvedInheritanceEdge(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	inherited := testScenePath(rootDir, "inherited.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "inherited.tscn"): resolvedResource("inherited.tscn", inherited),
		resolverKey(inherited.Canonical, "missing-base.tscn"): {
			Reason: project.ResolutionMissing,
			Path:   project.ResolvedPath{Original: "missing-base.tscn"},
		},
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: sceneWithMounts("Root", []resourceMount{{"1_inherited", "inherited.tscn"}}),
			inherited.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="missing-base.tscn" id="1_base"]
[node name="Inherited" instance=ExtResource("1_base")]
`,
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Graph.SceneDependencies != 1 || len(result.Graph.Nodes) != 2 || len(result.Graph.Edges) != 2 {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	edge := findGraphEdge(t, result.Graph, inherited.Canonical, EdgeInheritance)
	if edge.Resolved || edge.Classification != TargetUnresolvedPath || edge.ResolutionReason != project.ResolutionMissing || edge.RawTarget != "missing-base.tscn" {
		t.Fatalf("inheritance edge = %#v", edge)
	}
	if result.ParsedSceneFiles != 2 {
		t.Fatalf("ParsedSceneFiles = %d, want 2", result.ParsedSceneFiles)
	}
}

func TestRecursiveAnalyzerTraversesResolvedInheritanceWithoutApplyingItsMetrics(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	inherited := testScenePath(rootDir, "inherited.tscn")
	base := testScenePath(rootDir, "base.tscn")
	leaf := testScenePath(rootDir, "leaf.tscn")
	asset := testResourcePath(rootDir, "leaf.res")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "inherited.tscn"): resolvedResource("inherited.tscn", inherited),
		resolverKey(inherited.Canonical, "base.tscn"): resolvedResource("base.tscn", base),
		resolverKey(base.Canonical, "leaf.tscn"):      resolvedResource("leaf.tscn", leaf),
		resolverKey(leaf.Canonical, "leaf.res"):       resolvedResource("leaf.res", asset),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: sceneWithMounts("Root", []resourceMount{{"1_inherited", "inherited.tscn"}}),
			inherited.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[node name="Inherited" instance=ExtResource("1_base")]
`,
			base.Canonical: sceneWithMounts("Base", []resourceMount{{"1_leaf", "leaf.tscn"}}),
			leaf.Canonical: `[gd_scene format=3]
[ext_resource type="Resource" path="leaf.res" id="1_asset"]
[node name="Leaf" type="MeshInstance3D"]
`,
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Graph.SceneDependencies != 3 || len(result.Graph.Nodes) != 4 || len(result.Graph.Edges) != 3 {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	if findGraphEdge(t, result.Graph, root.Canonical, EdgeInstance).ToCanonical != inherited.Canonical ||
		findGraphEdge(t, result.Graph, inherited.Canonical, EdgeInheritance).ToCanonical != base.Canonical ||
		findGraphEdge(t, result.Graph, base.Canonical, EdgeInstance).ToCanonical != leaf.Canonical {
		t.Fatalf("Edges = %#v", result.Graph.Edges)
	}
	wantDependencies := []string{inherited.Canonical, base.Canonical, leaf.Canonical}
	sort.Strings(wantDependencies)
	if !reflect.DeepEqual(result.Summary.Dependencies, wantDependencies) {
		t.Fatalf("Dependencies = %#v, want %#v", result.Summary.Dependencies, wantDependencies)
	}
	if result.Summary.Metrics != (metrics.Values{Nodes: 2, TreeDepth: 2, SceneInstances: 1}) || result.Summary.Coverage.Unresolved != 1 {
		t.Fatalf("inherited metrics were applied exactly: %#v/%#v", result.Summary.Metrics, result.Summary.Coverage)
	}
	if result.ParsedSceneFiles != 4 {
		t.Fatalf("ParsedSceneFiles = %d, want 4", result.ParsedSceneFiles)
	}
	for _, canonical := range []string{inherited.Canonical, base.Canonical, leaf.Canonical, asset.Canonical} {
		if occurrences(result.Summary.ExternalResources, canonical) != 1 {
			t.Errorf("resource %q occurrences = %d", canonical, occurrences(result.Summary.ExternalResources, canonical))
		}
	}
	for _, path := range []project.ResolvedPath{root, inherited, base, leaf} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
	if len(resolver.calls) != 4 {
		t.Fatalf("resolver calls = %#v", resolver.calls)
	}
}

func TestRecursiveAnalyzerReportsSelfCycle(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "root.tscn"): resolvedResource("root.tscn", root),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{root.Canonical: sceneWithMounts("Root", []resourceMount{{"1_root", "root.tscn"}})},
		errors:  map[string]error{},
		calls:   map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	cycle := requireCycle(t, result, err)
	if !reflect.DeepEqual(cycle.Canonical, []string{root.Canonical, root.Canonical}) || !reflect.DeepEqual(cycle.Display, []string{root.Display, root.Display}) {
		t.Fatalf("Cycle = %#v", cycle)
	}
	wantMessage := "scene dependency cycle\n\n" + root.Display + "\n→ " + root.Display
	if diagnostic.MessageOf(err) != wantMessage {
		t.Fatalf("DiagnosticMessage = %q, want %q", diagnostic.MessageOf(err), wantMessage)
	}
}

func TestRecursiveAnalyzerReportsMixedInstanceInheritanceCycle(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	inherited := testScenePath(rootDir, "inherited.tscn")
	base := testScenePath(rootDir, "base.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "inherited.tscn"): resolvedResource("inherited.tscn", inherited),
		resolverKey(inherited.Canonical, "base.tscn"): resolvedResource("base.tscn", base),
		resolverKey(base.Canonical, "root.tscn"):      resolvedResource("root.tscn", root),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: sceneWithMounts("Root", []resourceMount{{"1_inherited", "inherited.tscn"}}),
			inherited.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[node name="Inherited" instance=ExtResource("1_base")]
`,
			base.Canonical: sceneWithMounts("Base", []resourceMount{{"1_root", "root.tscn"}}),
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	cycle := requireCycle(t, result, err)
	wantCanonical := []string{root.Canonical, inherited.Canonical, base.Canonical, root.Canonical}
	wantDisplay := []string{root.Display, inherited.Display, base.Display, root.Display}
	if !reflect.DeepEqual(cycle.Canonical, wantCanonical) || !reflect.DeepEqual(cycle.Display, wantDisplay) {
		t.Fatalf("Cycle = %#v, want %#v/%#v", cycle, wantCanonical, wantDisplay)
	}
}

func TestRecursiveAnalyzerSelectsCycleByDeterministicTargetOrder(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	a := testScenePath(rootDir, "a.tscn")
	z := testScenePath(rootDir, "z.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "z.tscn"): resolvedResource("z.tscn", z),
		resolverKey(root.Canonical, "a.tscn"): resolvedResource("a.tscn", a),
		resolverKey(a.Canonical, "a.tscn"):    resolvedResource("a.tscn", a),
		resolverKey(z.Canonical, "z.tscn"):    resolvedResource("z.tscn", z),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: sceneWithMounts("Root", []resourceMount{{"1_z", "z.tscn"}, {"2_a", "a.tscn"}}),
			a.Canonical:    sceneWithMounts("A", []resourceMount{{"1_a", "a.tscn"}}),
			z.Canonical:    sceneWithMounts("Z", []resourceMount{{"1_z", "z.tscn"}}),
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	cycle := requireCycle(t, result, err)
	if !reflect.DeepEqual(cycle.Canonical, []string{a.Canonical, a.Canonical}) {
		t.Fatalf("Cycle = %#v, want lexicographically first target A", cycle)
	}
	if loader.calls[z.Canonical] != 0 {
		t.Fatalf("later cycle target was visited first: calls = %#v", loader.calls)
	}
}

func TestGraphBuilderRejectsEdgeAndDependencyOverflow(t *testing.T) {
	root := testScenePath(t.TempDir(), "root.tscn")
	edge := GraphEdge{
		FromCanonical:  root.Canonical,
		FromDisplay:    root.Display,
		RawTarget:      "missing.tscn",
		Kind:           EdgeInstance,
		Classification: TargetUnresolvedPath,
		Occurrences:    math.MaxInt64,
	}
	builder := newGraphBuilder(root)
	if err := builder.addEdge(edge); err != nil {
		t.Fatalf("maximum edge occurrence error = %v", err)
	}
	edge.Occurrences = 1
	assertOverflow(t, builder.addEdge(edge))
	if len(builder.edges) != 1 {
		t.Fatalf("edge overflow mutated graph builder: %#v", builder.edges)
	}
	for _, cached := range builder.edges {
		if cached.Occurrences != math.MaxInt64 {
			t.Fatalf("edge overflow changed occurrences to %d", cached.Occurrences)
		}
	}

	builder = newGraphBuilder(root)
	builder.dependencyCount = math.MaxInt64
	assertOverflow(t, builder.addNode(testScenePath(filepath.Dir(root.Canonical), "child.tscn")))
	if len(builder.nodes) != 0 || builder.dependencyCount != math.MaxInt64 {
		t.Fatalf("dependency overflow mutated graph builder: %#v", builder)
	}
}

func TestRecursiveAnalyzerOwnsDeterministicGraphResults(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:  sceneWithMounts("Root", []resourceMount{{"1_child", "child.tscn"}}),
			child.Canonical: "[gd_scene format=3]\n[node name=\"Child\" type=\"Node3D\"]\n",
		},
		errors: map[string]error{},
		calls:  map[string]int{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	first, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("first Analyze() error = %v", err)
	}
	expected := cloneRecursiveResult(first)
	first.Graph.Nodes[0].Canonical = "mutated"
	first.Graph.Edges[0].RawTarget = "mutated"
	first.Summary.Dependencies[0] = "mutated"
	first.Summary.ExternalResources[0].Canonical = "mutated"

	second, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("second Analyze() error = %v", err)
	}
	if !reflect.DeepEqual(second, expected) {
		t.Fatalf("caller mutation changed later result:\nsecond: %#v\nwant: %#v", second, expected)
	}
}

func TestRecursiveAnalyzerGraphUsesRealResolverForRelativeInheritance(t *testing.T) {
	projectRoot := t.TempDir()
	rootFile := filepath.Join(projectRoot, "root.tscn")
	inheritedFile := filepath.Join(projectRoot, "scenes", "inherited.tscn")
	baseFile := filepath.Join(projectRoot, "base", "base.tscn")
	writeTestFile(t, rootFile, `[gd_scene format=3]
[ext_resource type="PackedScene" path="scenes/inherited.tscn" id="1_inherited"]
[node name="Root" type="Node3D"]
[node name="Inherited" parent="." instance=ExtResource("1_inherited")]
`)
	writeTestFile(t, inheritedFile, `[gd_scene format=3]
[ext_resource type="PackedScene" path="../base/base.tscn" id="1_base"]
[node name="Inherited" instance=ExtResource("1_base")]
`)
	writeTestFile(t, baseFile, `[gd_scene format=3]
[node name="Base" type="MeshInstance3D"]
`)

	resolver, err := project.NewResolver(projectRoot)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	root, err := resolver.ResolveSceneInput(rootFile, projectRoot)
	if err != nil {
		t.Fatalf("ResolveSceneInput() error = %v", err)
	}
	opener := func(path project.ResolvedPath) (io.ReadCloser, error) {
		return os.Open(path.Canonical)
	}
	analyzer, err := NewRecursiveAnalyzer(resolver, opener, tscn.Parse)
	if err != nil {
		t.Fatalf("NewRecursiveAnalyzer() error = %v", err)
	}

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	canonicalInherited, err := filepath.EvalSymlinks(inheritedFile)
	if err != nil {
		t.Fatalf("EvalSymlinks(inherited) error = %v", err)
	}
	canonicalBase, err := filepath.EvalSymlinks(baseFile)
	if err != nil {
		t.Fatalf("EvalSymlinks(base) error = %v", err)
	}
	if result.Graph.SceneDependencies != 2 ||
		findGraphEdge(t, result.Graph, canonicalInherited, EdgeInheritance).ToCanonical != canonicalBase {
		t.Fatalf("Graph = %#v", result.Graph)
	}
	if result.ParsedSceneFiles != 3 {
		t.Fatalf("ParsedSceneFiles = %d, want 3", result.ParsedSceneFiles)
	}
	if result.Summary.Metrics.MeshInstances != 0 || result.Summary.Metrics.SceneDependencies != 0 {
		t.Fatalf("deferred/final metrics leaked = %#v", result.Summary.Metrics)
	}
}

func findGraphEdge(t *testing.T, graph DependencyGraph, from string, kind EdgeKind) GraphEdge {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.FromCanonical == from && edge.Kind == kind {
			return edge
		}
	}
	t.Fatalf("edge from %q with kind %q not found in %#v", from, kind, graph.Edges)

	return GraphEdge{}
}

func requireCycle(t *testing.T, result RecursiveResult, err error) *CycleError {
	t.Helper()
	if !reflect.DeepEqual(result, RecursiveResult{}) {
		t.Fatalf("fatal cycle returned usable result: %#v", result)
	}
	var cycle *CycleError
	if !errors.As(err, &cycle) {
		t.Fatalf("error = %T %v, want *CycleError", err, err)
	}
	if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeSceneDependencyCycle {
		t.Fatalf("diagnostic code = %q, %v", code, ok)
	}

	return cycle
}

func testDecimal(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 3)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}

	return string(digits)
}
