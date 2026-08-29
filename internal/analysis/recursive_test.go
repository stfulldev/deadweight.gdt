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

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	summary := result.Summary
	wantMetrics := metrics.Values{
		Nodes:             6,
		TreeDepth:         6,
		SceneInstances:    2,
		MeshInstances:     1,
		Lights:            1,
		ShadowLights:      1,
		ExternalResources: 3,
		SceneDependencies: 2,
	}
	if summary.Metrics != wantMetrics {
		t.Fatalf("Metrics = %#v, want %#v", summary.Metrics, wantMetrics)
	}
	if err := ValidateContributionEvidence(result); err != nil {
		t.Fatalf("contribution evidence = %v", err)
	}
	rootContribution := requireContribution(t, summary.Contributions, ContributionRoot, root.Canonical, "")
	childContribution := requireContribution(t, summary.Contributions, ContributionScene, child.Canonical, root.Canonical)
	leafContribution := requireContribution(t, summary.Contributions, ContributionScene, leaf.Canonical, child.Canonical)
	if rootContribution.Values != (ContributionValues{Nodes: 2}) ||
		childContribution.Values != (ContributionValues{Nodes: 2, SceneInstances: 1, MeshInstances: 1}) ||
		leafContribution.Values != (ContributionValues{Nodes: 2, SceneInstances: 1, Lights: 1, ShadowLights: 1}) ||
		leafContribution.DepthCandidate != (OptionalDepth{Value: 6, Known: true}) {
		t.Fatalf("chain contributions = %#v", summary.Contributions)
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
		Nodes:             201,
		TreeDepth:         3,
		SceneInstances:    200,
		MeshInstances:     100,
		ExternalResources: 2,
		SceneDependencies: 2,
	}
	if first.Metrics != wantMetrics || first.Coverage != (SceneInstanceCoverage{Resolved: 200}) {
		t.Fatalf("first summary = %#v / %#v", first.Metrics, first.Coverage)
	}
	if err := validateContributions(first.Contributions, first.Metrics, first.DepthPartial); err != nil {
		t.Fatalf("first contributions = %v", err)
	}
	leafContribution := requireContribution(t, first.Contributions, ContributionScene, leaf.Canonical, child.Canonical)
	if leafContribution.Occurrences != 100 || leafContribution.Values != (ContributionValues{Nodes: 100, SceneInstances: 100}) {
		t.Fatalf("repeated leaf contribution = %#v", leafContribution)
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
	first.Contributions[0].Values.Nodes = -1
	second, err := analyzer.Expand(root)
	if err != nil {
		t.Fatalf("second Expand() error = %v", err)
	}
	if secondContribution := requireContribution(t, second.Contributions, ContributionScene, leaf.Canonical, child.Canonical); secondContribution.Values.Nodes < 0 {
		t.Fatalf("second contribution aliases first result: %#v", secondContribution)
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

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	summary := result.Summary
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
	if summary.Metrics.ExternalResources != 4 || summary.Metrics.SceneDependencies != 3 {
		t.Fatalf("unique metrics = %#v", summary.Metrics)
	}
	if err := ValidateContributionEvidence(result); err != nil {
		t.Fatalf("contribution evidence = %v", err)
	}
	sharedDependency := requireUniqueEvidence(t, result.UniqueEvidence, metrics.SceneDependencies, shared.Canonical)
	if len(sharedDependency.Referrers) != 2 || sharedDependency.Referrers[0].SceneCanonical != left.Canonical || sharedDependency.Referrers[1].SceneCanonical != right.Canonical {
		t.Fatalf("shared dependency referrers = %#v", sharedDependency.Referrers)
	}
	sharedResource := requireUniqueEvidence(t, result.UniqueEvidence, metrics.ExternalResources, asset.Canonical)
	if len(sharedResource.Referrers) != 1 || sharedResource.Referrers[0].SceneCanonical != shared.Canonical {
		t.Fatalf("shared resource referrers = %#v", sharedResource.Referrers)
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

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	summary := result.Summary
	if summary.Metrics != (metrics.Values{
		Nodes: 13, TreeDepth: 2, SceneInstances: 12, MeshInstances: 1,
		ExternalResources: 10, SceneDependencies: 2,
	}) {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
	if summary.Coverage != (SceneInstanceCoverage{Resolved: 2, Unresolved: 10}) {
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
	if inheritedEvidence.DeclaringScene != inherited.Canonical || inheritedEvidence.BaseCanonical != base.Canonical || inheritedEvidence.BaseClassification != TargetUnavailableScene || inheritedEvidence.BaseResolutionReason != project.ResolutionFilesystem || inheritedEvidence.Occurrences != 1 || inheritedEvidence.Classification != TargetInheritedScene {
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
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityApproximate ||
		result.Coverage != (Coverage{
			ResolvedSceneInstances: 2, UnresolvedSceneInstances: 10,
			ParsedSceneFiles: 3, InheritedScenes: 1,
		}) || len(result.Diagnostics) != 11 {
		t.Fatalf("completeness = %q/%q/%#v/%#v", result.Status, result.Reliability, result.Coverage, result.Diagnostics)
	}
	unresolvedContributions := 0
	for _, contribution := range summary.Contributions {
		if contribution.Kind == ContributionUnresolved {
			unresolvedContributions++
			if contribution.Reliability != ReliabilityLowerBound {
				t.Errorf("unresolved contribution reliability = %q", contribution.Reliability)
			}
		}
	}
	if unresolvedContributions != 10 {
		t.Fatalf("unresolved contributions = %d, want 10", unresolvedContributions)
	}
}

func TestRecursiveAnalyzerAppliesResolvedInheritedBaseWithLocalAdditions(t *testing.T) {
	rootDir := t.TempDir()
	derived := testScenePath(rootDir, "derived.tscn")
	base := testScenePath(rootDir, "base.tscn")
	nested := testScenePath(rootDir, "nested.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(derived.Canonical, "base.tscn"):   resolvedResource("base.tscn", base),
		resolverKey(derived.Canonical, "nested.tscn"): resolvedResource("nested.tscn", nested),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			derived.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[ext_resource type="PackedScene" path="nested.tscn" id="2_nested"]
[node name="Derived" instance=ExtResource("1_base")]
[node name="Body" parent="."]
[node name="Hat" type="MeshInstance3D" parent="Body"]
[node name="Nested" parent="Body" instance=ExtResource("2_nested")]
[editable path="Body"]
`,
			base.Canonical: `[gd_scene format=3]
[node name="Base" type="Node3D"]
[node name="Body" type="MeshInstance3D" parent="."]
[node name="Sun" type="OmniLight3D" parent="Body"]
shadow_enabled = true
`,
			nested.Canonical: `[gd_scene format=3]
[node name="Nested" type="MeshInstance3D"]
`,
		},
		errors: map[string]error{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(derived)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	wantMetrics := metrics.Values{
		Nodes: 5, TreeDepth: 3, SceneInstances: 1, MeshInstances: 3,
		Lights: 1, ShadowLights: 1, ExternalResources: 2, SceneDependencies: 2,
	}
	if result.Summary.Metrics != wantMetrics || result.Summary.Coverage != (SceneInstanceCoverage{Resolved: 1}) {
		t.Fatalf("summary = %#v / %#v, want %#v", result.Summary.Metrics, result.Summary.Coverage, wantMetrics)
	}
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityApproximate ||
		result.Coverage != (Coverage{ResolvedSceneInstances: 1, ParsedSceneFiles: 3, InheritedScenes: 1}) {
		t.Fatalf("completeness = %q/%q/%#v", result.Status, result.Reliability, result.Coverage)
	}
	if len(result.Summary.InheritedTargets) != 1 {
		t.Fatalf("InheritedTargets = %#v", result.Summary.InheritedTargets)
	}
	baseContribution := requireContribution(t, result.Summary.Contributions, ContributionInherited, base.Canonical, derived.Canonical)
	if baseContribution.Reliability != ReliabilityApproximate || baseContribution.Values.SceneInstances != 0 {
		t.Fatalf("inherited base contribution = %#v", baseContribution)
	}
	evidence := result.Summary.InheritedTargets[0]
	if evidence.DeclaringScene != derived.Canonical || evidence.BaseResourceID != "1_base" ||
		evidence.BaseRawTarget != "base.tscn" || evidence.BaseCanonical != base.Canonical ||
		evidence.BaseResolutionReason != project.ResolutionResolved || evidence.BaseClassification != "" ||
		!evidence.HasOverrideStubs || !evidence.HasEditable || evidence.Occurrences != 1 ||
		evidence.InheritedRootPosition.Line != 4 {
		t.Fatalf("inheritance evidence = %#v", evidence)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != diagnostic.CodeInheritedScene ||
		result.Diagnostics[0].Resource != base.Display || result.Diagnostics[0].Occurrences != 1 {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
	for _, path := range []project.ResolvedPath{derived, base, nested} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerRetainsOneRootForUnsupportedInheritedBases(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		rootInstance   string
		resolution     project.ResolutionReason
		extension      string
		unreadable     bool
		classification TargetClassification
		wantReason     project.ResolutionReason
		wantRaw        string
	}{
		{name: "missing", raw: "missing.tscn", resolution: project.ResolutionMissing, classification: TargetUnresolvedPath, wantReason: project.ResolutionMissing, wantRaw: "missing.tscn"},
		{name: "unreadable", raw: "locked.tscn", extension: ".tscn", unreadable: true, classification: TargetUnavailableScene, wantReason: project.ResolutionFilesystem, wantRaw: "locked.tscn"},
		{name: "imported glb", raw: "model.glb", extension: ".glb", classification: TargetImportedScene, wantReason: project.ResolutionResolved, wantRaw: "model.glb"},
		{name: "binary scn", raw: "model.scn", extension: ".scn", classification: TargetImportedScene, wantReason: project.ResolutionResolved, wantRaw: "model.scn"},
		{name: "sub resource", rootInstance: `SubResource("Scene_1")`, classification: TargetSubResource, wantRaw: "Scene_1"},
		{name: "uid", raw: "uid://base", resolution: project.ResolutionUIDOnly, classification: TargetUnresolvedPath, wantReason: project.ResolutionUIDOnly, wantRaw: "uid://base"},
		{name: "user data", raw: "user://base.tscn", resolution: project.ResolutionUserData, classification: TargetUnresolvedPath, wantReason: project.ResolutionUserData, wantRaw: "user://base.tscn"},
		{name: "outside project", raw: "../base.tscn", resolution: project.ResolutionOutsideProject, classification: TargetUnresolvedPath, wantReason: project.ResolutionOutsideProject, wantRaw: "../base.tscn"},
		{name: "unsupported resource", raw: "base.res", extension: ".res", classification: TargetUnsupportedScene, wantReason: project.ResolutionResolved, wantRaw: "base.res"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rootDir := t.TempDir()
			root := testScenePath(rootDir, "derived.tscn")
			target := testResourcePath(rootDir, "base"+test.extension)
			resolver := &memoryResolver{results: map[string]project.Resolution{}}
			instance := test.rootInstance
			source := "[gd_scene format=3]\n"
			if instance == "" {
				instance = `ExtResource("1_base")`
				source += fmt.Sprintf("[ext_resource type=\"PackedScene\" path=\"%s\" id=\"1_base\"]\n", test.raw)
				resolution := project.Resolution{Reason: test.resolution, Path: project.ResolvedPath{Original: test.raw}}
				if test.extension != "" {
					resolution = resolvedResource(test.raw, target)
				}
				resolver.results[resolverKey(root.Canonical, test.raw)] = resolution
			}
			source += fmt.Sprintf("[node name=\"Derived\" instance=%s]\n", instance)
			source += "[node name=\"Local\" type=\"MeshInstance3D\" parent=\".\"]\n"
			loader := &memorySceneEffects{
				sources: map[string]string{root.Canonical: source},
				errors:  map[string]error{},
			}
			if test.unreadable {
				loader.errors[target.Canonical] = errors.New("permission denied")
			}
			analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

			result, err := analyzer.Analyze(root)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}
			wantExternal := int64(1)
			if test.rootInstance != "" {
				wantExternal = 0
			}
			if result.Summary.Metrics != (metrics.Values{
				Nodes: 2, TreeDepth: 2, MeshInstances: 1, ExternalResources: wantExternal,
			}) || result.Summary.Coverage != (SceneInstanceCoverage{}) {
				t.Fatalf("summary = %#v/%#v", result.Summary.Metrics, result.Summary.Coverage)
			}
			if result.Status != AnalysisPartial || result.Reliability != ReliabilityApproximate ||
				result.Coverage != (Coverage{ParsedSceneFiles: 1, InheritedScenes: 1}) {
				t.Fatalf("completeness = %q/%q/%#v", result.Status, result.Reliability, result.Coverage)
			}
			if len(result.Summary.InheritedTargets) != 1 {
				t.Fatalf("InheritedTargets = %#v", result.Summary.InheritedTargets)
			}
			evidence := result.Summary.InheritedTargets[0]
			if evidence.BaseClassification != test.classification || evidence.BaseResolutionReason != test.wantReason ||
				evidence.BaseRawTarget != test.wantRaw || evidence.Occurrences != 1 {
				t.Fatalf("evidence = %#v", evidence)
			}
			if test.extension != "" && evidence.BaseCanonical != target.Canonical {
				t.Fatalf("BaseCanonical = %q, want %q", evidence.BaseCanonical, target.Canonical)
			}
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != diagnostic.CodeInheritedScene {
				t.Fatalf("diagnostics = %#v", result.Diagnostics)
			}
			requireMemorySceneEffects(t, loader, root, 1)
			if test.unreadable && (loader.calls[target.Canonical] != 1 || loader.parses[target.Canonical] != 0) {
				t.Fatalf("unreadable effects = %#v/%#v", loader.calls, loader.parses)
			}
		})
	}
}

func TestRecursiveAnalyzerMultipliesRepeatedInheritedSummaryAndOwnsEvidence(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	derived := testScenePath(rootDir, "derived.tscn")
	base := testScenePath(rootDir, "base.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "derived.tscn"): resolvedResource("derived.tscn", derived),
		resolverKey(derived.Canonical, "base.tscn"): resolvedResource("base.tscn", base),
	}}
	var mounts strings.Builder
	for index := 0; index < 100; index++ {
		fmt.Fprintf(&mounts, "[node name=\"Derived%d\" parent=\".\" instance=ExtResource(\"1_derived\")]\n", index)
	}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="derived.tscn" id="1_derived"]
[node name="Root" type="Node3D"]
` + mounts.String(),
			derived.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[node name="Derived" instance=ExtResource("1_base")]
`,
			base.Canonical: `[gd_scene format=3]
[node name="Base" type="MeshInstance3D"]
`,
		},
		errors: map[string]error{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	first, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if first.Summary.Metrics != (metrics.Values{
		Nodes: 101, TreeDepth: 2, SceneInstances: 100, MeshInstances: 100,
		ExternalResources: 2, SceneDependencies: 2,
	}) || first.Summary.Coverage != (SceneInstanceCoverage{Resolved: 100}) {
		t.Fatalf("summary = %#v/%#v", first.Summary.Metrics, first.Summary.Coverage)
	}
	if first.Coverage != (Coverage{ResolvedSceneInstances: 100, ParsedSceneFiles: 3, InheritedScenes: 100}) ||
		len(first.Summary.InheritedTargets) != 1 || first.Summary.InheritedTargets[0].Occurrences != 100 ||
		len(first.Diagnostics) != 1 || first.Diagnostics[0].Occurrences != 100 {
		t.Fatalf("repeated evidence = %#v/%#v/%#v", first.Coverage, first.Summary.InheritedTargets, first.Diagnostics)
	}
	for _, path := range []project.ResolvedPath{root, derived, base} {
		requireMemorySceneEffects(t, loader, path, 1)
	}

	wantSecond := cloneRecursiveResult(first)
	first.Summary.InheritedTargets[0].BaseCanonical = "mutated"
	first.Diagnostics[0].Resource = "mutated"
	first.Summary.Dependencies[0] = "mutated"
	second, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("second Analyze() error = %v", err)
	}
	if !reflect.DeepEqual(second, wantSecond) {
		t.Fatalf("caller mutation changed later result:\nsecond: %#v\nwant: %#v", second, wantSecond)
	}
}

func TestRecursiveAnalyzerComposesTransitiveInheritance(t *testing.T) {
	rootDir := t.TempDir()
	a := testScenePath(rootDir, "a.tscn")
	b := testScenePath(rootDir, "b.tscn")
	c := testScenePath(rootDir, "c.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(c.Canonical, "b.tscn"): resolvedResource("b.tscn", b),
		resolverKey(b.Canonical, "a.tscn"): resolvedResource("a.tscn", a),
	}}
	loader := &memorySceneEffects{
		sources: map[string]string{
			a.Canonical: `[gd_scene format=3]
[node name="A" type="MeshInstance3D"]
`,
			b.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="a.tscn" id="1_base"]
[node name="B" instance=ExtResource("1_base")]
[node name="Lamp" type="OmniLight3D" parent="."]
`,
			c.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="b.tscn" id="1_base"]
[node name="C" instance=ExtResource("1_base")]
[node name="Mesh" type="MeshInstance3D" parent="."]
`,
		},
		errors: map[string]error{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(c)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Summary.Metrics != (metrics.Values{
		Nodes: 3, TreeDepth: 2, MeshInstances: 2, Lights: 1,
		ExternalResources: 2, SceneDependencies: 2,
	}) || result.Summary.Coverage != (SceneInstanceCoverage{}) {
		t.Fatalf("summary = %#v/%#v", result.Summary.Metrics, result.Summary.Coverage)
	}
	if result.Coverage != (Coverage{ParsedSceneFiles: 3, InheritedScenes: 2}) ||
		len(result.Summary.InheritedTargets) != 2 || len(result.Diagnostics) != 2 {
		t.Fatalf("transitive evidence = %#v/%#v/%#v", result.Coverage, result.Summary.InheritedTargets, result.Diagnostics)
	}
	for _, path := range []project.ResolvedPath{a, b, c} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerReusesInheritedBaseAcrossDiamond(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	left := testScenePath(rootDir, "left.tscn")
	right := testScenePath(rootDir, "right.tscn")
	base := testScenePath(rootDir, "base.tscn")
	resolver := &memoryResolver{results: map[string]project.Resolution{
		resolverKey(root.Canonical, "left.tscn"):  resolvedResource("left.tscn", left),
		resolverKey(root.Canonical, "right.tscn"): resolvedResource("right.tscn", right),
		resolverKey(left.Canonical, "base.tscn"):  resolvedResource("base.tscn", base),
		resolverKey(right.Canonical, "base.tscn"): resolvedResource("base.tscn", base),
	}}
	derivedSource := func(name string) string {
		return fmt.Sprintf(`[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[node name="%s" instance=ExtResource("1_base")]
`, name)
	}
	loader := &memorySceneEffects{
		sources: map[string]string{
			root.Canonical:  sceneWithMounts("Root", []resourceMount{{"1_left", "left.tscn"}, {"2_right", "right.tscn"}}),
			left.Canonical:  derivedSource("Left"),
			right.Canonical: derivedSource("Right"),
			base.Canonical: `[gd_scene format=3]
[node name="Base" type="MeshInstance3D"]
`,
		},
		errors: map[string]error{},
	}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Summary.Metrics != (metrics.Values{
		Nodes: 3, TreeDepth: 2, SceneInstances: 2, MeshInstances: 2,
		ExternalResources: 3, SceneDependencies: 3,
	}) || result.Summary.Coverage != (SceneInstanceCoverage{Resolved: 2}) {
		t.Fatalf("summary = %#v/%#v", result.Summary.Metrics, result.Summary.Coverage)
	}
	if result.Coverage != (Coverage{ResolvedSceneInstances: 2, ParsedSceneFiles: 4, InheritedScenes: 2}) ||
		len(result.Summary.InheritedTargets) != 2 || len(result.Diagnostics) != 2 {
		t.Fatalf("diamond evidence = %#v/%#v/%#v", result.Coverage, result.Summary.InheritedTargets, result.Diagnostics)
	}
	for _, canonical := range []string{left.Canonical, right.Canonical, base.Canonical} {
		if occurrences(result.Summary.ExternalResources, canonical) != 1 {
			t.Errorf("resource %q occurrences = %d", canonical, occurrences(result.Summary.ExternalResources, canonical))
		}
	}
	for _, path := range []project.ResolvedPath{root, left, right, base} {
		requireMemorySceneEffects(t, loader, path, 1)
	}
}

func TestRecursiveAnalyzerRejectsMalformedInheritedBaseAndInheritanceCycle(t *testing.T) {
	t.Run("malformed base", func(t *testing.T) {
		rootDir := t.TempDir()
		root := testScenePath(rootDir, "root.tscn")
		base := testScenePath(rootDir, "base.tscn")
		resolver := &memoryResolver{results: map[string]project.Resolution{
			resolverKey(root.Canonical, "base.tscn"): resolvedResource("base.tscn", base),
		}}
		loader := &memorySceneEffects{
			sources: map[string]string{
				root.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="base.tscn" id="1_base"]
[node name="Root" instance=ExtResource("1_base")]
`,
				base.Canonical: "not a scene",
			},
			errors: map[string]error{},
		}
		analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

		result, err := analyzer.Analyze(root)
		if err == nil || !reflect.DeepEqual(result, RecursiveResult{}) {
			t.Fatalf("Analyze() = %#v, %v; want fatal zero result", result, err)
		}
		if code, ok := diagnostic.CodeOf(err); !ok || code != diagnostic.CodeInvalidTSCNRoot {
			t.Fatalf("diagnostic code = %q, %v", code, ok)
		}
	})

	t.Run("inheritance cycle", func(t *testing.T) {
		rootDir := t.TempDir()
		a := testScenePath(rootDir, "a.tscn")
		b := testScenePath(rootDir, "b.tscn")
		resolver := &memoryResolver{results: map[string]project.Resolution{
			resolverKey(a.Canonical, "b.tscn"): resolvedResource("b.tscn", b),
			resolverKey(b.Canonical, "a.tscn"): resolvedResource("a.tscn", a),
		}}
		loader := &memorySceneEffects{
			sources: map[string]string{
				a.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="b.tscn" id="1_base"]
[node name="A" instance=ExtResource("1_base")]
`,
				b.Canonical: `[gd_scene format=3]
[ext_resource type="PackedScene" path="a.tscn" id="1_base"]
[node name="B" instance=ExtResource("1_base")]
`,
			},
			errors: map[string]error{},
		}
		analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

		result, err := analyzer.Analyze(a)
		cycle := requireCycle(t, result, err)
		if !reflect.DeepEqual(cycle.Canonical, []string{a.Canonical, b.Canonical, a.Canonical}) {
			t.Fatalf("cycle = %#v", cycle)
		}
	})
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

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	summary := result.Summary
	if summary.Metrics != (metrics.Values{
		Nodes: 3, TreeDepth: 1, SceneInstances: 1, MeshInstances: 1,
		ExternalResources: 1, SceneDependencies: 1,
	}) {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
	if !summary.DepthPartial || len(summary.ParentFindings) != 1 || summary.ParentFindings[0].DeclaringScene != root.Canonical {
		t.Fatalf("depth evidence = %v/%#v", summary.DepthPartial, summary.ParentFindings)
	}
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityLowerBound ||
		len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != diagnostic.CodeUnsupportedParent {
		t.Fatalf("completeness = %q/%q/%#v", result.Status, result.Reliability, result.Diagnostics)
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
			name:    "unsupported inherited node",
			builder: summaryBuilder{summary: ExpandedSummary{Metrics: metrics.Values{Nodes: math.MaxInt64}}},
			apply: func(builder *summaryBuilder) error {
				return builder.addUnsupportedInheritance(InheritedTarget{Occurrences: 1})
			},
		},
		{
			name: "resolved inherited metrics",
			builder: summaryBuilder{
				summary:      ExpandedSummary{Metrics: metrics.Values{Nodes: math.MaxInt64}},
				resources:    map[ResourceIdentity]struct{}{},
				dependencies: map[string]struct{}{},
			},
			apply: func(builder *summaryBuilder) error {
				return builder.applyInheritedBase(
					project.ResolvedPath{Canonical: filepath.Join(t.TempDir(), "base.tscn")},
					ExpandedSummary{Metrics: metrics.Values{Nodes: 1}},
					InheritedTarget{Occurrences: 1},
				)
			},
		},
		{
			name: "resolved inherited coverage",
			builder: summaryBuilder{
				summary:      ExpandedSummary{Coverage: SceneInstanceCoverage{Resolved: math.MaxInt64}},
				resources:    map[ResourceIdentity]struct{}{},
				dependencies: map[string]struct{}{},
			},
			apply: func(builder *summaryBuilder) error {
				return builder.applyInheritedBase(
					project.ResolvedPath{Canonical: filepath.Join(t.TempDir(), "base.tscn")},
					ExpandedSummary{Coverage: SceneInstanceCoverage{Resolved: 1}},
					InheritedTarget{Occurrences: 1},
				)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := cloneExpandedSummary(test.builder.summary)
			assertOverflow(t, test.apply(&test.builder))
			if !reflect.DeepEqual(test.builder.summary, before) || len(test.builder.resources) != 0 || len(test.builder.dependencies) != 0 {
				t.Fatalf("overflow mutated builder: %#v", test.builder)
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
	if summary.Metrics != (metrics.Values{
		Nodes: 2, TreeDepth: 2, SceneInstances: 1, MeshInstances: 1,
		ExternalResources: 2, SceneDependencies: 1,
	}) {
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
