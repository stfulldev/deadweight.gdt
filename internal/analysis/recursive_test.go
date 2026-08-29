package analysis

import (
	"errors"
	"fmt"
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

type resolverCall struct {
	from string
	raw  string
}

type memoryResolver struct {
	results map[string]project.Resolution
	calls   []resolverCall
}

func (resolver *memoryResolver) ResolveResource(fromScene, raw string) project.Resolution {
	resolver.calls = append(resolver.calls, resolverCall{from: fromScene, raw: raw})
	if result, exists := resolver.results[resolverKey(fromScene, raw)]; exists {
		return result
	}

	return project.Resolution{
		Reason: project.ResolutionMissing,
		Path:   project.ResolvedPath{Original: raw},
	}
}

type memorySceneEffects struct {
	sources     map[string]string
	errors      map[string]error
	parseErrors map[string]error
	closeErrors map[string]error
	calls       map[string]int
	parses      map[string]int
	closes      map[string]int
}

type memorySceneReader struct {
	io.Reader
	canonical string
	effects   *memorySceneEffects
}

func (reader *memorySceneReader) Close() error {
	reader.effects.ensureCounters()
	reader.effects.closes[reader.canonical]++

	return reader.effects.closeErrors[reader.canonical]
}

func (effects *memorySceneEffects) open(path project.ResolvedPath) (io.ReadCloser, error) {
	effects.ensureCounters()
	effects.calls[path.Canonical]++
	if err := effects.errors[path.Canonical]; err != nil {
		return nil, err
	}
	source, exists := effects.sources[path.Canonical]
	if !exists {
		return nil, fmt.Errorf("no in-memory scene for %s", path.Canonical)
	}

	return &memorySceneReader{
		Reader:    strings.NewReader(source),
		canonical: path.Canonical,
		effects:   effects,
	}, nil
}

func (effects *memorySceneEffects) parse(reader io.Reader, source string) (*tscn.Document, error) {
	effects.ensureCounters()
	memoryReader, ok := reader.(*memorySceneReader)
	if !ok {
		return nil, fmt.Errorf("unexpected in-memory reader %T", reader)
	}
	effects.parses[memoryReader.canonical]++
	if err := effects.parseErrors[memoryReader.canonical]; err != nil {
		return nil, err
	}

	return tscn.Parse(reader, source)
}

func (effects *memorySceneEffects) ensureCounters() {
	if effects.calls == nil {
		effects.calls = make(map[string]int)
	}
	if effects.parses == nil {
		effects.parses = make(map[string]int)
	}
	if effects.closes == nil {
		effects.closes = make(map[string]int)
	}
}

func TestRecursiveAnalyzerExpandsChainFromEachDeclaringScene(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, filepath.Join("nested", "child.tscn"))
	leaf := testScenePath(rootDir, "leaf.tscn")
	texture := testResourcePath(rootDir, "texture.png")

	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "nested/child.tscn"): resolvedResource("nested/child.tscn", child),
		resolverKey(child.Canonical, "../leaf.tscn"):     resolvedResource("../leaf.tscn", leaf),
		resolverKey(leaf.Canonical, "texture.png"):       resolvedResource("texture.png", texture),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="nested/child.tscn" id="1_child"]
