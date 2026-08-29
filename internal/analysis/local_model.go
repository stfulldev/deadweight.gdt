// Package analysis converts parsed scenes into deterministic analyzer-domain
// values. Filesystem resolution and recursive scene aggregation live outside
// this package.
package analysis

import (
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

// OptionalDepth distinguishes an unknown depth from a known numeric value.
// The zero value represents an unknown depth.
type OptionalDepth struct {
	Value int64
	Known bool
}

// NodeRecord contains source and local-tree evidence shared by every node
// classification.
type NodeRecord struct {
	Name     string
	Path     string
	Parent   string
	Depth    OptionalDepth
	Position tscn.Position
}

// OrdinaryNode is a typed local node that contributes to ordinary-node and
// literal-type metrics.
type OrdinaryNode struct {
	NodeRecord
	Type          string
	ShadowEnabled *bool
}

// MountKind identifies how a later graph layer must handle one local instance
// mount without claiming that any target has already resolved.
type MountKind string

const (
	MountPackedSceneCandidate MountKind = "packed_scene_candidate"
	MountExternalResource     MountKind = "external_resource"
	MountSubResource          MountKind = "sub_resource"
	MountPlaceholder          MountKind = "placeholder"
)

// ResourceReference is an owned copy of a node's parsed instance reference.
type ResourceReference struct {
	Kind string
	ID   string
}

// ExternalResource is one document-local declaration. ID is unique only
// within its declaring document; later layers decide cross-scene uniqueness.
type ExternalResource struct {
	ID       string
	Type     string
	UID      string
	Path     string
	Position tscn.Position
}

// InstanceMount represents one non-expanded scene-instance occurrence.
type InstanceMount struct {
	NodeRecord
	Kind        MountKind
	Reference   ResourceReference
	Placeholder string
	Candidate   *ExternalResource
}

// InheritedRoot preserves the root instance evidence that later aggregation
// uses as an inheritance edge rather than a nested scene occurrence.
type InheritedRoot struct {
	NodeRecord
	Reference ResourceReference
	Candidate *ExternalResource
}

// OverrideStub is a node header without a literal type or instance. It is
// evidence for inherited override semantics, not a new ordinary node.
type OverrideStub struct {
	NodeRecord
}

// ParentFindingKind is a stable local reason for an unknown tree depth.
// Whole-analysis orchestration maps these findings into user diagnostics.
type ParentFindingKind string

const (
	ParentInvalid   ParentFindingKind = "invalid_parent"
	ParentMissing   ParentFindingKind = "missing_parent"
	ParentAmbiguous ParentFindingKind = "ambiguous_parent"
)

// ParentFinding retains actionable source evidence without assigning an
// unrelated user-visible diagnostic code.
type ParentFinding struct {
	Kind     ParentFindingKind
	NodeName string
	NodePath string
	Parent   string
	Position tscn.Position
}

// LocalSummary describes the statically known contribution of one parsed
// scene without expanding nested scenes.
type LocalSummary struct {
	// Metrics contains local occurrence/depth contributions. ExternalResources
	// and SceneDependencies are always zero because their final values require
	// canonical path resolution and graph-wide unique unions.
	Metrics metrics.Values

	Nodes             []OrdinaryNode
	Mounts            []InstanceMount
	InheritedRoot     *InheritedRoot
	OverrideStubs     []OverrideStub
	ExternalResources []ExternalResource
	Findings          []ParentFinding
	DepthPartial      bool
}
