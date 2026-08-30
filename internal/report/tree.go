package report

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

type dependencyTree struct {
	Root    string
	Entries []dependencyTreeEntry
}

type dependencyTreeEntry struct {
	Depth            int64
	Source           string
	Target           string
	Kind             analysis.EdgeKind
	Resolved         bool
	Occurrences      int64
	Reliability      analysis.Reliability
	BackReference    bool
	Classification   analysis.TargetClassification
	ResourceID       string
	RawTarget        string
	ResolutionReason project.ResolutionReason
	AncestorLast     []bool
	Last             bool
}

type projectedTreeEdge struct {
	entry     dependencyTreeEntry
	canonical string
	sortKey   string
}

// Tree renders one bounded, deterministic explanation of the authoritative
// dependency graph retained by recursive analysis.
func Tree(result application.TreeResult, options Options) (string, error) {
	tree, err := projectDependencyTree(result)
	if err != nil {
		return "", err
	}
	options = normalizedOptions(options)
	style := styler{enabled: options.Color}

	var output strings.Builder
	writeVersion(&output, options)
	fmt.Fprintf(&output, "%-11s%s\n", "Scene:", tree.Root)
	fmt.Fprintf(&output, "%-11s%s\n", "Project:", "res://")
	fmt.Fprintf(
		&output,
		"%-11s%s\n",
		"Analysis:",
		style.status(strings.ToUpper(string(result.Inspect.Analysis.Status))),
	)
	fmt.Fprintf(
		&output,
		"%-11s%s\n",
		"Accuracy:",
		accuracyLabel(result.Inspect.Analysis.Reliability),
	)

	output.WriteString("\nDependencies\n")
	output.WriteString(tree.Root)
	output.WriteByte('\n')
	for _, entry := range tree.Entries {
		for _, last := range entry.AncestorLast {
			if last {
				output.WriteString("    ")
			} else {
				output.WriteString("│   ")
			}
		}
		if entry.Last {
			output.WriteString("└── ")
		} else {
			output.WriteString("├── ")
		}
		fmt.Fprintf(
			&output,
			"%s ×%s [%s] %s",
			entry.Kind,
			formatInteger(entry.Occurrences),
			entry.Reliability,
			entry.Target,
		)
		if entry.BackReference {
			output.WriteString(" (back-reference)")
		}
		if !entry.Resolved {
			fmt.Fprintf(&output, " (%s)", entry.Classification)
		}
		output.WriteByte('\n')
	}

	writeDiagnostics(&output, result.Inspect.Analysis.Diagnostics, style)
	writeReliabilityWarning(&output, result.Inspect.Analysis, style)

	return output.String(), nil
}