[node name="Root" type="Node3D"]
[node name="Container" type="Node3D" parent="."]
[node name="Child" parent="Container" instance=ExtResource("1_child")]
`,
			child.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="../leaf.tscn" id="1_leaf"]
[node name="ChildRoot" type="MeshInstance3D"]
[node name="Branch" type="Node3D" parent="."]
[node name="Leaf" parent="Branch" instance=ExtResource("1_leaf")]
`,
			leaf.Canonical: `[gd_scene format=3]
[ext_resource type="Texture2D" path="texture.png" id="1_texture"]
[node name="LeafRoot" type="Node3D"]
[node name="Lamp" type="OmniLight3D" parent="."]
shadow_enabled = true
`,
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)
	builds := make(map[*tscn.Document]int)
	analyzer.summarize = func(document *tscn.Document) (LocalSummary, error) {
		builds[document]++
		return BuildLocalSummary(document)
	}

	summary, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("Expand() error = %v", err)
	}
	wantMetrics := metrics.Values{
		Nodes:          6,
		TreeDepth:      6,
		SceneInstances: 2,
		MeshInstances:  1,
		Lights:         1,
		ShadowLights:   1,
	}
	if summary.Metrics != wantMetrics {
		t.Fatalf("Metrics = %#v, want %#v", summary.Metrics, wantMetrics)
	}
	if summary.Coverage != (SceneInstanceCoverage{Resolved: 2}) {
		t.Fatalf("Coverage = %#v", summary.Coverage)
	}
	wantDependencies := []string{child.Canonical, leaf.Canonical}
	sort.Strings(wantDependencies)
	if !reflect.DeepEqual(summary.Dependencies, wantDependencies) {
		t.Fatalf("Dependencies = %#v, want %#v", summary.Dependencies, wantDependencies)
	}
	wantResources := []string{child.Canonical, leaf.Canonical, texture.Canonical}
	sort.Strings(wantResources)
	if got := resourceCanonicals(summary.ExternalResources); !reflect.DeepEqual(got, wantResources) {
		t.Fatalf("resolved resources = %#v, want %#v", got, wantResources)
	}
	for _, path := range []project.ResolvedPath{root, child, leaf} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
	if len(builds) != 3 {
		t.Fatalf("local summary document count = %d, want 3", len(builds))
	}
	for document, count := range builds {
		if document == nil || count != 1 {
			t.Fatalf("local summary builds for %p = %d, want 1", document, count)
		}
	}
	wantCalls := []resolverCall{
		{from: root.Canonical, raw: "nested/child.tscn"},
		{from: child.Canonical, raw: "../leaf.tscn"},
		{from: leaf.Canonical, raw: "texture.png"},
	}
	if !reflect.DeepEqual(resolver.calls, wantCalls) {
		t.Fatalf("resolver calls = %#v, want %#v", resolver.calls, wantCalls)
	}
}

func TestRecursiveAnalyzerAppliesRepeatedSummaryOneHundredTimesAndResetsInvocation(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	leaf := testScenePath(rootDir, "leaf.tscn")

	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
		resolverKey(child.Canonical, "leaf.tscn"): resolvedResource("leaf.tscn", leaf),
	}}
	var mounts strings.Builder
	for index := 0; index < 100; index++ {
		fmt.Fprintf(&mounts, "[node name=\"Child%d\" parent=\".\" instance=ExtResource(\"1_child\")]\n", index)
	}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="child.tscn" id="1_child"]
[node name="Root" type="Node3D"]
` + mounts.String(),
			child.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="leaf.tscn" id="1_leaf"]
[node name="ChildRoot" type="MeshInstance3D"]
[node name="Nested" parent="." instance=ExtResource("1_leaf")]
`,
			leaf.Canonical: `[gd_scene format=3]
[node name="LeafRoot" type="Node3D"]
`,
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)
	builds := make(map[*tscn.Document]int)
	analyzer.summarize = func(document *tscn.Document) (LocalSummary, error) {
		builds[document]++
		return BuildLocalSummary(document)
	}

	first, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("first Expand() error = %v", err)
	}
	wantMetrics := metrics.Values{
		Nodes:          201,
		TreeDepth:      3,
		SceneInstances: 200,
		MeshInstances:  100,
	}
	if first.Metrics != wantMetrics || first.Coverage != (SceneInstanceCoverage{Resolved: 200}) {
		t.Fatalf("first summary = %#v / %#v", first.Metrics, first.Coverage)
	}
	wantDependencies := []string{child.Canonical, leaf.Canonical}
	sort.Strings(wantDependencies)
	if !reflect.DeepEqual(first.Dependencies, wantDependencies) {
		t.Fatalf("dependencies = %#v, want %#v", first.Dependencies, wantDependencies)
	}
	if len(builds) != 3 {
		t.Fatalf("built document count = %d, want 3", len(builds))
	}
	for document, count := range builds {
		if document == nil || count != 1 {
			t.Fatalf("build count = %d for %#v", count, document)
		}
	}
	for _, path := range []project.ResolvedPath{root, child, leaf} {
		requireMemorySceneEffects(t, loader, path, 1)
	}

	first.Dependencies[0] = "mutated"
	first.ExternalResources[0].Canonical = "mutated"
	second, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("second Expand() error = %v", err)
	}
	if second.Metrics != wantMetrics || !reflect.DeepEqual(second.Dependencies, wantDependencies) {
		t.Fatalf("second summary changed after caller mutation: %#v", second)
	}
	for _, path := range []project.ResolvedPath{root, child, leaf} {
		requireMemorySceneEffects(t, loader, path, 2)
	}
}

