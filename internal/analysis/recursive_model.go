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

// SceneInstanceCoverage counts scene-instance occurrences whose complete
// nested contribution is or is not statically available.
type SceneInstanceCoverage struct {
	Resolved   int64
	Unresolved int64
}

// ResourceIdentity is either a canonical resolved resource identity or a
// document-local unresolved declaration identity.
type ResourceIdentity struct {
	Resolved       bool
	Canonical      string
	DeclaringScene string
	ResourceID     string
	RawPath        string
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

// InheritedTarget records a parsed nested scene whose inherited base remains
// intentionally unexpanded until the inherited-scene analysis capability.
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

// ExpandedSummary is the deterministic one-occurrence recursive contribution
// of a canonical scene. The two unique metric fields remain zero until the
// later final-metrics layer derives them from the retained sets.
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
