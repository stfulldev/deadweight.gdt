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

// RecursiveResult pairs occurrence aggregation with its authoritative graph.
type RecursiveResult struct {
	Summary ExpandedSummary
	Graph   DependencyGraph
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
	return RecursiveResult{
		Summary: cloneExpandedSummary(result.Summary),
		Graph:   cloneDependencyGraph(result.Graph),
	}
}