func TestRecursiveAnalyzerReusesDiamondDescendantAndUnionsEvidence(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	left := testScenePath(rootDir, "left.tscn")
	right := testScenePath(rootDir, "right.tscn")
	shared := testScenePath(rootDir, "shared.tscn")
	asset := testResourcePath(rootDir, "shared.res")

	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "left.tscn"):    resolvedResource("left.tscn", left),
		resolverKey(root.Canonical, "right.tscn"):   resolvedResource("right.tscn", right),
		resolverKey(left.Canonical, "shared.tscn"):  resolvedResource("shared.tscn", shared),
		resolverKey(right.Canonical, "shared.tscn"): resolvedResource("shared.tscn", shared),
		resolverKey(shared.Canonical, "shared.res"): resolvedResource("shared.res", asset),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:  sceneWithMounts("Root", []resourceMount{{"1_left", "left.tscn"}, {"2_right", "right.tscn"}}),
			left.Canonical:  sceneWithMounts("Left", []resourceMount{{"1_shared", "shared.tscn"}}),
			right.Canonical: sceneWithMounts("Right", []resourceMount{{"1_shared", "shared.tscn"}}),
			shared.Canonical: `[gd_scene format=3]
[ext_resource type="Resource" path="shared.res" id="1_asset"]
[node name="Shared" type="Node3D"]
`,
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)
	builds := make(map[*tscn.Document]int)
	analyzer.summarize = func(document *tscn.Document) (LocalSummary, error) {
		builds[document]++
		return BuildLocalSummary(document)
	}

	summary, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("Expand() error = %v", err)
	}
	if summary.Metrics.Nodes != 5 || summary.Metrics.SceneInstances != 4 || summary.Coverage.Resolved != 4 {
		t.Fatalf("summary metrics/coverage = %#v/%#v", summary.Metrics, summary.Coverage)
	}
	wantDependencies := []string{left.Canonical, right.Canonical, shared.Canonical}
	sort.Strings(wantDependencies)
	if !reflect.DeepEqual(summary.Dependencies, wantDependencies) {
		t.Fatalf("dependencies = %#v, want %#v", summary.Dependencies, wantDependencies)
	}
	if occurrences(summary.ExternalResources, shared.Canonical) != 1 || occurrences(summary.ExternalResources, asset.Canonical) != 1 {
		t.Fatalf("resource union = %#v", summary.ExternalResources)
	}
	for _, path := range []project.ResolvedPath{root, left, right, shared} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
	if len(builds) != 4 {
		t.Fatalf("local summary document count = %d, want 4", len(builds))
	}
	for document, count := range builds {
		if document == nil || count != 1 {
			t.Fatalf("local summary builds for %p = %d, want 1", document, count)
		}
	}
}

func TestRecursiveAnalyzerClassifiesEveryUnresolvedTarget(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	imported := testResourcePath(rootDir, "model.glb")
	unsupported := testResourcePath(rootDir, "logic.gd")
	unavailable := testScenePath(rootDir, "unavailable.tscn")
	inherited := testScenePath(rootDir, "inherited.tscn")
	base := testScenePath(rootDir, "base.tscn")
	wrongType := testScenePath(rootDir, "wrong-type.tscn")

	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "missing.tscn"): {
			Reason: project.ResolutionMissing,
			Path:   project.ResolvedPath{Original: "missing.tscn"},
		},
		resolverKey(root.Canonical, "../outside.tscn"): {
			Reason: project.ResolutionOutsideProject,
			Path:   project.ResolvedPath{Original: "../outside.tscn"},
		},
		resolverKey(root.Canonical, "uid://abc"): {
			Reason: project.ResolutionUIDOnly,
			Path:   project.ResolvedPath{Original: "uid://abc"},
		},
		resolverKey(root.Canonical, "user://save.tscn"): {
			Reason: project.ResolutionUserData,
			Path:   project.ResolvedPath{Original: "user://save.tscn"},
		},
		resolverKey(root.Canonical, "model.glb"):        resolvedResource("model.glb", imported),
		resolverKey(root.Canonical, "logic.gd"):         resolvedResource("logic.gd", unsupported),
		resolverKey(root.Canonical, "unavailable.tscn"): resolvedResource("unavailable.tscn", unavailable),
		resolverKey(root.Canonical, "inherited.tscn"):   resolvedResource("inherited.tscn", inherited),
		resolverKey(inherited.Canonical, "base.tscn"):   resolvedResource("base.tscn", base),
		resolverKey(root.Canonical, "wrong-type.tscn"):  resolvedResource("wrong-type.tscn", wrongType),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="missing.tscn" id="1_missing"]
