package tscn

// Document is the minimal Godot 4 text-scene representation required by MVP 0.1.
type Document struct {
	Header       SceneHeader
	ExtResources map[string]ExtResource
	Nodes        []Node
	Features     Features
}

// SceneHeader contains supported [gd_scene] attributes.
type SceneHeader struct {
	Format int
	UID    string
}

// ExtResource describes one [ext_resource] declaration.
type ExtResource struct {
	ID       string
	Type     string
	UID      string
	Path     string
	Position Position
}

// ResourceRef identifies an ExtResource(...) or SubResource(...) reference.
type ResourceRef struct {
	Kind string
	ID   string
}

const (
	ResourceRefExternal = "ExtResource"
	ResourceRefInternal = "SubResource"
)

// Node contains only header fields and properties used by static MVP metrics.
type Node struct {
	Name                string
	Type                string
	Parent              string
	Owner               string
	Index               *int
	Instance            *ResourceRef
	InstancePlaceholder string
	ShadowEnabled       *bool
	Position            Position
}

// Features records syntax that requires special handling by the analyzer.
type Features struct {
	HasInheritedRoot bool
	HasOverrideNodes bool
	HasEditable      bool
}