func projectDependencyTree(result application.TreeResult) (dependencyTree, error) {
	inspect := result.Inspect
	if err := validateInspect(inspect); err != nil {
		return dependencyTree{}, err
	}
	graph := inspect.Analysis.Graph
	if graph.RootCanonical == "" {
		return dependencyTree{}, errors.New("dependency graph root is required")
	}
	if inspect.Scene.Canonical != "" && inspect.Scene.Canonical != graph.RootCanonical {
		return dependencyTree{}, errors.New("dependency graph root does not match analyzed scene")
	}
	if len(graph.Nodes) == 0 {
		return dependencyTree{}, errors.New("dependency graph must contain its root node")
	}
	if graph.SceneDependencies != int64(len(graph.Nodes)-1) {
		return dependencyTree{}, fmt.Errorf(
			"dependency graph count is %d, want %d",
			graph.SceneDependencies,
			len(graph.Nodes)-1,
		)
	}
	if graph.SceneDependencies != inspect.Analysis.Summary.Metrics.SceneDependencies {
		return dependencyTree{}, fmt.Errorf(
			"dependency graph count is %d, but analysis reports %d",
			graph.SceneDependencies,
			inspect.Analysis.Summary.Metrics.SceneDependencies,
		)
	}
	if inspect.Analysis.Coverage.ParsedSceneFiles != int64(len(graph.Nodes)) {
		return dependencyTree{}, fmt.Errorf(
			"dependency graph contains %d parsed nodes, but coverage reports %d",
			len(graph.Nodes),
			inspect.Analysis.Coverage.ParsedSceneFiles,
		)
	}

	nodes := make(map[string]string, len(graph.Nodes))
	portableNodes := make(map[string]string, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if node.Canonical == "" {
			return dependencyTree{}, errors.New("dependency graph node canonical identity is required")
		}
		if _, duplicate := nodes[node.Canonical]; duplicate {
			return dependencyTree{}, fmt.Errorf("duplicate dependency graph node %q", node.Canonical)
		}
		portable, err := portableGraphIdentity(inspect.Project.Directory, node.Display, node.Canonical)
		if err != nil {
			return dependencyTree{}, fmt.Errorf("project dependency graph node: %w", err)
		}
		if canonical, duplicate := portableNodes[portable]; duplicate {
			return dependencyTree{}, fmt.Errorf(
				"dependency graph nodes %q and %q share portable identity %q",
				canonical,
				node.Canonical,
				portable,
			)
		}
		nodes[node.Canonical] = portable
		portableNodes[portable] = node.Canonical
	}
	root, present := nodes[graph.RootCanonical]
	if !present {
		return dependencyTree{}, errors.New("dependency graph root node is missing")
	}
	if graph.RootDisplay != "" {
		display, err := portableGraphIdentity(inspect.Project.Directory, graph.RootDisplay)
		if err != nil || display != root {
			return dependencyTree{}, errors.New("dependency graph root display is inconsistent")
		}
	}
	if scene, err := portableSceneIdentity(inspect); err != nil || scene != root {
		return dependencyTree{}, errors.New("dependency graph root has no matching portable scene identity")
	}

	outgoing := make(map[string][]projectedTreeEdge, len(nodes))
	seenEdges := make(map[string]struct{}, len(graph.Edges))
	for _, edge := range graph.Edges {
		projected, signature, err := projectTreeEdge(inspect.Project.Directory, nodes, edge)
		if err != nil {
			return dependencyTree{}, err
		}
		if _, duplicate := seenEdges[signature]; duplicate {
			return dependencyTree{}, fmt.Errorf(
				"dependency graph contains duplicate compacted edge from %q to %q",
				projected.entry.Source,
				projected.entry.Target,
			)
		}
		seenEdges[signature] = struct{}{}
		outgoing[edge.FromCanonical] = append(outgoing[edge.FromCanonical], projected)
	}
	for source := range outgoing {
		edges := outgoing[source]
		sort.Slice(edges, func(left, right int) bool { return edges[left].sortKey < edges[right].sortKey })
		for index := 1; index < len(edges); index++ {
			if edges[index-1].sortKey == edges[index].sortKey {
				return dependencyTree{}, fmt.Errorf(
					"dependency graph contains ambiguous portable edges from %q",
					nodes[source],
				)
			}
		}
		outgoing[source] = edges
	}

	if err := validateAcyclicGraph(graph.RootCanonical, outgoing); err != nil {
		return dependencyTree{}, err
	}

	entries := make([]dependencyTreeEntry, 0, len(graph.Edges))
	expanded := map[string]bool{graph.RootCanonical: true}
	consumed := 0
	var visit func(string, []bool)
	visit = func(source string, ancestors []bool) {
		edges := outgoing[source]
		for index, edge := range edges {
			entry := edge.entry
			entry.Depth = int64(len(ancestors) + 1)
			entry.Last = index == len(edges)-1
			entry.AncestorLast = append([]bool(nil), ancestors...)
			if entry.Resolved && expanded[edge.canonical] {
				entry.BackReference = true
			}
			entries = append(entries, entry)
			consumed++
			if entry.Resolved && !entry.BackReference {
				expanded[edge.canonical] = true
				visit(edge.canonical, append(append([]bool(nil), ancestors...), entry.Last))
			}
		}
	}
	visit(graph.RootCanonical, nil)
	if consumed != len(graph.Edges) {
		return dependencyTree{}, fmt.Errorf(
			"dependency graph has %d unreachable edges",
			len(graph.Edges)-consumed,
		)
	}
	if len(expanded) != len(nodes) {
		return dependencyTree{}, fmt.Errorf(
			"dependency graph has %d unreachable nodes",
			len(nodes)-len(expanded),
		)
	}

	return dependencyTree{Root: root, Entries: entries}, nil
}