[ext_resource type="PackedScene" path="../outside.tscn" id="2_outside"]
[ext_resource type="PackedScene" path="uid://abc" id="3_uid"]
[ext_resource type="PackedScene" path="user://save.tscn" id="4_user"]
[ext_resource type="PackedScene" path="model.glb" id="5_imported"]
[ext_resource type="PackedScene" path="logic.gd" id="6_unsupported"]
[ext_resource type="PackedScene" path="unavailable.tscn" id="7_unavailable"]
[ext_resource type="PackedScene" path="inherited.tscn" id="8_inherited"]
[ext_resource type="Script" path="wrong-type.tscn" id="9_wrong_type"]
[node name="Root" type="Node3D"]
[node name="WrongID" parent="." instance=ExtResource("0_unknown")]
[node name="Missing" parent="." instance=ExtResource("1_missing")]
[node name="Outside" parent="." instance=ExtResource("2_outside")]
[node name="UID" parent="." instance=ExtResource("3_uid")]
[node name="User" parent="." instance=ExtResource("4_user")]
[node name="Embedded" parent="." instance=SubResource("Scene_1")]
[node name="Placeholder" parent="." instance_placeholder="res://later.tscn"]
[node name="Imported" parent="." instance=ExtResource("5_imported")]
[node name="Unsupported" parent="." instance=ExtResource("6_unsupported")]
[node name="Unavailable" parent="." instance=ExtResource("7_unavailable")]
[node name="Inherited" parent="." instance=ExtResource("8_inherited")]
[node name="WrongType" parent="." instance=ExtResource("9_wrong_type")]
`,
			inherited.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[node name="InheritedRoot" instance=ExtResource("1_base")]
`,
			wrongType.Canonical: `[gd_scene format=3]
[node name="WrongTypeRoot" type="MeshInstance3D"]
`,
		},
		errors: map[string]error{unavailable.Canonical: errors.New("permission denied")},
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	summary, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("Expand() error = %v", err)
	}
	if summary.Metrics != (metrics.Values{Nodes: 13, TreeDepth: 2, SceneInstances: 12, MeshInstances: 1}) {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
	if summary.Coverage != (SceneInstanceCoverage{Resolved: 1, Unresolved: 11}) {
		t.Fatalf("Coverage = %#v", summary.Coverage)
	}
	wantClassifications := map[TargetClassification]int{
		TargetMissingExternalResource: 1,
		TargetUnresolvedPath:          4,
		TargetSubResource:             1,
		TargetPlaceholder:             1,
		TargetImportedScene:           1,
		TargetUnsupportedScene:        1,
		TargetUnavailableScene:        1,
	}
	gotClassifications := make(map[TargetClassification]int)
	for _, evidence := range summary.Unresolved {
		gotClassifications[evidence.Classification]++
		if evidence.DeclaringScene != root.Canonical || evidence.DeclaringDisplay != root.Display || evidence.MountName == "" || evidence.MountPath == "" || !evidence.MountDepth.Known || evidence.Position.Line == 0 || evidence.Occurrences != 1 {
			t.Errorf("incomplete unresolved evidence: %#v", evidence)
		}
	}
	if !reflect.DeepEqual(gotClassifications, wantClassifications) {
		t.Fatalf("classifications = %#v, want %#v", gotClassifications, wantClassifications)
	}
	if len(summary.InheritedTargets) != 1 {
		t.Fatalf("InheritedTargets = %#v", summary.InheritedTargets)
	}
	inheritedEvidence := summary.InheritedTargets[0]
	if inheritedEvidence.TargetCanonical != inherited.Canonical || inheritedEvidence.BaseCanonical != base.Canonical || inheritedEvidence.Occurrences != 1 || inheritedEvidence.Classification != TargetInheritedScene {
		t.Fatalf("inherited evidence = %#v", inheritedEvidence)
	}
	wantDependencies := []string{inherited.Canonical, wrongType.Canonical}
	sort.Strings(wantDependencies)
	if !reflect.DeepEqual(summary.Dependencies, wantDependencies) {
		t.Fatalf("Dependencies = %#v, want %#v", summary.Dependencies, wantDependencies)
	}
	if loader.calls[unavailable.Canonical] != 1 || loader.calls[inherited.Canonical] != 1 || loader.calls[wrongType.Canonical] != 1 {
		t.Fatalf("loader calls = %#v", loader.calls)
	}
}

func TestRecursiveAnalyzerPreservesUnknownDepthWhileExpandingMetrics(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="child.tscn" id="1_child"]
[node name="Root" type="Node3D"]
[node name="Child" parent="Missing" instance=ExtResource("1_child")]
`,
			child.Canonical: `[gd_scene format=3]
