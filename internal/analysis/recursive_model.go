package analysis

import (
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

// TargetClassification is a stable lower-level reason why one scene mount
// could not be expanded as a supported text scene.
type TargetClassification string

const (
	TargetMissingExternalResource TargetClassification = "missing_external_resource"
	TargetUnresolvedPath          TargetClassification = "unresolved_path"
	TargetImportedScene           TargetClassification = "imported_scene"
	TargetUnsupportedScene        TargetClassification = "unsupported_scene"
	TargetSubResource             TargetClassification = "sub_resource"
	TargetPlaceholder             TargetClassification = "placeholder"
	TargetUnavailableScene        TargetClassification = "unavailable_scene"
	TargetInheritedScene          TargetClassification = "inherited_scene"
)

// Valid reports whether classification is part of the MVP target taxonomy.
func (classification TargetClassification) Valid() bool {
	switch classification {
	case TargetMissingExternalResource,
		TargetUnresolvedPath,
		TargetImportedScene,
		TargetUnsupportedScene,
		TargetSubResource,
		TargetPlaceholder,
		TargetUnavailableScene,
		TargetInheritedScene:
		return true
	default:
		return false
	}
}

// SceneInstanceCoverage counts scene-instance occurrences whose complete
// nested contribution is or is not statically available.
type SceneInstanceCoverage struct {
	Resolved   int64
	Unresolved int64
}

// ResourceIdentity is either a canonical resolved resource identity or a
// document-local unresolved declaration identity.
type ResourceIdentity struct {
	Resolved         bool
	Canonical        string
	DeclaringScene   string
	ResourceID       string
	RawPath          string
	ResolutionReason project.ResolutionReason
}

// UnresolvedInstance preserves source and target evidence for one or more
// equivalent unresolved scene-instance occurrences.
type UnresolvedInstance struct {
	Classification   TargetClassification
	ResolutionReason project.ResolutionReason
	DeclaringScene   string
	DeclaringDisplay string
	ResourceID       string
	RawTarget        string
	TargetCanonical  string
	TargetDisplay    string
	TargetOriginal   string
	MountName        string
	MountPath        string
	MountDepth       OptionalDepth
	Position         tscn.Position
	Occurrences      int64
}

// InheritedTarget records one inherited-scene occurrence and the evidence
// used to classify or approximately expand its base scene.
type InheritedTarget struct {
	Classification        TargetClassification
	DeclaringScene        string
	DeclaringDisplay      string
	TargetCanonical       string
	TargetDisplay         string
	TargetOriginal        string
	BaseResourceID        string
	BaseRawTarget         string
	BaseCanonical         string
	BaseDisplay           string
	BaseClassification    TargetClassification
	BaseResolutionReason  project.ResolutionReason
	HasOverrideStubs      bool
	HasEditable           bool
	MountName             string
	MountPath             string
	MountDepth            OptionalDepth
	MountPosition         tscn.Position
	InheritedRootPosition tscn.Position
	Occurrences           int64
}

// SceneParentFinding attaches one local parent finding to its declaring scene
// and preserves its occurrence multiplicity through recursive application.
type SceneParentFinding struct {
	DeclaringScene   string
	DeclaringDisplay string
	Finding          ParentFinding
	Occurrences      int64
}

// ExpandedSummary is the deterministic recursive contribution of a canonical
// scene. Cached one-occurrence summaries keep the two unique metric fields at
// zero; the public root result finalizes them from the retained sets and graph.
type ExpandedSummary struct {
	Metrics           metrics.Values
	ExternalResources []ResourceIdentity
	Dependencies      []string
	Coverage          SceneInstanceCoverage
	Unresolved        []UnresolvedInstance
	InheritedTargets  []InheritedTarget
	ParentFindings    []SceneParentFinding
	DepthPartial      bool
}
