package analysis

import (
	"fmt"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

// EdgeKind identifies why one scene depends on another scene target.
type EdgeKind string

const (
	EdgeInstance    EdgeKind = "instance"
	EdgeInheritance EdgeKind = "inheritance"
)

// GraphNode is one successfully loaded canonical text-scene identity.
type GraphNode struct {
	Canonical string
	Display   string
}

// GraphEdge preserves one compacted resolved or unresolved scene dependency.
type GraphEdge struct {
	FromCanonical    string
	FromDisplay      string
	ToCanonical      string
	ToDisplay        string
	RawTarget        string
	ResourceID       string
	Kind             EdgeKind
	Resolved         bool
	Classification   TargetClassification
	ResolutionReason project.ResolutionReason
	Occurrences      int64
}

// DependencyGraph is the deterministic graph reachable from one root scene.
type DependencyGraph struct {
	RootCanonical     string
	RootDisplay       string
	Nodes             []GraphNode
	Edges             []GraphEdge
	SceneDependencies int64
}

// AnalysisStatus describes whether a successful static result is complete.
type AnalysisStatus string

const (
	AnalysisComplete AnalysisStatus = "complete"
	AnalysisPartial  AnalysisStatus = "partial"
)

// Valid reports whether status is part of the MVP analysis taxonomy.
func (status AnalysisStatus) Valid() bool {
	return status == AnalysisComplete || status == AnalysisPartial
}

// Reliability describes how known metrics relate to unavailable static data.
type Reliability string

const (
	ReliabilityExact       Reliability = "exact"
	ReliabilityLowerBound  Reliability = "lower_bound"
	ReliabilityApproximate Reliability = "approximate"
)

// Valid reports whether reliability is part of the MVP analysis taxonomy.
func (reliability Reliability) Valid() bool {
	return reliability == ReliabilityExact ||
		reliability == ReliabilityLowerBound ||
		reliability == ReliabilityApproximate
}

// Coverage publishes checked root-level analysis coverage.
type Coverage struct {
	ResolvedSceneInstances   int64
	UnresolvedSceneInstances int64
	ParsedSceneFiles         int64
	InheritedScenes          int64
}

// Validate checks the non-negative coverage domain and subset relationship.
func (coverage Coverage) Validate() error {
	fields := []struct {
		name  string
		value int64
	}{
		{name: "resolved_scene_instances", value: coverage.ResolvedSceneInstances},
		{name: "unresolved_scene_instances", value: coverage.UnresolvedSceneInstances},
		{name: "parsed_scene_files", value: coverage.ParsedSceneFiles},
		{name: "inherited_scenes", value: coverage.InheritedScenes},
	}
	for _, field := range fields {
		if field.value < 0 {
			return fmt.Errorf("coverage %s must be non-negative, got %d", field.name, field.value)
		}
	}
	return nil
}

// RecursiveResult pairs occurrence aggregation with its authoritative graph.
type RecursiveResult struct {
	Summary          ExpandedSummary
	Graph            DependencyGraph
	ParsedSceneFiles int64
	Status           AnalysisStatus
	Reliability      Reliability
	Coverage         Coverage
	Diagnostics      []diagnostic.Diagnostic
}

// CycleError is a fatal, explainable resolved-scene dependency cycle.
type CycleError struct {
	Canonical []string
	Display   []string
}

func (err *CycleError) Error() string {
	return fmt.Sprintf("%s: %s", diagnostic.CodeSceneDependencyCycle, err.DiagnosticMessage())
}

// DiagnosticCode exposes the stable scene-dependency-cycle code.
func (err *CycleError) DiagnosticCode() diagnostic.Code {
	return diagnostic.CodeSceneDependencyCycle
}

// DiagnosticMessage returns deterministic code-free cycle text for the CLI.
func (err *CycleError) DiagnosticMessage() string {
	const heading = "scene dependency cycle"
	if len(err.Display) == 0 {
		return heading
	}

	return heading + "\n\n" + strings.Join(err.Display, "\n→ ")
}

func cloneDependencyGraph(graph DependencyGraph) DependencyGraph {
	cloned := graph
	cloned.Nodes = append([]GraphNode(nil), graph.Nodes...)
	cloned.Edges = append([]GraphEdge(nil), graph.Edges...)

	return cloned
}

func cloneRecursiveResult(result RecursiveResult) RecursiveResult {
	cloned := result
	cloned.Summary = cloneExpandedSummary(result.Summary)
	cloned.Graph = cloneDependencyGraph(result.Graph)
	cloned.Diagnostics = append([]diagnostic.Diagnostic(nil), result.Diagnostics...)

	return cloned
}