[node name="ChildRoot" type="Node3D"]
[node name="Deep" type="MeshInstance3D" parent="."]
`,
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	summary, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("Expand() error = %v", err)
	}
	if summary.Metrics != (metrics.Values{Nodes: 3, TreeDepth: 1, SceneInstances: 1, MeshInstances: 1}) {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
	if !summary.DepthPartial || len(summary.ParentFindings) != 1 || summary.ParentFindings[0].DeclaringScene != root.Canonical {
		t.Fatalf("depth evidence = %v/%#v", summary.DepthPartial, summary.ParentFindings)
	}
}

func TestRecursiveAnalyzerReturnsFatalNestedParseErrorWithoutResult(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:  sceneWithMounts("Root", []resourceMount{{"1_child", "child.tscn"}}),
			child.Canonical: "this is not a scene",
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err == nil || !reflect.DeepEqual(result, RecursiveResult{}) {
		t.Fatalf("Analyze() = %#v, %v; want zero result and error", result, err)
	}
	if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeInvalidTSCNRoot {
		t.Fatalf("diagnostic code = %q, %v", code, ok)
	}
	var parseErr *tscn.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error type = %T, want *tscn.ParseError", err)
	}
	for _, path := range []project.ResolvedPath{root, child} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerReportsRecursiveReferenceAsSB2002Cycle(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	child := testScenePath(rootDir, "child.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "child.tscn"): resolvedResource("child.tscn", child),
		resolverKey(child.Canonical, "root.tscn"): resolvedResource("root.tscn", root),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:  sceneWithMounts("Root", []resourceMount{{"1_child", "child.tscn"}}),
			child.Canonical: sceneWithMounts("Child", []resourceMount{{"1_root", "root.tscn"}}),
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	summary, err := analyzer.Expand(root)
	var cycle *CycleError
	if !errors.As(err, &cycle) || !reflect.DeepEqual(summary, ExpandedSummary{}) {
		t.Fatalf("Expand() = %#v, %T %v", summary, err, err)
	}
	if code, coded := diagnostic.CodeOf(err); !coded || code != diagnostic.CodeSceneDependencyCycle {
		t.Fatalf("cycle diagnostic code = %q, %v", code, coded)
	}
	if !reflect.DeepEqual(cycle.Canonical, []string{root.Canonical, child.Canonical, root.Canonical}) ||
		!reflect.DeepEqual(cycle.Display, []string{root.Display, child.Display, root.Display}) {
		t.Fatalf("cycle = %#v", cycle)
	}
	if loader.calls[root.Canonical] != 1 || loader.calls[child.Canonical] != 1 {
		t.Fatalf("loader calls = %#v", loader.calls)
	}
}

func TestRecursiveAnalyzerOutputIsDeterministic(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	a := testResourcePath(rootDir, "a.gd")
	z := testResourcePath(rootDir, "z.gd")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "z.gd"): resolvedResource("z.gd", z),
		resolverKey(root.Canonical, "a.gd"): resolvedResource("a.gd", a),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="Script" path="z.gd" id="2_z"]
[ext_resource type="Script" path="a.gd" id="1_a"]
[node name="Root" type="Node3D"]
[node name="Z" parent="." instance=ExtResource("2_z")]
[node name="A" parent="." instance=ExtResource("1_a")]
`,
		},
		errors: make(map[string]error),
		calls:  make(map[string]int),
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	first, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("first Expand() error = %v", err)
	}
	second, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("second Expand() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("results differ:\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func TestCheckedRecursiveArithmeticBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		operation func() (int64, error)
		want      int64
		wantError bool
	}{
		{name: "maximum addition", operation: func() (int64, error) { return checkedAdd(math.MaxInt64, 0) }, want: math.MaxInt64},
		{name: "addition overflow", operation: func() (int64, error) { return checkedAdd(math.MaxInt64, 1) }, wantError: true},
		{name: "maximum multiply", operation: func() (int64, error) { return checkedMultiply(math.MaxInt64, 1) }, want: math.MaxInt64},
		{name: "multiply overflow", operation: func() (int64, error) { return checkedMultiply(math.MaxInt64, 2) }, wantError: true},
		{name: "maximum depth", operation: func() (int64, error) { return checkedDepth(math.MaxInt64, 1) }, want: math.MaxInt64},
		{name: "depth overflow", operation: func() (int64, error) { return checkedDepth(math.MaxInt64, 2) }, wantError: true},
		{name: "maximum cardinality", operation: func() (int64, error) { return checkedCardinality(uint64(math.MaxInt64)) }, want: math.MaxInt64},
		{name: "cardinality overflow", operation: func() (int64, error) { return checkedCardinality(uint64(math.MaxInt64) + 1) }, wantError: true},
		{name: "negative operand", operation: func() (int64, error) { return checkedAdd(-1, 1) }, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.operation()
			if test.wantError {
				assertOverflow(t, err)
				if got != 0 {
					t.Fatalf("result = %d, want zero on failure", got)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("operation = %d, %v; want %d", got, err, test.want)
			}
		})
	}
}

