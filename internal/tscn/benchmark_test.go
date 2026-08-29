package tscn_test

import (
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

const representativeScene = `[gd_scene load_steps=4 format=3 uid="uid://benchmark"]

[ext_resource type="PackedScene" path="res://props/lamp.tscn" id="1_lamp"]
[ext_resource type="Texture2D" path="res://art/shared.png" id="2_texture"]

[sub_resource type="ArrayMesh" id="Mesh"]
_surface_data = PackedByteArray(1, 2, 3, 4, 5, 6, 7, 8)

[node name="City" type="Node3D"]
metadata = {
  "transforms": [Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 2, 3, 4)],
  "labels": ["one;still-string", "two"]
}

[node name="Sun" type="DirectionalLight3D" parent="."]
shadow_enabled = true

[node name="LampA" parent="." instance=ExtResource("1_lamp")]
[node name="LampB" parent="." instance=ExtResource("1_lamp")]
`

var benchmarkDocument *tscn.Document

func BenchmarkParseRepresentativeScene(b *testing.B) {
	document, err := tscn.Parse(strings.NewReader(representativeScene), "benchmark.tscn")
	if err != nil {
		b.Fatalf("validate representative scene: %v", err)
	}
	if document.Header.Format != 3 || len(document.ExtResources) != 2 || len(document.Nodes) != 4 ||
		document.Nodes[1].ShadowEnabled == nil || !*document.Nodes[1].ShadowEnabled {
		b.Fatalf("unexpected representative document: %#v", document)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(representativeScene)))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		parsed, parseErr := tscn.Parse(strings.NewReader(representativeScene), "benchmark.tscn")
		if parseErr != nil {
			b.Fatal(parseErr)
		}
		benchmarkDocument = parsed
	}
}