func projectTreeEdge(
	projectRoot string,
	nodes map[string]string,
	edge analysis.GraphEdge,
) (projectedTreeEdge, string, error) {
	if edge.Kind != analysis.EdgeInstance && edge.Kind != analysis.EdgeInheritance {
		return projectedTreeEdge{}, "", fmt.Errorf("invalid dependency edge kind %q", edge.Kind)
	}
	if edge.Occurrences <= 0 {
		return projectedTreeEdge{}, "", fmt.Errorf(
			"dependency edge occurrences must be positive, got %d",
			edge.Occurrences,
		)
	}
	source, present := nodes[edge.FromCanonical]
	if !present {
		return projectedTreeEdge{}, "", fmt.Errorf(
			"dependency edge source node %q is missing",
			edge.FromCanonical,
		)
	}
	if edge.FromDisplay != "" {
		display, err := portableGraphIdentity(projectRoot, edge.FromDisplay)
		if err != nil || display != source {
			return projectedTreeEdge{}, "", errors.New("dependency edge source display is inconsistent")
		}
	}

	entry := dependencyTreeEntry{
		Source:      source,
		Kind:        edge.Kind,
		Resolved:    edge.Resolved,
		Occurrences: edge.Occurrences,
		ResourceID:  safeTreeLiteral(edge.ResourceID),
	}
	if raw, ok := portableOptionalResource(projectRoot, edge.RawTarget); ok {
		entry.RawTarget = raw
	}
	canonical := ""
	if edge.Resolved {
		if edge.ToCanonical == "" {
			return projectedTreeEdge{}, "", errors.New("resolved dependency edge target is required")
		}
		target, ok := nodes[edge.ToCanonical]
		if !ok {
			return projectedTreeEdge{}, "", fmt.Errorf(
				"resolved dependency edge target node %q is missing",
				edge.ToCanonical,
			)
		}
		if edge.ResolutionReason != project.ResolutionResolved {
			return projectedTreeEdge{}, "", fmt.Errorf(
				"resolved dependency edge has resolution reason %q",
				edge.ResolutionReason,
			)
		}
		if edge.Classification != "" {
			return projectedTreeEdge{}, "", errors.New("resolved dependency edge has unresolved classification")
		}
		if edge.ToDisplay != "" {
			display, err := portableGraphIdentity(projectRoot, edge.ToDisplay)
			if err != nil || display != target {
				return projectedTreeEdge{}, "", errors.New("dependency edge target display is inconsistent")
			}
		}
		entry.Target = target
		canonical = edge.ToCanonical
	} else {
		if edge.ToCanonical != "" {
			return projectedTreeEdge{}, "", errors.New("unresolved dependency edge has canonical target")
		}
		if !edge.Classification.Valid() {
			return projectedTreeEdge{}, "", fmt.Errorf(
				"invalid unresolved dependency classification %q",
				edge.Classification,
			)
		}
		if !edge.ResolutionReason.Valid() {
			return projectedTreeEdge{}, "", fmt.Errorf(
				"invalid unresolved dependency reason %q",
				edge.ResolutionReason,
			)
		}
		targetDisplay := ""
		if portable, ok := portableOptionalResource(projectRoot, edge.ToDisplay); ok {
			targetDisplay = portable
		}
		entry.Target = firstNonEmpty(
			targetDisplay,
			entry.RawTarget,
			resourceTarget(entry.ResourceID),
			"<"+string(edge.Classification)+">",
		)
		entry.Classification = edge.Classification
		entry.ResolutionReason = edge.ResolutionReason
	}
	if edge.Kind == analysis.EdgeInheritance {
		entry.Reliability = analysis.ReliabilityApproximate
	} else if edge.Resolved {
		entry.Reliability = analysis.ReliabilityExact
	} else {
		entry.Reliability = analysis.ReliabilityLowerBound
	}

	sortKey := strings.Join([]string{
		entry.Target,
		string(entry.Kind),
		fmt.Sprintf("%t", entry.Resolved),
		entry.ResourceID,
		entry.RawTarget,
		string(entry.Classification),
		string(entry.ResolutionReason),
	}, "\x00")
	signature := strings.Join([]string{
		edge.FromCanonical,
		edge.FromDisplay,
		edge.ToCanonical,
		edge.ToDisplay,
		edge.RawTarget,
		edge.ResourceID,
		string(edge.Kind),
		fmt.Sprintf("%t", edge.Resolved),
		string(edge.Classification),
		string(edge.ResolutionReason),
	}, "\x00")

	return projectedTreeEdge{entry: entry, canonical: canonical, sortKey: sortKey}, signature, nil
}

func portableGraphIdentity(projectRoot string, candidates ...string) (string, error) {
	for _, candidate := range candidates {
		if portable, ok := portableOptionalPath(projectRoot, candidate); ok {
			return portable, nil
		}
	}

	return "", errors.New("graph identity is not a portable in-project path")
}

func safeTreeLiteral(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || strings.ContainsAny(value, "\\\r\n\x00") {
		return ""
	}

	return value
}

func resourceTarget(resourceID string) string {
	if resourceID == "" {
		return ""
	}

	return "resource:" + resourceID
}

func validateAcyclicGraph(root string, outgoing map[string][]projectedTreeEdge) error {
	const (
		unvisited = iota
		visiting
		visited
	)
	states := make(map[string]int)
	var visit func(string) error
	visit = func(node string) error {
		states[node] = visiting
		for _, edge := range outgoing[node] {
			if !edge.entry.Resolved {
				continue
			}
			switch states[edge.canonical] {
			case visiting:
				return errors.New("dependency graph contains a resolved cycle")
			case unvisited:
				if err := visit(edge.canonical); err != nil {
					return err
				}
			}
		}
		states[node] = visited
		return nil
	}

	return visit(root)
}