func TestResolvedApplicationRejectsMetricCoverageAndDepthOverflow(t *testing.T) {
	tests := []struct {
		name         string
		builder      summaryBuilder
		key          resolvedApplicationKey
		multiplicity int64
		child        ExpandedSummary
	}{
		{
			name:    "metric",
			builder: summaryBuilder{summary: ExpandedSummary{Metrics: metrics.Values{Nodes: math.MaxInt64}}, resources: map[ResourceIdentity]struct{}{}, dependencies: map[string]struct{}{}},
			key:     resolvedApplicationKey{known: true, depth: 1},
			child:   ExpandedSummary{Metrics: metrics.Values{Nodes: 1, TreeDepth: 1}},
		},
		{
			name:    "coverage",
			builder: summaryBuilder{resources: map[ResourceIdentity]struct{}{}, dependencies: map[string]struct{}{}},
			key:     resolvedApplicationKey{known: true, depth: 1},
			child:   ExpandedSummary{Metrics: metrics.Values{TreeDepth: 1}, Coverage: SceneInstanceCoverage{Resolved: math.MaxInt64}},
		},
		{
			name:    "depth",
			builder: summaryBuilder{resources: map[ResourceIdentity]struct{}{}, dependencies: map[string]struct{}{}},
			key:     resolvedApplicationKey{known: true, depth: math.MaxInt64},
			child:   ExpandedSummary{Metrics: metrics.Values{TreeDepth: 2}},
		},
		{
			name:         "unresolved evidence",
			builder:      summaryBuilder{resources: map[ResourceIdentity]struct{}{}, dependencies: map[string]struct{}{}},
			key:          resolvedApplicationKey{known: true, depth: 1},
			multiplicity: 2,
			child: ExpandedSummary{
				Metrics:    metrics.Values{TreeDepth: 1},
				Unresolved: []UnresolvedInstance{{Occurrences: math.MaxInt64}},
			},
		},
		{
			name:         "inherited evidence",
			builder:      summaryBuilder{resources: map[ResourceIdentity]struct{}{}, dependencies: map[string]struct{}{}},
			key:          resolvedApplicationKey{known: true, depth: 1},
			multiplicity: 2,
			child: ExpandedSummary{
				Metrics:          metrics.Values{TreeDepth: 1},
				InheritedTargets: []InheritedTarget{{Occurrences: math.MaxInt64}},
			},
		},
		{
			name:         "parent finding evidence",
			builder:      summaryBuilder{resources: map[ResourceIdentity]struct{}{}, dependencies: map[string]struct{}{}},
			key:          resolvedApplicationKey{known: true, depth: 1},
			multiplicity: 2,
			child: ExpandedSummary{
				Metrics:        metrics.Values{TreeDepth: 1},
				ParentFindings: []SceneParentFinding{{Occurrences: math.MaxInt64}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := cloneExpandedSummary(test.builder.summary)
			multiplicity := test.multiplicity
			if multiplicity == 0 {
				multiplicity = 1
			}
			err := test.builder.applyResolved(
				project.ResolvedPath{Canonical: filepath.Join(t.TempDir(), "child.tscn")},
				test.key,
				multiplicity,
				test.child,
			)
			assertOverflow(t, err)
			if !reflect.DeepEqual(test.builder.summary, before) || len(test.builder.resources) != 0 || len(test.builder.dependencies) != 0 {
				t.Fatalf("overflow mutated builder: %#v", test.builder)
			}
		})
	}
}

func TestEvidenceOverflowDoesNotMutateSummaryBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder summaryBuilder
		apply   func(*summaryBuilder) error
	}{
		{
			name:    "unresolved node",
			builder: summaryBuilder{summary: ExpandedSummary{Metrics: metrics.Values{Nodes: math.MaxInt64}}},
			apply: func(builder *summaryBuilder) error {
				return builder.addUnresolved(UnresolvedInstance{Occurrences: 1})
			},
		},
		{
			name:    "unresolved coverage",
			builder: summaryBuilder{summary: ExpandedSummary{Coverage: SceneInstanceCoverage{Unresolved: math.MaxInt64}}},
			apply: func(builder *summaryBuilder) error {
				return builder.addUnresolved(UnresolvedInstance{Occurrences: 1})
			},
		},
		{
			name:    "inherited coverage",
			builder: summaryBuilder{summary: ExpandedSummary{Coverage: SceneInstanceCoverage{Unresolved: math.MaxInt64}}},
			apply: func(builder *summaryBuilder) error {
				return builder.addInherited(
					project.ResolvedPath{},
					InstanceMount{},
					&inheritedSceneError{},
				)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := cloneExpandedSummary(test.builder.summary)
			assertOverflow(t, test.apply(&test.builder))
			if !reflect.DeepEqual(test.builder.summary, before) {
				t.Fatalf("overflow mutated builder: %#v", test.builder.summary)
			}
		})
	}
}

