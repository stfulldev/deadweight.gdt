package analysis

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

type benchmarkResolver map[string]project.Resolution

func (resolver benchmarkResolver) ResolveResource(fromScene, raw string) project.Resolution {
	if result, exists := resolver[fromScene+"\x00"+raw]; exists {
		return result
	}
	return project.Resolution{Reason: project.ResolutionMissing, Path: project.ResolvedPath{Original: raw}}
}

var benchmarkAnalysisResult RecursiveResult

func BenchmarkRecursiveRepeatedScene100(b *testing.B) {
	rootDir := b.TempDir()
	city := project.ResolvedPath{
		Canonical: filepath.Join(rootDir, "city.tscn"),
		Display:   "res://city.tscn",
		Original:  "res://city.tscn",
	}
	lamp := project.ResolvedPath{
		Canonical: filepath.Join(rootDir, "lamp.tscn"),
		Display:   "res://lamp.tscn",
		Original:  "lamp.tscn",
	}

	var citySource strings.Builder
	citySource.WriteString("[gd_scene load_steps=2 format=3]\n")
	citySource.WriteString("[ext_resource type=\"PackedScene\" path=\"lamp.tscn\" id=\"1_lamp\"]\n")
	citySource.WriteString("[node name=\"City\" type=\"Node3D\"]\n")
	for index := 0; index < 100; index++ {
		fmt.Fprintf(&citySource, "[node name=\"Lamp%d\" parent=\".\" instance=ExtResource(\"1_lamp\")]\n", index)
	}
	sources := map[string]string{
		city.Canonical: citySource.String(),
		lamp.Canonical: `[gd_scene format=3]
[node name="Lamp" type="OmniLight3D"]
shadow_enabled = true
[node name="Mesh" type="MeshInstance3D" parent="."]
`,
	}
	resolver := benchmarkResolver{
		city.Canonical + "\x00lamp.tscn": {
			Reason: project.ResolutionResolved,
			Path:   lamp,
		},
	}
	analyzer, err := NewRecursiveAnalyzer(
		resolver,
		func(path project.ResolvedPath) (io.ReadCloser, error) {
			source, exists := sources[path.Canonical]
			if !exists {
				return nil, fmt.Errorf("missing benchmark scene %s", path.Display)
			}
			return io.NopCloser(strings.NewReader(source)), nil
		},
		tscn.Parse,
	)
	if err != nil {
		b.Fatalf("construct analyzer: %v", err)
	}

	result, err := analyzer.Analyze(city)
	if err != nil {
		b.Fatalf("validate repeated scene: %v", err)
	}
	wantMetrics := metrics.Values{
		Nodes: 201, TreeDepth: 3, SceneInstances: 100, MeshInstances: 100,
		Lights: 100, ShadowLights: 100, ExternalResources: 1, SceneDependencies: 1,
	}
	if result.Summary.Metrics != wantMetrics ||
		result.Coverage != (Coverage{ResolvedSceneInstances: 100, ParsedSceneFiles: 2}) {
		b.Fatalf("unexpected repeated result: %#v", result)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		measured, analyzeErr := analyzer.Analyze(city)
		if analyzeErr != nil {
			b.Fatal(analyzeErr)
		}
		benchmarkAnalysisResult = measured
	}
}