func TestRecursiveAnalyzerUsesRealProjectResolverAndParsedScenes(t *testing.T) {
	projectRoot := t.TempDir()
	rootFile := filepath.Join(projectRoot, "root.tscn")
	childFile := filepath.Join(projectRoot, "scenes", "nested", "child.tscn")
	assetFile := filepath.Join(projectRoot, "assets", "mesh.res")
	writeTestFile(t, rootFile, `[gd_scene format=3]
[ext_resource type="PackedScene" path="scenes/nested/child.tscn" id="1_child"]
[node name="Root" type="Node3D"]
[node name="Child" parent="." instance=ExtResource("1_child")]
`)
	writeTestFile(t, childFile, `[gd_scene format=3]
[ext_resource type="Resource" path="../../assets/mesh.res" id="1_mesh"]
[node name="ChildRoot" type="MeshInstance3D"]
`)
	writeTestFile(t, assetFile, "resource placeholder")

	resolver, err := project.NewResolver(projectRoot)
	if err != nil {
		t.Fatalf("NewResolver() error = %v", err)
	}
	root, err := resolver.ResolveSceneInput(rootFile, projectRoot)
	if err != nil {
		t.Fatalf("ResolveSceneInput() error = %v", err)
	}
	opens := make(map[string]int)
	parses := make(map[string]int)
	opener := func(path project.ResolvedPath) (io.ReadCloser, error) {
		opens[path.Canonical]++
		return os.Open(path.Canonical)
	}
	parser := func(reader io.Reader, source string) (*tscn.Document, error) {
		parses[source]++
		return tscn.Parse(reader, source)
	}
	analyzer, err := NewRecursiveAnalyzer(resolver, opener, parser)
	if err != nil {
		t.Fatalf("NewRecursiveAnalyzer() error = %v", err)
	}

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	summary := result.Summary
	canonicalChild, err := filepath.EvalSymlinks(childFile)
	if err != nil {
		t.Fatalf("EvalSymlinks(child) error = %v", err)
	}
	canonicalAsset, err := filepath.EvalSymlinks(assetFile)
	if err != nil {
		t.Fatalf("EvalSymlinks(asset) error = %v", err)
	}
	if summary.Metrics != (metrics.Values{Nodes: 2, TreeDepth: 2, SceneInstances: 1, MeshInstances: 1}) {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
	if !reflect.DeepEqual(summary.Dependencies, []string{canonicalChild}) {
		t.Fatalf("Dependencies = %#v", summary.Dependencies)
	}
	if occurrences(summary.ExternalResources, canonicalChild) != 1 || occurrences(summary.ExternalResources, canonicalAsset) != 1 {
		t.Fatalf("ExternalResources = %#v", summary.ExternalResources)
	}
	if result.ParsedSceneFiles != 2 || opens[root.Canonical] != 1 || opens[canonicalChild] != 1 ||
		parses[root.Display] != 1 || parses["res://scenes/nested/child.tscn"] != 1 {
		t.Fatalf("cache instrumentation = parsed %d, opens %#v, parses %#v", result.ParsedSceneFiles, opens, parses)
	}
}

func TestRecursiveAnalyzerValidatesDependenciesAndCanonicalRoot(t *testing.T) {
	opener := func(project.ResolvedPath) (io.ReadCloser, error) { return nil, nil }
	parser := func(io.Reader, string) (*tscn.Document, error) { return nil, nil }
	resolver := &memoryResolver{}
	if analyzer, err := NewRecursiveAnalyzer(nil, opener, parser); err == nil || analyzer != nil {
		t.Fatalf("NewRecursiveAnalyzer(nil, opener, parser) = %#v, %v", analyzer, err)
	}
	if analyzer, err := NewRecursiveAnalyzer(resolver, nil, parser); err == nil || analyzer != nil {
		t.Fatalf("NewRecursiveAnalyzer(resolver, nil, parser) = %#v, %v", analyzer, err)
	}
	if analyzer, err := NewRecursiveAnalyzer(resolver, opener, nil); err == nil || analyzer != nil {
		t.Fatalf("NewRecursiveAnalyzer(resolver, opener, nil) = %#v, %v", analyzer, err)
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, &memorySceneEffects{sources: map[string]string{}, errors: map[string]error{}, calls: map[string]int{}})
	invalidRoots := []project.ResolvedPath{
		{},
		{Canonical: "relative.tscn", Display: "res://relative.tscn"},
		{Canonical: filepath.Join(t.TempDir(), "scene.scn"), Display: "res://scene.scn"},
		{Canonical: filepath.Join(t.TempDir(), "scene.tscn")},
	}
	for _, root := range invalidRoots {
		if summary, err := analyzer.Expand(root); err == nil || !reflect.DeepEqual(summary, ExpandedSummary{}) {
			t.Errorf("Expand(%#v) = %#v, %v", root, summary, err)
		}
	}
}

type resourceMount struct {
	id   string
	path string
}

func sceneWithMounts(rootName string, mounts []resourceMount) string {
	var source strings.Builder
	source.WriteString("[gd_scene format=3]\n")
	for _, mount := range mounts {
		fmt.Fprintf(&source, "[ext_resource type=\"PackedScene\" path=\"%s\" id=\"%s\"]\n", mount.path, mount.id)
	}
	fmt.Fprintf(&source, "[node name=\"%s\" type=\"Node3D\"]\n", rootName)
	for index, mount := range mounts {
		fmt.Fprintf(&source, "[node name=\"Mount%d\" parent=\".\" instance=ExtResource(\"%s\")]\n", index, mount.id)
	}

	return source.String()
}

func newTestRecursiveAnalyzer(t *testing.T, resolver ResourceResolver, effects *memorySceneEffects) *RecursiveAnalyzer {
	t.Helper()
	analyzer, err := NewRecursiveAnalyzer(resolver, effects.open, effects.parse)
	if err != nil {
		t.Fatalf("NewRecursiveAnalyzer() error = %v", err)
	}

	return analyzer
}

func requireMemorySceneEffects(
	t *testing.T,
	effects *memorySceneEffects,
	path project.ResolvedPath,
	want int,
) {
	t.Helper()
	if effects.calls[path.Canonical] != want ||
		effects.parses[path.Canonical] != want ||
		effects.closes[path.Canonical] != want {
		t.Fatalf(
			"effects for %s = open %d, parse %d, close %d; want %d each",
			path.Display,
			effects.calls[path.Canonical],
			effects.parses[path.Canonical],
			effects.closes[path.Canonical],
			want,
		)
	}
}

func resolverKey(from, raw string) string {
	return from + "\x00" + raw
}

func resolvedResource(raw string, path project.ResolvedPath) project.Resolution {
	path.Original = raw

	return project.Resolution{Reason: project.ResolutionResolved, Path: path}
}

func testScenePath(root, relative string) project.ResolvedPath {
	canonical := filepath.Clean(filepath.Join(root, relative))

	return project.ResolvedPath{
		Canonical: canonical,
		Display:   "res://" + filepath.ToSlash(relative),
		Original:  filepath.ToSlash(relative),
	}
}

func testResourcePath(root, relative string) project.ResolvedPath {
	return testScenePath(root, relative)
}

func resourceCanonicals(resources []ResourceIdentity) []string {
	canonicals := make([]string, 0, len(resources))
	for _, resource := range resources {
		if resource.Resolved {
			canonicals = append(canonicals, resource.Canonical)
		}
	}
	sort.Strings(canonicals)

	return canonicals
}

func occurrences(resources []ResourceIdentity, canonical string) int {
	count := 0
	for _, resource := range resources {
		if resource.Resolved && resource.Canonical == canonical {
			count++
		}
	}

	return count
}

func assertOverflow(t *testing.T, err error) {
	t.Helper()
	var overflow *OverflowError
	if !errors.As(err, &overflow) {
		t.Fatalf("error = %T %v, want *OverflowError", err, err)
	}
	if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeArithmeticOverflow {
		t.Fatalf("diagnostic code = %q, %v", code, ok)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
